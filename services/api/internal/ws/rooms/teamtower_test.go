package rooms

import (
	"testing"
)

// -------------------------------------------------------------------
// TestValidatePlacement_WrongTurn
// player_2 submits when ActivePlayer is "player_1" → error "not_your_turn"
// -------------------------------------------------------------------
func TestValidatePlacement_WrongTurn(t *testing.T) {
	state := NewTowerState(42)
	action := TeamTowerAction{Type: "place_block", PositionX: 0.5, ClientTimestamp: 1000}

	result := ValidatePlacement(state, action, "player_2", 42)
	if result.Error == nil {
		t.Fatal("expected error for wrong turn, got nil")
	}
	if result.Error.Error() != "not_your_turn" {
		t.Errorf("expected 'not_your_turn', got '%s'", result.Error.Error())
	}
}

// -------------------------------------------------------------------
// TestValidatePlacement_OutOfBounds
// position_x = 0.03 → error "out_of_bounds"
// position_x = 0.97 → error "out_of_bounds"
// position_x = 0.50 → no error
// -------------------------------------------------------------------
func TestValidatePlacement_OutOfBounds(t *testing.T) {
	state := NewTowerState(42)

	tests := []struct {
		name      string
		posX      float64
		expectErr bool
	}{
		{"too far left", 0.03, true},
		{"too far right", 0.97, true},
		{"valid centre", 0.50, false},
		{"left boundary", 0.05, false},
		{"right boundary", 0.95, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := TeamTowerAction{Type: "place_block", PositionX: tt.posX, ClientTimestamp: 1000}
			result := ValidatePlacement(state, action, "player_1", 42)

			if tt.expectErr {
				if result.Error == nil {
					t.Fatalf("expected error for posX=%.2f, got nil", tt.posX)
				}
				if result.Error.Error() != "out_of_bounds" {
					t.Errorf("expected 'out_of_bounds', got '%s'", result.Error.Error())
				}
			} else {
				if result.Error != nil {
					t.Fatalf("expected no error for posX=%.2f, got '%s'", tt.posX, result.Error)
				}
			}
		})
	}
}

// -------------------------------------------------------------------
// TestValidatePlacement_StacksCorrectly
// Place block at x=0.5 → Y = block_height/2
// Place second block at x=0.5 → Y > first block Y
// -------------------------------------------------------------------
func TestValidatePlacement_StacksCorrectly(t *testing.T) {
	seed := int64(42)
	state := NewTowerState(seed)

	// Place first block.
	action1 := TeamTowerAction{Type: "place_block", PositionX: 0.5, ClientTimestamp: 1000}
	result1 := ValidatePlacement(state, action1, "player_1", seed)
	if result1.Error != nil {
		t.Fatalf("first placement error: %v", result1.Error)
	}

	firstBlock := result1.UpdatedState.Blocks[0]
	_, firstHeight := GetBlockDimensions(firstBlock.Shape)

	// First block should rest on floor: Y = height/2.
	if firstBlock.Y != firstHeight/2.0 {
		t.Errorf("first block Y=%.4f, expected %.4f (height/2)", firstBlock.Y, firstHeight/2.0)
	}

	// Place second block — now it's player_2's turn.
	action2 := TeamTowerAction{Type: "place_block", PositionX: 0.5, ClientTimestamp: 2000}
	result2 := ValidatePlacement(result1.UpdatedState, action2, "player_2", seed)
	if result2.Error != nil {
		t.Fatalf("second placement error: %v", result2.Error)
	}

	secondBlock := result2.UpdatedState.Blocks[1]
	if secondBlock.Y <= firstBlock.Y {
		t.Errorf("second block Y=%.4f should be > first block Y=%.4f", secondBlock.Y, firstBlock.Y)
	}
}

