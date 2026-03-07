package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/E-Timileyin/skill-island/services/api/internal/api/handlers"
	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/config"
)

func newTestHandler() *handlers.Handler {
	return &handlers.Handler{
		Cfg: config.Config{
			JWTSecret:        "test-secret-minimum-32-characters-long",
			JWTRefreshSecret: "test-refresh-secret-minimum-32-chars",
			Env:              "development",
		},
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "admin",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp handlers.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "short",
		"role":     "student",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp handlers.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLogout(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.MaxAge != -1 {
			t.Fatalf("expected cookie %s MaxAge -1, got %d", c.Name, c.MaxAge)
		}
	}
}

func TestRefresh_NoCookie(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestMe_NoAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestMe_WithMiddleware_NoCookie(t *testing.T) {
	h := newTestHandler()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.Me))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
