package game

import "testing"

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setupGame creates a game with the given number of players and starts it.
// Players are named p0, p1, ... with StartingChips each.
func setupGame(n int) *Game {
	g := NewGame()
	for i := 0; i < n; i++ {
		id := "p" + string(rune('0'+i))
		g.Players = append(g.Players, NewPlayer(id, "Player"+id, StartingChips, i == 0))
	}
	return g
}

// startedGame returns a started game with n players.
func startedGame(n int) *Game {
	g := setupGame(n)
	g.Start()
	return g
}

// currentID returns the ID of the current turn player.
func currentID(g *Game) string {
	if g.CurrentTurn < len(g.PlayerOrder) {
		return g.PlayerOrder[g.CurrentTurn]
	}
	return ""
}

// tokenPtr is a convenience helper to get a pointer to a ShiftToken.
func tokenPtr(t ShiftToken) *ShiftToken {
	return &t
}

// ---------------------------------------------------------------------------
// NewGame tests
// ---------------------------------------------------------------------------

func TestNewGame(t *testing.T) {
	g := NewGame()
	if g.Phase != PhaseLobby {
		t.Errorf("new game should be in lobby phase, got %s", g.Phase)
	}
	if len(g.Players) != 0 {
		t.Errorf("new game should have 0 players, got %d", len(g.Players))
	}
}

// ---------------------------------------------------------------------------
// PlayerByID tests
// ---------------------------------------------------------------------------

func TestPlayerByIDFound(t *testing.T) {
	g := setupGame(3)
	p := g.PlayerByID("p0")
	if p == nil {
		t.Fatal("expected to find p0")
	}
	if p.ID != "p0" {
		t.Errorf("expected p0, got %s", p.ID)
	}
}

func TestPlayerByIDNotFound(t *testing.T) {
	g := setupGame(2)
	p := g.PlayerByID("nonexistent")
	if p != nil {
		t.Error("expected nil for unknown player ID")
	}
}

// ---------------------------------------------------------------------------
// Start tests
// ---------------------------------------------------------------------------

func TestStartMinPlayers(t *testing.T) {
	g := setupGame(1)
	err := g.Start()
	if err == nil {
		t.Error("Start with 1 player should return error")
	}
}

func TestStartTwoPlayers(t *testing.T) {
	g := setupGame(2)
	err := g.Start()
	if err != nil {
		t.Fatalf("Start with 2 players should succeed: %v", err)
	}
	if g.Phase != PhaseTurn {
		t.Errorf("expected PhaseTurn after start, got %s", g.Phase)
	}
	if g.Round != 1 {
		t.Errorf("expected round 1, got %d", g.Round)
	}
}

func TestStartMaxPlayers(t *testing.T) {
	g := setupGame(MaxPlayers)
	err := g.Start()
	if err != nil {
		t.Fatalf("Start with MaxPlayers should succeed: %v", err)
	}
	if len(g.PlayerOrder) != MaxPlayers {
		t.Errorf("expected %d players in order, got %d", MaxPlayers, len(g.PlayerOrder))
	}
}

func TestStartTooManyPlayers(t *testing.T) {
	g := setupGame(MaxPlayers + 1)
	err := g.Start()
	if err == nil {
		t.Error("Start with too many players should return error")
	}
}

func TestStartPlayerOrderRandomized(t *testing.T) {
	// Run many times and check that at least once the order differs from input
	originalOrder := []string{"p0", "p1", "p2", "p3"}
	diffSeen := false
	for i := 0; i < 50; i++ {
		g := setupGame(4)
		g.Start()
		for j := range g.PlayerOrder {
			if g.PlayerOrder[j] != originalOrder[j] {
				diffSeen = true
				break
			}
		}
		if diffSeen {
			break
		}
	}
	if !diffSeen {
		t.Error("player order was never shuffled in 50 attempts")
	}
}

func TestStartDealsCards(t *testing.T) {
	g := startedGame(2)

	for _, id := range g.PlayerOrder {
		p := g.PlayerByID(id)
		if p.SandCard == nil {
			t.Errorf("player %s should have a sand card after start", id)
		}
		if p.BloodCard == nil {
			t.Errorf("player %s should have a blood card after start", id)
		}
	}
}

