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

func TestSetTokenCookies(t *testing.T) {
	w := httptest.NewRecorder()
	auth.SetTokenCookies(w, "access-val", "refresh-val", true)

	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	var access, refresh *http.Cookie
	for _, c := range cookies {
		switch c.Name {
		case "access_token":
			access = c
		case "refresh_token":
			refresh = c
		}
	}

	if access == nil {
		t.Fatal("missing access_token cookie")
	}
	if access.Value != "access-val" {
		t.Fatalf("expected access_token value 'access-val', got %s", access.Value)
	}
	if !access.HttpOnly {
		t.Fatal("access_token should be HttpOnly")
	}
	if !access.Secure {
		t.Fatal("access_token should be Secure")
	}
	if access.SameSite != http.SameSiteStrictMode {
		t.Fatal("access_token should be SameSite=Strict")
	}
	if access.MaxAge != 3600 {
		t.Fatalf("expected access_token MaxAge 3600, got %d", access.MaxAge)
	}

	if refresh == nil {
		t.Fatal("missing refresh_token cookie")
	}
	if refresh.Path != "/api/auth/refresh" {
		t.Fatalf("expected refresh_token path /api/auth/refresh, got %s", refresh.Path)
	}
}

func TestClearTokenCookies(t *testing.T) {
	w := httptest.NewRecorder()
	auth.ClearTokenCookies(w)

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.MaxAge != -1 {
			t.Fatalf("expected MaxAge -1 for cookie %s, got %d", c.Name, c.MaxAge)
		}
	}
}

func TestMiddleware_NoCookie(t *testing.T) {
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
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid-token"})
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
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
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
