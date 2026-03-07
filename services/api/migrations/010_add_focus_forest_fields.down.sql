-- Rollback: Remove Focus Forest fields from pending_sessions
ALTER TABLE pending_sessions DROP COLUMN IF EXISTS difficulty_level;
ALTER TABLE pending_sessions DROP COLUMN IF EXISTS used;
ALTER TABLE pending_sessions DROP COLUMN IF EXISTS session_duration_ms;
