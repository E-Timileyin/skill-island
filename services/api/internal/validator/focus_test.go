package validator

import (
	"math"
	"testing"
)

// -------------------------------------------------------------------
// TestGenerateSpawnManifest_Deterministic
// Same seed + same params always produces identical manifest.
// -------------------------------------------------------------------
func TestGenerateSpawnManifest_Deterministic(t *testing.T) {
	seed := int64(12345)
	m1 := GenerateSpawnManifest(seed, SESSION_DURATION, 2)
	m2 := GenerateSpawnManifest(seed, SESSION_DURATION, 2)

	if len(m1) != len(m2) {
		t.Fatalf("manifest lengths differ: %d vs %d", len(m1), len(m2))
	}
	for i := range m1 {
		if m1[i].TargetID != m2[i].TargetID {
			t.Errorf("TargetID mismatch at %d: %s vs %s", i, m1[i].TargetID, m2[i].TargetID)
		}
		if m1[i].TargetType != m2[i].TargetType {
			t.Errorf("TargetType mismatch at %d: %s vs %s", i, m1[i].TargetType, m2[i].TargetType)
		}
		if m1[i].SpawnTimeMs != m2[i].SpawnTimeMs {
			t.Errorf("SpawnTimeMs mismatch at %d: %d vs %d", i, m1[i].SpawnTimeMs, m2[i].SpawnTimeMs)
		}
		if m1[i].PositionX != m2[i].PositionX {
			t.Errorf("PositionX mismatch at %d", i)
		}
		if m1[i].PositionY != m2[i].PositionY {
			t.Errorf("PositionY mismatch at %d", i)
		}
	}
}

// -------------------------------------------------------------------
// TestGenerateSpawnManifest_DifficultyRatios
// Level 1 manifest has ~30% bees (±5% tolerance)
// Level 4 manifest has ~55% bees (±5% tolerance)
// -------------------------------------------------------------------
func TestGenerateSpawnManifest_DifficultyRatios(t *testing.T) {
	cases := []struct {
		level       int
		expectedPct float64
		tolerance   float64
	}{
		{1, 0.30, 0.05},
		{4, 0.55, 0.05},
	}

	for _, c := range cases {
		manifest := GenerateSpawnManifest(42, SESSION_DURATION, c.level)
		beeCount := 0
		for _, s := range manifest {
			if s.TargetType == "bee" {
				beeCount++
			}
		}
		actualPct := float64(beeCount) / float64(len(manifest))
		if math.Abs(actualPct-c.expectedPct) > c.tolerance {
			t.Errorf("Level %d: expected ~%.0f%% bees (±%.0f%%), got %.1f%% (%d/%d)",
				c.level, c.expectedPct*100, c.tolerance*100, actualPct*100, beeCount, len(manifest))
		}
	}
}

// -------------------------------------------------------------------
// TestGenerateSpawnManifest_PositionYRange
// All PositionY values must be in [0.1, 0.9].
// -------------------------------------------------------------------
func TestGenerateSpawnManifest_PositionYRange(t *testing.T) {
	manifest := GenerateSpawnManifest(99, SESSION_DURATION, 2)
	for i, s := range manifest {
		if s.PositionY < 0.1 || s.PositionY > 0.9 {
			t.Errorf("spawn %d: PositionY=%.4f out of [0.1, 0.9] range", i, s.PositionY)
		}
	}
}

// -------------------------------------------------------------------
// TestGenerateSpawnManifest_PositionXStartsAtZero
// All targets should spawn at PositionX = 0.0.
// -------------------------------------------------------------------
func TestGenerateSpawnManifest_PositionXStartsAtZero(t *testing.T) {
	manifest := GenerateSpawnManifest(77, SESSION_DURATION, 1)
	for i, s := range manifest {
		if s.PositionX != 0.0 {
			t.Errorf("spawn %d: PositionX=%.4f, expected 0.0", i, s.PositionX)
		}
	}
}

// -------------------------------------------------------------------
// TestGetTargetPositionAtTime_Interpolation
// At SpawnTimeMs: x = spawn.PositionX
// At SpawnTimeMs + 1000: x has moved by (speed/screenWidth)
// -------------------------------------------------------------------
func TestGetTargetPositionAtTime_Interpolation(t *testing.T) {
	spawn := SpawnEvent{
		TargetID:    "test-1",
		TargetType:  "butterfly_blue",
		SpawnTimeMs: 0,
		PositionX:   0.0,
		PositionY:   0.5,
	}

	// At spawn time, position should be at starting x.
	x0, y0 := GetTargetPositionAtTime(spawn, 0, 80)
	if x0 != spawn.PositionX {
		t.Errorf("at spawn time: expected x=%.4f, got x=%.4f", spawn.PositionX, x0)
	}
	if y0 != spawn.PositionY {
		t.Errorf("at spawn time: expected y=%.4f, got y=%.4f", spawn.PositionY, y0)
	}

	// At +1000ms with speed 80px/s: x should move by 80/800 = 0.1.
	x1, y1 := GetTargetPositionAtTime(spawn, 1000, 80)
	expectedX := 0.0 + 80.0/screenWidth*1.0
	if math.Abs(x1-expectedX) > 0.001 {
		t.Errorf("at +1000ms: expected x=%.4f, got x=%.4f", expectedX, x1)
	}
	if y1 != spawn.PositionY {
		t.Errorf("y should not change: expected %.4f, got %.4f", spawn.PositionY, y1)
	}
}

