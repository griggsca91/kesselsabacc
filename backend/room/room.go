package room

import (
	"math/rand"
	"sabacc/game"
	"sync"
	"time"
)

const codeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateCode() string {
	b := make([]byte, 4)
	for i := range b {
		b[i] = codeChars[rand.Intn(len(codeChars))]
	}
	return string(b)
}

type Room struct {
	Code           string
	Game           *game.Game
	Clients        map[string]*Client // playerID -> client (players only)
	Spectators     map[string]*Client // playerID -> client (spectators only)
	ResultSaved    bool               // true once game_over results have been persisted
	IsPublic       bool
	ChatTimestamps map[string][]int64 // playerID -> recent message unix-ms timestamps
	LastActive     time.Time          // updated on any client activity; used by cleanup
	mu             sync.RWMutex
}

func NewRoom() *Room {
	return &Room{
		Game:           game.NewGame(),
		Clients:        map[string]*Client{},
		Spectators:     map[string]*Client{},
		ChatTimestamps: map[string][]int64{},
		LastActive:     time.Now(),
	}
}

func (r *Room) AddClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[c.PlayerID] = c
	r.LastActive = time.Now()
}

func (r *Room) RemoveClient(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, playerID)
	r.LastActive = time.Now()
}

func (r *Room) AddSpectator(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Spectators[c.PlayerID] = c
	r.LastActive = time.Now()
}

func (r *Room) RemoveSpectator(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Spectators, playerID)
	r.LastActive = time.Now()
}

func (r *Room) Broadcast(msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.Clients {
		c.Send(msg)
	}
	for _, c := range r.Spectators {
		c.Send(msg)
	}
}

func (r *Room) SendTo(playerID string, msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.Clients[playerID]; ok {
		c.Send(msg)
	}
}
