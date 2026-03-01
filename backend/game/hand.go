package game

// HandRank represents the quality of a hand (lower = better).
type HandRank int

const (
	RankPureSabacc HandRank = iota // two sylops
	RankSabacc                     // matching pair (lower value = better)
	RankNoSabacc                   // non-matching (lower difference = better)
)

type HandResult struct {
	Rank      HandRank `json:"rank"`
	Value     int      `json:"value"` // pair value for Sabacc, difference for NoSabacc
	SandCard  Card     `json:"sandCard"`
	BloodCard Card     `json:"bloodCard"`
}

// ResolveHand computes the final HandResult for a pair of cards.
// sandValue and bloodValue are the effective values after impostor dice rolls
// and sylop resolution.
func ResolveHand(sand, blood Card, sandValue, bloodValue int) HandResult {
	if sand.Kind == KindSylop && blood.Kind == KindSylop {
		return HandResult{Rank: RankPureSabacc, Value: 0, SandCard: sand, BloodCard: blood}
	}

	if sandValue == bloodValue {
		return HandResult{Rank: RankSabacc, Value: sandValue, SandCard: sand, BloodCard: blood}
	}

	diff := sandValue - bloodValue
	if diff < 0 {
		diff = -diff
	}
	return HandResult{Rank: RankNoSabacc, Value: diff, SandCard: sand, BloodCard: blood}
}

// CompareHands returns -1 if a is better, 1 if b is better, 0 if tied.
func CompareHands(a, b HandResult) int {
	if a.Rank != b.Rank {
		if a.Rank < b.Rank {
			return -1
		}
		return 1
	}
	// Same rank: lower value is better for both Sabacc and NoSabacc
	if a.Value < b.Value {
		return -1
	}
	if a.Value > b.Value {
		return 1
	}
	return 0
}

// HandRankName returns a string name for a HandRank value.
func HandRankName(rank HandRank) string {
	switch rank {
	case RankPureSabacc:
		return "pure_sabacc"
	case RankSabacc:
		return "sabacc"
	case RankNoSabacc:
		return "no_sabacc"
	default:
		return "unknown"
	}
}

// EffectiveValue returns the resolved numeric value of a card given the
// other card in the hand (needed for Sylop which mirrors its partner).
func EffectiveValue(card, other Card, diceRoll int) int {
	switch card.Kind {
	case KindSylop:
		if other.Kind == KindSylop {
			return 0 // Pure Sabacc — value irrelevant for comparison
		}
		return EffectiveValue(other, card, diceRoll)
	case KindImpostor:
		return diceRoll
	default:
		return card.Value
	}
}
