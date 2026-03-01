package game

import "testing"

// ---------------------------------------------------------------------------
// ResolveHand tests
// ---------------------------------------------------------------------------

func TestResolveHandPureSabacc(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindSylop, Value: 0}
	blood := Card{Suit: SuitBlood, Kind: KindSylop, Value: 0}

	h := ResolveHand(sand, blood, 0, 0)
	if h.Rank != RankPureSabacc {
		t.Errorf("two sylops should be RankPureSabacc, got %d", h.Rank)
	}
	if h.Value != 0 {
		t.Errorf("PureSabacc value should be 0, got %d", h.Value)
	}
}

func TestResolveHandSabaccMatchingValues(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 3}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 3}

	h := ResolveHand(sand, blood, 3, 3)
	if h.Rank != RankSabacc {
		t.Errorf("matching values should be RankSabacc, got %d", h.Rank)
	}
	if h.Value != 3 {
		t.Errorf("Sabacc value should be 3, got %d", h.Value)
	}
}

func TestResolveHandSabaccValueOne(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 1}

	h := ResolveHand(sand, blood, 1, 1)
	if h.Rank != RankSabacc {
		t.Errorf("matching 1s should be RankSabacc, got %d", h.Rank)
	}
	if h.Value != 1 {
		t.Errorf("Sabacc value should be 1, got %d", h.Value)
	}
}

func TestResolveHandSabaccValueSix(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 6}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 6}

	h := ResolveHand(sand, blood, 6, 6)
	if h.Rank != RankSabacc {
		t.Errorf("matching 6s should be RankSabacc, got %d", h.Rank)
	}
	if h.Value != 6 {
		t.Errorf("Sabacc value should be 6, got %d", h.Value)
	}
}

func TestResolveHandNoSabacc(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 2}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 5}

	h := ResolveHand(sand, blood, 2, 5)
	if h.Rank != RankNoSabacc {
		t.Errorf("non-matching values should be RankNoSabacc, got %d", h.Rank)
	}
	if h.Value != 3 {
		t.Errorf("NoSabacc value should be |2-5|=3, got %d", h.Value)
	}
}

func TestResolveHandNoSabaccReverseDiff(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 5}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 2}

	h := ResolveHand(sand, blood, 5, 2)
	if h.Rank != RankNoSabacc {
		t.Errorf("non-matching values should be RankNoSabacc, got %d", h.Rank)
	}
	if h.Value != 3 {
		t.Errorf("NoSabacc value should be |5-2|=3, got %d", h.Value)
	}
}

func TestResolveHandNoSabaccMaxDifference(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 6}

	h := ResolveHand(sand, blood, 1, 6)
	if h.Value != 5 {
		t.Errorf("NoSabacc value should be |1-6|=5, got %d", h.Value)
	}
}

func TestResolveHandCardsPreserved(t *testing.T) {
	sand := Card{Suit: SuitSand, Kind: KindValue, Value: 4, ID: 10}
	blood := Card{Suit: SuitBlood, Kind: KindValue, Value: 4, ID: 30}

	h := ResolveHand(sand, blood, 4, 4)
	if h.SandCard.ID != 10 {
		t.Errorf("SandCard not preserved in HandResult")
	}
	if h.BloodCard.ID != 30 {
		t.Errorf("BloodCard not preserved in HandResult")
	}
}

// ---------------------------------------------------------------------------
// CompareHands tests
// ---------------------------------------------------------------------------

func TestCompareHandsPureSabaccBeatsAll(t *testing.T) {
	pure := HandResult{Rank: RankPureSabacc, Value: 0}
	sabacc := HandResult{Rank: RankSabacc, Value: 1}
	noSabacc := HandResult{Rank: RankNoSabacc, Value: 1}

	if CompareHands(pure, sabacc) != -1 {
		t.Error("PureSabacc should beat Sabacc")
	}
	if CompareHands(pure, noSabacc) != -1 {
		t.Error("PureSabacc should beat NoSabacc")
	}
}

func TestCompareHandsSabaccBeatsNoSabacc(t *testing.T) {
	sabacc := HandResult{Rank: RankSabacc, Value: 6}
	noSabacc := HandResult{Rank: RankNoSabacc, Value: 1}

	if CompareHands(sabacc, noSabacc) != -1 {
		t.Error("Sabacc (even value 6) should beat NoSabacc (value 1)")
	}
}

func TestCompareHandsNoSabaccLosesToSabacc(t *testing.T) {
	sabacc := HandResult{Rank: RankSabacc, Value: 1}
	noSabacc := HandResult{Rank: RankNoSabacc, Value: 1}

	if CompareHands(noSabacc, sabacc) != 1 {
		t.Error("NoSabacc should lose to Sabacc")
	}
}

