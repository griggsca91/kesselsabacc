package room

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sabacc/db"
	"sabacc/game"
	"strings"
	"sync"
	"time"
)

type IncomingMessage struct {
	Client *Client
	Data   []byte
}

type Hub struct {
	rooms        map[string]*Room
	mu           sync.RWMutex
	repo         *db.Repository // nil when running without a database
	Register     chan *Client
	Unregister   chan *Client
	Incoming     chan IncomingMessage
	matchmaker   *Matchmaker
	matchResults map[string]string // playerID -> room code
}

func NewHub(repo *db.Repository) *Hub {
	h := &Hub{
		rooms:        map[string]*Room{},
		repo:         repo,
		Register:     make(chan *Client, 16),
		Unregister:   make(chan *Client, 16),
		Incoming:     make(chan IncomingMessage, 128),
		matchResults: map[string]string{},
	}
	h.matchmaker = NewMatchmaker(h)
	go h.matchmaker.Run()
	return h
}

// GetMatchResult returns the room code for a matched player, if available.
func (h *Hub) GetMatchResult(playerID string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	code, ok := h.matchResults[playerID]
	return code, ok
}

// ClearMatchResult removes the match result for a player after they've consumed it.
func (h *Hub) ClearMatchResult(playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.matchResults, playerID)
}

// Matchmaker returns the hub's matchmaker (for use by API handlers).
func (h *Hub) Matchmaker() *Matchmaker {
	return h.matchmaker
}

func (h *Hub) createMatchedRoom(players []QueueEntry) {
	if len(players) < 2 {
		return
	}
	code, err := h.CreateRoom(players[0].PlayerID, players[0].PlayerName, false)
	if err != nil {
		slog.Error("matchmaker: failed to create room", "error", err)
		return
	}
	for _, p := range players[1:] {
		if err := h.JoinRoom(code, p.PlayerID, p.PlayerName); err != nil {
			slog.Error("matchmaker: failed to join player", "playerID", p.PlayerID, "error", err)
		}
	}
	// Store the code so clients can poll for it
	h.mu.Lock()
	for _, p := range players {
		h.matchResults[p.PlayerID] = code
	}
	h.mu.Unlock()
	slog.Info("matchmaker: created room", "code", code, "players", len(players))
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.RLock()
			room, ok := h.rooms[client.RoomCode]
			h.mu.RUnlock()
			if ok {
				if client.IsSpectator {
					room.AddSpectator(client)
				} else {
					room.AddClient(client)
				}
				h.broadcastState(room)
			}

		case client := <-h.Unregister:
			h.mu.RLock()
			room, ok := h.rooms[client.RoomCode]
			h.mu.RUnlock()
			if ok {
				if client.IsSpectator {
					room.RemoveSpectator(client.PlayerID)
				} else {
					room.RemoveClient(client.PlayerID)
				}
				close(client.send)
				h.broadcastState(room)
			}

		case msg := <-h.Incoming:
			h.handleMessage(msg.Client, msg.Data)
		}
	}
}

// CreateRoom creates a new room and adds the host player.
func (h *Hub) CreateRoom(playerID, playerName string, isPublic bool) (string, error) {
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
	r.IsPublic = isPublic
	r.Game.Players = append(r.Game.Players, game.NewPlayer(playerID, playerName, game.StartingChips, true))
	h.rooms[code] = r
	return code, nil
}

