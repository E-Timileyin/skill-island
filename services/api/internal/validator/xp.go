package validator

// starToXP maps stars earned (0–3) to the XP awarded.
var starToXP = map[int]int{
	0: 0,
	1: 10,
	2: 20,
	3: 35,
}

// ZoneThreshold defines the XP required to unlock a game zone.
type ZoneThreshold struct {
	Zone      string
	XPRequired int
}

// zoneThresholds is ordered by ascending XP requirement.
var zoneThresholds = []ZoneThreshold{
	{Zone: "memory_cove", XPRequired: 0},
	{Zone: "focus_forest", XPRequired: 30},
	{Zone: "team_tower", XPRequired: 80},
	{Zone: "pattern_plateau", XPRequired: 150},
	{Zone: "community_hub", XPRequired: 250},
}

// CalculateXP returns the XP earned for a given number of stars (0–3).
// Stars outside the 0–3 range return 0 XP.
func CalculateXP(starsEarned int) int {
	xp, ok := starToXP[starsEarned]
	if !ok {
		return 0
	}
	return xp
}

// CheckUnlockedZones returns the list of zone names that are unlocked
// at the given total XP level.
func CheckUnlockedZones(totalXP int) []string {
	var unlocked []string
	for _, zt := range zoneThresholds {
		if totalXP >= zt.XPRequired {
			unlocked = append(unlocked, zt.Zone)
		}
	}
	return unlocked
}
