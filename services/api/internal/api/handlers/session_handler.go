package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/E-Timileyin/skill-island/services/api/internal/validator"
)

// SessionSubmission is the expected JSON body for POST /api/sessions.
// The client submits actions only — scores are never accepted from the client.
type SessionSubmission struct {
	SessionToken  string            `json:"session_token,omitempty"`
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

	// Focus Forest specific validation.
	if sub.GameType == "focus_forest" {
		if len(sub.Actions) > 300 {
			writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_REJECTED", Message: "implausible action count for focus_forest"})
			return
		}
		if sub.DurationMs < 5000 {
			writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_REJECTED", Message: "impossibly short session"})
			return
		}
	}

	// Look up pending session for seed and difficulty.
	var seed int64
	var difficultyLevel int = 1
	if sub.SessionToken != "" {
		ps, err := db.GetPendingSession(r.Context(), h.DB, sub.SessionToken)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_TOKEN_NOT_FOUND", Message: "session token not found"})
			return
		}
		if ps.Used {
			writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_TOKEN_ALREADY_USED", Message: "session_token_already_used"})
			return
		}
		if time.Now().After(ps.ExpiresAt) {
			writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_TOKEN_EXPIRED", Message: "session_token_expired"})
			return
		}
		seed = ps.Seed
		difficultyLevel = ps.DifficultyLevel
	}

	// Validate actions and compute score server-side.
	valResult := validateActions(sub, seed, difficultyLevel)
	if valResult.Rejected {
		log.Printf("SubmitSession: rejected session for profile %s: %s (action_count=%d)", claims.ProfileID, valResult.RejectReason, len(sub.Actions))
		writeJSON(w, http.StatusUnprocessableEntity, APIError{Code: "SESSION_REJECTED", Message: valResult.RejectReason})
		return
	}

	// Calculate XP from stars.
	xpEarned := validator.CalculateXP(sub.GameType, valResult.StarsEarned, sub.DurationMs/1000)

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

	if sub.SessionToken != "" {
		if err := db.MarkPendingSessionUsed(r.Context(), h.DB, sub.SessionToken); err != nil {
			log.Printf("SubmitSession: MarkPendingSessionUsed error: %v", err)
			// Non-fatal, session was already saved
		}
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

// CoopSessionSubmission is the expected JSON body for POST /api/sessions/coop.
type CoopSessionSubmission struct {
	GameType      string `json:"game_type"`
	Mode          string `json:"mode"`
	RoomSessionID string `json:"room_session_id"`
	Outcome       string `json:"outcome"`
	DurationMs    int    `json:"duration_ms"`
}

// SubmitCoopSession handles POST /api/sessions/coop.
// The room's score is already persisted by the WS hub; this endpoint
// returns the caller's updated profile totals for immediate UI display.
func (h *Handler) SubmitCoopSession(w http.ResponseWriter, r *http.Request) {
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

	var sub CoopSessionSubmission
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "BAD_REQUEST", Message: "invalid request body"})
		return
	}

	if sub.GameType != "team_tower" {
		writeJSON(w, http.StatusBadRequest, APIError{Code: "VALIDATION_ERROR", Message: "coop endpoint only valid for team_tower"})
		return
	}

	// Fetch the current profile to return up-to-date totals.
	// XP/stars were already written by the WS room; we just read them back.
	profile, err := db.GetStudentProfileByID(r.Context(), h.DB, claims.ProfileID)
	if err != nil {
		log.Printf("SubmitCoopSession: GetProfileByID error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{Code: "INTERNAL_ERROR", Message: "failed to fetch profile"})
		return
	}

	unlockedZones := validator.CheckUnlockedZones(profile.TotalXP)

	writeJSON(w, http.StatusCreated, SessionResult{
		Score:         0,
		Accuracy:      0,
		StarsEarned:   1, // Coop stars already applied by WS room; return 1 as minimum display
		XPEarned:      0, // XP already credited by WS room
		TotalXP:       profile.TotalXP,
		UnlockedZones: unlockedZones,
	})
}

// validateActions dispatches to the appropriate validator based on game_type.
func validateActions(sub SessionSubmission, seed int64, difficultyLevel int) validator.ValidationResult {
	actionCount := len(sub.Actions)

	// Memory cove: max 500 actions.
	if sub.GameType == "memory_cove" && actionCount > 500 {
		return validator.ValidationResult{
			Rejected:     true,
			RejectReason: "implausible action count",
		}
	}

	if sub.GameType == "memory_cove" && sub.Mode == "solo" {
		var actions []validator.MemoryCoveAction
		for _, raw := range sub.Actions {
			var act validator.MemoryCoveAction
			if err := json.Unmarshal(raw, &act); err != nil {
				return validator.ValidationResult{
					Rejected:     true,
					RejectReason: "invalid action format",
				}
			}
			actions = append(actions, act)
		}
		roundsCompleted := 1 // TODO: derive from session context
		val := validator.ValidateActions(actions, seed, roundsCompleted)
		scoreRes := validator.CalculateScore(val, roundsCompleted)

		metrics := make([]validator.BehavioralMetric, len(actions))
		for i, act := range actions {
			metrics[i] = validator.BehavioralMetric{
				EventType:         act.Type,
				ReactionTimeMs:    nil,
				HesitationMs:      nil,
				RetryCount:        0,
				Correct:           val.Correct[i],
				TimestampOffsetMs: int(act.ClientTimestamp),
				Metadata:          sub.Actions[i],
			}
		}

		return validator.ValidationResult{
			Score:       scoreRes.Score,
			Accuracy:    scoreRes.Accuracy,
			StarsEarned: scoreRes.Stars,
			Metrics:     metrics,
		}
	}

	// Focus Forest validation.
	if sub.GameType == "focus_forest" && sub.Mode == "solo" {
		var actions []validator.FocusForestAction
		for _, raw := range sub.Actions {
			var act validator.FocusForestAction
			if err := json.Unmarshal(raw, &act); err != nil {
				return validator.ValidationResult{
					Rejected:     true,
					RejectReason: "invalid action format",
				}
			}
			// Clamp negative reaction times to 0.
			if act.ClientTimestamp < 0 {
				act.ClientTimestamp = 0
			}
			actions = append(actions, act)
		}

		// Regenerate manifest from stored seed + difficulty_level.
		manifest := validator.GenerateSpawnManifest(seed, validator.SESSION_DURATION, difficultyLevel)

		// Validate taps against manifest.
		tapResult := validator.ValidateTaps(actions, manifest, difficultyLevel)
		scoredResult := validator.CalculateAttentionScore(tapResult)

		// Build behavioral metrics for each tap action.
		metrics := make([]validator.BehavioralMetric, len(actions))
		for i, act := range actions {
			rt := int(tapResult.ReactionTimes[i])
			metrics[i] = validator.BehavioralMetric{
				EventType:         act.Type,
				ReactionTimeMs:    &rt,
				HesitationMs:      nil,
				RetryCount:        0,
				Correct:           tapResult.Correct[i],
				TimestampOffsetMs: int(act.ClientTimestamp),
				Metadata:          sub.Actions[i],
			}
		}

		// Use attention score as the score (scaled to 0-1000).
		score := int(scoredResult.AttentionScore * 1000)

		return validator.ValidationResult{
			Score:       score,
			Accuracy:    scoredResult.AttentionScore,
			StarsEarned: scoredResult.Stars,
			Metrics:     metrics,
		}
	}

	// Fallback: zero session (e.g., empty action log or unsupported game type).
	return validator.ValidationResult{
		Score:       0,
		Accuracy:    0,
		StarsEarned: 0,
		Metrics:     nil,
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
