package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
)

// WeeklySummary is the JSON response for GET /api/analytics/overview.
type WeeklySummary struct {
	AttentionScore        *float64 `json:"attention_score"`
	MemoryScore           *float64 `json:"memory_score"`
	EngagementFrequency   int      `json:"engagement_frequency"`
	CoopParticipationRate *float64 `json:"coop_participation_rate"`
	AvgReactionTimeMs     *float64 `json:"avg_reaction_time_ms"`
	TotalStars            int      `json:"total_stars"`
	TotalXP               int      `json:"total_xp"`
	SessionsThisWeek      int      `json:"sessions_this_week"`
	SnapshotDate          string   `json:"snapshot_date"`
	Message               string   `json:"message,omitempty"`
}

// requireDashboardRole extracts claims and checks the role is "parent" or "educator".
// Returns the claims on success, or writes a JSON error and returns nil.
func requireDashboardRole(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, APIError{Code: "UNAUTHORIZED", Message: "not authenticated"})
		return nil
	}
	if claims.Role != "parent" && claims.Role != "educator" {
		writeJSON(w, http.StatusForbidden, APIError{Code: "FORBIDDEN", Message: "parent or educator role required"})
		return nil
	}
	return claims
}

// AnalyticsOverview handles GET /api/analytics/overview.
// Requires auth with parent or educator role.
// Accepts optional ?profile_id= query parameter for the student profile to query.
func (h *Handler) AnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	claims := requireDashboardRole(w, r)
	if claims == nil {
		return
	}

	profileID := r.URL.Query().Get("profile_id")
	if profileID == "" {
		writeJSON(w, http.StatusBadRequest, APIError{
			Code:    "VALIDATION_ERROR",
			Message: "profile_id query parameter is required",
		})
		return
	}

	// Set cache header — 15-minute TTL.
	w.Header().Set("Cache-Control", "private, max-age=900")

	snapshot, err := db.GetLatestSnapshot(r.Context(), h.DB, profileID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeJSON(w, http.StatusOK, WeeklySummary{
				Message: "No data yet",
			})
			return
		}
		log.Printf("AnalyticsOverview: db error: %v", err)
		writeJSON(w, http.StatusInternalServerError, APIError{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch analytics",
		})
		return
	}

	// Look up profile for total_stars and total_xp.
	var totalStars, totalXP int
	profile, err := db.GetStudentProfileByID(r.Context(), h.DB, profileID)
	if err == nil {
		totalStars = profile.TotalStars
		totalXP = profile.TotalXP
	} else if !errors.Is(err, db.ErrNotFound) {
		log.Printf("AnalyticsOverview: GetStudentProfileByID error: %v", err)
	}

	engagementFreq := 0
	if snapshot.EngagementFrequency != nil {
		engagementFreq = *snapshot.EngagementFrequency
	}

	writeJSON(w, http.StatusOK, WeeklySummary{
		AttentionScore:        snapshot.AttentionScore,
		MemoryScore:           snapshot.MemoryScore,
		EngagementFrequency:   engagementFreq,
		CoopParticipationRate: snapshot.CoopParticipationRate,
		AvgReactionTimeMs:     snapshot.AvgReactionTimeMs,
		TotalStars:            totalStars,
		TotalXP:               totalXP,
		SessionsThisWeek:      engagementFreq,
		SnapshotDate:          snapshot.SnapshotDate.Format("2006-01-02"),
	})
}
