package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/config"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// APIError is the standard error response format.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	DB  *pgxpool.Pool
	Cfg config.Config
}

// Health returns the service health status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// registerRequest is the expected JSON body for POST /api/auth/register.
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// loginRequest is the expected JSON body for POST /api/auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register creates a new user account and sets JWT cookies.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" || req.Role == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "email, password, and role are required"})
		return
	}

	validRoles := map[string]bool{"student": true, "parent": true, "educator": true}
	if !validRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "role must be student, parent, or educator"})
		return
	}

	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "password must be at least 8 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Register: bcrypt error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to process registration"})
		return
	}

	user, err := db.CreateUser(r.Context(), h.DB, req.Email, string(hash), req.Role)
	if err != nil {
		if errors.Is(err, db.ErrDuplicateEmail) {
			writeJSON(w, http.StatusConflict, APIError{Code: "DUPLICATE_EMAIL", Message: "email already registered"})
			return
		}
		log.Printf("Register: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to create user"})
		return
	}

	profileID := ""
	if user.Role == "student" {
		profile, err := db.GetStudentProfileByUserID(r.Context(), h.DB, user.ID)
		if err == nil {
			profileID = profile.ID
		}
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Role, profileID, h.Cfg.JWTSecret)
	if err != nil {
		log.Printf("Register: token error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, h.Cfg.JWTRefreshSecret)
	if err != nil {
		log.Printf("Register: refresh token error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to generate token"})
		return
	}

	secure := h.Cfg.Env != "development"
	auth.SetTokenCookies(w, accessToken, refreshToken, secure)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login authenticates a user and sets JWT cookies.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "email and password are required"})
		return
	}

	user, err := db.GetUserByEmail(r.Context(), h.DB, req.Email)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, APIError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"})
			return
		}
		log.Printf("Login: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "login failed"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"})
		return
	}

	if err := db.UpdateLastLogin(r.Context(), h.DB, user.ID); err != nil {
		log.Printf("Login: update last login error: %v", err)
	}

	profileID := ""
	if user.Role == "student" {
		profile, err := db.GetStudentProfileByUserID(r.Context(), h.DB, user.ID)
		if err == nil {
			profileID = profile.ID
		}
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Role, profileID, h.Cfg.JWTSecret)
	if err != nil {
		log.Printf("Login: token error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, h.Cfg.JWTRefreshSecret)
	if err != nil {
		log.Printf("Login: refresh token error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to generate token"})
		return
	}

	secure := h.Cfg.Env != "development"
	auth.SetTokenCookies(w, accessToken, refreshToken, secure)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":   user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Logout clears the JWT cookies.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearTokenCookies(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Refresh validates the refresh token and issues new access/refresh tokens.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "missing refresh token"})
		return
	}

	userID, err := auth.ValidateRefreshToken(cookie.Value, h.Cfg.JWTRefreshSecret)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "invalid or expired refresh token"})
		return
	}

	user, err := db.GetUserByID(r.Context(), h.DB, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "user not found"})
			return
		}
		log.Printf("Refresh: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "refresh failed"})
		return
	}

	profileID := ""
	if user.Role == "student" {
		profile, err := db.GetStudentProfileByUserID(r.Context(), h.DB, user.ID)
		if err == nil {
			profileID = profile.ID
		}
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Role, profileID, h.Cfg.JWTSecret)
	if err != nil {
		log.Printf("Refresh: token error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, h.Cfg.JWTRefreshSecret)
	if err != nil {
		log.Printf("Refresh: refresh token error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to generate token"})
		return
	}

	secure := h.Cfg.Env != "development"
	auth.SetTokenCookies(w, accessToken, refreshToken, secure)

	writeJSON(w, http.StatusOK, map[string]string{"message": "tokens refreshed"})
}

// Me returns the current authenticated user's information.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "not authenticated"})
		return
	}

	user, err := db.GetUserByID(r.Context(), h.DB, claims.UserID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "user not found"})
			return
		}
		log.Printf("Me: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to fetch user"})
		return
	}

	resp := map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "student" {
		profile, err := db.GetStudentProfileByUserID(r.Context(), h.DB, user.ID)
		if err == nil {
			resp["profile"] = map[string]interface{}{
				"id":          profile.ID,
				"nickname":    profile.Nickname,
				"avatar_id":   profile.AvatarID,
				"total_stars": profile.TotalStars,
				"total_xp":    profile.TotalXP,
				"play_mode":   profile.PlayMode,
			}
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
