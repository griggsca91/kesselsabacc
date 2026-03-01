package game

import "testing"

// ---------------------------------------------------------------------------
// NewPlayer tests
// ---------------------------------------------------------------------------

func TestNewPlayerStartingState(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, true)

	if p.ID != "p1" {
		t.Errorf("expected ID p1, got %s", p.ID)
	}
	if p.Name != "Alice" {
		t.Errorf("expected Name Alice, got %s", p.Name)
	}
	if p.Chips != 6 {
		t.Errorf("expected 6 chips, got %d", p.Chips)
	}
	if p.IsHost != true {
		t.Error("expected IsHost to be true")
	}
	if p.Eliminated {
		t.Error("new player should not be eliminated")
	}
	if p.Stood {
		t.Error("new player should not have stood")
	}
	if p.Invested != 0 {
		t.Errorf("new player should have 0 invested, got %d", p.Invested)
	}
	if p.SandCard != nil {
		t.Error("new player should have nil SandCard")
	}
	if p.BloodCard != nil {
		t.Error("new player should have nil BloodCard")
	}
}

func TestNewPlayerNonHost(t *testing.T) {
	p := NewPlayer("p2", "Bob", 6, false)
	if p.IsHost {
		t.Error("expected IsHost to be false")
	}
}

func TestNewPlayerStartingTokens(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, true)

	if len(p.ShiftTokens) != len(StartingTokens) {
		t.Fatalf("expected %d starting tokens, got %d", len(StartingTokens), len(p.ShiftTokens))
	}
	for i, tok := range StartingTokens {
		if p.ShiftTokens[i] != tok {
			t.Errorf("starting token %d: expected %s, got %s", i, tok, p.ShiftTokens[i])
		}
	}
}

func TestNewPlayerTokensAreIndependentCopy(t *testing.T) {
	p1 := NewPlayer("p1", "Alice", 6, true)
	p2 := NewPlayer("p2", "Bob", 6, false)

	p1.ShiftTokens[0] = "modified"
	if p2.ShiftTokens[0] == "modified" {
		t.Error("player tokens should be independent copies")
	}
}

func TestNewPlayerEmptyUsedTokens(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, true)
	if len(p.UsedTokens) != 0 {
		t.Errorf("expected 0 used tokens, got %d", len(p.UsedTokens))
	}
}

// ---------------------------------------------------------------------------
// HasToken tests
// ---------------------------------------------------------------------------

func TestHasTokenPresent(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	if !p.HasToken(TokenFreeDraw) {
		t.Error("new player should have TokenFreeDraw")
	}
	if !p.HasToken(TokenRefund) {
		t.Error("new player should have TokenRefund")
	}
	if !p.HasToken(TokenGeneralTariff) {
		t.Error("new player should have TokenGeneralTariff")
	}
}

func TestHasTokenAbsent(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	if p.HasToken(TokenMarkdown) {
		t.Error("new player should not have TokenMarkdown")
	}
	if p.HasToken(TokenCookTheBooks) {
		t.Error("new player should not have TokenCookTheBooks")
	}
	if p.HasToken(TokenImmunity) {
		t.Error("new player should not have TokenImmunity")
	}
}

func TestHasTokenAfterUse(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)
	p.UseToken(TokenFreeDraw)

	if p.HasToken(TokenFreeDraw) {
		t.Error("should not have TokenFreeDraw after using it")
	}
}

// ---------------------------------------------------------------------------
// UseToken tests
// ---------------------------------------------------------------------------

func TestUseTokenSuccess(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	ok := p.UseToken(TokenFreeDraw)
	if !ok {
		t.Error("UseToken should return true for available token")
	}
	if len(p.ShiftTokens) != 2 {
		t.Errorf("expected 2 shift tokens after use, got %d", len(p.ShiftTokens))
	}
	if len(p.UsedTokens) != 1 {
		t.Errorf("expected 1 used token, got %d", len(p.UsedTokens))
	}
	if p.UsedTokens[0] != TokenFreeDraw {
		t.Errorf("expected used token to be TokenFreeDraw, got %s", p.UsedTokens[0])
	}
}

func TestUseTokenFail(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	ok := p.UseToken(TokenMarkdown)
	if ok {
		t.Error("UseToken should return false for token player doesn't have")
	}
	if len(p.ShiftTokens) != 3 {
		t.Errorf("shift tokens should be unchanged, got %d", len(p.ShiftTokens))
	}
}

func TestUseTokenTwiceFails(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	p.UseToken(TokenFreeDraw)
	ok := p.UseToken(TokenFreeDraw)
	if ok {
		t.Error("UseToken should fail when token already used")
	}
}

func TestUseAllTokens(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	p.UseToken(TokenFreeDraw)
	p.UseToken(TokenRefund)
	p.UseToken(TokenGeneralTariff)

	if len(p.ShiftTokens) != 0 {
		t.Errorf("expected 0 tokens left, got %d", len(p.ShiftTokens))
	}
	if len(p.UsedTokens) != 3 {
		t.Errorf("expected 3 used tokens, got %d", len(p.UsedTokens))
	}
}

// ---------------------------------------------------------------------------
// Invest tests
// ---------------------------------------------------------------------------