func TestStartExcludesEliminatedPlayers(t *testing.T) {
	g := setupGame(3)
	g.Players[2].Eliminated = true

	err := g.Start()
	if err != nil {
		t.Fatalf("Start should succeed with 2 active players: %v", err)
	}
	if len(g.PlayerOrder) != 2 {
		t.Errorf("expected 2 players in order (eliminated excluded), got %d", len(g.PlayerOrder))
	}
}

// ---------------------------------------------------------------------------
// dealRound tests
// ---------------------------------------------------------------------------

func TestDealRoundResetsPlayerState(t *testing.T) {
	g := startedGame(2)

	// Modify player state
	p := g.PlayerByID(g.PlayerOrder[0])
	p.Stood = true
	p.Invested = 3

	g.dealRound()

	if p.Stood {
		t.Error("dealRound should reset Stood")
	}
	if p.Invested != 0 {
		t.Error("dealRound should reset Invested")
	}
	if p.SandCard == nil || p.BloodCard == nil {
		t.Error("dealRound should deal new cards")
	}
}

func TestDealRoundResetsCookAndMarkdown(t *testing.T) {
	g := startedGame(2)
	g.CookTheBooksActive = true
	g.MarkdownActive = true

	g.dealRound()

	if g.CookTheBooksActive {
		t.Error("dealRound should reset CookTheBooksActive")
	}
	if g.MarkdownActive {
		t.Error("dealRound should reset MarkdownActive")
	}
}

func TestDealRoundSetsPhaseTurn(t *testing.T) {
	g := startedGame(2)
	g.Phase = PhaseRoundEnd

	g.dealRound()

	if g.Phase != PhaseTurn {
		t.Errorf("expected PhaseTurn after dealRound, got %s", g.Phase)
	}
	if g.CurrentTurn != 0 {
		t.Errorf("expected CurrentTurn=0, got %d", g.CurrentTurn)
	}
	if g.TurnInRound != 0 {
		t.Errorf("expected TurnInRound=0, got %d", g.TurnInRound)
	}
}

// ---------------------------------------------------------------------------
// ActionDraw tests
// ---------------------------------------------------------------------------

func TestActionDrawValid(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)
	oldChips := p.Chips

	err := g.ActionDraw(id, SuitSand, nil)
	if err != nil {
		t.Fatalf("ActionDraw should succeed: %v", err)
	}
	if p.Chips != oldChips-1 {
		t.Errorf("expected %d chips after draw (cost 1), got %d", oldChips-1, p.Chips)
	}
	if p.Invested != 1 {
		t.Errorf("expected 1 invested, got %d", p.Invested)
	}
}

func TestActionDrawBloodSuit(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)

	err := g.ActionDraw(id, SuitBlood, nil)
	if err != nil {
		t.Fatalf("ActionDraw blood should succeed: %v", err)
	}
	if p.BloodCard == nil {
		t.Error("BloodCard should be set after drawing blood")
	}
}

func TestActionDrawInvalidSuit(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)

	err := g.ActionDraw(id, "fire", nil)
	if err == nil {
		t.Error("ActionDraw with invalid suit should return error")
	}
}

func TestActionDrawNotYourTurn(t *testing.T) {
	g := startedGame(2)
	// Get the player who is NOT current
	otherId := g.PlayerOrder[1]
	if otherId == currentID(g) {
		otherId = g.PlayerOrder[0]
	}

	err := g.ActionDraw(otherId, SuitSand, nil)
	if err == nil {
		t.Error("ActionDraw by wrong player should return error")
	}
}

func TestActionDrawNotInTurnPhase(t *testing.T) {
	g := startedGame(2)
	g.Phase = PhaseReveal

	err := g.ActionDraw(currentID(g), SuitSand, nil)
	if err == nil {
		t.Error("ActionDraw outside turn phase should return error")
	}
}

func TestActionDrawNotEnoughChips(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)
	p.Chips = 0

	err := g.ActionDraw(id, SuitSand, nil)
	if err == nil {
		t.Error("ActionDraw with 0 chips should return error")
	}
}