// -------------------------------------------------------------------
// TestValidatePlacement_BalanceCheck
// Stack 5 blocks at x=0.9 → Stable = false, Outcome = "lose"
// Stack 5 blocks alternating x=0.3 and x=0.7 → Stable = true
// -------------------------------------------------------------------
func TestValidatePlacement_BalanceCheck(t *testing.T) {
	t.Run("unbalanced tower falls", func(t *testing.T) {
		seed := int64(42)
		state := NewTowerState(seed)
		player := "player_1"

		var result PlacementResult
		for i := 0; i < 5; i++ {
			action := TeamTowerAction{Type: "place_block", PositionX: 0.9, ClientTimestamp: int64(i * 1000)}
			result = ValidatePlacement(state, action, player, seed)
			if result.Error != nil {
				t.Fatalf("placement %d error: %v", i, result.Error)
			}
			if result.Outcome == "lose" {
				// Tower fell — expected.
				if result.UpdatedState.Stable != false {
					t.Error("expected Stable=false on tower fall")
				}
				return
			}
			state = result.UpdatedState
			if player == "player_1" {
				player = "player_2"
			} else {
				player = "player_1"
			}
		}
		t.Error("expected tower to fall after 5 blocks at x=0.9, but it did not")
	})

	t.Run("balanced tower stays stable", func(t *testing.T) {
		seed := int64(42)
		state := NewTowerState(seed)
		player := "player_1"
		positions := []float64{0.3, 0.7, 0.3, 0.7, 0.3}

		for i, posX := range positions {
			action := TeamTowerAction{Type: "place_block", PositionX: posX, ClientTimestamp: int64(i * 1000)}
			result := ValidatePlacement(state, action, player, seed)
			if result.Error != nil {
				t.Fatalf("placement %d error: %v", i, result.Error)
			}
			if result.Outcome == "lose" {
				t.Fatalf("tower fell at placement %d with balanced positions", i)
			}
			state = result.UpdatedState
			if player == "player_1" {
				player = "player_2"
			} else {
				player = "player_1"
			}
		}
		if !state.Stable {
			t.Error("expected tower to remain stable with balanced placements")
		}
	})
}

// -------------------------------------------------------------------
// TestValidatePlacement_WinCondition
// Simulate placements until CurrentHeight >= TargetHeight → Outcome = "win"
// -------------------------------------------------------------------
func TestValidatePlacement_WinCondition(t *testing.T) {
	seed := int64(42)
	state := NewTowerState(seed)
	// Set a low target height to make win achievable within test.
	state.TargetHeight = 0.15
	player := "player_1"

	for i := 0; i < 100; i++ {
		action := TeamTowerAction{Type: "place_block", PositionX: 0.5, ClientTimestamp: int64(i * 1000)}
		result := ValidatePlacement(state, action, player, seed)
		if result.Error != nil {
			t.Fatalf("placement %d error: %v", i, result.Error)
		}
		if result.Outcome == "win" {
			if result.UpdatedState.CurrentHeight < state.TargetHeight {
				t.Errorf("won but height %.4f < target %.4f", result.UpdatedState.CurrentHeight, state.TargetHeight)
			}
			return
		}
		if result.Outcome == "lose" {
			t.Fatal("tower fell before win — reduce test target height")
		}
		state = result.UpdatedState
		if player == "player_1" {
			player = "player_2"
		} else {
			player = "player_1"
		}
	}
	t.Error("did not reach win condition after 100 placements")
}

// -------------------------------------------------------------------
// TestGetNextBlockShape_Deterministic
// Same seed + turnNumber always returns same shape.
// -------------------------------------------------------------------
func TestGetNextBlockShape_Deterministic(t *testing.T) {
	seed := int64(12345)
	for turn := 1; turn <= 20; turn++ {
		shape1 := GetNextBlockShape(seed, turn)
		shape2 := GetNextBlockShape(seed, turn)
		if shape1 != shape2 {
			t.Errorf("turn %d: shape mismatch: %s vs %s", turn, shape1, shape2)
		}
		// Verify it's a valid shape.
		if _, ok := blockDimensions[shape1]; !ok {
			t.Errorf("turn %d: invalid shape '%s'", turn, shape1)
		}
	}
}

// -------------------------------------------------------------------
// TestCalculateTeamTowerResult_Stars
// Win + extra height → 3 stars
// Win normal → 2 stars
// Incomplete → 1 star
// Lose (tower fall) → 1 star (not 0)
// -------------------------------------------------------------------
func TestCalculateTeamTowerResult_Stars(t *testing.T) {
	cases := []struct {
		name          string
		outcome       string
		currentHeight float64
		targetHeight  float64
		expectedStars int
	}{
		{
			"win with extra height",
			"win",
			11.0,  // >= 10.0 * 1.1
			10.0,
			3,
		},
		{
			"win normal",
			"win",
			10.0, // >= 10.0 but < 11.0
			10.0,
			2,
		},
		{
			"incomplete",
			"incomplete",
			5.0,
			10.0,
			1,
		},
		{
			"lose tower fall",
			"lose",
			5.0,
			10.0,
			1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := TowerState{
				CurrentHeight: c.currentHeight,
				TargetHeight:  c.targetHeight,
				GroupXP:       50,
			}
			result := CalculateTeamTowerResult(state, c.outcome)
			if result.Stars != c.expectedStars {
				t.Errorf("expected %d stars, got %d", c.expectedStars, result.Stars)
			}
		})
	}
}