// -------------------------------------------------------------------
// TestGetTargetPositionAtTime_ClampAt1_2
// X should clamp at 1.2.
// -------------------------------------------------------------------
func TestGetTargetPositionAtTime_ClampAt1_2(t *testing.T) {
	spawn := SpawnEvent{
		TargetID:    "test-2",
		TargetType:  "butterfly_blue",
		SpawnTimeMs: 0,
		PositionX:   0.0,
		PositionY:   0.5,
	}

	// At a very late time, x should cap at 1.2.
	x, _ := GetTargetPositionAtTime(spawn, 30000, 160)
	if x > 1.2 {
		t.Errorf("x should be clamped at 1.2, got %.4f", x)
	}
	if x != 1.2 {
		t.Errorf("expected x=1.2 when overshot, got x=%.4f", x)
	}
}

// -------------------------------------------------------------------
// TestValidateTaps_ButterflyHit
// Tap within HIT_RADIUS of active butterfly → correct = true.
// -------------------------------------------------------------------
func TestValidateTaps_ButterflyHit(t *testing.T) {
	manifest := []SpawnEvent{
		{TargetID: "b1", TargetType: "butterfly_blue", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5},
	}
	// At 200ms with speed 80: x = 0 + (80/800)*(0.2) = 0.02
	actions := []FocusForestAction{
		{Type: "tap", TapX: 0.02, TapY: 0.5, ClientTimestamp: 200},
	}
	res := ValidateTaps(actions, manifest, 1)
	if res.ButterflyHits != 1 {
		t.Errorf("expected 1 butterfly hit, got %d", res.ButterflyHits)
	}
	if !res.Correct[0] {
		t.Error("expected correct[0] = true for butterfly hit")
	}
	if res.ReactionTimes[0] != 200 {
		t.Errorf("expected reaction time 200, got %d", res.ReactionTimes[0])
	}
}

// -------------------------------------------------------------------
// TestValidateTaps_BeeHit
// Tap within HIT_RADIUS of active bee → correct = false (distraction, not miss).
// -------------------------------------------------------------------
func TestValidateTaps_BeeHit(t *testing.T) {
	manifest := []SpawnEvent{
		{TargetID: "bee1", TargetType: "bee", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5},
	}
	actions := []FocusForestAction{
		{Type: "tap", TapX: 0.02, TapY: 0.5, ClientTimestamp: 200},
	}
	res := ValidateTaps(actions, manifest, 1)
	if res.BeeHits != 1 {
		t.Errorf("expected 1 bee hit, got %d", res.BeeHits)
	}
	if res.Correct[0] {
		t.Error("expected correct[0] = false for bee hit")
	}
	if res.Misses != 0 {
		t.Errorf("expected 0 misses, got %d (bee hit should not be a miss)", res.Misses)
	}
}

// -------------------------------------------------------------------
// TestValidateTaps_Miss
// Tap with no target within HIT_RADIUS → recorded as missed tap.
// -------------------------------------------------------------------
func TestValidateTaps_Miss(t *testing.T) {
	manifest := []SpawnEvent{
		{TargetID: "b1", TargetType: "butterfly_blue", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5},
	}
	// Tap very far from any target.
	actions := []FocusForestAction{
		{Type: "tap", TapX: 0.9, TapY: 0.9, ClientTimestamp: 500},
	}
	res := ValidateTaps(actions, manifest, 1)
	if res.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", res.Misses)
	}
	if res.ButterflyHits != 0 {
		t.Errorf("expected 0 butterfly hits, got %d", res.ButterflyHits)
	}
	if res.Correct[0] {
		t.Error("expected correct[0] = false for miss")
	}
}

// -------------------------------------------------------------------
// TestValidateTaps_ExpiredTarget
// Tap on target after DESPAWN_AFTER_MS → treated as missed tap (target gone).
// -------------------------------------------------------------------
func TestValidateTaps_ExpiredTarget(t *testing.T) {
	manifest := []SpawnEvent{
		{TargetID: "b1", TargetType: "butterfly_blue", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5},
	}
	// Tap at 3500ms — target despawned at 3000ms.
	actions := []FocusForestAction{
		{Type: "tap", TapX: 0.01, TapY: 0.5, ClientTimestamp: 3500},
	}
	res := ValidateTaps(actions, manifest, 1)
	if res.ButterflyHits != 0 {
		t.Errorf("expected 0 butterfly hits for expired target, got %d", res.ButterflyHits)
	}
	if res.Misses != 1 {
		t.Errorf("expected 1 miss for expired target, got %d", res.Misses)
	}
}