func TestActionDrawWithFreeDrawToken(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)
	oldChips := p.Chips

	err := g.ActionDraw(id, SuitSand, tokenPtr(TokenFreeDraw))
	if err != nil {
		t.Fatalf("ActionDraw with free_draw token should succeed: %v", err)
	}
	if p.Chips != oldChips {
		t.Errorf("free_draw should not cost chips: expected %d, got %d", oldChips, p.Chips)
	}
	if p.HasToken(TokenFreeDraw) {
		t.Error("free_draw token should be consumed")
	}
}

func TestActionDrawWithFreeDrawTokenAlreadyUsed(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)
	p.UseToken(TokenFreeDraw) // use it first

	err := g.ActionDraw(id, SuitSand, tokenPtr(TokenFreeDraw))
	if err == nil {
		t.Error("ActionDraw with already-used free_draw should return error")
	}
}

func TestActionDrawWithRefundToken(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)
	// First invest some chips so refund has something to refund
	p.Invest(2)
	oldChips := p.Chips

	err := g.ActionDraw(id, SuitSand, tokenPtr(TokenRefund))
	if err != nil {
		t.Fatalf("ActionDraw with refund should succeed: %v", err)
	}
	// Refund gives back 2, then draw costs 1 => net +1
	if p.Chips != oldChips+2-1 {
		t.Errorf("expected %d chips (refund +2, draw -1), got %d", oldChips+2-1, p.Chips)
	}
}

func TestActionDrawEmptyDeck(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)

	// Drain the sand deck
	g.SandDeck.Cards = nil

	err := g.ActionDraw(id, SuitSand, nil)
	if err == nil {
		t.Error("ActionDraw from empty deck should return error")
	}
}

func TestActionDrawAdvancesTurn(t *testing.T) {
	g := startedGame(2)
	firstPlayer := currentID(g)

	g.ActionDraw(firstPlayer, SuitSand, nil)

	if currentID(g) == firstPlayer && g.Phase == PhaseTurn {
		t.Error("turn should advance after draw")
	}
}

// ---------------------------------------------------------------------------
// ActionStand tests
// ---------------------------------------------------------------------------

func TestActionStandValid(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)

	err := g.ActionStand(id, nil)
	if err != nil {
		t.Fatalf("ActionStand should succeed: %v", err)
	}
	if !p.Stood {
		t.Error("player should be marked as stood")
	}
}

func TestActionStandNotYourTurn(t *testing.T) {
	g := startedGame(2)
	otherId := g.PlayerOrder[1]
	if otherId == currentID(g) {
		otherId = g.PlayerOrder[0]
	}

	err := g.ActionStand(otherId, nil)
	if err == nil {
		t.Error("ActionStand by wrong player should return error")
	}
}

func TestActionStandNotInTurnPhase(t *testing.T) {
	g := startedGame(2)
	g.Phase = PhaseReveal

	err := g.ActionStand(currentID(g), nil)
	if err == nil {
		t.Error("ActionStand outside turn phase should return error")
	}
}

func TestActionStandWithToken(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)

	err := g.ActionStand(id, tokenPtr(TokenRefund))
	if err != nil {
		t.Fatalf("ActionStand with refund token should succeed: %v", err)
	}
	p := g.PlayerByID(id)
	if p.HasToken(TokenRefund) {
		t.Error("refund token should be consumed")
	}
}

func TestActionStandWithTokenNotOwned(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)

	err := g.ActionStand(id, tokenPtr(TokenMarkdown))
	if err == nil {
		t.Error("ActionStand with unowned token should return error")
	}
}

// ---------------------------------------------------------------------------
// advanceTurn tests
// ---------------------------------------------------------------------------

func TestAdvanceTurnCyclesToNextPlayer(t *testing.T) {
	g := startedGame(2)
	first := currentID(g)

	g.advanceTurn()

	if g.Phase == PhaseTurn && currentID(g) == first {
		t.Error("turn should advance to different player")
	}
}

func TestAdvanceTurnWrapsAround(t *testing.T) {
	g := startedGame(2)

	// Advance through all players to wrap
	g.advanceTurn()
	g.advanceTurn()

	if g.TurnInRound != 1 {
		t.Errorf("expected TurnInRound=1 after wrapping, got %d", g.TurnInRound)
	}
}

