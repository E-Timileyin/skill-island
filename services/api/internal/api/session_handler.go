package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/E-Timileyin/skill-island/services/api/internal/validator"
)

// SessionSubmission is the expected JSON body for POST /api/sessions.
// The client submits actions only — scores are never accepted from the client.
type SessionSubmission struct {
	GameType      string            `json:"game_type"`
	Mode          string            `json:"mode"`
	Actions       []json.RawMessage `json:"actions"`
	DurationMs    int               `json:"duration_ms"`
	RoomSessionID string            `json:"room_session_id,omitempty"`
}

// SessionResult is the response returned after session validation and storage.
type SessionResult struct {
	Score                  int      `json:"score"`
	Accuracy               float64  `json:"accuracy"`
	StarsEarned            int      `json:"stars_earned"`
	XPEarned               int      `json:"xp_earned"`
	TotalXP                int      `json:"total_xp"`
	UnlockedZones          []string `json:"unlocked_zones"`
	BehavioralMetricsCount int      `json:"behavioral_metrics_count"`
}

// maxActionsPerSession is the upper bound on actions a client may submit.
// Sessions exceeding this are rejected as implausible (anti-DoS measure).
const maxActionsPerSession = 10000

// validGameTypes is the set of accepted game_type values.
var validGameTypes = map[string]bool{
	"memory_cove":  true,
	"focus_forest": true,
	"team_tower":   true,
}

// validModes is the set of accepted mode values.
var validModes = map[string]bool{
	"solo":        true,
	"cooperative": true,
}

// SubmitSession handles POST /api/sessions.
// It validates the client action log, computes score/stars/XP server-side,
// writes the session and behavioral metrics in a transaction, and returns the result.
func (h *Handler) SubmitSession(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "not authenticated"})
		return
	}

	if claims.Role != "student" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "only students can submit sessions"})
		return
	}

	if claims.ProfileID == "" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "student profile required"})
		return
	}

	var sub SessionSubmission
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	if !validGameTypes[sub.GameType] {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "invalid game_type"})
		return
	}

	if !validModes[sub.Mode] {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "invalid mode"})
		return
	}

	if sub.DurationMs <= 0 {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "duration_ms must be positive"})
		return
	}

	if len(sub.Actions) == 0 {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "actions must not be empty"})
		return
	}

	// Validate actions and compute score server-side.
	valResult := validateActions(sub)
	if valResult.Rejected {
		log.Printf("SubmitSession: rejected session for profile %s: %s (action_count=%d)", claims.ProfileID, valResult.RejectReason, len(sub.Actions))
		writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_REJECTED", Message: valResult.RejectReason})
		return
	}

	// Calculate XP from stars.
	xpEarned := validator.CalculateXP(valResult.StarsEarned)

	// Convert duration from ms to seconds (round up so sub-second sessions are at least 1s).
	durationSeconds := (sub.DurationMs + 999) / 1000

	// Write game_sessions + behavioral_metrics in a single transaction.
	sessionInput := db.CreateSessionInput{
		ProfileID:       claims.ProfileID,
		GameType:        sub.GameType,
		Mode:            sub.Mode,
		Score:           valResult.Score,
		DurationSeconds: durationSeconds,
		Accuracy:        valResult.Accuracy,
		StarsEarned:     valResult.StarsEarned,
	}

	_, metricsCount, err := db.CreateGameSessionWithMetrics(r.Context(), h.DB, sessionInput, toDBMetrics(valResult.Metrics))
	if err != nil {
		log.Printf("SubmitSession: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to save session"})
		return
	}

	// Update profile XP and stars.
	newTotalXP, err := db.AddXPToProfile(r.Context(), h.DB, claims.ProfileID, xpEarned)
	if err != nil {
		log.Printf("SubmitSession: AddXPToProfile error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to update XP"})
		return
	}

	if _, err := db.AddStarsToProfile(r.Context(), h.DB, claims.ProfileID, valResult.StarsEarned); err != nil {
		log.Printf("SubmitSession: AddStarsToProfile error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to update stars"})
		return
	}

	// Determine unlocked zones.
	unlockedZones := validator.CheckUnlockedZones(newTotalXP)

	writeJSON(w, http.StatusCreated, SessionResult{
		Score:                  valResult.Score,
		Accuracy:               valResult.Accuracy,
		StarsEarned:            valResult.StarsEarned,
		XPEarned:               xpEarned,
		TotalXP:                newTotalXP,
		UnlockedZones:          unlockedZones,
		BehavioralMetricsCount: metricsCount,
	})
}

// validateActions dispatches to the appropriate validator based on game_type.
// Currently returns a placeholder result — full per-game validators (memory, focus, tower)
// will be implemented in later phases.
func validateActions(sub SessionSubmission) validator.ValidationResult {
	actionCount := len(sub.Actions)

	// Reject sessions with implausible action counts.
	if actionCount > maxActionsPerSession {
		return validator.ValidationResult{
			Rejected:     true,
			RejectReason: "implausible action count",
		}
	}

	// Placeholder: compute basic score from action count.
	// Per-game validators (memory.go, focus.go, tower.go) will override this.
	score := actionCount * 10
	accuracy := 0.0
	if actionCount > 0 {
		accuracy = 1.0
	}

	// Star thresholds (generic placeholder until per-game validators exist).
	starsEarned := 0
	switch {
	case accuracy >= 0.90:
		starsEarned = 3
	case accuracy >= 0.70:
		starsEarned = 2
	case accuracy >= 0.50:
		starsEarned = 1
	}

	// Build placeholder behavioral metrics from actions.
	metrics := make([]validator.BehavioralMetric, 0, actionCount)
	for i := range sub.Actions {
		offsetMs := (i + 1) * 100
		metrics = append(metrics, validator.BehavioralMetric{
			EventType:         "action",
			ReactionTimeMs:    nil,
			HesitationMs:      nil,
			RetryCount:        0,
			Correct:           true,
			TimestampOffsetMs: offsetMs,
			Metadata:          sub.Actions[i],
		})
	}

	return validator.ValidationResult{
		Score:       score,
		Accuracy:    accuracy,
		StarsEarned: starsEarned,
		Metrics:     metrics,
	}
}

// toDBMetrics converts validator metrics to db metrics.
func toDBMetrics(vMetrics []validator.BehavioralMetric) []db.BehavioralMetric {
	dbMetrics := make([]db.BehavioralMetric, len(vMetrics))
	for i, m := range vMetrics {
		dbMetrics[i] = db.BehavioralMetric{
			EventType:         m.EventType,
			ReactionTimeMs:    m.ReactionTimeMs,
			HesitationMs:      m.HesitationMs,
			RetryCount:        m.RetryCount,
			Correct:           m.Correct,
			TimestampOffsetMs: m.TimestampOffsetMs,
			Metadata:          m.Metadata,
		}
	}
	return dbMetrics
}
