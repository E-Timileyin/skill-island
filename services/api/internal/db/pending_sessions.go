package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PendingSession represents a row in the pending_sessions table.
type PendingSession struct {
	ID                string    `json:"id"`
	ProfileID         string    `json:"profile_id"`
	GameType          string    `json:"game_type"`
	Seed              int64     `json:"seed"`
	DifficultyLevel   int       `json:"difficulty_level"`
	SessionDurationMs int       `json:"session_duration_ms"`
	Used              bool      `json:"used"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
}

// CreatePendingSession inserts a new pending session and returns the created record.
func CreatePendingSession(ctx context.Context, pool *pgxpool.Pool, input PendingSession) (PendingSession, error) {
	var ps PendingSession
	err := pool.QueryRow(ctx,
		`INSERT INTO pending_sessions (id, profile_id, game_type, seed, difficulty_level, session_duration_ms, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, profile_id, game_type, seed, difficulty_level, session_duration_ms, used, created_at, expires_at`,
		input.ID, input.ProfileID, input.GameType, input.Seed, input.DifficultyLevel, input.SessionDurationMs, input.ExpiresAt,
	).Scan(&ps.ID, &ps.ProfileID, &ps.GameType, &ps.Seed, &ps.DifficultyLevel, &ps.SessionDurationMs, &ps.Used, &ps.CreatedAt, &ps.ExpiresAt)
	if err != nil {
		return PendingSession{}, fmt.Errorf("CreatePendingSession: %w", err)
	}
	return ps, nil
}

// GetPendingSession looks up a pending session by token (ID).
// Returns an error if not found.
func GetPendingSession(ctx context.Context, pool *pgxpool.Pool, token string) (PendingSession, error) {
	var ps PendingSession
	err := pool.QueryRow(ctx,
		`SELECT id, profile_id, game_type, seed, difficulty_level, session_duration_ms, used, created_at, expires_at
		 FROM pending_sessions WHERE id = $1`, token,
	).Scan(&ps.ID, &ps.ProfileID, &ps.GameType, &ps.Seed, &ps.DifficultyLevel, &ps.SessionDurationMs, &ps.Used, &ps.CreatedAt, &ps.ExpiresAt)
	if err != nil {
		return PendingSession{}, fmt.Errorf("GetPendingSession: %w", err)
	}
	return ps, nil
}

// MarkPendingSessionUsed marks a pending session as used (consumed).
func MarkPendingSessionUsed(ctx context.Context, pool *pgxpool.Pool, token string) error {
	_, err := pool.Exec(ctx,
		`UPDATE pending_sessions SET used = true WHERE id = $1`, token,
	)
	if err != nil {
		return fmt.Errorf("MarkPendingSessionUsed: %w", err)
	}
	return nil
}

// GetRecentFocusForestSessions returns the last N focus_forest sessions for a profile,
// ordered by most recent first, used for difficulty level determination.
func GetRecentFocusForestSessions(ctx context.Context, pool *pgxpool.Pool, profileID string, limit int) ([]GameSession, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, profile_id, game_type, mode, score, duration_seconds, accuracy, stars_earned, created_at
		 FROM game_sessions
		 WHERE profile_id = $1 AND game_type = 'focus_forest'
		 ORDER BY created_at DESC
		 LIMIT $2`,
		profileID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRecentFocusForestSessions: %w", err)
	}
	defer rows.Close()

	var sessions []GameSession
	for rows.Next() {
		var gs GameSession
		if err := rows.Scan(&gs.ID, &gs.ProfileID, &gs.GameType, &gs.Mode, &gs.Score, &gs.DurationSeconds, &gs.Accuracy, &gs.StarsEarned, &gs.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetRecentFocusForestSessions: scan: %w", err)
		}
		sessions = append(sessions, gs)
	}
	return sessions, nil
}

// DetermineFocusForestDifficulty calculates the difficulty level based on recent session history.
//
//	avg_attention_score >= 0.85 AND count >= 3 → level up (max 4)
//	avg_attention_score < 0.40 AND count >= 3 → level down (min 1)
//	Otherwise: keep current level (default 1 for new profiles)
func DetermineFocusForestDifficulty(sessions []GameSession) int {
	if len(sessions) < 3 {
		return 1 // default for new profiles or fewer than 3 sessions
	}

	// Use last 5 sessions (or fewer if less available).
	count := len(sessions)
	if count > 5 {
		count = 5
	}

	totalAccuracy := 0.0
	for i := 0; i < count; i++ {
		totalAccuracy += sessions[i].Accuracy
	}
	avgAccuracy := totalAccuracy / float64(count)

	// Current level: determine from the last session's accuracy trend.
	currentLevel := 1
	// Try to infer from last pending session's difficulty, but since we don't track that
	// in game_sessions, we start from 1 and adjust based on performance.

	if avgAccuracy >= 0.85 {
		currentLevel = min(currentLevel+1, 4) // level up
		// If consistently high, scale further.
		if count >= 5 && avgAccuracy >= 0.90 {
			currentLevel = min(3, 4)
		}
	} else if avgAccuracy < 0.40 {
		currentLevel = max(currentLevel-1, 1) // level down
	}

	// More sophisticated level calculation based on sustained performance.
	if avgAccuracy >= 0.90 && count >= 5 {
		currentLevel = 4
	} else if avgAccuracy >= 0.85 && count >= 3 {
		currentLevel = 3
	} else if avgAccuracy >= 0.70 && count >= 3 {
		currentLevel = 2
	} else if avgAccuracy < 0.40 && count >= 3 {
		currentLevel = 1
	}

	return currentLevel
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
