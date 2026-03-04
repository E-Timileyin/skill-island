package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/E-Timileyin/skill-island/services/api/internal/api"
	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
)

func TestCreateProfile_NoAuth(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"nickname":  "TestPlayer",
		"avatar_id": 1,
		"play_mode": "solo",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.CreateProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestCreateProfile_NonStudentRole(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"nickname":  "TestPlayer",
		"avatar_id": 1,
		"play_mode": "solo",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Use real middleware to inject parent claims
	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-456", "parent", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.CreateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp api.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN code, got %s", resp.Code)
	}
}

func TestCreateProfile_InvalidBody(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.CreateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateProfile_MissingNickname(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"avatar_id": 1,
		"play_mode": "solo",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.CreateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp api.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestCreateProfile_InvalidPlayMode(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"nickname":  "TestPlayer",
		"avatar_id": 1,
		"play_mode": "invalid",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.CreateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp api.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestGetProfile_NoAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/me", nil)
	w := httptest.NewRecorder()

	h.GetProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestGetProfile_NonStudentRole(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/me", nil)
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-456", "educator", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.GetProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestUpdateProfile_NoAuth(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"nickname": "NewName",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/profiles/me", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.UpdateProfile(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUpdateProfile_InvalidBody(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPatch, "/api/profiles/me", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.UpdateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateProfile_EmptyNickname(t *testing.T) {
	h := newTestHandler()

	empty := ""
	body, _ := json.Marshal(map[string]interface{}{
		"nickname": empty,
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/profiles/me", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.UpdateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp api.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestUpdateProfile_InvalidPlayMode(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{
		"play_mode": "invalid",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/profiles/me", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.UpdateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp api.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestUpdateProfile_NoFields(t *testing.T) {
	h := newTestHandler()

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPatch, "/api/profiles/me", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(h.UpdateProfile))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp api.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR code, got %s", resp.Code)
	}
}

func TestCreateProfile_DefaultPlayMode(t *testing.T) {
	h := newTestHandler()

	// When play_mode is omitted, it should default to "solo" and pass validation.
	// Without a DB, the handler will panic on pool access; we verify validation
	// passes by checking that the handler does NOT return 400.
	// We wrap the call to recover from the nil-pool panic.
	body, _ := json.Marshal(map[string]interface{}{
		"nickname":  "TestPlayer",
		"avatar_id": 1,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mw := auth.Middleware(h.Cfg.JWTSecret)
	token, _ := auth.GenerateAccessToken("user-123", "student", "", h.Cfg.JWTSecret)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	handler := mw(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Expected: nil pool causes panic after validation passes.
				// Write a 500 so test can verify it was not a 400.
				rw.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.CreateProfile(rw, r)
	}))

	handler.ServeHTTP(w, req)

	// Validation should pass (not 400); we expect 500 from nil pool or the panic recovery.
	if w.Code == http.StatusBadRequest {
		t.Fatalf("expected request to pass validation, got status 400")
	}
}
