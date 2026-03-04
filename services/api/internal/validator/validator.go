package validator

import (
	"encoding/json"
)

// Action represents a single game action submitted by the client.
// Payload is excluded from JSON marshaling — game-specific validators
// unmarshal each action's raw JSON into type-specific structs directly.
type Action struct {
	Type            string          `json:"type"`
	ClientTimestamp int             `json:"client_timestamp"`
	Payload         json.RawMessage `json:"-"`
}

// BehavioralMetric holds data for a single behavioral metric produced by validation.
type BehavioralMetric struct {
	EventType         string
	ReactionTimeMs    *int
	HesitationMs      *int
	RetryCount        int
	Correct           bool
	TimestampOffsetMs int
	Metadata          json.RawMessage
}

// ValidationResult holds the computed results of validating a game session.
type ValidationResult struct {
	Score        int
	Accuracy     float64
	StarsEarned  int
	XPEarned     int
	Metrics      []BehavioralMetric
	Rejected     bool
	RejectReason string
}
