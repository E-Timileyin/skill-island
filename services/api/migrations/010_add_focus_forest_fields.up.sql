-- Migration: Add difficulty_level and used fields to pending_sessions for Focus Forest
ALTER TABLE pending_sessions ADD COLUMN IF NOT EXISTS difficulty_level INTEGER NOT NULL DEFAULT 1;
ALTER TABLE pending_sessions ADD COLUMN IF NOT EXISTS used BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE pending_sessions ADD COLUMN IF NOT EXISTS session_duration_ms INTEGER NOT NULL DEFAULT 60000;