func TestAdvanceTurnTriggersRevealAfterThreePasses(t *testing.T) {
	g := startedGame(2)
	numPlayers := len(g.PlayerOrder)

	// 3 passes * numPlayers advances should trigger reveal
	for i := 0; i < 3*numPlayers; i++ {
		if g.Phase != PhaseTurn {
			break
		}
		g.advanceTurn()
	}

	if g.Phase != PhaseReveal {
		t.Errorf("expected PhaseReveal after 3 full passes, got %s", g.Phase)
	}
}

// ---------------------------------------------------------------------------
// allStoodThisPass tests
// ---------------------------------------------------------------------------

func TestAllStoodThisPassAllStood(t *testing.T) {
	g := startedGame(2)

	for _, id := range g.PlayerOrder {
		g.PlayerByID(id).Stood = true
	}

	if !g.allStoodThisPass() {
		t.Error("allStoodThisPass should return true when all players stood")
	}
}

func TestAllStoodThisPassNoneStood(t *testing.T) {
	g := startedGame(2)

	if g.allStoodThisPass() {
		t.Error("allStoodThisPass should return false when no one stood")
	}
}

func TestAllStoodThisPassPartial(t *testing.T) {
	g := startedGame(3)

	g.PlayerByID(g.PlayerOrder[0]).Stood = true

	if g.allStoodThisPass() {
		t.Error("allStoodThisPass should return false when only some stood")
	}
}

func TestAllStoodTriggersRevealEarly(t *testing.T) {
	g := startedGame(2)

	// Both players stand
	id0 := currentID(g)
	g.ActionStand(id0, nil)

	id1 := currentID(g)
	g.ActionStand(id1, nil)

	if g.Phase != PhaseReveal {
		t.Errorf("all players standing should trigger reveal, got phase %s", g.Phase)
	}
}

// ---------------------------------------------------------------------------
// Reveal tests
// ---------------------------------------------------------------------------

func TestRevealNotInRevealPhase(t *testing.T) {
	g := startedGame(2)
	// Phase is PhaseTurn, not PhaseReveal

	_, err := g.Reveal()
	if err == nil {
		t.Error("Reveal should fail when not in reveal phase")
	}
}

func TestRevealBasicFlow(t *testing.T) {
	g := startedGame(2)
	g.StartReveal()

	result, err := g.Reveal()
	if err != nil {
		t.Fatalf("Reveal should succeed: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.WinnerIDs) == 0 {
		t.Error("should have at least one winner")
	}
	if g.Phase != PhaseRoundEnd {
		t.Errorf("expected PhaseRoundEnd after reveal, got %s", g.Phase)
	}
}

func TestRevealSetsImpostorDiceRolls(t *testing.T) {
	g := startedGame(2)

	// Give both players impostor cards
	for _, p := range g.activePlayers() {
		impostorSand := Card{Suit: SuitSand, Kind: KindImpostor, Value: 0}
		impostorBlood := Card{Suit: SuitBlood, Kind: KindImpostor, Value: 0}
		p.SandCard = &impostorSand
		p.BloodCard = &impostorBlood
	}

	g.StartReveal()
	g.Reveal()

	for _, p := range g.activePlayers() {
		if p.SandDiceRoll < 1 || p.SandDiceRoll > 6 {
			t.Errorf("SandDiceRoll should be 1-6, got %d", p.SandDiceRoll)
		}
		if p.BloodDiceRoll < 1 || p.BloodDiceRoll > 6 {
			t.Errorf("BloodDiceRoll should be 1-6, got %d", p.BloodDiceRoll)
		}
	}
}

func TestRevealWinnerGetsChipsBack(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	// Give p0 a Pure Sabacc (best hand) and p1 a weak NoSabacc
	p0.SandCard = &Card{Suit: SuitSand, Kind: KindSylop, Value: 0}
	p0.BloodCard = &Card{Suit: SuitBlood, Kind: KindSylop, Value: 0}
	p1.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	p1.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 6}

	// Simulate investment
	p0.Invest(1)
	p1.Invest(1)

	g.StartReveal()
	result, err := g.Reveal()
	if err != nil {
		t.Fatalf("Reveal failed: %v", err)
	}

	// p0 should be the winner
	if len(result.WinnerIDs) != 1 || result.WinnerIDs[0] != "p0" {
		t.Errorf("expected p0 as sole winner, got %v", result.WinnerIDs)
	}

	// Winner gets invested chips back
	if result.ChipChanges["p0"] != 1 {
		t.Errorf("winner chip change should be +1 (invested back), got %d", result.ChipChanges["p0"])
	}
	if p0.Invested != 0 {
		t.Errorf("winner invested should be reset to 0, got %d", p0.Invested)
	}
}

