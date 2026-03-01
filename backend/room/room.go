package room

import (
	"math/rand"
	"sabacc/game"
	"sync"
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
	Code    string
	Game    *game.Game
	Clients map[string]*Client // playerID -> client
	mu      sync.RWMutex
}

func NewRoom() *Room {
	return &Room{
		Game:    game.NewGame(),
		Clients: map[string]*Client{},
	}
}

func (r *Room) AddClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[c.PlayerID] = c
}

func (r *Room) RemoveClient(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, playerID)
}

func (r *Room) Broadcast(msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.Clients {
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