func TestCompareHandsSameRankLowerValueWins(t *testing.T) {
	a := HandResult{Rank: RankSabacc, Value: 2}
	b := HandResult{Rank: RankSabacc, Value: 5}

	if CompareHands(a, b) != -1 {
		t.Error("Sabacc(2) should beat Sabacc(5)")
	}
	if CompareHands(b, a) != 1 {
		t.Error("Sabacc(5) should lose to Sabacc(2)")
	}
}

func TestCompareHandsNoSabaccTiebreaker(t *testing.T) {
	a := HandResult{Rank: RankNoSabacc, Value: 1}
	b := HandResult{Rank: RankNoSabacc, Value: 4}

	if CompareHands(a, b) != -1 {
		t.Error("NoSabacc(1) should beat NoSabacc(4)")
	}
}

func TestCompareHandsTie(t *testing.T) {
	a := HandResult{Rank: RankSabacc, Value: 3}
	b := HandResult{Rank: RankSabacc, Value: 3}

	if CompareHands(a, b) != 0 {
		t.Error("identical hands should tie")
	}
}

func TestCompareHandsTiePureSabacc(t *testing.T) {
	a := HandResult{Rank: RankPureSabacc, Value: 0}
	b := HandResult{Rank: RankPureSabacc, Value: 0}

	if CompareHands(a, b) != 0 {
		t.Error("two PureSabacc hands should tie")
	}
}

func TestCompareHandsTieNoSabacc(t *testing.T) {
	a := HandResult{Rank: RankNoSabacc, Value: 2}
	b := HandResult{Rank: RankNoSabacc, Value: 2}

	if CompareHands(a, b) != 0 {
		t.Error("two NoSabacc hands with same value should tie")
	}
}

// ---------------------------------------------------------------------------
// EffectiveValue tests
// ---------------------------------------------------------------------------

func TestEffectiveValueForValueCard(t *testing.T) {
	card := Card{Kind: KindValue, Value: 4}
	other := Card{Kind: KindValue, Value: 2}

	if v := EffectiveValue(card, other, 0); v != 4 {
		t.Errorf("value card should return its own value, got %d", v)
	}
}

func TestEffectiveValueForValueCardIgnoresDiceRoll(t *testing.T) {
	card := Card{Kind: KindValue, Value: 3}
	other := Card{Kind: KindValue, Value: 5}

	if v := EffectiveValue(card, other, 6); v != 3 {
		t.Errorf("value card should ignore dice roll, expected 3, got %d", v)
	}
}

func TestEffectiveValueForImpostor(t *testing.T) {
	card := Card{Kind: KindImpostor, Value: 0}
	other := Card{Kind: KindValue, Value: 3}

	if v := EffectiveValue(card, other, 5); v != 5 {
		t.Errorf("impostor should return dice roll 5, got %d", v)
	}
}

func TestEffectiveValueForImpostorAllDiceValues(t *testing.T) {
	card := Card{Kind: KindImpostor, Value: 0}
	other := Card{Kind: KindValue, Value: 1}

	for roll := 1; roll <= 6; roll++ {
		if v := EffectiveValue(card, other, roll); v != roll {
			t.Errorf("impostor with dice roll %d should return %d, got %d", roll, roll, v)
		}
	}
}

func TestEffectiveValueForSylopMirrorsValuePartner(t *testing.T) {
	sylop := Card{Kind: KindSylop, Value: 0}
	partner := Card{Kind: KindValue, Value: 4}

	if v := EffectiveValue(sylop, partner, 0); v != 4 {
		t.Errorf("sylop should mirror partner value 4, got %d", v)
	}
}

func TestEffectiveValueForSylopMirrorsImpostorPartner(t *testing.T) {
	sylop := Card{Kind: KindSylop, Value: 0}
	partner := Card{Kind: KindImpostor, Value: 0}

	// Sylop mirrors impostor, which uses diceRoll
	if v := EffectiveValue(sylop, partner, 3); v != 3 {
		t.Errorf("sylop mirroring impostor with dice 3 should be 3, got %d", v)
	}
}

func TestEffectiveValueForDoubleSylop(t *testing.T) {
	sylop1 := Card{Kind: KindSylop, Value: 0}
	sylop2 := Card{Kind: KindSylop, Value: 0}

	if v := EffectiveValue(sylop1, sylop2, 0); v != 0 {
		t.Errorf("double sylop effective value should be 0, got %d", v)
	}
}

func TestEffectiveValueForValueCardRange(t *testing.T) {
	other := Card{Kind: KindValue, Value: 1}
	for val := 1; val <= 6; val++ {
		card := Card{Kind: KindValue, Value: val}
		if v := EffectiveValue(card, other, 0); v != val {
			t.Errorf("value card %d should return %d, got %d", val, val, v)
		}
	}
}
