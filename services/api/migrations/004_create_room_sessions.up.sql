CREATE TABLE room_sessions (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_type             game_zone NOT NULL,
  player_1_profile_id   UUID NOT NULL REFERENCES student_profiles(id),
  player_2_profile_id   UUID NOT NULL REFERENCES student_profiles(id),
  group_xp_earned       INTEGER NOT NULL DEFAULT 0,
  completed             BOOLEAN NOT NULL DEFAULT false,
  started_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at              TIMESTAMPTZ,
  disconnect_reason     TEXT
);
