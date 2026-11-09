package httpapi

import (
	"net/http"
	"os"

	"github.com/mrscorpio/uptimeapp/backend/internal/auth"
	"github.com/mrscorpio/uptimeapp/backend/internal/config"
	"github.com/mrscorpio/uptimeapp/backend/internal/httpjson"
	"github.com/mrscorpio/uptimeapp/backend/internal/monitors"
	"github.com/mrscorpio/uptimeapp/backend/internal/users"

	"gorm.io/gorm"
)

type API struct {
	DB     *gorm.DB
	Config config.Config
}

func New(db *gorm.DB, cfg config.Config) *API {
	return &API{DB: db, Config: cfg}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	if a.Config.AvatarDir == "" {
		a.Config.AvatarDir = "./avatars"
	}
	if err := os.MkdirAll(a.Config.AvatarDir, 0750); err != nil {
		panic("create avatar directory: " + err.Error())
	}
	mux.Handle("GET /avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir(a.Config.AvatarDir))))
	auth.NewController(a.DB, a.Config).RegisterRoutes(mux)
	tokens := auth.Tokens{Secret: []byte(a.Config.JWTSecret), AccessTTL: a.Config.AccessTTL, RefreshTTL: a.Config.RefreshTTL}
	users.NewController(a.DB, tokens, a.Config.AvatarDir).RegisterRoutes(mux)
	monitors.NewController(a.DB, tokens).RegisterRoutes(mux)
	return a.cors(mux)
}

func (a *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin != a.Config.AllowedOrigin {
			httpjson.WriteError(w, http.StatusForbidden, "origin not allowed")
			return
		}
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
