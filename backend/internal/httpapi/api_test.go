package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mrscorpio/uptimeapp/backend/internal/auth"
	"github.com/mrscorpio/uptimeapp/backend/internal/config"
	"github.com/mrscorpio/uptimeapp/backend/internal/store"
)

func testAPI(t *testing.T) (*API, http.Handler) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL integration tests")
	}
	db, err := store.Open(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{JWTSecret: "a-secret-of-at-least-thirty-two-characters", AllowedOrigin: "http://localhost:3000", AccessTTL: 15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour, AvatarDir: t.TempDir()}
	a := New(db, cfg)
	return a, a.Handler()
}

func TestAvatarUpload(t *testing.T) {
	a, h := testAPI(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	reg := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":"%s@example.com","login":"user%s","password":"password123"}`, suffix, suffix), "", "")
	if reg.Code != 201 {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}
	token := access(t, reg)
	var imageBytes bytes.Buffer
	if err := png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(imageBytes.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/api/v1/users/me/avatar", &body)
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("avatar upload: %d %s", w.Code, w.Body.String())
	}
	var user store.User
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	if user.AvatarURL == "" {
		t.Fatal("avatar URL missing from response")
	}
	file := filepath.Join(a.Config.AvatarDir, filepath.Base(user.AvatarURL))
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("avatar not saved: %v", err)
	}
	if got := request(h, "GET", user.AvatarURL, "", "", ""); got.Code != http.StatusOK {
		t.Fatalf("avatar not served: %d", got.Code)
	}
	if got := request(h, "GET", "/api/v1/users/me", "", token, ""); got.Code != 200 || !strings.Contains(got.Body.String(), user.AvatarURL) {
		t.Fatalf("profile avatar missing: %d %s", got.Code, got.Body.String())
	}
	badBody := strings.NewReader(`invalid`)
	badReq := httptest.NewRequest("POST", "/api/v1/users/me/avatar", badBody)
	badReq.Header.Set("Authorization", "Bearer "+token)
	badReq.Header.Set("Content-Type", "multipart/form-data; boundary=bad")
	badResponse := httptest.NewRecorder()
	h.ServeHTTP(badResponse, badReq)
	if badResponse.Code != 400 {
		t.Fatalf("invalid avatar accepted: %d", badResponse.Code)
	}
}

func TestMonitorCreateAndList(t *testing.T) {
	_, h := testAPI(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	reg := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":"%s@example.com","login":"user%s","password":"password123"}`, suffix, suffix), "", "")
	if reg.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}
	token := access(t, reg)
	created := request(h, "POST", "/api/v1/monitors", `{"url":"https://example.com","interval_seconds":60}`, token, "")
	if created.Code != http.StatusCreated {
		t.Fatalf("create monitor: %d %s", created.Code, created.Body.String())
	}
	var monitor store.Monitor
	if err := json.Unmarshal(created.Body.Bytes(), &monitor); err != nil {
		t.Fatal(err)
	}
	if monitor.ID == 0 || monitor.URL != "https://example.com" || monitor.Interval != 60 {
		t.Fatalf("unexpected monitor: %+v", monitor)
	}
	listed := request(h, "GET", "/api/v1/monitors", "", token, "")
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), `"id":`+fmt.Sprint(monitor.ID)) {
		t.Fatalf("list monitors: %d %s", listed.Code, listed.Body.String())
	}
	for _, body := range []string{`{"url":"ftp://example.com","interval_seconds":60}`, `{"url":"https://example.com","interval_seconds":0}`} {
		if response := request(h, "POST", "/api/v1/monitors", body, token, ""); response.Code != http.StatusBadRequest {
			t.Fatalf("invalid monitor accepted: %d %s", response.Code, response.Body.String())
		}
	}
	if response := request(h, "GET", "/api/v1/monitors", "", "", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated list accepted: %d", response.Code)
	}
}

