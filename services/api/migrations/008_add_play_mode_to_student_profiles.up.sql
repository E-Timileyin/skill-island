ALTER TABLE student_profiles
  ADD COLUMN play_mode TEXT NOT NULL DEFAULT 'solo'
  CHECK (play_mode IN ('solo', 'team'));