func TestRevealLoserPenalty(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	// p0 gets PureSabacc, p1 gets NoSabacc with diff=5
	p0.SandCard = &Card{Suit: SuitSand, Kind: KindSylop, Value: 0}
	p0.BloodCard = &Card{Suit: SuitBlood, Kind: KindSylop, Value: 0}
	p1.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	p1.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 6}

	p0.Invest(1)
	p1.Invest(1)

	g.StartReveal()
	result, _ := g.Reveal()

	// NoSabacc penalty = value (difference) = 5, plus invested = 1, total = -6
	if result.ChipChanges["p1"] != -6 {
		t.Errorf("loser chip change should be -6 (invested 1 + penalty 5), got %d", result.ChipChanges["p1"])
	}
}

func TestRevealSabaccLoserPenaltyIsOne(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	// p0 gets Sabacc(1), p1 gets Sabacc(3) — p0 wins
	p0.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	p0.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 1}
	p1.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 3}
	p1.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 3}

	p1.Invest(1)

	g.StartReveal()
	result, _ := g.Reveal()

	// Sabacc loser penalty = 1, plus invested = 1, total = -2
	if result.ChipChanges["p1"] != -2 {
		t.Errorf("sabacc loser chip change should be -2, got %d", result.ChipChanges["p1"])
	}
}

func TestRevealEliminatesPlayerAtZeroChips(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", 1, false) // only 1 chip
	g.Players = []*Player{p0, p1}
	g.Start()

	// p0 wins with PureSabacc
	p0.SandCard = &Card{Suit: SuitSand, Kind: KindSylop, Value: 0}
	p0.BloodCard = &Card{Suit: SuitBlood, Kind: KindSylop, Value: 0}
	p1.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	p1.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 6}

	p1.Invest(1) // p1 now has 0 chips

	g.StartReveal()
	g.Reveal()

	if !p1.Eliminated {
		t.Error("player with 0 chips after penalties should be eliminated")
	}
	if p1.Chips != 0 {
		t.Errorf("eliminated player chips should be 0, got %d", p1.Chips)
	}
}

func TestRevealTiedHands(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	// Both get Sabacc(3) — tied
	p0.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 3}
	p0.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 3}
	p1.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 3}
	p1.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 3}

	g.StartReveal()
	result, _ := g.Reveal()

	if len(result.WinnerIDs) != 2 {
		t.Errorf("tied hands should produce 2 winners, got %d", len(result.WinnerIDs))
	}
}

func TestRevealSetsLastResult(t *testing.T) {
	g := startedGame(2)
	g.StartReveal()

	result, _ := g.Reveal()
	if g.LastResult != result {
		t.Error("LastResult should be set after Reveal")
	}
}

func TestRevealHandResults(t *testing.T) {
	g := startedGame(2)
	g.StartReveal()

	result, _ := g.Reveal()
	for _, id := range g.PlayerOrder {
		if _, ok := result.PlayerHands[id]; !ok {
			t.Errorf("player %s should have a hand result", id)
		}
	}
}

// ---------------------------------------------------------------------------
// CookTheBooks token test
// ---------------------------------------------------------------------------

