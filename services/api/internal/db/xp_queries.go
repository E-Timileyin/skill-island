package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AddXPToProfile atomically adds XP to a student profile and returns the new total.
func AddXPToProfile(ctx context.Context, pool *pgxpool.Pool, profileID string, xpAmount int) (int, error) {
	var newTotalXP int
	err := pool.QueryRow(ctx,
		`UPDATE student_profiles
		 SET total_xp = total_xp + $1
		 WHERE id = $2
		 RETURNING total_xp`,
		xpAmount, profileID,
	).Scan(&newTotalXP)
	if err != nil {
		return 0, fmt.Errorf("AddXPToProfile: %w", err)
	}
	return newTotalXP, nil
}

// AddStarsToProfile atomically adds stars to a student profile and returns the new total.
func AddStarsToProfile(ctx context.Context, pool *pgxpool.Pool, profileID string, stars int) (int, error) {
	var newTotalStars int
	err := pool.QueryRow(ctx,
		`UPDATE student_profiles
		 SET total_stars = total_stars + $1
		 WHERE id = $2
		 RETURNING total_stars`,
		stars, profileID,
	).Scan(&newTotalStars)
	if err != nil {
		return 0, fmt.Errorf("AddStarsToProfile: %w", err)
	}
	return newTotalStars, nil
}
