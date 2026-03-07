package validator

import (
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// Focus Forest constants.
const (
	HIT_RADIUS       = 0.08 // normalised units — server-defined, client cannot change
	DESPAWN_AFTER_MS = 3000 // target removed after 3 seconds if not tapped
	SESSION_DURATION = 60000
	screenWidth      = 800.0 // reference width for normalised speed calculation
)

// SpawnEvent represents a target spawn in Focus Forest.
type SpawnEvent struct {
	TargetID    string  `json:"target_id"`
	TargetType  string  `json:"target_type"`    // "butterfly_blue"|"butterfly_orange"|"butterfly_red"|"bee"
	SpawnTimeMs int     `json:"spawn_time_ms"`  // ms from session start
	PositionX   float64 `json:"position_x"`     // 0.0–1.0 normalised
	PositionY   float64 `json:"position_y"`     // 0.0–1.0 normalised
}

// FocusForestAction represents a tap action from the client.
type FocusForestAction struct {
	Type            string  `json:"type"`             // always "tap"
	TapX            float64 `json:"tap_x"`            // 0.0–1.0 normalised
	TapY            float64 `json:"tap_y"`            // 0.0–1.0 normalised
	ClientTimestamp int64   `json:"client_timestamp"` // ms since session start
}

// difficultyConfig holds parameters for a given difficulty level.
type difficultyConfig struct {
	BeePct   int
	Speed    float64 // px/s
	Interval int     // ms between spawns
}

var difficultyLevels = map[int]difficultyConfig{
	1: {BeePct: 30, Speed: 80, Interval: 1200},
	2: {BeePct: 40, Speed: 100, Interval: 1000},
	3: {BeePct: 50, Speed: 130, Interval: 800},
	4: {BeePct: 55, Speed: 160, Interval: 700},
}

// butterflyTypes are the possible butterfly sub-types.
var butterflyTypes = []string{"butterfly_blue", "butterfly_orange", "butterfly_red"}

// getDifficultyConfig returns the config for a difficulty level (defaults to level 1).
func getDifficultyConfig(level int) difficultyConfig {
	cfg, ok := difficultyLevels[level]
	if !ok {
		return difficultyLevels[1]
	}
	return cfg
}

// GenerateSpawnManifest returns a deterministic list of targets for a session.
// Uses seeded random source — same seed always produces identical manifest.
func GenerateSpawnManifest(seed int64, durationMs int, difficultyLevel int) []SpawnEvent {
	cfg := getDifficultyConfig(difficultyLevel)
	r := rand.New(rand.NewSource(seed))

	manifest := []SpawnEvent{}
	for t := 0; t < durationMs; t += cfg.Interval {
		isBee := r.Intn(100) < cfg.BeePct

		var targetType string
		if isBee {
			targetType = "bee"
		} else {
			// Pick a random butterfly sub-type.
			targetType = butterflyTypes[r.Intn(len(butterflyTypes))]
		}

		// PositionX: spawn off left edge (0.0); target moves right to 1.0.
		posX := 0.0

		// PositionY: random value in [0.1, 0.9] to avoid edges.
		posY := 0.1 + r.Float64()*0.8

		manifest = append(manifest, SpawnEvent{
			TargetID:    uuid.NewSHA1(uuid.NameSpaceDNS, []byte(seed2bytes(seed, t))).String(),
			TargetType:  targetType,
			SpawnTimeMs: t,
			PositionX:   posX,
			PositionY:   posY,
		})
	}
	return manifest
}

// seed2bytes creates a deterministic byte slice from seed and time for UUID generation.
func seed2bytes(seed int64, t int) string {
	return string(rune(seed)) + "_" + string(rune(t))
}

// GetTargetPositionAtTime returns the interpolated position of a target.
// x = spawn.PositionX + (speedPxPerS / screenWidth) * (timestampMs - spawn.SpawnTimeMs) / 1000
// y = spawn.PositionY (constant — targets move horizontally only)
// Clamp x to [0.0, 1.2] — allow slightly off screen before despawn.
func GetTargetPositionAtTime(spawn SpawnEvent, timestampMs int, speedPxPerS float64) (float64, float64) {
	deltaSeconds := float64(timestampMs-spawn.SpawnTimeMs) / 1000.0
	x := spawn.PositionX + (speedPxPerS/screenWidth)*deltaSeconds
	if x > 1.2 {
		x = 1.2
	}
	if x < 0.0 {
		x = 0.0
	}
	return x, spawn.PositionY
}

// FocusForestValidationResult holds per-tap validation data.
type FocusForestValidationResult struct {
	ButterflyHits    int
	BeeHits          int
	Misses           int
	TotalButterflies int
	ReactionTimes    []int64
	Correct          []bool
	TotalActions     int
}

// isButterfly returns true if the target type is any butterfly variant.
func isButterfly(targetType string) bool {
	return targetType == "butterfly_blue" ||
		targetType == "butterfly_orange" ||
		targetType == "butterfly_red"
}

// ValidateTaps checks tap actions against the spawn manifest.
func ValidateTaps(actions []FocusForestAction, manifest []SpawnEvent, difficultyLevel int) FocusForestValidationResult {
	cfg := getDifficultyConfig(difficultyLevel)
	speed := cfg.Speed

	butterflyHits, beeHits, misses := 0, 0, 0
	totalButterflies := 0
	reactionTimes := make([]int64, len(actions))
	correct := make([]bool, len(actions))

	// Track which targets have already been hit (prevent double-scoring).
	hitTargets := make(map[int]bool)

	for _, s := range manifest {
		if isButterfly(s.TargetType) {
			totalButterflies++
		}
	}

	for i, action := range actions {
		minDist := math.MaxFloat64
		closest := -1

		for j, spawn := range manifest {
			// Skip already-hit targets.
			if hitTargets[j] {
				continue
			}

			// Target must have spawned before/at tap time.
			if action.ClientTimestamp < int64(spawn.SpawnTimeMs) {
				continue
			}

			// Target must not have despawned (expired after DESPAWN_AFTER_MS).
			if action.ClientTimestamp > int64(spawn.SpawnTimeMs)+DESPAWN_AFTER_MS {
				continue
			}

			x, y := GetTargetPositionAtTime(spawn, int(action.ClientTimestamp), speed)
			dist := math.Sqrt(math.Pow(action.TapX-x, 2) + math.Pow(action.TapY-y, 2))
			if dist < minDist {
				minDist = dist
				closest = j
			}
		}

		if closest >= 0 && minDist <= HIT_RADIUS {
			target := manifest[closest]
			hitTargets[closest] = true
			reactionTimes[i] = action.ClientTimestamp - int64(target.SpawnTimeMs)

			if isButterfly(target.TargetType) {
				butterflyHits++
				correct[i] = true
			} else {
				// Bee hit — distraction event, correct=false but NO score penalty.
				beeHits++
				correct[i] = false
			}
		} else {
			// Nothing within radius — missed tap.
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

// FocusForestScoredResult holds the final computed scores.
type FocusForestScoredResult struct {
	AttentionScore float64
	Stars          int
	XPEarned       int
}

// CalculateAttentionScore computes the attention score, stars, and XP.
//
//	attention_score = (butterfly_hits - (bee_hits * 0.5)) / butterflies_total
//	Clamped to [0.0, 1.0]
//	3 stars → attention_score >= 0.85
//	2 stars → attention_score >= 0.65
//	1 star  → attention_score >= 0.40
//	0 stars → attention_score <  0.40
//	xp_earned = (stars * 10) + floor(attention_score * 50)
func CalculateAttentionScore(result FocusForestValidationResult) FocusForestScoredResult {
	if result.TotalButterflies == 0 {
		return FocusForestScoredResult{AttentionScore: 0, Stars: 0, XPEarned: 0}
	}

	attn := (float64(result.ButterflyHits) - float64(result.BeeHits)*0.5) / float64(result.TotalButterflies)

	// Clamp to [0.0, 1.0].
	if attn < 0 {
		attn = 0
	}
	if attn > 1 {
		attn = 1
	}

	stars := 0
	if attn >= 0.85 {
		stars = 3
	} else if attn >= 0.65 {
		stars = 2
	} else if attn >= 0.40 {
		stars = 1
	}

	xp := stars*10 + int(math.Floor(attn*50))

	return FocusForestScoredResult{
		AttentionScore: attn,
		Stars:          stars,
		XPEarned:       xp,
	}
}
