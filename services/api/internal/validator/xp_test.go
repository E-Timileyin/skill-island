package validator

import (
	"testing"
)

func TestCalculateXP(t *testing.T) {
	tests := []struct {
		stars    int
		expected int
	}{
		{0, 0},
		{1, 10},
		{2, 20},
		{3, 35},
		{-1, 0},
		{4, 0},
	}

	for _, tc := range tests {
		got := CalculateXP("focus_forest", tc.stars, 0)
		if got != tc.expected {
			t.Errorf("CalculateXP(%d) = %d, want %d", tc.stars, got, tc.expected)
		}
	}
}

func TestCheckUnlockedZones(t *testing.T) {
	tests := []struct {
		totalXP  int
		expected []string
	}{
		{0, []string{"memory_cove"}},
		{29, []string{"memory_cove"}},
		{30, []string{"memory_cove", "focus_forest"}},
		{79, []string{"memory_cove", "focus_forest"}},
		{80, []string{"memory_cove", "focus_forest", "team_tower"}},
		{149, []string{"memory_cove", "focus_forest", "team_tower"}},
		{150, []string{"memory_cove", "focus_forest", "team_tower", "pattern_plateau"}},
		{249, []string{"memory_cove", "focus_forest", "team_tower", "pattern_plateau"}},
		{250, []string{"memory_cove", "focus_forest", "team_tower", "pattern_plateau", "community_hub"}},
		{1000, []string{"memory_cove", "focus_forest", "team_tower", "pattern_plateau", "community_hub"}},
	}

	for _, tc := range tests {
		got := CheckUnlockedZones(tc.totalXP)
		if len(got) != len(tc.expected) {
			t.Errorf("CheckUnlockedZones(%d) = %v, want %v", tc.totalXP, got, tc.expected)
			continue
		}
		for i, zone := range got {
			if zone != tc.expected[i] {
				t.Errorf("CheckUnlockedZones(%d)[%d] = %s, want %s", tc.totalXP, i, zone, tc.expected[i])
			}
		}
	}
}
