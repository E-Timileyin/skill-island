package validator

import (
	"math"
	"math/rand"
	"strconv"
)

const HIT_RADIUS = 0.08
const SESSION_DURATION = 60000 // ms

// SpawnEvent represents a target spawn in Focus Forest.
type SpawnEvent struct {
	TargetID    string
	TargetType  string // "butterfly" or "bee"
	SpawnTimeMs int
	PositionX   float64
	PositionY   float64
}

// FocusForestAction represents a tap action from the client.
type FocusForestAction struct {
	Type            string
	TapX            float64
	TapY            float64
	ClientTimestamp int64
}

// GenerateSpawnManifest returns a deterministic list of targets for a session.
func GenerateSpawnManifest(seed int64, durationMs int, difficultyLevel int) []SpawnEvent {
	var beePct, interval int
	switch difficultyLevel {
	case 1:
		beePct, interval = 30, 1200
	case 2:
		beePct, interval = 40, 1000
	case 3:
		beePct, interval = 50, 800
	case 4:
		beePct, interval = 55, 700
	default:
		beePct, interval = 30, 1200
	}
	r := rand.New(rand.NewSource(seed))
	manifest := []SpawnEvent{}
	for t := 0; t < durationMs; t += interval {
		isBee := r.Intn(100) < beePct
		typeStr := "bee"
		if !isBee {
			typeStr = "butterfly"
		}
		id := typeStr + "_" + strconv.Itoa(t)
		posY := r.Float64()
		manifest = append(manifest, SpawnEvent{
			TargetID:    id,
			TargetType:  typeStr,
			SpawnTimeMs: t,
			PositionX:   0.0,
			PositionY:   posY,
		})
	}
	return manifest
}

// GetTargetPositionAtTime returns the interpolated position of a target.
func GetTargetPositionAtTime(spawn SpawnEvent, timestampMs int, speedPxPerS float64) (float64, float64) {
	delta := float64(timestampMs-spawn.SpawnTimeMs) / 1000.0
	x := spawn.PositionX + speedPxPerS*delta/800.0 // normalised width
	if x > 1.0 {
		x = 1.0
	}
	return x, spawn.PositionY
}

// ValidationResult for Focus Forest.
type FocusForestValidationResult struct {
	ButterflyHits    int
	BeeHits          int
	Misses           int
	TotalButterflies int
	ReactionTimes    []int64
	Correct          []bool
	TotalActions     int
}

// ValidateTaps checks tap actions against the spawn manifest.
func ValidateTaps(actions []FocusForestAction, manifest []SpawnEvent, difficultyLevel int) FocusForestValidationResult {
	var speed int
	switch difficultyLevel {
	case 1:
		speed = 80
	case 2:
		speed = 100
	case 3:
		speed = 130
	case 4:
		speed = 160
	default:
		speed = 80
	}

	butterflyHits, beeHits, misses := 0, 0, 0
	totalButterflies := 0
	reactionTimes := make([]int64, len(actions))
	correct := make([]bool, len(actions))
	for _, s := range manifest {
		if s.TargetType == "butterfly" {
			totalButterflies++
		}
	}
	for i, action := range actions {
		minDist := math.MaxFloat64
		closest := -1
		for j, spawn := range manifest {
			if int64(action.ClientTimestamp) < int64(spawn.SpawnTimeMs) {
				continue
			}
			x, y := GetTargetPositionAtTime(spawn, int(action.ClientTimestamp), float64(speed))
			dist := math.Sqrt(math.Pow(action.TapX-x, 2) + math.Pow(action.TapY-y, 2))
			if dist < minDist {
				minDist = dist
				closest = j
			}
		}
		if closest >= 0 && minDist <= HIT_RADIUS {
			target := manifest[closest]
			if target.TargetType == "butterfly" {
				butterflyHits++
				correct[i] = true
				reactionTimes[i] = int64(action.ClientTimestamp) - int64(target.SpawnTimeMs)
			} else {
				beeHits++
				correct[i] = false
				reactionTimes[i] = int64(action.ClientTimestamp) - int64(target.SpawnTimeMs)
			}
		} else {
			misses++
			correct[i] = false
			reactionTimes[i] = 0
		}
	}
	return FocusForestValidationResult{
		ButterflyHits:    butterflyHits,
		BeeHits:          beeHits,
		Misses:           misses,
		TotalButterflies: totalButterflies,
		ReactionTimes:    reactionTimes,
		Correct:          correct,
		TotalActions:     len(actions),
	}
}

// ScoredResult for Focus Forest.
type FocusForestScoredResult struct {
	AttentionScore float64
	Stars          int
	XPEarned       int
}

// CalculateAttentionScore computes the attention score and stars.
func CalculateAttentionScore(result FocusForestValidationResult) FocusForestScoredResult {
	if result.TotalButterflies == 0 {
		return FocusForestScoredResult{AttentionScore: 0, Stars: 0, XPEarned: 0}
	}
	attn := (float64(result.ButterflyHits) - float64(result.BeeHits)*0.5) / float64(result.TotalButterflies)
	if attn < 0 {
		attn = 0
	}
	if attn > 1 {
		attn = 1
	}
	stars := 0
	if attn >= 0.85 {
		stars = 3
	} else if attn >= 0.60 {
		stars = 2
	} else if attn >= 0.30 {
		stars = 1
	}
	xp := stars*10 + int(math.Floor(attn*50))
	return FocusForestScoredResult{
		AttentionScore: attn,
		Stars:          stars,
		XPEarned:       xp,
	}
}
