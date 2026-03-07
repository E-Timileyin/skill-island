-- Migration: Create pending_sessions table
CREATE TABLE IF NOT EXISTS pending_sessions (
    id UUID PRIMARY KEY,
    profile_id UUID REFERENCES student_profiles(id) ON DELETE CASCADE,
    game_type TEXT NOT NULL,
    seed BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_pending_sessions_profile_id ON pending_sessions(profile_id);
CREATE INDEX IF NOT EXISTS idx_pending_sessions_expires_at ON pending_sessions(expires_at);
