package auth

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/mrscorpio/uptimeapp/backend/internal/config"
	"github.com/mrscorpio/uptimeapp/backend/internal/httpjson"
	"github.com/mrscorpio/uptimeapp/backend/internal/store"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const cookieName = "refresh_token"

type Controller struct {
	DB     *gorm.DB
	Config config.Config
	Tokens Tokens
}

func NewController(db *gorm.DB, cfg config.Config) *Controller {
	return &Controller{DB: db, Config: cfg, Tokens: Tokens{Secret: []byte(cfg.JWTSecret), AccessTTL: cfg.AccessTTL, RefreshTTL: cfg.RefreshTTL}}
}

func (a *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", a.register)
	mux.HandleFunc("POST /api/v1/auth/login", a.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", a.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", a.logout)
}

type registerRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (a *Controller) register(w http.ResponseWriter, r *http.Request) {
	var in registerRequest
	if err := httpjson.Decode(r, &in); err != nil {
		httpjson.WriteError(w, 400, "invalid request body")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Login = strings.ToLower(strings.TrimSpace(in.Login))
	addr, err := mail.ParseAddress(in.Email)
	if err != nil || addr.Address != in.Email || len(in.Email) > 254 || len(in.Login) < 3 || len(in.Login) > 32 || !validLogin(in.Login) || len(in.Password) < 8 || len(in.Password) > 72 {
		httpjson.WriteError(w, 400, "invalid email, login or password")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	u := store.User{Email: in.Email, Login: in.Login, Name: in.Login, PasswordHash: string(hash)}
	if err := a.DB.WithContext(r.Context()).Create(&u).Error; err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			httpjson.WriteError(w, 409, "email or login already exists")
			return
		}
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	a.issue(w, r, &u, http.StatusCreated)
}

func validLogin(s string) bool {
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func (a *Controller) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if err := httpjson.Decode(r, &in); err != nil {
		httpjson.WriteError(w, 400, "invalid request body")
		return
	}
	identifier := strings.ToLower(strings.TrimSpace(in.Identifier))
	if identifier == "" || in.Password == "" {
		httpjson.WriteError(w, 400, "identifier and password are required")
		return
	}
	var u store.User
	err := a.DB.WithContext(r.Context()).Where("email = ? OR login = ?", identifier, identifier).First(&u).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		httpjson.WriteError(w, 401, "invalid credentials")
		return
	}
	a.issue(w, r, &u, http.StatusOK)
}

func (a *Controller) issue(w http.ResponseWriter, r *http.Request, u *store.User, status int) {
	access, err := a.Tokens.Access(u.ID)
	if err != nil {
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	raw, hash, err := Refresh()
	if err != nil {
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	s := store.Session{UserID: u.ID, TokenHash: hash, ExpiresAt: time.Now().Add(a.Config.RefreshTTL)}
	if err := a.DB.WithContext(r.Context()).Create(&s).Error; err != nil {
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	a.setCookie(w, raw, a.Config.RefreshTTL)
	httpjson.WriteJSON(w, status, map[string]any{"access_token": access, "token_type": "Bearer", "user": u})
}

func (a *Controller) refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		httpjson.WriteError(w, 401, "invalid refresh token")
		return
	}
	newRaw, newHash, err := Refresh()
	if err != nil {
		httpjson.WriteError(w, 500, "internal error")
		return
	}
	var s store.Session
	var access string
	err = a.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ?", Hash(c.Value)).First(&s).Error; err != nil {
			return err
		}
		if s.RevokedAt != nil || !s.ExpiresAt.After(time.Now()) {
			return gorm.ErrRecordNotFound
		}
		var err error
		access, err = a.Tokens.Access(s.UserID)
		if err != nil {
			return err
		}
		return tx.Model(&s).Updates(map[string]any{"token_hash": newHash, "expires_at": time.Now().Add(a.Config.RefreshTTL)}).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpjson.WriteError(w, 401, "invalid refresh token")
		} else {
			httpjson.WriteError(w, 500, "internal error")
		}
		return
	}
	a.setCookie(w, newRaw, a.Config.RefreshTTL)
	httpjson.WriteJSON(w, 200, map[string]any{"access_token": access, "token_type": "Bearer"})
}

func (a *Controller) logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err == nil && c.Value != "" {
		now := time.Now()
		if err := a.DB.WithContext(r.Context()).Model(&store.Session{}).Where("token_hash = ? AND revoked_at IS NULL", Hash(c.Value)).Update("revoked_at", &now).Error; err != nil {
			httpjson.WriteError(w, 500, "internal error")
			return
		}
	}
	a.setCookie(w, "", -time.Second)
	w.WriteHeader(http.StatusNoContent)
}

func (a *Controller) setCookie(w http.ResponseWriter, raw string, ttl time.Duration) {
	c := &http.Cookie{Name: cookieName, Value: raw, Path: "/api/v1/auth", HttpOnly: true, Secure: a.Config.CookieSecure, SameSite: http.SameSiteLaxMode}
	if ttl < 0 {
		c.MaxAge = -1
		c.Expires = time.Unix(0, 0)
	} else {
		c.MaxAge = int(ttl.Seconds())
		c.Expires = time.Now().Add(ttl)
	}
	http.SetCookie(w, c)
}
