package monitors

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mrscorpio/uptimeapp/backend/internal/auth"
	"github.com/mrscorpio/uptimeapp/backend/internal/httpjson"
	"github.com/mrscorpio/uptimeapp/backend/internal/store"
	"gorm.io/gorm"
)

type Controller struct {
	DB     *gorm.DB
	Tokens auth.Tokens
}

func NewController(db *gorm.DB, tokens auth.Tokens) *Controller {
	return &Controller{DB: db, Tokens: tokens}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/monitors", c.requireAccess(http.HandlerFunc(c.list)))
	mux.Handle("POST /api/v1/monitors", c.requireAccess(http.HandlerFunc(c.create)))
}

type userIDKey struct{}

func (c *Controller) requireAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			httpjson.WriteError(w, http.StatusUnauthorized, "invalid access token")
			return
		}
		id, err := c.Tokens.Verify(header[1])
		if err != nil {
			httpjson.WriteError(w, http.StatusUnauthorized, "invalid access token")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey{}, id)))
	})
}

func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(userIDKey{}).(uint)
	var monitors []store.Monitor
	if err := c.DB.WithContext(r.Context()).Where("user_id = ?", id).Order("created_at DESC").Find(&monitors).Error; err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not load monitors")
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, monitors)
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL      string `json:"url"`
		Interval int    `json:"interval_seconds"`
	}
	if err := httpjson.Decode(r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	parsed, err := url.ParseRequestURI(input.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || strings.ContainsAny(input.URL, "\r\n") {
		httpjson.WriteError(w, http.StatusBadRequest, "URL must be a valid HTTP or HTTPS address")
		return
	}
	if input.Interval < 1 || input.Interval > 86400 {
		httpjson.WriteError(w, http.StatusBadRequest, "interval_seconds must be between 1 and 86400")
		return
	}
	monitor := store.Monitor{UserID: r.Context().Value(userIDKey{}).(uint), URL: input.URL, Interval: input.Interval}
	if err := c.DB.WithContext(r.Context()).Create(&monitor).Error; err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not create monitor")
		return
	}
	w.Header().Set("Location", "/api/v1/monitors/"+strconv.FormatUint(uint64(monitor.ID), 10))
	httpjson.WriteJSON(w, http.StatusCreated, monitor)
}
