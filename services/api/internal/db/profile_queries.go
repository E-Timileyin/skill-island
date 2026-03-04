package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateProfile inserts a new student profile and returns the created row.
func CreateProfile(ctx context.Context, pool *pgxpool.Pool, userID, nickname string, avatarID int, playMode string) (StudentProfile, error) {
	var p StudentProfile
	err := pool.QueryRow(ctx,
		`INSERT INTO student_profiles (user_id, nickname, avatar_id, play_mode)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, nickname, avatar_id, total_stars, total_xp, play_mode, created_at`,
		userID, nickname, avatarID, playMode,
	).Scan(&p.ID, &p.UserID, &p.Nickname, &p.AvatarID, &p.TotalStars, &p.TotalXP, &p.PlayMode, &p.CreatedAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return StudentProfile{}, ErrDuplicateProfile
		}
		return StudentProfile{}, fmt.Errorf("CreateProfile: %w", err)
	}
	return p, nil
}

// ProfileUpdate holds the optional fields for a partial profile update.
type ProfileUpdate struct {
	Nickname *string
	AvatarID *int
	PlayMode *string
}

// UpdateProfile applies a partial update to a student profile and returns the updated row.
func UpdateProfile(ctx context.Context, pool *pgxpool.Pool, profileID string, fields ProfileUpdate) (StudentProfile, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if fields.Nickname != nil {
		setClauses = append(setClauses, fmt.Sprintf("nickname = $%d", argIdx))
		args = append(args, *fields.Nickname)
		argIdx++
	}
	if fields.AvatarID != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_id = $%d", argIdx))
		args = append(args, *fields.AvatarID)
		argIdx++
	}
	if fields.PlayMode != nil {
		setClauses = append(setClauses, fmt.Sprintf("play_mode = $%d", argIdx))
		args = append(args, *fields.PlayMode)
		argIdx++
	}

	if len(setClauses) == 0 {
		return StudentProfile{}, fmt.Errorf("UpdateProfile: no fields to update")
	}

	query := fmt.Sprintf(
		`UPDATE student_profiles SET %s WHERE id = $%d
		 RETURNING id, user_id, nickname, avatar_id, total_stars, total_xp, play_mode, created_at`,
		strings.Join(setClauses, ", "), argIdx,
	)
	args = append(args, profileID)

	var p StudentProfile
	err := pool.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.UserID, &p.Nickname, &p.AvatarID, &p.TotalStars, &p.TotalXP, &p.PlayMode, &p.CreatedAt,
	)
	if err != nil {
		return StudentProfile{}, fmt.Errorf("UpdateProfile: %w", err)
	}
	return p, nil
}
