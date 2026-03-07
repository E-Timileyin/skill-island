package rooms

import (
	"errors"
	"math"
	"math/rand"

	"github.com/google/uuid"
)

// ---------- Types ----------

// Block represents a placed block in the tower.
type Block struct {
	ID       string  `json:"id"`
	Shape    string  `json:"shape"`     // "1x1"|"2x1"|"1x2"|"L_shape"|"T_shape"
	Colour   string  `json:"colour"`    // "blue"|"red"|"green"|"orange"|"grey"
	X        float64 `json:"x"`         // centre position, normalised 0.0–1.0
	Y        float64 `json:"y"`         // height from tower base, normalised
	Rotation float64 `json:"rotation"`  // degrees; 0 = upright
	PlacedBy string  `json:"placed_by"` // "player_1" | "player_2"
}

// TowerState holds the authoritative state of a Team Tower game.
type TowerState struct {
	Blocks        []Block `json:"blocks"`
	CurrentHeight float64 `json:"current_height"`
	TargetHeight  float64 `json:"target_height"`
	ActivePlayer  string  `json:"active_player"`   // "player_1" | "player_2"
	NextBlockShape string `json:"next_block_shape"`
	Stable        bool    `json:"stable"`           // false = tower fallen
	GroupXP       int     `json:"group_xp"`
	TurnNumber    int     `json:"turn_number"`
}

// TeamTowerAction represents a client action (block placement).
type TeamTowerAction struct {
	Type            string  `json:"type"`             // always "place_block"
	PositionX       float64 `json:"position_x"`       // 0.0–1.0
	ClientTimestamp int64   `json:"client_timestamp"`
}

// PlacementResult is the outcome of a ValidatePlacement call.
type PlacementResult struct {
	UpdatedState TowerState
	Outcome      string // "" | "win" | "lose"
	Error        error
}

// ScoredResult holds the final computed scores for any game type.
type ScoredResult struct {
	Stars        int
	XPPerPlayer  int
	GroupXP      int
}

// ---------- Block Shape Config ----------

type blockDims struct {
	Width  float64
	Height float64
}

type shapeWeight struct {
	CumulativeWeight int
	Shape            string
}

var blockDimensions = map[string]blockDims{
	"1x1":     {Width: 0.10, Height: 0.05},
	"2x1":     {Width: 0.20, Height: 0.05},
	"1x2":     {Width: 0.10, Height: 0.10},
	"L_shape": {Width: 0.15, Height: 0.10},
	"T_shape": {Width: 0.20, Height: 0.08},
}

var blockWeights = map[string]float64{
	"1x1":     1.0,
	"2x1":     2.0,
	"1x2":     1.5,
	"L_shape": 2.5,
	"T_shape": 2.5,
}

// Weighted distribution for shape selection.
// "1x1" = 30%, "2x1" = 25%, "1x2" = 20%, "L_shape" = 15%, "T_shape" = 10%.
var shapeDistribution = []shapeWeight{
	{30, "1x1"},
	{55, "2x1"},     // 30 + 25
	{75, "1x2"},     // 55 + 20
	{90, "L_shape"}, // 75 + 15
	{100, "T_shape"},// 90 + 10
}

var colours = []string{"blue", "red", "green", "orange", "grey"}

// ---------- Functions ----------

// NewTowerState creates the initial TowerState for a new game.
func NewTowerState(seed int64) TowerState {
	return TowerState{
		Blocks:         []Block{},
		CurrentHeight:  0.0,
		TargetHeight:   10.0,
		ActivePlayer:   "player_1",
		NextBlockShape: GetNextBlockShape(seed, 1),
		Stable:         true,
		GroupXP:        0,
		TurnNumber:     1,
	}
}

// GetNextBlockShape returns a deterministic block shape for a given seed and turn.
func GetNextBlockShape(seed int64, turnNumber int) string {
	r := rand.New(rand.NewSource(seed + int64(turnNumber)))
	roll := r.Intn(100)
	for _, sw := range shapeDistribution {
		if roll < sw.CumulativeWeight {
			return sw.Shape
		}
	}
	return "1x1" // fallback
}

// GetBlockDimensions returns the normalised width and height for a block shape.
func GetBlockDimensions(shape string) (width, height float64) {
	dims, ok := blockDimensions[shape]
	if !ok {
		return 0.10, 0.05 // default to 1x1
	}
	return dims.Width, dims.Height
}

// GetBlockWeight returns the weight of a block shape.
func GetBlockWeight(shape string) float64 {
	w, ok := blockWeights[shape]
	if !ok {
		return 1.0
	}
	return w
}

