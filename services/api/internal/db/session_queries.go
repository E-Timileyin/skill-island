package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GameSession represents a row in the game_sessions table.
type GameSession struct {
	ID              string    `json:"id"`
	ProfileID       string    `json:"profile_id"`
	GameType        string    `json:"game_type"`
	Mode            string    `json:"mode"`
	Score           int       `json:"score"`
	DurationSeconds int       `json:"duration_seconds"`
	Accuracy        float64   `json:"accuracy"`
	StarsEarned     int       `json:"stars_earned"`
	CreatedAt       time.Time `json:"created_at"`
}

// BehavioralMetric represents a row in the behavioral_metrics table.
type BehavioralMetric struct {
	SessionID         string          `json:"session_id"`
	EventType         string          `json:"event_type"`
	ReactionTimeMs    *int            `json:"reaction_time_ms"`
	HesitationMs      *int            `json:"hesitation_ms"`
	RetryCount        int             `json:"retry_count"`
	Correct           bool            `json:"correct"`
	TimestampOffsetMs int             `json:"timestamp_offset_ms"`
	Metadata          json.RawMessage `json:"metadata"`
}

// CreateSessionInput holds the data needed to create a game session.
type CreateSessionInput struct {
	ProfileID       string
	GameType        string
	Mode            string
	Score           int
	DurationSeconds int
	Accuracy        float64
	StarsEarned     int
}

// CreateGameSessionWithMetrics creates a game session and its associated
// behavioral metrics rows in a single database transaction. It returns the
// created GameSession and the number of metrics written.
func CreateGameSessionWithMetrics(ctx context.Context, pool *pgxpool.Pool, input CreateSessionInput, metrics []BehavioralMetric) (GameSession, int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return GameSession{}, 0, fmt.Errorf("CreateGameSessionWithMetrics: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert game session.
	var gs GameSession
	err = tx.QueryRow(ctx,
		`INSERT INTO game_sessions (profile_id, game_type, mode, score, duration_seconds, accuracy, stars_earned)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, profile_id, game_type, mode, score, duration_seconds, accuracy, stars_earned, created_at`,
		input.ProfileID, input.GameType, input.Mode, input.Score, input.DurationSeconds, input.Accuracy, input.StarsEarned,
	).Scan(&gs.ID, &gs.ProfileID, &gs.GameType, &gs.Mode, &gs.Score, &gs.DurationSeconds, &gs.Accuracy, &gs.StarsEarned, &gs.CreatedAt)
	if err != nil {
		return GameSession{}, 0, fmt.Errorf("CreateGameSessionWithMetrics: insert session: %w", err)
	}

	// Insert behavioral metrics in batch.
	if len(metrics) > 0 {
		rows := make([][]interface{}, len(metrics))
		for i, m := range metrics {
			rows[i] = []interface{}{
				gs.ID,
				m.EventType,
				m.ReactionTimeMs,
				m.HesitationMs,
				m.RetryCount,
				m.Correct,
				m.TimestampOffsetMs,
				m.Metadata,
			}
		}

		_, err = tx.CopyFrom(ctx,
			pgx.Identifier{"behavioral_metrics"},
			[]string{"session_id", "event_type", "reaction_time_ms", "hesitation_ms", "retry_count", "correct", "timestamp_offset_ms", "metadata"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return GameSession{}, 0, fmt.Errorf("CreateGameSessionWithMetrics: insert metrics: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return GameSession{}, 0, fmt.Errorf("CreateGameSessionWithMetrics: commit: %w", err)
	}

	return gs, len(metrics), nil
}
