package validator

import (
	"math/rand"
)

// SequenceElement represents a shape-colour pair in the sequence.
type SequenceElement struct {
	Shape  string
	Colour string
}

// MemoryCoveAction represents a single action from the client.
type MemoryCoveAction struct {
	Type            string
	ButtonID        string
	ElementIndex    int
	ClientTimestamp int64
}

// MemoryCoveSession represents a session for Memory Cove.
type MemoryCoveSession struct {
	Seed   int64
	Rounds []RoundResult
}

// RoundResult represents the result of a single round.
type RoundResult struct {
	CorrectCount int
	TotalActions int
	Accuracy     float64
}

// MemoryCoveValidationResult holds the outcome of action validation.
type MemoryCoveValidationResult struct {
	Correct      []bool
	ReactionTime []int64
	TotalActions int
	CorrectCount int
}

// ScoredResult holds the scoring outcome.
type ScoredResult struct {
	Score    int
	Stars    int
	XPEarned int
	Accuracy float64
}

var shapes = []string{"circle", "square", "triangle", "star"}
var colours = []string{"red", "blue", "green", "yellow"}

// GenerateSequence returns a deterministic sequence for a given seed and length.
func GenerateSequence(seed int64, length int) []SequenceElement {
	r := rand.New(rand.NewSource(seed))
	seq := make([]SequenceElement, length)
	for i := 0; i < length; i++ {
		seq[i] = SequenceElement{
			Shape:  shapes[r.Intn(len(shapes))],
			Colour: colours[r.Intn(len(colours))],
		}
	}
	return seq
}

// GetSequenceLength determines the sequence length for a round.
func GetSequenceLength(roundNumber int, recentAccuracy float64) int {
	length := 3
	if roundNumber >= 11 {
		length = 8
	} else if roundNumber >= 9 {
		length = 7
	} else if roundNumber >= 7 {
		length = 6
	} else if roundNumber >= 5 {
		length = 5
	} else if roundNumber >= 3 {
		length = 4
	}
	return length
}

// ValidateActions checks client actions against the expected sequence.
func ValidateActions(actions []MemoryCoveAction, seed int64, roundsCompleted int) MemoryCoveValidationResult {
	correct := make([]bool, len(actions))
	reaction := make([]int64, len(actions))
	correctCount := 0

	seqLen := len(actions)
	for _, a := range actions {
		if a.ElementIndex >= seqLen {
			seqLen = a.ElementIndex + 1
		}
	}
	seq := GenerateSequence(seed, seqLen)

	for i, action := range actions {
		expectedID := seq[action.ElementIndex].Shape + "-" + seq[action.ElementIndex].Colour
		correct[i] = action.ButtonID == expectedID
		if correct[i] {
			correctCount++
		}
		if i > 0 {
			reaction[i] = action.ClientTimestamp - actions[i-1].ClientTimestamp
		} else {
			reaction[i] = 0
		}
	}
	return MemoryCoveValidationResult{
		Correct:      correct,
		ReactionTime: reaction,
		TotalActions: len(actions),
		CorrectCount: correctCount,
	}
}

// CalculateScore computes score, stars, and XP earned.
func CalculateScore(result MemoryCoveValidationResult, roundsCompleted int) ScoredResult {
	accuracy := 0.0
	if result.TotalActions > 0 {
		accuracy = float64(result.CorrectCount) / float64(result.TotalActions)
	}
	score := int(accuracy * 1000)
	stars := 0
	if accuracy >= 0.90 {
		stars = 3
	} else if accuracy >= 0.70 {
		stars = 2
	} else if accuracy >= 0.50 {
		stars = 1
	}
	xp := stars*10 + roundsCompleted*5
	return ScoredResult{
		Score:    score,
		Stars:    stars,
		XPEarned: xp,
		Accuracy: accuracy,
	}
}