// getRandomColour returns a deterministic colour for a given seed and turn.
func getRandomColour(seed int64, turnNumber int) string {
	r := rand.New(rand.NewSource(seed + int64(turnNumber)*97))
	return colours[r.Intn(len(colours))]
}

// ValidatePlacement processes a block placement action and returns the updated state.
func ValidatePlacement(state TowerState, action TeamTowerAction, playerRole string, seed int64) PlacementResult {
	// Step 1 — Turn validation.
	if playerRole != state.ActivePlayer {
		return PlacementResult{Error: errors.New("not_your_turn")}
	}

	// Step 2 — Bounds validation.
	if action.PositionX < 0.05 || action.PositionX > 0.95 {
		return PlacementResult{Error: errors.New("out_of_bounds")}
	}

	// Step 3 — Calculate landing Y.
	newWidth, newHeight := GetBlockDimensions(state.NextBlockShape)
	landingY := newHeight / 2.0 // rests on floor if no blocks

	for _, b := range state.Blocks {
		bw, bh := GetBlockDimensions(b.Shape)
		// Check if X ranges overlap.
		blockLeft := b.X - bw/2.0
		blockRight := b.X + bw/2.0
		newLeft := action.PositionX - newWidth/2.0
		newRight := action.PositionX + newWidth/2.0

		if newRight > blockLeft && newLeft < blockRight {
			// Overlap — stack on top.
			topOfBlock := b.Y + bh/2.0 + newHeight/2.0
			if topOfBlock > landingY {
				landingY = topOfBlock
			}
		}
	}

	// Step 4 — Create new block.
	newBlock := Block{
		ID:       uuid.New().String(),
		Shape:    state.NextBlockShape,
		Colour:   getRandomColour(seed, state.TurnNumber),
		X:        action.PositionX,
		Y:        landingY,
		Rotation: 0,
		PlacedBy: playerRole,
	}
	state.Blocks = append(state.Blocks, newBlock)

	// Step 5 — Balance check (centre of mass).
	totalWeight := 0.0
	weightedX := 0.0
	for _, b := range state.Blocks {
		w := GetBlockWeight(b.Shape)
		totalWeight += w
		weightedX += b.X * w
	}
	centreOfMass := weightedX / totalWeight
	if math.Abs(centreOfMass-0.5) > 0.35 {
		state.Stable = false
		return PlacementResult{UpdatedState: state, Outcome: "lose"}
	}

	// Step 6 — Update height.
	maxHeight := 0.0
	for _, b := range state.Blocks {
		_, bh := GetBlockDimensions(b.Shape)
		top := b.Y + bh/2.0
		if top > maxHeight {
			maxHeight = top
		}
	}
	state.CurrentHeight = maxHeight

	// Step 7 — Win check.
	if state.CurrentHeight >= state.TargetHeight {
		return PlacementResult{UpdatedState: state, Outcome: "win"}
	}

	// Step 8 — Advance turn.
	state.GroupXP += 5
	state.TurnNumber++
	if state.ActivePlayer == "player_1" {
		state.ActivePlayer = "player_2"
	} else {
		state.ActivePlayer = "player_1"
	}
	state.NextBlockShape = GetNextBlockShape(seed, state.TurnNumber)

	return PlacementResult{UpdatedState: state, Outcome: ""}
}

// CalculateTeamTowerResult computes the final stars and XP for a Team Tower game.
//
//	Win AND currentHeight >= targetHeight * 1.1 → 3 stars
//	Win (any)                                   → 2 stars
//	Incomplete (disconnect, idle timeout)        → 1 star (SEND rule)
//	Lose (tower fall)                            → 1 star (never 0 in co-op)
//	groupXP = state.GroupXP + (20 if win, else 0)
//	xpPerPlayer = groupXP / 2
func CalculateTeamTowerResult(state TowerState, outcome string) ScoredResult {
	stars := 1 // minimum 1 star in co-op (SEND rule)
	winBonus := 0

	switch outcome {
	case "win":
		winBonus = 20
		if state.CurrentHeight >= state.TargetHeight*1.1 {
			stars = 3
		} else {
			stars = 2
		}
	case "lose":
		stars = 1 // tower fall — never 0 in co-op
	case "incomplete":
		stars = 1 // disconnect/timeout — effort reward
	}

	groupXP := state.GroupXP + winBonus
	xpPerPlayer := groupXP / 2

	return ScoredResult{
		Stars:       stars,
		XPPerPlayer: xpPerPlayer,
		GroupXP:     groupXP,
	}
}
