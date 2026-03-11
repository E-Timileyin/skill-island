package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/E-Timileyin/skill-island/services/api/internal/api/handlers"
	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
)

func TestAnalyticsOverview_NoAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/overview", nil)
	w := httptest.NewRecorder()

	h.AnalyticsOverview(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAnalyticsOverview_StudentRole(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "student", "profile-1", h.Cfg.JWTSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.AnalyticsOverview))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var resp handlers.APIError
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN code, got %s", resp.Code)
	}
}

func TestAnalyticsOverview_ParentRole_MissingProfileID(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "parent", "", h.Cfg.JWTSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.AnalyticsOverview))

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

func TestAnalyticsOverview_EducatorRole_MissingProfileID(t *testing.T) {
	h := newTestHandler()

	token, _ := auth.GenerateAccessToken("user-1", "educator", "", h.Cfg.JWTSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	mw := auth.Middleware(h.Cfg.JWTSecret)
	handler := mw(http.HandlerFunc(h.AnalyticsOverview))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
