CREATE TYPE game_zone AS ENUM (
  'memory_cove', 'focus_forest', 'team_tower', 'pattern_plateau'
);

CREATE TYPE play_mode AS ENUM ('solo', 'cooperative');

CREATE TABLE game_sessions (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profile_id       UUID NOT NULL REFERENCES student_profiles(id),
  game_type        game_zone NOT NULL,
  mode             play_mode NOT NULL DEFAULT 'solo',
  score            INTEGER NOT NULL DEFAULT 0,
  duration_seconds INTEGER NOT NULL,
  accuracy         FLOAT NOT NULL DEFAULT 0.0,
  stars_earned     INTEGER NOT NULL DEFAULT 0 CHECK (stars_earned BETWEEN 0 AND 3),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_game_sessions_profile ON game_sessions(profile_id);
CREATE INDEX idx_game_sessions_created ON game_sessions(created_at);
