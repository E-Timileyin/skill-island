package handlers

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/E-Timileyin/skill-island/services/api/internal/validator"
	"github.com/google/uuid"
)

// SessionInitRequest is the expected JSON body for POST /api/sessions/init.
type SessionInitRequest struct {
	GameType string `json:"game_type"`
	Mode     string `json:"mode"`
}

// SessionInitResponse is the response for POST /api/sessions/init.
type SessionInitResponse struct {
	SessionToken    string `json:"session_token"`
	Seed            int64  `json:"seed"`
	DifficultyLevel int    `json:"difficulty_level,omitempty"`
	SessionDurationMs int  `json:"session_duration_ms,omitempty"`
}

// ManifestResponse is a spawn event returned from GET /api/sessions/manifest.
type ManifestResponse struct {
	TargetID    string  `json:"target_id"`
	TargetType  string  `json:"target_type"`
	SpawnTimeMs int     `json:"spawn_time_ms"`
	PositionX   float64 `json:"position_x"`
	PositionY   float64 `json:"position_y"`
}

// InitSession handles POST /api/sessions/init.
// Creates a pending session with a server-generated seed and returns the session token.
func (h *Handler) InitSession(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "not authenticated"})
		return
	}
	if claims.Role != "student" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "only students can init sessions"})
		return
	}
	if claims.ProfileID == "" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "student profile required"})
		return
	}

	var req SessionInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	if !validGameTypes[req.GameType] {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "invalid game_type"})
		return
	}

	// Generate a cryptographically random seed.
	seed := generateSeed()
	token := uuid.New().String()

	resp := SessionInitResponse{
		SessionToken: token,
		Seed:         seed,
	}

	difficultyLevel := 1
	sessionDurationMs := 60000

	if req.GameType == "focus_forest" {
		// Determine difficulty level from last 5 focus_forest sessions.
		recentSessions, err := db.GetRecentFocusForestSessions(r.Context(), h.DB, claims.ProfileID, 5)
		if err != nil {
			log.Printf("InitSession: GetRecentFocusForestSessions error: %v", err)
			// Non-fatal — default to level 1.
		} else {
			difficultyLevel = db.DetermineFocusForestDifficulty(recentSessions)
		}
		resp.DifficultyLevel = difficultyLevel
		resp.SessionDurationMs = sessionDurationMs
	}

	// Store pending session.
	ps := db.PendingSession{
		ID:                token,
		ProfileID:         claims.ProfileID,
		GameType:          req.GameType,
		Seed:              seed,
		DifficultyLevel:   difficultyLevel,
		SessionDurationMs: sessionDurationMs,
		ExpiresAt:         time.Now().Add(10 * time.Minute), // token valid for 10 minutes
	}

	if _, err := db.CreatePendingSession(r.Context(), h.DB, ps); err != nil {
		log.Printf("InitSession: CreatePendingSession error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to create session"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetManifest handles GET /api/sessions/manifest.
// Regenerates the spawn manifest from the stored seed and difficulty.
func (h *Handler) GetManifest(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "not authenticated"})
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "token query parameter required"})
		return
	}

	// Look up pending session by token.
	ps, err := db.GetPendingSession(r.Context(), h.DB, token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, APIError{Code: "NOT_FOUND", Message: "session token not found"})
		return
	}

	// Validate token not expired.
	if time.Now().After(ps.ExpiresAt) {
		writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_TOKEN_EXPIRED", Message: "session_token_expired"})
		return
	}

	// Validate token belongs to this profile.
	if ps.ProfileID != claims.ProfileID {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "token does not belong to this profile"})
		return
	}

	if ps.GameType != "focus_forest" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "manifest only available for focus_forest"})
		return
	}

	// Regenerate manifest from stored seed + difficulty_level.
	manifest := validator.GenerateSpawnManifest(ps.Seed, ps.SessionDurationMs, ps.DifficultyLevel)

	// Convert to response format.
	response := make([]ManifestResponse, len(manifest))
	for i, s := range manifest {
		response[i] = ManifestResponse{
			TargetID:    s.TargetID,
			TargetType:  s.TargetType,
			SpawnTimeMs: s.SpawnTimeMs,
			PositionX:   s.PositionX,
			PositionY:   s.PositionY,
		}
	}

	// Do NOT consume token yet — token consumed on final POST /api/sessions.
	writeJSON(w, http.StatusOK, response)
}

// generateSeed creates a cryptographically random int64 seed.
func generateSeed() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback to time-based seed if crypto/rand fails.
		return time.Now().UnixNano()
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}
