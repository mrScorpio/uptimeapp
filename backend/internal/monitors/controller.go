package monitors

import (
	"context"
	"errors"
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

type createRequest struct {
	URL      string `json:"url"`
	Interval int    `json:"interval_seconds"`
}

func NewController(db *gorm.DB, tokens auth.Tokens) *Controller {
	return &Controller{DB: db, Tokens: tokens}
}

func (c *Controller) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/monitors", c.requireAccess(http.HandlerFunc(c.list)))
	mux.Handle("POST /api/v1/monitors", c.requireAccess(http.HandlerFunc(c.create)))
	mux.Handle("PATCH /api/v1/monitors/{id}", c.requireAccess(http.HandlerFunc(c.update)))
	mux.Handle("DELETE /api/v1/monitors/{id}", c.requireAccess(http.HandlerFunc(c.delete)))
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

// list returns the authenticated user's monitors.
// @Summary List monitors
// @Tags monitors
// @Produce json
// @Security BearerAuth
// @Success 200 {array} store.Monitor
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitors [get]
func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(userIDKey{}).(uint)
	var monitors []store.Monitor
	if err := c.DB.WithContext(r.Context()).Where("user_id = ?", id).Order("created_at DESC").Find(&monitors).Error; err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not load monitors")
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, monitors)
}

// create adds a monitor for the authenticated user.
// @Summary Create monitor
// @Tags monitors
// @Accept json
// @Produce json
// @Param request body createRequest true "Monitor configuration"
// @Security BearerAuth
// @Success 201 {object} store.Monitor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitors [post]
func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	var input createRequest
	if err := httpjson.Decode(r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate(&input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
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

// update changes a monitor owned by the authenticated user.
// @Summary Update monitor
// @Tags monitors
// @Accept json
// @Produce json
// @Param id path int true "Monitor ID"
// @Param request body createRequest true "Updated monitor configuration"
// @Security BearerAuth
// @Success 200 {object} store.Monitor
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitors/{id} [patch]
func (c *Controller) update(w http.ResponseWriter, r *http.Request) {
	id, err := monitorID(r)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}
	var input createRequest
	if err := httpjson.Decode(r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate(&input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var monitor store.Monitor
	result := c.DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", id, r.Context().Value(userIDKey{}).(uint)).First(&monitor)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			httpjson.WriteError(w, http.StatusNotFound, "monitor not found")
			return
		}
		httpjson.WriteError(w, http.StatusInternalServerError, "could not load monitor")
		return
	}
	monitor.URL = input.URL
	monitor.Interval = input.Interval
	if err := c.DB.WithContext(r.Context()).Save(&monitor).Error; err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not update monitor")
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, monitor)
}

// delete soft-deletes a monitor owned by the authenticated user.
// @Summary Delete monitor
// @Tags monitors
// @Produce json
// @Param id path int true "Monitor ID"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitors/{id} [delete]
func (c *Controller) delete(w http.ResponseWriter, r *http.Request) {
	id, err := monitorID(r)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}
	result := c.DB.WithContext(r.Context()).Where("id = ? AND user_id = ?", id, r.Context().Value(userIDKey{}).(uint)).Delete(&store.Monitor{})
	if result.Error != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "could not delete monitor")
		return
	}
	if result.RowsAffected == 0 {
		httpjson.WriteError(w, http.StatusNotFound, "monitor not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func monitorID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, strconv.ErrSyntax
	}
	return uint(id), nil
}

func validate(input *createRequest) error {
	input.URL = strings.TrimSpace(input.URL)
	parsed, err := url.ParseRequestURI(input.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || strings.ContainsAny(input.URL, "\r\n") {
		return errors.New("URL must be a valid HTTP or HTTPS address")
	}
	if input.Interval < 1 || input.Interval > 86400 {
		return errors.New("interval_seconds must be between 1 and 86400")
	}
	return nil
}
