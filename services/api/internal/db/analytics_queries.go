package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsSnapshot represents a row in the analytics_snapshots table.
type AnalyticsSnapshot struct {
	ProfileID             string    `json:"profile_id"`
	SnapshotDate          time.Time `json:"snapshot_date"`
	AttentionScore        *float64  `json:"attention_score"`
	MemoryScore           *float64  `json:"memory_score"`
	EngagementFrequency   *int      `json:"engagement_frequency"`
	CoopParticipationRate *float64  `json:"coop_participation_rate"`
	AvgReactionTimeMs     *float64  `json:"avg_reaction_time_ms"`
	AvgHesitationMs       *float64  `json:"avg_hesitation_ms"`
	RetryRate             *float64  `json:"retry_rate"`
}

// GetLatestSnapshot retrieves the most recent analytics snapshot for a profile.
func GetLatestSnapshot(ctx context.Context, pool *pgxpool.Pool, profileID string) (AnalyticsSnapshot, error) {
	var s AnalyticsSnapshot
	err := pool.QueryRow(ctx,
		`SELECT profile_id, snapshot_date, attention_score, memory_score,
		        engagement_frequency, coop_participation_rate, avg_reaction_time_ms,
		        avg_hesitation_ms, retry_rate
		 FROM analytics_snapshots
		 WHERE profile_id = $1
		 ORDER BY snapshot_date DESC
		 LIMIT 1`,
		profileID,
	).Scan(
		&s.ProfileID, &s.SnapshotDate, &s.AttentionScore, &s.MemoryScore,
		&s.EngagementFrequency, &s.CoopParticipationRate, &s.AvgReactionTimeMs,
		&s.AvgHesitationMs, &s.RetryRate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AnalyticsSnapshot{}, ErrNotFound
		}
		return AnalyticsSnapshot{}, fmt.Errorf("GetLatestSnapshot: %w", err)
	}
	return s, nil
}

// GetSnapshotRange retrieves analytics snapshots for the last N days for a profile.
func GetSnapshotRange(ctx context.Context, pool *pgxpool.Pool, profileID string, days int) ([]AnalyticsSnapshot, error) {
	rows, err := pool.Query(ctx,
		`SELECT profile_id, snapshot_date, attention_score, memory_score,
		        engagement_frequency, coop_participation_rate, avg_reaction_time_ms,
		        avg_hesitation_ms, retry_rate
		 FROM analytics_snapshots
		 WHERE profile_id = $1
		   AND snapshot_date >= CURRENT_DATE - $2::int
		 ORDER BY snapshot_date DESC`,
		profileID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSnapshotRange: %w", err)
	}
	defer rows.Close()

	var snapshots []AnalyticsSnapshot
	for rows.Next() {
		var s AnalyticsSnapshot
		if err := rows.Scan(
			&s.ProfileID, &s.SnapshotDate, &s.AttentionScore, &s.MemoryScore,
			&s.EngagementFrequency, &s.CoopParticipationRate, &s.AvgReactionTimeMs,
			&s.AvgHesitationMs, &s.RetryRate,
		); err != nil {
			return nil, fmt.Errorf("GetSnapshotRange: scan: %w", err)
		}
		snapshots = append(snapshots, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetSnapshotRange: rows: %w", err)
	}

	return snapshots, nil
}
