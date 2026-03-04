package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a query returns no rows.
var ErrNotFound = errors.New("record not found")

// ErrDuplicateEmail is returned when a user with the same email already exists.
var ErrDuplicateEmail = errors.New("email already registered")

// User represents a row in the users table.
type User struct {
	ID           string
	Email        string
	Role         string
	PasswordHash string
	CreatedAt    time.Time
	LastLoginAt  *time.Time
}

// StudentProfile represents a row in the student_profiles table.
type StudentProfile struct {
	ID         string
	UserID     string
	Nickname   string
	AvatarID   int
	TotalStars int
	TotalXP    int
	CreatedAt  time.Time
}

// CreateUser inserts a new user and returns the created row.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, email, passwordHash, role string) (User, error) {
	var u User
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, role, password_hash, created_at, last_login_at`,
		email, passwordHash, role,
	).Scan(&u.ID, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		if isDuplicateKeyError(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("CreateUser: %w", err)
	}
	return u, nil
}

// GetUserByEmail retrieves a user by email address.
func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (User, error) {
	var u User
	err := pool.QueryRow(ctx,
		`SELECT id, email, role, password_hash, created_at, last_login_at
		 FROM users
		 WHERE email = $1 AND deleted_at IS NULL`,
		email,
	).Scan(&u.ID, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("GetUserByEmail: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by ID.
func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id string) (User, error) {
	var u User
	err := pool.QueryRow(ctx,
		`SELECT id, email, role, password_hash, created_at, last_login_at
		 FROM users
		 WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&u.ID, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

// UpdateLastLogin sets the last_login_at field to now.
func UpdateLastLogin(ctx context.Context, pool *pgxpool.Pool, id string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET last_login_at = now() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("UpdateLastLogin: %w", err)
	}
	return nil
}

// GetStudentProfileByUserID retrieves the student profile for a user.
func GetStudentProfileByUserID(ctx context.Context, pool *pgxpool.Pool, userID string) (StudentProfile, error) {
	var p StudentProfile
	err := pool.QueryRow(ctx,
		`SELECT id, user_id, nickname, avatar_id, total_stars, total_xp, created_at
		 FROM student_profiles
		 WHERE user_id = $1`,
		userID,
	).Scan(&p.ID, &p.UserID, &p.Nickname, &p.AvatarID, &p.TotalStars, &p.TotalXP, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StudentProfile{}, ErrNotFound
		}
		return StudentProfile{}, fmt.Errorf("GetStudentProfileByUserID: %w", err)
	}
	return p, nil
}

// isDuplicateKeyError checks if a pgx error is a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	return err != nil && containsString(err.Error(), "duplicate key")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
