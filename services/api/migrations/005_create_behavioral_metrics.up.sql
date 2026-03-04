CREATE TYPE behavior_event AS ENUM ('action', 'hesitation', 'retry', 'distraction');

CREATE TABLE behavioral_metrics (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id          UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
  event_type          behavior_event NOT NULL,
  reaction_time_ms    INTEGER,
  hesitation_ms       INTEGER,
  retry_count         INTEGER NOT NULL DEFAULT 0,
  correct             BOOLEAN NOT NULL,
  timestamp_offset_ms INTEGER NOT NULL,
  metadata            JSONB
);

CREATE INDEX idx_behavioral_session ON behavioral_metrics(session_id);