// PublicRoomInfo is the data exposed by the room browser endpoint.
type PublicRoomInfo struct {
	Code        string `json:"code"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
}

// ListPublicRooms returns rooms that are public and still in the lobby phase.
func (h *Hub) ListPublicRooms() []PublicRoomInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var list []PublicRoomInfo
	for _, r := range h.rooms {
		r.mu.RLock()
		isPublic := r.IsPublic
		phase := r.Game.Phase
		playerCount := len(r.Game.Players)
		r.mu.RUnlock()

		if isPublic && phase == game.PhaseLobby {
			list = append(list, PublicRoomInfo{
				Code:        r.Code,
				PlayerCount: playerCount,
				MaxPlayers:  game.MaxPlayers,
			})
		}
	}
	if list == nil {
		list = []PublicRoomInfo{}
	}
	return list
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

// JoinAsSpectator registers a spectator in an existing room without adding them as a player.
func (h *Hub) JoinAsSpectator(code, playerID, playerName string) error {
	h.mu.RLock()
	_, ok := h.rooms[code]
	h.mu.RUnlock()
	if !ok {
		return errors.New("room not found")
	}
	// Spectators are allowed regardless of game phase — store the name for display purposes
	// but do NOT add them to the game's player list.
	_ = playerName
	return nil
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

	room.LastActive = time.Now()

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
	case "chat":
		err = h.handleChat(c, room, action.Payload)
		if err != nil {
			h.sendError(c, err.Error())
		}
		return // chat does not trigger a game state broadcast
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

// --- Chat ---

type ChatPayload struct {
	Text string `json:"text"`
}

type ChatMessage struct {
	Type       string `json:"type"` // "chat"
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Text       string `json:"text"`
	Timestamp  int64  `json:"timestamp"` // unix ms
}

func (h *Hub) handleChat(c *Client, room *Room, payload json.RawMessage) error {
	var p ChatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return errors.New("invalid chat payload")
	}

	text := strings.TrimSpace(p.Text)
	if text == "" {
		return errors.New("empty message")
	}
	if len([]rune(text)) > 200 {
		runes := []rune(text)
		text = string(runes[:200])
	}

	// Rate limiting: max 3 messages per 5 seconds per player
	now := time.Now().UnixMilli()
	windowStart := now - 5000
	timestamps := room.ChatTimestamps[c.PlayerID]
	// Filter to only timestamps within the window
	filtered := timestamps[:0]
	for _, ts := range timestamps {
		if ts >= windowStart {
			filtered = append(filtered, ts)
		}
	}
	if len(filtered) >= 3 {
		return errors.New("sending too fast — slow down")
	}
	room.ChatTimestamps[c.PlayerID] = append(filtered, now)

	// Find player name
	player := room.Game.PlayerByID(c.PlayerID)
	name := c.PlayerID // fallback
	if player != nil {
		name = player.Name
	}

	msg := ChatMessage{
		Type:       "chat",
		PlayerID:   c.PlayerID,
		PlayerName: name,
		Text:       text,
		Timestamp:  now,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return errors.New("internal error")
	}

	// Broadcast to all clients and spectators in the room (lock already held as read lock for clients map access via Broadcast)
	// We hold room.mu.Lock() here, so we must send directly to avoid deadlock with Broadcast's RLock.
	for _, client := range room.Clients {
		client.Send(data)
	}
	for _, client := range room.Spectators {
		client.Send(data)
	}
	return nil
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
	SpectatorCount      int               `json:"spectatorCount"`
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

	spectatorCount := len(room.Spectators)

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
			SpectatorCount:      spectatorCount,
		}

		msg, err := json.Marshal(Envelope{Type: "game_state", Payload: view})
		if err != nil {
			slog.Error("marshal error", "error", err)
			continue
		}
		client.Send(msg)
	}

	// Spectator view — same state but no hand info (spectators never see hands)
	spectatorView := GameStateView{
		Phase:               g.Phase,
		Round:               g.Round,
		TurnInRound:         g.TurnInRound,
		CurrentTurnPlayerID: currentTurnID,
		Players:             playerViews,
		YourHand:            nil,
		LastResult:          g.LastResult,
		WinnerID:            g.WinnerID,
		SandRemaining:       g.SandDeck.Remaining(),
		BloodRemaining:      g.BloodDeck.Remaining(),
		SpectatorCount:      spectatorCount,
	}
	spectatorMsg, sErr := json.Marshal(Envelope{Type: "game_state", Payload: spectatorView})
	if sErr != nil {
		slog.Error("spectator marshal error", "error", sErr)
	} else {
		for _, sc := range room.Spectators {
			sc.Send(spectatorMsg)
		}
	}
}

// persistGameResult saves the completed game and player results to the database.
func (h *Hub) persistGameResult(room *Room) {
	g := room.Game
	ctx := context.Background()

	// Create the game record
	gameID, err := h.repo.SaveGameState(ctx, room.Code, g.Round, string(g.Phase), nil)
	if err != nil {
		slog.Error("failed to save game state", "error", err)
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
		slog.Error("failed to record game result", "error", err)
	}
}

func (h *Hub) sendError(c *Client, msg string) {
	data, _ := json.Marshal(Envelope{Type: "error", Payload: map[string]string{"message": msg}})
	c.Send(data)
}
