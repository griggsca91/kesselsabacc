package game

import (
	"errors"
	"math/rand"
)

type Phase string

const (
	PhaseLobby    Phase = "lobby"
	PhaseDealing  Phase = "dealing"
	PhaseTurn     Phase = "turn"
	PhaseReveal   Phase = "reveal"
	PhaseRoundEnd Phase = "round_end"
	PhaseGameOver Phase = "game_over"
)

const StartingChips = 6
const MaxPlayers = 4

type RoundResult struct {
	WinnerIDs    []string        `json:"winnerIds"`
	PlayerHands  map[string]HandResult `json:"playerHands"`
	ChipChanges  map[string]int  `json:"chipChanges"` // negative = lost chips
}

type Game struct {
	Phase         Phase     `json:"phase"`
	Players       []*Player `json:"players"`
	PlayerOrder   []string  `json:"playerOrder"` // IDs in turn order
	CurrentTurn   int       `json:"currentTurn"` // index into PlayerOrder
	Round         int       `json:"round"`
	TurnInRound   int       `json:"turnInRound"` // 0-2 (3 turns per round)
	SandDeck      Deck      `json:"-"`
	BloodDeck     Deck     `json:"-"`
	LastResult    *RoundResult `json:"lastResult"`
	WinnerID      string    `json:"winnerId"`
	// CookTheBooksActive inverts Sabacc rankings this round
	CookTheBooksActive bool `json:"cookTheBooksActive"`
	// MarkdownActive sets sylop value to 0 this round
	MarkdownActive bool `json:"markdownActive"`
}

func NewGame() *Game {
	return &Game{
		Phase:   PhaseLobby,
		Players: []*Player{},
	}
}

func (g *Game) PlayerByID(id string) *Player {
	for _, p := range g.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) activePlayers() []*Player {
	var out []*Player
	for _, p := range g.Players {
		if !p.Eliminated {
			out = append(out, p)
		}
	}
	return out
}

func (g *Game) Start() error {
	active := g.activePlayers()
	if len(active) < 2 {
		return errors.New("need at least 2 players to start")
	}
	if len(active) > MaxPlayers {
		return errors.New("too many players")
	}

	g.PlayerOrder = make([]string, len(active))
	for i, p := range active {
		g.PlayerOrder[i] = p.ID
	}
	rand.Shuffle(len(g.PlayerOrder), func(i, j int) {
		g.PlayerOrder[i], g.PlayerOrder[j] = g.PlayerOrder[j], g.PlayerOrder[i]
	})

	g.Round = 1
	g.dealRound()
	return nil
}

func (g *Game) dealRound() {
	g.SandDeck, g.BloodDeck = NewDecks()
	g.Phase = PhaseTurn
	g.TurnInRound = 0
	g.CurrentTurn = 0
	g.CookTheBooksActive = false
	g.MarkdownActive = false

	for _, p := range g.activePlayers() {
		p.ResetRound()
		sand, _ := g.SandDeck.Draw()
		blood, _ := g.BloodDeck.Draw()
		p.SandCard = &sand
		p.BloodCard = &blood
	}
}

func (g *Game) currentPlayer() *Player {
	if g.CurrentTurn >= len(g.PlayerOrder) {
		return nil
	}
	return g.PlayerByID(g.PlayerOrder[g.CurrentTurn])
}

// ActionDraw handles a player drawing a new card.
func (g *Game) ActionDraw(playerID string, suit CardSuit, tokenUsed *ShiftToken) error {
	if g.Phase != PhaseTurn {
		return errors.New("not in turn phase")
	}
	cp := g.currentPlayer()
	if cp == nil || cp.ID != playerID {
		return errors.New("not your turn")
	}

	freeDraw := false
	if tokenUsed != nil {
		if *tokenUsed == TokenFreeDraw {
			if !cp.UseToken(TokenFreeDraw) {
				return errors.New("you don't have that token")
			}
			freeDraw = true
		} else {
			if err := g.applyToken(cp, *tokenUsed); err != nil {
				return err
			}
		}
	}

	if !freeDraw {
		if !cp.Invest(1) {
			return errors.New("not enough chips")
		}
	}

	switch suit {
	case SuitSand:
		card, ok := g.SandDeck.Draw()
		if !ok {
			return errors.New("sand deck is empty")
		}
		cp.SandCard = &card
	case SuitBlood:
		card, ok := g.BloodDeck.Draw()
		if !ok {
			return errors.New("blood deck is empty")
		}
		cp.BloodCard = &card
	default:
		return errors.New("invalid suit")
	}

	g.advanceTurn()
	return nil
}

// ActionStand handles a player standing.
func (g *Game) ActionStand(playerID string, tokenUsed *ShiftToken) error {
	if g.Phase != PhaseTurn {
		return errors.New("not in turn phase")
	}
	cp := g.currentPlayer()
	if cp == nil || cp.ID != playerID {
		return errors.New("not your turn")
	}

	if tokenUsed != nil {
		if err := g.applyToken(cp, *tokenUsed); err != nil {
			return err
		}
	}

	cp.Stood = true
	g.advanceTurn()
	return nil
}