func request(h http.Handler, method, path, body, token, cookie string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if cookie != "" {
		r.Header.Set("Cookie", cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func access(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.AccessToken == "" {
		t.Fatalf("missing access token: %s: %v", w.Body.String(), err)
	}
	return result.AccessToken
}

func cookie(w *httptest.ResponseRecorder) string { return w.Result().Cookies()[0].String() }

func TestAuthFlow(t *testing.T) {
	a, h := testAPI(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	login := "user" + suffix
	email := login + "@example.com"
	reg := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":%q,"login":%q,"password":"password123"}`, email, login), "", "")
	if reg.Code != 201 {
		t.Fatalf("register %d: %s", reg.Code, reg.Body.String())
	}
	if reg.Result().Cookies()[0].HttpOnly != true {
		t.Fatal("refresh cookie is not HttpOnly")
	}
	jwt := access(t, reg)
	refresh := cookie(reg)
	if w := request(h, "GET", "/api/v1/users/me", "", jwt, ""); w.Code != 200 {
		t.Fatalf("me: %d %s", w.Code, w.Body.String())
	}
	if w := request(h, "GET", "/api/v1/users/me", "", "", ""); w.Code != 401 {
		t.Fatal("me without token accepted")
	}
	if w := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":%q,"login":%q,"password":"password123"}`, email, login), "", ""); w.Code != 409 {
		t.Fatalf("duplicate: %d", w.Code)
	}
	if w := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":%q,"login":"another%s","password":"password123"}`, strings.ToUpper(email), suffix), "", ""); w.Code != 409 {
		t.Fatalf("duplicate email: %d", w.Code)
	}
	if w := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":"another%s@example.com","login":%q,"password":"password123"}`, suffix, strings.ToUpper(login)), "", ""); w.Code != 409 {
		t.Fatalf("duplicate login: %d", w.Code)
	}
	for _, identifier := range []string{email, login} {
		w := request(h, "POST", "/api/v1/auth/login", fmt.Sprintf(`{"identifier":%q,"password":"password123"}`, identifier), "", "")
		if w.Code != 200 {
			t.Fatalf("login %s: %d %s", identifier, w.Code, w.Body.String())
		}
	}
	if w := request(h, "POST", "/api/v1/auth/login", fmt.Sprintf(`{"identifier":%q,"password":"wrong"}`, login), "", ""); w.Code != 401 {
		t.Fatal("wrong password accepted")
	}
	if w := request(h, "POST", "/api/v1/auth/login", `{"identifier":"missing-user","password":"password123"}`, "", ""); w.Code != 401 {
		t.Fatal("unknown user accepted")
	}
	if w := request(h, "POST", "/api/v1/auth/login", fmt.Sprintf(`{"identifier":%q,"password":"password123"}`, strings.ToUpper(login)), "", ""); w.Code != 200 {
		t.Fatal("case-insensitive login failed")
	}
	rotated := request(h, "POST", "/api/v1/auth/refresh", "", "", refresh)
	if rotated.Code != 200 {
		t.Fatalf("refresh: %d %s", rotated.Code, rotated.Body.String())
	}
	if cookie(rotated) == refresh {
		t.Fatal("refresh cookie did not rotate")
	}
	if w := request(h, "POST", "/api/v1/auth/refresh", "", "", refresh); w.Code != 401 {
		t.Fatal("old refresh token accepted")
	}
	if w := request(h, "POST", "/api/v1/auth/logout", "", "", cookie(rotated)); w.Code != 204 {
		t.Fatal("logout failed")
	}
	if w := request(h, "POST", "/api/v1/auth/refresh", "", "", cookie(rotated)); w.Code != 401 {
		t.Fatal("logged out refresh token accepted")
	}
	var session store.Session
	if err := a.DB.Where("token_hash = ?", auth.Hash(reg.Result().Cookies()[0].Value)).First(&session).Error; err == nil {
		t.Fatal("old token hash still stored")
	}
}

func TestRegistrationValidationAndCORS(t *testing.T) {
	_, h := testAPI(t)
	for _, body := range []string{
		`{"email":"bad","login":"alice","password":"password123"}`,
		`{"email":"a@example.com","login":"a","password":"password123"}`,
		`{"email":"a@example.com","login":"alice","password":"short"}`,
	} {
		if w := request(h, "POST", "/api/v1/auth/register", body, "", ""); w.Code != 400 {
			t.Fatalf("invalid registration accepted: %s", body)
		}
	}
	r := httptest.NewRequest("OPTIONS", "/api/v1/auth/login", nil)
	r.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("allowed origin preflight failed")
	}
	r.Header.Set("Origin", "http://evil.example")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("unexpected origin allowed")
	}
}

func TestConcurrentRefresh(t *testing.T) {
	_, h := testAPI(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	reg := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":"%s@example.com","login":"user%s","password":"password123"}`, suffix, suffix), "", "")
	if reg.Code != 201 {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}
	c := cookie(reg)
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); codes <- request(h, "POST", "/api/v1/auth/refresh", "", "", c).Code }()
	}
	wg.Wait()
	close(codes)
	ok, unauthorized := 0, 0
	for code := range codes {
		if code == 200 {
			ok++
		}
		if code == 401 {
			unauthorized++
		}
	}
	if ok != 1 || unauthorized != 1 {
		t.Fatalf("expected one success and one rejection, got %d and %d", ok, unauthorized)
	}
}

func TestExpiredRefresh(t *testing.T) {
	a, h := testAPI(t)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	reg := request(h, "POST", "/api/v1/auth/register", fmt.Sprintf(`{"email":"%s@example.com","login":"user%s","password":"password123"}`, suffix, suffix), "", "")
	if reg.Code != 201 {
		t.Fatalf("register: %d", reg.Code)
	}
	if err := a.DB.Model(&store.Session{}).Where("token_hash = ?", auth.Hash(reg.Result().Cookies()[0].Value)).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if w := request(h, "POST", "/api/v1/auth/refresh", "", "", cookie(reg)); w.Code != 401 {
		t.Fatal("expired refresh token accepted")
	}
}
