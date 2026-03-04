CREATE TABLE achievements (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  profile_id   UUID NOT NULL REFERENCES student_profiles(id),
  type         TEXT NOT NULL,
  earned_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_achievements_profile ON achievements(profile_id);