func TestRevealCookTheBooksInvertsRanking(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	// p0 gets PureSabacc (normally the best), p1 gets NoSabacc diff=5 (normally the worst)
	p0.SandCard = &Card{Suit: SuitSand, Kind: KindSylop, Value: 0}
	p0.BloodCard = &Card{Suit: SuitBlood, Kind: KindSylop, Value: 0}
	p1.SandCard = &Card{Suit: SuitSand, Kind: KindValue, Value: 1}
	p1.BloodCard = &Card{Suit: SuitBlood, Kind: KindValue, Value: 6}

	// Activate CookTheBooks which inverts rankings
	g.CookTheBooksActive = true

	g.StartReveal()
	result, _ := g.Reveal()

	// With cook_the_books, the worst hand should win
	if len(result.WinnerIDs) != 1 || result.WinnerIDs[0] != "p1" {
		t.Errorf("cook_the_books should invert rankings, making p1 the winner, got %v", result.WinnerIDs)
	}
}

// ---------------------------------------------------------------------------
// GeneralTariff token test
// ---------------------------------------------------------------------------

func TestApplyTokenGeneralTariff(t *testing.T) {
	g := startedGame(3)
	id := currentID(g)
	p := g.PlayerByID(id)

	// Give the player the token
	p.ShiftTokens = append(p.ShiftTokens, TokenGeneralTariff)

	otherChipsBefore := map[string]int{}
	for _, oid := range g.PlayerOrder {
		if oid != id {
			otherChipsBefore[oid] = g.PlayerByID(oid).Chips
		}
	}

	err := g.ActionStand(id, tokenPtr(TokenGeneralTariff))
	if err != nil {
		t.Fatalf("ActionStand with general_tariff should succeed: %v", err)
	}

	for _, oid := range g.PlayerOrder {
		if oid != id {
			op := g.PlayerByID(oid)
			expected := otherChipsBefore[oid] - 1
			if op.Chips != expected {
				t.Errorf("general_tariff should take 1 chip from %s: expected %d, got %d", oid, expected, op.Chips)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Markdown token test
// ---------------------------------------------------------------------------

func TestApplyTokenMarkdown(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)

	p.ShiftTokens = append(p.ShiftTokens, TokenMarkdown)

	err := g.ActionStand(id, tokenPtr(TokenMarkdown))
	if err != nil {
		t.Fatalf("ActionStand with markdown should succeed: %v", err)
	}
	if !g.MarkdownActive {
		t.Error("MarkdownActive should be true after markdown token")
	}
}

// ---------------------------------------------------------------------------
// CookTheBooks token application test
// ---------------------------------------------------------------------------

func TestApplyTokenCookTheBooks(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)

	p.ShiftTokens = append(p.ShiftTokens, TokenCookTheBooks)

	err := g.ActionStand(id, tokenPtr(TokenCookTheBooks))
	if err != nil {
		t.Fatalf("ActionStand with cook_the_books should succeed: %v", err)
	}
	if !g.CookTheBooksActive {
		t.Error("CookTheBooksActive should be true after cook_the_books token")
	}
}

// ---------------------------------------------------------------------------
// Refund token application test
// ---------------------------------------------------------------------------

func TestApplyTokenRefund(t *testing.T) {
	g := startedGame(2)
	id := currentID(g)
	p := g.PlayerByID(id)

	p.Invest(2)
	chipsBefore := p.Chips

	err := g.ActionStand(id, tokenPtr(TokenRefund))
	if err != nil {
		t.Fatalf("ActionStand with refund should succeed: %v", err)
	}
	if p.Chips != chipsBefore+2 {
		t.Errorf("refund should give back 2 chips: expected %d, got %d", chipsBefore+2, p.Chips)
	}
}

// ---------------------------------------------------------------------------
// NextRound tests
// ---------------------------------------------------------------------------

func TestNextRoundNotInRoundEndPhase(t *testing.T) {
	g := startedGame(2)
	// Phase is PhaseTurn

	err := g.NextRound()
	if err == nil {
		t.Error("NextRound should fail when not in round_end phase")
	}
}

func TestNextRoundAdvancesRound(t *testing.T) {
	g := startedGame(2)
	g.StartReveal()
	g.Reveal()

	oldRound := g.Round
	err := g.NextRound()
	if err != nil {
		t.Fatalf("NextRound should succeed: %v", err)
	}
	if g.Round != oldRound+1 {
		t.Errorf("expected round %d, got %d", oldRound+1, g.Round)
	}
	if g.Phase != PhaseTurn {
		t.Errorf("expected PhaseTurn after NextRound, got %s", g.Phase)
	}
}

func TestNextRoundGameOverWhenOnePlayerLeft(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	// Eliminate p1
	p1.Eliminated = true
	g.Phase = PhaseRoundEnd

	err := g.NextRound()
	if err != nil {
		t.Fatalf("NextRound should succeed: %v", err)
	}
	if g.Phase != PhaseGameOver {
		t.Errorf("expected PhaseGameOver, got %s", g.Phase)
	}
	if g.WinnerID != "p0" {
		t.Errorf("expected winner p0, got %s", g.WinnerID)
	}
}

func TestNextRoundGameOverNoPlayersLeft(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}
	g.Start()

	p0.Eliminated = true
	p1.Eliminated = true
	g.Phase = PhaseRoundEnd

	err := g.NextRound()
	if err != nil {
		t.Fatalf("NextRound should succeed: %v", err)
	}
	if g.Phase != PhaseGameOver {
		t.Errorf("expected PhaseGameOver, got %s", g.Phase)
	}
	if g.WinnerID != "" {
		t.Errorf("expected no winner when all eliminated, got %s", g.WinnerID)
	}
}

func TestNextRoundDealsNewCards(t *testing.T) {
	g := startedGame(2)
	g.StartReveal()
	g.Reveal()
	g.NextRound()

	for _, id := range g.PlayerOrder {
		p := g.PlayerByID(id)
		if p.Eliminated {
			continue
		}
		if p.SandCard == nil {
			t.Errorf("player %s should have sand card after next round", id)
		}
		if p.BloodCard == nil {
			t.Errorf("player %s should have blood card after next round", id)
		}
	}
}

// ---------------------------------------------------------------------------
// Full game lifecycle test
// ---------------------------------------------------------------------------

func TestFullGameLifecycle(t *testing.T) {
	g := NewGame()
	p0 := NewPlayer("p0", "Alice", StartingChips, true)
	p1 := NewPlayer("p1", "Bob", StartingChips, false)
	g.Players = []*Player{p0, p1}

	// Start
	if err := g.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if g.Phase != PhaseTurn {
		t.Fatalf("expected PhaseTurn, got %s", g.Phase)
	}

	// Both players stand to trigger reveal
	first := currentID(g)
	if err := g.ActionStand(first, nil); err != nil {
		t.Fatalf("first stand failed: %v", err)
	}

	second := currentID(g)
	if err := g.ActionStand(second, nil); err != nil {
		t.Fatalf("second stand failed: %v", err)
	}

	if g.Phase != PhaseReveal {
		t.Fatalf("expected PhaseReveal after all stand, got %s", g.Phase)
	}

	// Reveal
	result, err := g.Reveal()
	if err != nil {
		t.Fatalf("Reveal failed: %v", err)
	}
	if len(result.WinnerIDs) == 0 {
		t.Fatal("should have winner(s)")
	}
	if g.Phase != PhaseRoundEnd {
		t.Fatalf("expected PhaseRoundEnd, got %s", g.Phase)
	}

	// NextRound
	if err := g.NextRound(); err != nil {
		t.Fatalf("NextRound failed: %v", err)
	}

	active := g.activePlayers()
	if len(active) >= 2 {
		if g.Phase != PhaseTurn {
			t.Errorf("expected PhaseTurn for round 2, got %s", g.Phase)
		}
		if g.Round != 2 {
			t.Errorf("expected round 2, got %d", g.Round)
		}
	}
}

// ---------------------------------------------------------------------------
// StartReveal exported method test
// ---------------------------------------------------------------------------

func TestStartRevealSetsPhase(t *testing.T) {
	g := startedGame(2)
	g.StartReveal()

	if g.Phase != PhaseReveal {
		t.Errorf("expected PhaseReveal, got %s", g.Phase)
	}
}

// ---------------------------------------------------------------------------
// currentPlayer edge case
// ---------------------------------------------------------------------------

func TestCurrentPlayerOutOfBounds(t *testing.T) {
	g := startedGame(2)
	g.CurrentTurn = 999

	if p := g.currentPlayer(); p != nil {
		t.Error("currentPlayer should return nil when CurrentTurn is out of bounds")
	}
}
