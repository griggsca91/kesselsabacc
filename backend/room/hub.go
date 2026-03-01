package room

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sabacc/db"
	"sabacc/game"
	"sync"
)

type IncomingMessage struct {
	Client *Client
	Data   []byte
}

type Hub struct {
	rooms      map[string]*Room
	mu         sync.RWMutex
	repo       *db.Repository // nil when running without a database
	Register   chan *Client
	Unregister chan *Client
	Incoming   chan IncomingMessage
}

func NewHub(repo *db.Repository) *Hub {
	return &Hub{
		rooms:      map[string]*Room{},
		repo:       repo,
		Register:   make(chan *Client, 16),
		Unregister: make(chan *Client, 16),
		Incoming:   make(chan IncomingMessage, 128),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.RLock()
			room, ok := h.rooms[client.RoomCode]
			h.mu.RUnlock()
			if ok {
				room.AddClient(client)
				h.broadcastState(room)
			}

		case client := <-h.Unregister:
			h.mu.RLock()
			room, ok := h.rooms[client.RoomCode]
			h.mu.RUnlock()
			if ok {
				room.RemoveClient(client.PlayerID)
				close(client.send)
				h.broadcastState(room)
			}

		case msg := <-h.Incoming:
			h.handleMessage(msg.Client, msg.Data)
		}
	}
}

// CreateRoom creates a new room and adds the host player.
func (h *Hub) CreateRoom(playerID, playerName string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var code string
	for {
		code = generateCode()
		if _, exists := h.rooms[code]; !exists {
			break
		}
	}

	r := NewRoom()
	r.Code = code
	r.Game.Players = append(r.Game.Players, game.NewPlayer(playerID, playerName, game.StartingChips, true))
	h.rooms[code] = r
	return code, nil
}

// JoinRoom adds a player to an existing room.
func (h *Hub) JoinRoom(code, playerID, playerName string) error {
	h.mu.RLock()
	r, ok := h.rooms[code]
	h.mu.RUnlock()
	if !ok {
		return errors.New("room not found")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.Phase != game.PhaseLobby {
		return errors.New("game already in progress")
	}
	if len(r.Game.Players) >= game.MaxPlayers {
		return errors.New("room is full")
	}
	for _, p := range r.Game.Players {
		if p.ID == playerID {
			return nil // already in room (reconnect)
		}
	}
	r.Game.Players = append(r.Game.Players, game.NewPlayer(playerID, playerName, game.StartingChips, false))
	return nil
}

func (h *Hub) GetRoom(code string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.rooms[code]
	return r, ok
}

// --- Message handling ---

type Action struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Envelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func (h *Hub) handleMessage(c *Client, data []byte) {
	var action Action
	if err := json.Unmarshal(data, &action); err != nil {
		h.sendError(c, "invalid message format")
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[c.RoomCode]
	h.mu.RUnlock()
	if !ok {
		h.sendError(c, "room not found")
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	var err error
	switch action.Type {
	case "start_game":
		err = h.handleStartGame(c, room)
	case "draw":
		err = h.handleDraw(c, room, action.Payload)
	case "stand":
		err = h.handleStand(c, room, action.Payload)
	case "next_round":
		err = h.handleNextRound(c, room)
	default:
		err = errors.New("unknown action: " + action.Type)
	}

	if err != nil {
		h.sendError(c, err.Error())
		return
	}

	h.broadcastStateUnlocked(room)
}

func (h *Hub) handleStartGame(c *Client, room *Room) error {
	player := room.Game.PlayerByID(c.PlayerID)
	if player == nil || !player.IsHost {
		return errors.New("only the host can start the game")
	}
	return room.Game.Start()
}

type DrawPayload struct {
	Suit      game.CardSuit    `json:"suit"`
	TokenUsed *game.ShiftToken `json:"tokenUsed,omitempty"`
}

func (h *Hub) handleDraw(c *Client, room *Room, payload json.RawMessage) error {
	var p DrawPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return errors.New("invalid draw payload")
	}
	return room.Game.ActionDraw(c.PlayerID, p.Suit, p.TokenUsed)
}

type StandPayload struct {
	TokenUsed *game.ShiftToken `json:"tokenUsed,omitempty"`
}

func (h *Hub) handleStand(c *Client, room *Room, payload json.RawMessage) error {
	var p StandPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return errors.New("invalid stand payload")
	}
	// After all players act in a turn pass, auto-reveal if needed
	err := room.Game.ActionStand(c.PlayerID, p.TokenUsed)
	if err != nil {
		return err
	}
	if room.Game.Phase == game.PhaseReveal {
		_, err = room.Game.Reveal()
	}
	return err
}

func (h *Hub) handleNextRound(c *Client, room *Room) error {
	// Trigger reveal if still needed
	if room.Game.Phase == game.PhaseReveal {
		if _, err := room.Game.Reveal(); err != nil {
			return err
		}
	}
	return room.Game.NextRound()
}

// --- State broadcasting ---

// GameStateView is the sanitized game state sent to each player.
// Other players' cards are hidden unless it's the reveal phase.
type GameStateView struct {
	Phase               game.Phase        `json:"phase"`
	Round               int               `json:"round"`
	TurnInRound         int               `json:"turnInRound"`
	CurrentTurnPlayerID string            `json:"currentTurnPlayerId"`
	Players             []PlayerView      `json:"players"`
	YourHand            *HandView         `json:"yourHand"`
	LastResult          *game.RoundResult `json:"lastResult"`
	WinnerID            string            `json:"winnerId"`
	SandRemaining       int               `json:"sandRemaining"`
	BloodRemaining      int               `json:"bloodRemaining"`
}

type PlayerView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Chips      int    `json:"chips"`
	Invested   int    `json:"invested"`
	IsHost     bool   `json:"isHost"`
	Eliminated bool   `json:"eliminated"`
	TokensLeft int    `json:"tokensLeft"`
	Stood      bool   `json:"stood"`
	// Cards only visible during reveal
	SandCard  *game.Card `json:"sandCard,omitempty"`
	BloodCard *game.Card `json:"bloodCard,omitempty"`
}