func TestInvestSuccess(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	ok := p.Invest(2)
	if !ok {
		t.Error("Invest should succeed with enough chips")
	}
	if p.Chips != 4 {
		t.Errorf("expected 4 chips after investing 2, got %d", p.Chips)
	}
	if p.Invested != 2 {
		t.Errorf("expected 2 invested, got %d", p.Invested)
	}
}

func TestInvestExactChips(t *testing.T) {
	p := NewPlayer("p1", "Alice", 3, false)

	ok := p.Invest(3)
	if !ok {
		t.Error("Invest should succeed with exactly enough chips")
	}
	if p.Chips != 0 {
		t.Errorf("expected 0 chips, got %d", p.Chips)
	}
	if p.Invested != 3 {
		t.Errorf("expected 3 invested, got %d", p.Invested)
	}
}

func TestInvestNotEnoughChips(t *testing.T) {
	p := NewPlayer("p1", "Alice", 2, false)

	ok := p.Invest(3)
	if ok {
		t.Error("Invest should fail with not enough chips")
	}
	if p.Chips != 2 {
		t.Errorf("chips should be unchanged at 2, got %d", p.Chips)
	}
	if p.Invested != 0 {
		t.Errorf("invested should be unchanged at 0, got %d", p.Invested)
	}
}

func TestInvestZeroChips(t *testing.T) {
	p := NewPlayer("p1", "Alice", 0, false)

	ok := p.Invest(1)
	if ok {
		t.Error("Invest should fail with zero chips")
	}
}

func TestInvestMultiple(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	p.Invest(1)
	p.Invest(2)

	if p.Chips != 3 {
		t.Errorf("expected 3 chips after investing 1+2, got %d", p.Chips)
	}
	if p.Invested != 3 {
		t.Errorf("expected 3 invested, got %d", p.Invested)
	}
}

// ---------------------------------------------------------------------------
// Refund tests
// ---------------------------------------------------------------------------

func TestRefund(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)
	p.Invest(3)

	p.Refund(2)
	if p.Chips != 5 {
		t.Errorf("expected 5 chips after refund 2, got %d", p.Chips)
	}
	if p.Invested != 1 {
		t.Errorf("expected 1 invested after refund 2, got %d", p.Invested)
	}
}

func TestRefundMoreThanInvested(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)
	p.Invest(1)

	p.Refund(3) // refund more than invested
	if p.Invested != 0 {
		t.Errorf("invested should clamp to 0, got %d", p.Invested)
	}
	// Chips increase by the full refund amount
	if p.Chips != 8 {
		t.Errorf("expected 8 chips (5+3), got %d", p.Chips)
	}
}

func TestRefundWithNoInvestment(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)

	p.Refund(2)
	if p.Invested != 0 {
		t.Errorf("invested should stay 0, got %d", p.Invested)
	}
	if p.Chips != 8 {
		t.Errorf("expected 8 chips, got %d", p.Chips)
	}
}

// ---------------------------------------------------------------------------
// ResetRound tests
// ---------------------------------------------------------------------------

func TestResetRound(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)
	p.Invested = 3
	p.Stood = true
	sandCard := Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	bloodCard := Card{Suit: SuitBlood, Kind: KindValue, Value: 2}
	p.SandCard = &sandCard
	p.BloodCard = &bloodCard
	p.SandDiceRoll = 4
	p.BloodDiceRoll = 5

	p.ResetRound()

	if p.Invested != 0 {
		t.Errorf("expected Invested=0 after reset, got %d", p.Invested)
	}
	if p.Stood {
		t.Error("expected Stood=false after reset")
	}
	if p.SandCard != nil {
		t.Error("expected SandCard=nil after reset")
	}
	if p.BloodCard != nil {
		t.Error("expected BloodCard=nil after reset")
	}
	if p.SandDiceRoll != 0 {
		t.Errorf("expected SandDiceRoll=0 after reset, got %d", p.SandDiceRoll)
	}
	if p.BloodDiceRoll != 0 {
		t.Errorf("expected BloodDiceRoll=0 after reset, got %d", p.BloodDiceRoll)
	}
}

func TestResetRoundPreservesChipsAndTokens(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, false)
	p.UseToken(TokenFreeDraw)

	p.ResetRound()

	if p.Chips != 6 {
		t.Errorf("ResetRound should not change chips, got %d", p.Chips)
	}
	if len(p.ShiftTokens) != 2 {
		t.Errorf("ResetRound should not change shift tokens, got %d", len(p.ShiftTokens))
	}
	if len(p.UsedTokens) != 1 {
		t.Errorf("ResetRound should not change used tokens, got %d", len(p.UsedTokens))
	}
}

func TestResetRoundPreservesEliminatedAndIdentity(t *testing.T) {
	p := NewPlayer("p1", "Alice", 6, true)
	p.Eliminated = true

	p.ResetRound()

	if !p.Eliminated {
		t.Error("ResetRound should not change Eliminated")
	}
	if p.ID != "p1" {
		t.Error("ResetRound should not change ID")
	}
	if p.Name != "Alice" {
		t.Error("ResetRound should not change Name")
	}
	if !p.IsHost {
		t.Error("ResetRound should not change IsHost")
	}
}