func (g *Game) applyToken(player *Player, token ShiftToken) error {
	if !player.UseToken(token) {
		return errors.New("you don't have that token")
	}
	switch token {
	case TokenRefund:
		player.Refund(2)
	case TokenGeneralTariff:
		for _, p := range g.activePlayers() {
			if p.ID != player.ID {
				p.Invest(1) // forces other players to pay 1
			}
		}
	case TokenMarkdown:
		g.MarkdownActive = true
	case TokenMajorFraud:
		// Sets impostor value to 6 — handled at reveal
	case TokenCookTheBooks:
		g.CookTheBooksActive = true
	case TokenDirectTransaction:
		// Requires target player — handled separately
	case TokenImmunity:
		// Passive — handled at token application time
	}
	return nil
}

func (g *Game) advanceTurn() {
	g.CurrentTurn++

	// Wrapped around to next turn-in-round
	if g.CurrentTurn >= len(g.PlayerOrder) {
		g.CurrentTurn = 0
		g.TurnInRound++
	}

	// Check if round ends: 3 turns completed or all players stood in the same pass
	if g.TurnInRound >= 3 || g.allStoodThisPass() {
		g.startReveal()
		return
	}
}

func (g *Game) allStoodThisPass() bool {
	// Check if every active player stood during the current pass
	for _, id := range g.PlayerOrder {
		p := g.PlayerByID(id)
		if p != nil && !p.Eliminated && !p.Stood {
			return false
		}
	}
	return true
}

// ActionResolveImpostor is called at reveal time for each player with an impostor.
// The dice values are provided by the client but validated server-side (we re-roll).
func (g *Game) Reveal() (*RoundResult, error) {
	if g.Phase != PhaseReveal {
		return nil, errors.New("not in reveal phase")
	}

	active := g.activePlayers()

	// Roll dice for any impostor cards
	for _, p := range active {
		if p.SandCard != nil && p.SandCard.Kind == KindImpostor {
			p.SandDiceRoll = rand.Intn(6) + 1
		}
		if p.BloodCard != nil && p.BloodCard.Kind == KindImpostor {
			p.BloodDiceRoll = rand.Intn(6) + 1
		}
	}

	// Compute hands
	hands := map[string]HandResult{}
	for _, p := range active {
		if p.SandCard == nil || p.BloodCard == nil {
			continue
		}
		sv := EffectiveValue(*p.SandCard, *p.BloodCard, p.SandDiceRoll)
		bv := EffectiveValue(*p.BloodCard, *p.SandCard, p.BloodDiceRoll)
		hands[p.ID] = ResolveHand(*p.SandCard, *p.BloodCard, sv, bv)
	}

	// Find winner(s)
	var bestHand *HandResult
	winnerIDs := []string{}
	for _, p := range active {
		h, ok := hands[p.ID]
		if !ok {
			continue
		}
		if bestHand == nil {
			bestHand = &h
			winnerIDs = []string{p.ID}
		} else {
			cmp := CompareHands(h, *bestHand)
			if g.CookTheBooksActive {
				cmp = -cmp // invert for cook the books
			}
			if cmp < 0 {
				bestHand = &h
				winnerIDs = []string{p.ID}
			} else if cmp == 0 {
				winnerIDs = append(winnerIDs, p.ID)
			}
		}
	}

	// Compute chip changes
	chipChanges := map[string]int{}
	isWinner := func(id string) bool {
		for _, w := range winnerIDs {
			if w == id {
				return true
			}
		}
		return false
	}

	for _, p := range active {
		if isWinner(p.ID) {
			// Win back invested chips
			chipChanges[p.ID] = p.Invested
			p.Chips += p.Invested
			p.Invested = 0
		} else {
			h := hands[p.ID]
			penalty := 0
			switch h.Rank {
			case RankPureSabacc, RankSabacc:
				penalty = 1
			case RankNoSabacc:
				penalty = h.Value
			}
			total := p.Invested + penalty
			chipChanges[p.ID] = -total
			p.Chips -= penalty
			p.Invested = 0
			if p.Chips <= 0 {
				p.Chips = 0
				p.Eliminated = true
			}
		}
	}

	result := &RoundResult{
		WinnerIDs:   winnerIDs,
		PlayerHands: hands,
		ChipChanges: chipChanges,
	}
	g.LastResult = result
	g.Phase = PhaseRoundEnd
	return result, nil
}

func (g *Game) StartReveal() {
	g.startReveal()
}

func (g *Game) startReveal() {
	g.Phase = PhaseReveal
}

// NextRound advances to the next round or ends the game.
func (g *Game) NextRound() error {
	if g.Phase != PhaseRoundEnd {
		return errors.New("not in round end phase")
	}

	active := g.activePlayers()
	if len(active) <= 1 {
		g.Phase = PhaseGameOver
		if len(active) == 1 {
			g.WinnerID = active[0].ID
		}
		return nil
	}

	g.Round++
	g.dealRound()
	return nil
}
