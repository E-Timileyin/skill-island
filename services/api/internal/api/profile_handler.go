package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
)

// validPlayModes defines the allowed values for play_mode on student profiles.
var validPlayModes = map[string]bool{"solo": true, "team": true}

// requireStudent extracts claims and checks the role is "student".
// Returns the claims on success, or writes a JSON error and returns nil.
func requireStudent(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "not authenticated"})
		return nil
	}
	if claims.Role != "student" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "student role required"})
		return nil
	}
	return claims
}

// profileToResponse converts a db.StudentProfile to a ProfileResponse.
func profileToResponse(p db.StudentProfile) ProfileResponse {
	return ProfileResponse{
		ID:         p.ID,
		Nickname:   p.Nickname,
		AvatarID:   p.AvatarID,
		TotalStars: p.TotalStars,
		TotalXP:    p.TotalXP,
		PlayMode:   p.PlayMode,
		CreatedAt:  p.CreatedAt,
	}
}

// CreateProfile handles POST /api/profiles.
func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	claims := requireStudent(w, r)
	if claims == nil {
		return
	}

	var req createProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	if req.Nickname == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "nickname is required"})
		return
	}

	if req.PlayMode == "" {
		req.PlayMode = "solo"
	}

	if !validPlayModes[req.PlayMode] {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "play_mode must be solo or team"})
		return
	}

	avatarID := 0
	if req.AvatarID != nil {
		avatarID = *req.AvatarID
	}

	profile, err := db.CreateProfile(r.Context(), h.DB, claims.UserID, req.Nickname, avatarID, req.PlayMode)
	if err != nil {
		if errors.Is(err, db.ErrDuplicateProfile) {
			writeJSON(w, http.StatusConflict, APIError{Code: "DUPLICATE_PROFILE", Message: "profile already exists for this user"})
			return
		}
		log.Printf("CreateProfile: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to create profile"})
		return
	}

	writeJSON(w, http.StatusCreated, profileToResponse(profile))
}

// GetProfile handles GET /api/profiles/me.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := requireStudent(w, r)
	if claims == nil {
		return
	}

	profile, err := db.GetStudentProfileByUserID(r.Context(), h.DB, claims.UserID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "profile not found"})
			return
		}
		log.Printf("GetProfile: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to fetch profile"})
		return
	}

	writeJSON(w, http.StatusOK, profileToResponse(profile))
}

// UpdateProfile handles PATCH /api/profiles/me.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := requireStudent(w, r)
	if claims == nil {
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	if req.Nickname != nil && *req.Nickname == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "nickname cannot be empty"})
		return
	}

	if req.PlayMode != nil {
		if !validPlayModes[*req.PlayMode] {
			writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "play_mode must be solo or team"})
			return
		}
	}

	if req.Nickname == nil && req.AvatarID == nil && req.PlayMode == nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "at least one field must be provided"})
		return
	}

	// Look up the profile to get the profile ID.
	profile, err := db.GetStudentProfileByUserID(r.Context(), h.DB, claims.UserID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "profile not found"})
			return
		}
		log.Printf("UpdateProfile: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to fetch profile"})
		return
	}

	fields := db.ProfileUpdate{
		Nickname: req.Nickname,
		AvatarID: req.AvatarID,
		PlayMode: req.PlayMode,
	}

	updated, err := db.UpdateProfile(r.Context(), h.DB, profile.ID, fields)
	if err != nil {
		log.Printf("UpdateProfile: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to update profile"})
		return
	}

	writeJSON(w, http.StatusOK, profileToResponse(updated))
}
