package handlers

import "time"

// ProfileResponse is the public representation of a student profile.
// Never exposes user_id or password_hash.
type ProfileResponse struct {
	ID         string    `json:"id"`
	Nickname   string    `json:"nickname"`
	AvatarID   int       `json:"avatar_id"`
	TotalStars int       `json:"total_stars"`
	TotalXP    int       `json:"total_xp"`
	PlayMode   string    `json:"play_mode"`
	CreatedAt  time.Time `json:"created_at"`
}

// createProfileRequest is the expected JSON body for POST /api/profiles.
type createProfileRequest struct {
	Nickname string `json:"nickname"`
	AvatarID *int   `json:"avatar_id"`
	PlayMode string `json:"play_mode"`
}

// updateProfileRequest is the expected JSON body for PATCH /api/profiles/me.
type updateProfileRequest struct {
	Nickname *string `json:"nickname"`
	AvatarID *int    `json:"avatar_id"`
	PlayMode *string `json:"play_mode"`
}
