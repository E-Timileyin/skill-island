package validator

import (
	"testing"
)

func TestGenerateSpawnManifestDeterministic(t *testing.T) {
	seed := int64(12345)
	m1 := GenerateSpawnManifest(seed, SESSION_DURATION, 2)
	m2 := GenerateSpawnManifest(seed, SESSION_DURATION, 2)
	if len(m1) != len(m2) {
		t.Fatalf("Manifest lengths differ: %d vs %d", len(m1), len(m2))
	}
	for i := range m1 {
		if m1[i] != m2[i] {
			t.Errorf("Manifest not deterministic at %d: %v vs %v", i, m1[i], m2[i])
		}
	}
}

func TestGetTargetPositionAtTimeLerp(t *testing.T) {
	spawn := SpawnEvent{TargetID: "b1", TargetType: "butterfly", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5}
	x, y := GetTargetPositionAtTime(spawn, 400, 80)
	if x <= 0.0 || x > 1.0 {
		t.Errorf("X out of bounds: %f", x)
	}
	if y != 0.5 {
		t.Errorf("Y should not change: %f", y)
	}
}

func TestValidateTapsCorrectness(t *testing.T) {
	manifest := []SpawnEvent{
		{TargetID: "b1", TargetType: "butterfly", SpawnTimeMs: 0, PositionX: 0.0, PositionY: 0.5},
		{TargetID: "bee1", TargetType: "bee", SpawnTimeMs: 1000, PositionX: 0.0, PositionY: 0.6},
	}
	actions := []FocusForestAction{
		{Type: "tap", TapX: 0.05, TapY: 0.5, ClientTimestamp: 200},  // butterfly hit
		{Type: "tap", TapX: 0.05, TapY: 0.6, ClientTimestamp: 1200}, // bee hit
		{Type: "tap", TapX: 0.9, TapY: 0.9, ClientTimestamp: 500},   // miss
	}
	res := ValidateTaps(actions, manifest, 1)
	if res.ButterflyHits != 1 {
		t.Errorf("Expected 1 butterfly hit, got %d", res.ButterflyHits)
	}
	if res.BeeHits != 1 {
		t.Errorf("Expected 1 bee hit, got %d", res.BeeHits)
	}
	if res.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", res.Misses)
	}
}

func TestCalculateAttentionScoreThresholds(t *testing.T) {
	cases := []struct {
		butterflyHits    int
		beeHits          int
		totalButterflies int
		stars            int
	}{
		{17, 0, 20, 3}, // 0.85
		{13, 2, 20, 2}, // 0.65
		{8, 4, 20, 1},  // 0.40
		{5, 10, 20, 0}, // 0.0
	}
	for _, c := range cases {
		vr := FocusForestValidationResult{
			ButterflyHits:    c.butterflyHits,
			BeeHits:          c.beeHits,
			TotalButterflies: c.totalButterflies,
		}
		res := CalculateAttentionScore(vr)
		if res.Stars != c.stars {
			t.Errorf("Expected %d stars, got %d (butterfly=%d, bee=%d, total=%d)", c.stars, res.Stars, c.butterflyHits, c.beeHits, c.totalButterflies)
		}
	}
}

func TestBeeHitsDoNotReduceBelowZero(t *testing.T) {
	vr := FocusForestValidationResult{
		ButterflyHits:    0,
		BeeHits:          10,
		TotalButterflies: 10,
	}
	res := CalculateAttentionScore(vr)
	if res.AttentionScore < 0 {
		t.Errorf("Attention score below zero: %f", res.AttentionScore)
	}
}
