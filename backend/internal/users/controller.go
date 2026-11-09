package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/mrscorpio/uptimeapp/backend/internal/auth"
	"github.com/mrscorpio/uptimeapp/backend/internal/httpjson"
	"github.com/mrscorpio/uptimeapp/backend/internal/store"
	"gorm.io/gorm"
)

type Controller struct {
	DB        *gorm.DB
	Tokens    auth.Tokens
	AvatarDir string
}

func NewController(db *gorm.DB, tokens auth.Tokens, avatarDir string) *Controller {
	return &Controller{DB: db, Tokens: tokens, AvatarDir: avatarDir}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/users/me", c.requireAccess(http.HandlerFunc(c.me)))
	mux.Handle("PATCH /api/v1/users/me", c.requireAccess(http.HandlerFunc(c.updateMe)))
	mux.Handle("POST /api/v1/users/me/avatar", c.requireAccess(http.HandlerFunc(c.updateAvatar)))
}

const maxAvatarSize = 5 << 20

func (c *Controller) updateAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize+(1<<20))
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid or oversized avatar")
		return
	}
	file, _, err := r.FormFile("avatar")
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "avatar image is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil || len(data) == 0 || len(data) > maxAvatarSize {
		httpjson.WriteError(w, http.StatusBadRequest, "avatar must be an image up to 5 MB")
		return
	}
	mime := http.DetectContentType(data)
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	extension, ok := extensions[mime]
	if !ok {
		httpjson.WriteError(w, http.StatusBadRequest, "avatar must be PNG, JPEG, or WebP")
		return
	}
	id := r.Context().Value(userIDKey{}).(uint)
	var user store.User
	if err := c.DB.WithContext(r.Context()).First(&user, id).Error; err != nil {
		httpjson.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	oldAvatarURL := user.AvatarURL
	var nameBytes [16]byte
	if _, err := rand.Read(nameBytes[:]); err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not save avatar")
		return
	}
	name := hex.EncodeToString(nameBytes[:]) + extension
	if err := os.MkdirAll(c.AvatarDir, 0750); err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not save avatar")
		return
	}
	path := filepath.Join(c.AvatarDir, name)
	if err := os.WriteFile(path, data, 0640); err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not save avatar")
		return
	}
	user.AvatarURL = "/avatars/" + name
	if err := c.DB.WithContext(r.Context()).Model(&user).Update("avatar_url", user.AvatarURL).Error; err != nil {
		_ = os.Remove(path)
		httpjson.WriteError(w, http.StatusInternalServerError, "could not save avatar")
		return
	}
	if old := filepath.Base(strings.TrimPrefix(oldAvatarURL, "/avatars/")); old != name && old != "." && old != "/" {
		_ = os.Remove(filepath.Join(c.AvatarDir, old))
	}
	httpjson.WriteJSON(w, http.StatusOK, user)
}

type userIDKey struct{}

func (c *Controller) requireAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			httpjson.WriteError(w, 401, "invalid access token")
			return
		}
		id, err := c.Tokens.Verify(header[1])
		if err != nil {
			httpjson.WriteError(w, 401, "invalid access token")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, id)))
	})
}

func (c *Controller) me(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(userIDKey{}).(uint)
	var u store.User
	if err := c.DB.WithContext(r.Context()).First(&u, id).Error; err != nil {
		httpjson.WriteError(w, 401, "invalid access token")
		return
	}
	httpjson.WriteJSON(w, 200, u)
}

func (c *Controller) updateMe(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := httpjson.Decode(r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 100 {
		httpjson.WriteError(w, http.StatusBadRequest, "name must contain 1 to 100 characters")
		return
	}

	id := r.Context().Value(userIDKey{}).(uint)
	result := c.DB.WithContext(r.Context()).Model(&store.User{}).Where("id = ?", id).Update("name", input.Name)
	if result.Error != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not update profile")
		return
	}
	if result.RowsAffected == 0 {
		httpjson.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	var user store.User
	if err := c.DB.WithContext(r.Context()).First(&user, id).Error; err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not load profile")
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, user)
}