// -------------------------------------------------------------------
// TestValidateTaps_DoubleTapSameTarget
// Tapping the same target twice should not count twice.
// -------------------------------------------------------------------
func TestValidateTaps_DoubleTapSameTarget(t *testing.T) {
	manifest := []SpawnEvent{
		{TargetID: "b1", TargetType: "butterfly_blue", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5},
	}
	actions := []FocusForestAction{
		{Type: "tap", TapX: 0.02, TapY: 0.5, ClientTimestamp: 200},
		{Type: "tap", TapX: 0.04, TapY: 0.5, ClientTimestamp: 400},
	}
	res := ValidateTaps(actions, manifest, 1)
	if res.ButterflyHits != 1 {
		t.Errorf("expected 1 butterfly hit (no double count), got %d", res.ButterflyHits)
	}
	if res.Misses != 1 {
		t.Errorf("expected 1 miss (second tap on same target), got %d", res.Misses)
	}
}

// -------------------------------------------------------------------
// TestCalculateAttentionScore_StarThresholds
// -------------------------------------------------------------------
func TestCalculateAttentionScore_StarThresholds(t *testing.T) {
	cases := []struct {
		name             string
		butterflyHits    int
		beeHits          int
		totalButterflies int
		expectedStars    int
	}{
		// attention_score = (17 - 0*0.5)/20 = 0.85 → 3 stars
		{"3 stars at 0.90", 18, 0, 20, 3},
		// attention_score = (14 - 0*0.5)/20 = 0.70 → 2 stars
		{"2 stars at 0.70", 14, 0, 20, 2},
		// attention_score = (9 - 0*0.5)/20 = 0.45 → 1 star
		{"1 star at 0.45", 9, 0, 20, 1},
		// attention_score = (4 - 0*0.5)/20 = 0.20 → 0 stars
		{"0 stars at 0.20", 4, 0, 20, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vr := FocusForestValidationResult{
				ButterflyHits:    c.butterflyHits,
				BeeHits:          c.beeHits,
				TotalButterflies: c.totalButterflies,
			}
			res := CalculateAttentionScore(vr)
			if res.Stars != c.expectedStars {
				t.Errorf("expected %d stars, got %d (attention=%.4f)",
					c.expectedStars, res.Stars, res.AttentionScore)
			}
		})
	}
}

// -------------------------------------------------------------------
// TestCalculateAttentionScore_BeeHitsNeverGoNegative
// All bees hit, no butterflies → attention_score clamped to 0.0.
// -------------------------------------------------------------------
func TestCalculateAttentionScore_BeeHitsNeverGoNegative(t *testing.T) {
	vr := FocusForestValidationResult{
		ButterflyHits:    0,
		BeeHits:          10,
		TotalButterflies: 10,
	}
	res := CalculateAttentionScore(vr)
	if res.AttentionScore < 0 {
		t.Errorf("attention score should never be negative, got %f", res.AttentionScore)
	}
	if res.AttentionScore != 0 {
		t.Errorf("expected attention_score = 0.0 when all bees, got %.4f", res.AttentionScore)
	}
	if res.Stars != 0 {
		t.Errorf("expected 0 stars, got %d", res.Stars)
	}
}

// -------------------------------------------------------------------
// TestCalculateAttentionScore_XPCalculation
// 3 stars + attention_score 0.90 → xp = 30 + floor(0.90*50) = 30 + 45 = 75
// 0 stars + attention_score 0.20 → xp = 0 + floor(0.20*50) = 10
// -------------------------------------------------------------------
func TestCalculateAttentionScore_XPCalculation(t *testing.T) {
	cases := []struct {
		name             string
		butterflyHits    int
		beeHits          int
		totalButterflies int
		expectedXP       int
	}{
		{
			"3 stars high attention",
			// attention = 18/20 = 0.90 → 3 stars → xp = 30 + 45 = 75
			18, 0, 20, 75,
		},
		{
			"0 stars low attention",
			// attention = 4/20 = 0.20 → 0 stars → xp = 0 + 10 = 10
			4, 0, 20, 10,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vr := FocusForestValidationResult{
				ButterflyHits:    c.butterflyHits,
				BeeHits:          c.beeHits,
				TotalButterflies: c.totalButterflies,
			}
			res := CalculateAttentionScore(vr)
			if res.XPEarned != c.expectedXP {
				t.Errorf("expected XP=%d, got XP=%d (attention=%.4f, stars=%d)",
					c.expectedXP, res.XPEarned, res.AttentionScore, res.Stars)
			}
		})
	}
}

// -------------------------------------------------------------------
// TestCalculateAttentionScore_ZeroButterflies
// Edge case: no butterflies in manifest → 0 all.
// -------------------------------------------------------------------
func TestCalculateAttentionScore_ZeroButterflies(t *testing.T) {
	vr := FocusForestValidationResult{
		ButterflyHits:    0,
		BeeHits:          5,
		TotalButterflies: 0,
	}
	res := CalculateAttentionScore(vr)
	if res.AttentionScore != 0 {
		t.Errorf("expected 0 attention with no butterflies, got %.4f", res.AttentionScore)
	}
	if res.Stars != 0 {
		t.Errorf("expected 0 stars, got %d", res.Stars)
	}
	if res.XPEarned != 0 {
		t.Errorf("expected 0 XP, got %d", res.XPEarned)
	}
}
