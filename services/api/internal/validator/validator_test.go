package validator

import (
	"testing"
)

func TestGenerateSequenceDeterministic(t *testing.T) {
	seed := int64(12345)
	seq1 := GenerateSequence(seed, 5)
	seq2 := GenerateSequence(seed, 5)
	for i := range seq1 {
		if seq1[i] != seq2[i] {
			t.Errorf("Sequence not deterministic at index %d: %v vs %v", i, seq1[i], seq2[i])
		}
	}
}

func TestValidateActionsCorrectness(t *testing.T) {
	seed := int64(42)
	seq := GenerateSequence(seed, 3)
	actions := []MemoryCoveAction{
		{Type: "press", ButtonID: seq[0].Shape + "-" + seq[0].Colour, ElementIndex: 0, ClientTimestamp: 1000},
		{Type: "press", ButtonID: seq[1].Shape + "-" + seq[1].Colour, ElementIndex: 1, ClientTimestamp: 2000},
		{Type: "press", ButtonID: "wrong-shape-wrong-colour", ElementIndex: 2, ClientTimestamp: 3000},
	}
	result := ValidateActions(actions, seed, 1)
	if result.CorrectCount != 2 {
		t.Errorf("Expected 2 correct, got %d", result.CorrectCount)
	}
	if result.TotalActions != 3 {
		t.Errorf("Expected 3 actions, got %d", result.TotalActions)
	}
}

func TestCalculateScoreThresholds(t *testing.T) {
	cases := []struct {
		correct int
		total   int
		stars   int
	}{
		{9, 10, 3}, // 0.90
		{7, 10, 2}, // 0.70
		{5, 10, 1}, // 0.50
		{4, 10, 0}, // 0.40
	}
	for _, c := range cases {
		vr := MemoryCoveValidationResult{CorrectCount: c.correct, TotalActions: c.total}
		res := CalculateScore(vr, 5)
		if res.Stars != c.stars {
			t.Errorf("Expected %d stars, got %d (correct=%d, total=%d)", c.stars, res.Stars, c.correct, c.total)
		}
	}
}

func TestGetSequenceLengthScaling(t *testing.T) {
	cases := []struct {
		round int
		acc   float64
		len   int
	}{
		{1, 0.85, 3},
		{3, 0.85, 4},
		{5, 0.85, 5},
		{7, 0.85, 6},
		{9, 0.85, 7},
		{11, 0.85, 8},
	}
	for _, c := range cases {
		l := GetSequenceLength(c.round, c.acc)
		if l != c.len {
			t.Errorf("Round %d, accuracy %.2f: expected length %d, got %d", c.round, c.acc, c.len, l)
		}
	}
}
