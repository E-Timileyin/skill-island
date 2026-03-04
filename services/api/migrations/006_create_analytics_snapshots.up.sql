CREATE TABLE analytics_snapshots (
  profile_id              UUID NOT NULL REFERENCES student_profiles(id),
  snapshot_date           DATE NOT NULL,
  attention_score         FLOAT,
  memory_score            FLOAT,
  engagement_frequency    INTEGER,
  coop_participation_rate FLOAT,
  avg_reaction_time_ms    FLOAT,
  avg_hesitation_ms       FLOAT,
  retry_rate              FLOAT,
  PRIMARY KEY (profile_id, snapshot_date)
);
