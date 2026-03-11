package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-minimum-32-characters-long"
const testRefreshSecret = "test-refresh-secret-minimum-32-chars"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	userID := "user-123"
	role := "student"
	profileID := "profile-456"

	token, err := auth.GenerateAccessToken(userID, role, profileID, testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := auth.ValidateAccessToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Fatalf("expected role %s, got %s", role, claims.Role)
	}
	if claims.ProfileID != profileID {
		t.Fatalf("expected profile_id %s, got %s", profileID, claims.ProfileID)
	}
	if claims.Subject != userID {
		t.Fatalf("expected subject %s, got %s", userID, claims.Subject)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	token, err := auth.GenerateAccessToken("user-1", "student", "", testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	_, err = auth.ValidateAccessToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	// Create an expired token manually
	claims := auth.Claims{
		UserID: "user-1",
		Role:   "student",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "user-1",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = auth.ValidateAccessToken(signed, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	userID := "user-789"

	token, err := auth.GenerateRefreshToken(userID, testRefreshSecret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	subject, err := auth.ValidateRefreshToken(token, testRefreshSecret)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}

	if subject != userID {
		t.Fatalf("expected subject %s, got %s", userID, subject)
	}
}

func TestValidateRefreshToken_WrongSecret(t *testing.T) {
	token, err := auth.GenerateRefreshToken("user-1", testRefreshSecret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	_, err = auth.ValidateRefreshToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestSetRefreshTokenCookie(t *testing.T) {
	w := httptest.NewRecorder()
	auth.SetRefreshTokenCookie(w, "refresh-val", true)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	refresh := cookies[0]
	if refresh.Name != "refresh_token" {
		t.Fatalf("expected cookie name 'refresh_token', got %s", refresh.Name)
	}
	if refresh.Value != "refresh-val" {
		t.Fatalf("expected refresh_token value 'refresh-val', got %s", refresh.Value)
	}
	if !refresh.HttpOnly {
		t.Fatal("refresh_token should be HttpOnly")
	}
	if !refresh.Secure {
		t.Fatal("refresh_token should be Secure")
	}
	if refresh.Path != "/api/auth/refresh" {
		t.Fatalf("expected refresh_token path /api/auth/refresh, got %s", refresh.Path)
	}
	if refresh.MaxAge != 7*24*3600 {
		t.Fatalf("expected refresh_token MaxAge %d, got %d", 7*24*3600, refresh.MaxAge)
	}
}

func TestClearRefreshTokenCookie(t *testing.T) {
	w := httptest.NewRecorder()
	auth.ClearRefreshTokenCookie(w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	c := cookies[0]
	if c.Name != "refresh_token" {
		t.Fatalf("expected cookie name 'refresh_token', got %s", c.Name)
	}
	if c.MaxAge != -1 {
		t.Fatalf("expected MaxAge -1, got %d", c.MaxAge)
	}
}

func TestMiddleware_NoAuthHeader(t *testing.T) {
	mw := auth.Middleware(testSecret)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	mw := auth.Middleware(testSecret)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestMiddleware_ValidToken(t *testing.T) {
	token, err := auth.GenerateAccessToken("user-1", "student", "profile-1", testSecret)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	var gotClaims *auth.Claims
	mw := auth.Middleware(testSecret)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = auth.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if gotClaims == nil {
		t.Fatal("expected claims in context, got nil")
	}
	if gotClaims.UserID != "user-1" {
		t.Fatalf("expected user_id user-1, got %s", gotClaims.UserID)
	}
	if gotClaims.Role != "student" {
		t.Fatalf("expected role student, got %s", gotClaims.Role)
	}
	if gotClaims.ProfileID != "profile-1" {
		t.Fatalf("expected profile_id profile-1, got %s", gotClaims.ProfileID)
	}
}

func TestClaimsFromContext_NoClaimsReturnsNil(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	claims := auth.ClaimsFromContext(req.Context())
	if claims != nil {
		t.Fatal("expected nil claims from empty context")
	}
}
