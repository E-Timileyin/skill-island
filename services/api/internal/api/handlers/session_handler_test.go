package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/E-Timileyin/skill-island/services/api/internal/api/handlers"
	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
)

// withClaims adds auth claims to the request context.
func withClaims(r *http.Request, claims *auth.Claims) *http.Request {
	ctx := context.WithValue(r.Context(), claimsKeyForTest, claims)
	return r.WithContext(ctx)
}

// claimsKeyForTest mirrors the unexported context key from the auth package.
// We use the auth.Middleware to set context in integration-style tests,
// but for unit tests we build the request through the middleware.
var claimsKeyForTest = contextKeyString("claims")

type contextKeyString string

func TestSubmitSession_NoAuth(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "memory_cove",
		"mode":        "solo",
		"actions":     []interface{}{map[string]interface{}{"type": "button_press"}},
		"duration_ms": 60000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.SubmitSession(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestSubmitSession_NonStudentRole(t *testing.T) {
	h := newTestHandler()

	// Create a valid token for a parent role.
	token, _ := auth.GenerateAccessToken("user-1", "parent", "", h.Cfg.JWTSecret)

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "memory_cove",
		"mode":        "solo",
		"actions":     []interface{}{map[string]interface{}{"type": "button_press"}},
		"duration_ms": 60000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestSubmitSession_NoProfile(t *testing.T) {
	h := newTestHandler()

	// Create a valid token for a student with no profile.
	token, _ := auth.GenerateAccessToken("user-1", "student", "", h.Cfg.JWTSecret)

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "memory_cove",
		"mode":        "solo",
		"actions":     []interface{}{map[string]interface{}{"type": "button_press"}},
		"duration_ms": 60000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestSubmitSession_InvalidBody(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "student", "profile-1", h.Cfg.JWTSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader([]byte("not json")))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestSubmitSession_InvalidGameType(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "student", "profile-1", h.Cfg.JWTSecret)

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "invalid_zone",
		"mode":        "solo",
		"actions":     []interface{}{map[string]interface{}{"type": "button_press"}},
		"duration_ms": 60000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp handlers.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestSubmitSession_InvalidMode(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "student", "profile-1", h.Cfg.JWTSecret)

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "memory_cove",
		"mode":        "invalid",
		"actions":     []interface{}{map[string]interface{}{"type": "button_press"}},
		"duration_ms": 60000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestSubmitSession_EmptyActions(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "student", "profile-1", h.Cfg.JWTSecret)

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "memory_cove",
		"mode":        "solo",
		"actions":     []interface{}{},
		"duration_ms": 60000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestSubmitSession_InvalidDuration(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "student", "profile-1", h.Cfg.JWTSecret)

	body, _ := json.Marshal(map[string]interface{}{
		"game_type":   "memory_cove",
		"mode":        "solo",
		"actions":     []interface{}{map[string]interface{}{"type": "button_press"}},
		"duration_ms": 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.SubmitSession))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
