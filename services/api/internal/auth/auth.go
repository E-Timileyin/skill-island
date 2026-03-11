package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getCookieDomain returns the cookie domain based on environment.
func getCookieDomain() string {
	env := os.Getenv("ENV")
	if env == "development" || env == "local" {
		fmt.Println("Cookie domain for local dev: ''")
		return "" // No domain for local dev
	}
	domain := os.Getenv("COOKIE_DOMAIN")
	if domain != "" {
		fmt.Printf("Cookie domain from env: %s\n", domain)
		return domain
	}
	fmt.Println("Cookie domain default: skill-island.vercel.app")
	return "skill-island.vercel.app" // Default production domain
}

// Claims represents the JWT claims for both access and refresh tokens.
type Claims struct {
	UserID    string `json:"user_id"`
	ProfileID string `json:"profile_id,omitempty"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

type contextKey string

const claimsKey contextKey = "claims"

// GenerateAccessToken creates a signed JWT access token with 1 hour expiry.
func GenerateAccessToken(userID, role, profileID, secret string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		ProfileID: profileID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("GenerateAccessToken: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken creates a signed JWT refresh token with 7 day expiry.
func GenerateRefreshToken(userID, secret string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("GenerateRefreshToken: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken parses and validates an access token, returning the claims.
func ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateAccessToken: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("ValidateAccessToken: invalid token claims")
	}
	return claims, nil
}

// ValidateRefreshToken parses and validates a refresh token, returning the subject (user ID).
func ValidateRefreshToken(tokenString, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("ValidateRefreshToken: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("ValidateRefreshToken: invalid token claims")
	}
	return claims.Subject, nil
}

// SetTokenCookies writes access and refresh token cookies to the response.
func SetTokenCookies(w http.ResponseWriter, accessToken, refreshToken string, secure bool) {
	env := os.Getenv("ENV")
	sameSite := http.SameSiteNoneMode
	if env == "development" || env == "local" {
		sameSite = http.SameSiteLaxMode
		secure = false
	}
	domain := getCookieDomain()
	fmt.Printf("[SetTokenCookies] ENV=%s Secure=%v SameSite=%v Domain='%s'\n", env, secure, sameSite, domain)
	fmt.Printf("[SetTokenCookies] access_token='%s' refresh_token='%s'\n", accessToken, refreshToken)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
		Path:     "/",
		MaxAge:   3600, // 1 hour
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
		Path:     "/api/auth/refresh",
		MaxAge:   7 * 24 * 3600, // 7 days
	})
}

// ClearTokenCookies removes the access and refresh token cookies.
func ClearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
	})
}

// Middleware returns an HTTP middleware that validates the access token cookie
// and adds claims to the request context.
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"missing access token"}`, http.StatusUnauthorized)
				return
			}

			claims, err := ValidateAccessToken(cookie.Value, secret)
			if err != nil {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves the JWT claims from the request context.
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}