// -------------------------------------------------------------------
// TestCalculateTeamTowerResult_XPSplit
// 10 placements + win → groupXP = 50 + 20 = 70 → 35 XP per player
// -------------------------------------------------------------------
func TestCalculateTeamTowerResult_XPSplit(t *testing.T) {
	state := TowerState{
		GroupXP:       50, // 10 placements × 5 XP each
		CurrentHeight: 10.0,
		TargetHeight:  10.0,
	}
	result := CalculateTeamTowerResult(state, "win")

	expectedGroupXP := 50 + 20 // 70
	if result.GroupXP != expectedGroupXP {
		t.Errorf("expected groupXP=%d, got %d", expectedGroupXP, result.GroupXP)
	}

	expectedXPPerPlayer := expectedGroupXP / 2 // 35
	if result.XPPerPlayer != expectedXPPerPlayer {
		t.Errorf("expected xpPerPlayer=%d, got %d", expectedXPPerPlayer, result.XPPerPlayer)
	}
}

// -------------------------------------------------------------------
// TestCalculateTeamTowerResult_LoseNeverZeroStars
// Lose should always give at least 1 star (SEND rule).
// -------------------------------------------------------------------
func TestCalculateTeamTowerResult_LoseNeverZeroStars(t *testing.T) {
	state := TowerState{GroupXP: 0, CurrentHeight: 0.5, TargetHeight: 10.0}
	result := CalculateTeamTowerResult(state, "lose")
	if result.Stars < 1 {
		t.Errorf("expected at least 1 star for lose, got %d", result.Stars)
	}
}

// -------------------------------------------------------------------
// TestGetBlockDimensions
// Verify all block shapes return correct dimensions.
// -------------------------------------------------------------------
func TestGetBlockDimensions(t *testing.T) {
	cases := []struct {
		shape    string
		width    float64
		height   float64
	}{
		{"1x1", 0.10, 0.05},
		{"2x1", 0.20, 0.05},
		{"1x2", 0.10, 0.10},
		{"L_shape", 0.15, 0.10},
		{"T_shape", 0.20, 0.08},
	}
	for _, c := range cases {
		w, h := GetBlockDimensions(c.shape)
		if w != c.width || h != c.height {
			t.Errorf("%s: expected (%.2f, %.2f), got (%.2f, %.2f)", c.shape, c.width, c.height, w, h)
		}
	}
}

// -------------------------------------------------------------------
// TestGetBlockWeight
// Verify all block weights.
// -------------------------------------------------------------------
func TestGetBlockWeight(t *testing.T) {
	cases := []struct {
		shape  string
		weight float64
	}{
		{"1x1", 1.0},
		{"2x1", 2.0},
		{"1x2", 1.5},
		{"L_shape", 2.5},
		{"T_shape", 2.5},
	}
	for _, c := range cases {
		w := GetBlockWeight(c.shape)
		if w != c.weight {
			t.Errorf("%s: expected weight %.1f, got %.1f", c.shape, c.weight, w)
		}
	}
}

// -------------------------------------------------------------------
// TestNewTowerState
// Verify initial state values.
// -------------------------------------------------------------------
func TestNewTowerState(t *testing.T) {
	state := NewTowerState(42)

	if state.TargetHeight != 10.0 {
		t.Errorf("expected TargetHeight=10.0, got %.1f", state.TargetHeight)
	}
	if state.CurrentHeight != 0.0 {
		t.Errorf("expected CurrentHeight=0.0, got %.1f", state.CurrentHeight)
	}
	if !state.Stable {
		t.Error("expected Stable=true")
	}
	if state.ActivePlayer != "player_1" {
		t.Errorf("expected ActivePlayer=player_1, got %s", state.ActivePlayer)
	}
	if state.TurnNumber != 1 {
		t.Errorf("expected TurnNumber=1, got %d", state.TurnNumber)
	}
	if state.GroupXP != 0 {
		t.Errorf("expected GroupXP=0, got %d", state.GroupXP)
	}
	if len(state.Blocks) != 0 {
		t.Errorf("expected empty blocks, got %d", len(state.Blocks))
	}
	if state.NextBlockShape == "" {
		t.Error("expected NextBlockShape to be set")
	}
}
