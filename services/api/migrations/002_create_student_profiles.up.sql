CREATE TABLE student_profiles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  nickname     TEXT NOT NULL,
  avatar_id    INTEGER NOT NULL DEFAULT 0,
  total_stars  INTEGER NOT NULL DEFAULT 0,
  total_xp     INTEGER NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id)
);