type HandView struct {
	SandCard  *game.Card        `json:"sandCard"`
	BloodCard *game.Card        `json:"bloodCard"`
	Tokens    []game.ShiftToken `json:"tokens"`
}

func (h *Hub) broadcastState(room *Room) {
	room.mu.RLock()
	defer room.mu.RUnlock()
	h.broadcastStateUnlocked(room)
}

func (h *Hub) broadcastStateUnlocked(room *Room) {
	g := room.Game

	// Persist game results when the game ends
	if g.Phase == game.PhaseGameOver && !room.ResultSaved && h.repo != nil {
		h.persistGameResult(room)
		room.ResultSaved = true
	}

	isReveal := g.Phase == game.PhaseReveal || g.Phase == game.PhaseRoundEnd || g.Phase == game.PhaseGameOver

	currentTurnID := ""
	if g.Phase == game.PhaseTurn && g.CurrentTurn < len(g.PlayerOrder) {
		currentTurnID = g.PlayerOrder[g.CurrentTurn]
	}

	playerViews := make([]PlayerView, len(g.Players))
	for i, p := range g.Players {
		pv := PlayerView{
			ID:         p.ID,
			Name:       p.Name,
			Chips:      p.Chips,
			Invested:   p.Invested,
			IsHost:     p.IsHost,
			Eliminated: p.Eliminated,
			TokensLeft: len(p.ShiftTokens),
			Stood:      p.Stood,
		}
		if isReveal {
			pv.SandCard = p.SandCard
			pv.BloodCard = p.BloodCard
		}
		playerViews[i] = pv
	}

	for playerID, client := range room.Clients {
		player := g.PlayerByID(playerID)
		var hand *HandView
		if player != nil && player.SandCard != nil {
			hand = &HandView{
				SandCard:  player.SandCard,
				BloodCard: player.BloodCard,
				Tokens:    player.ShiftTokens,
			}
		}

		view := GameStateView{
			Phase:               g.Phase,
			Round:               g.Round,
			TurnInRound:         g.TurnInRound,
			CurrentTurnPlayerID: currentTurnID,
			Players:             playerViews,
			YourHand:            hand,
			LastResult:          g.LastResult,
			WinnerID:            g.WinnerID,
			SandRemaining:       g.SandDeck.Remaining(),
			BloodRemaining:      g.BloodDeck.Remaining(),
		}

		msg, err := json.Marshal(Envelope{Type: "game_state", Payload: view})
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}
		client.Send(msg)
	}
}

// persistGameResult saves the completed game and player results to the database.
func (h *Hub) persistGameResult(room *Room) {
	g := room.Game
	ctx := context.Background()

	// Create the game record
	gameID, err := h.repo.SaveGameState(ctx, room.Code, g.Round, string(g.Phase), nil)
	if err != nil {
		log.Printf("failed to save game state: %v", err)
		return
	}

	// Build player results
	results := make([]db.PlayerResult, 0, len(g.Players))
	for _, p := range g.Players {
		pr := db.PlayerResult{
			UserID:     p.ID,
			FinalChips: p.Chips,
			IsWinner:   p.ID == g.WinnerID,
		}
		// Store the hand rank from the last round's result if available
		if g.LastResult != nil {
			if hand, ok := g.LastResult.PlayerHands[p.ID]; ok {
				rankName := game.HandRankName(hand.Rank)
				pr.HandRank = &rankName
			}
		}
		results = append(results, pr)
	}

	if err := h.repo.RecordGameResult(ctx, gameID, results); err != nil {
		log.Printf("failed to record game result: %v", err)
	}
}

func (h *Hub) sendError(c *Client, msg string) {
	data, _ := json.Marshal(Envelope{Type: "error", Payload: map[string]string{"message": msg}})
	c.Send(data)
}
