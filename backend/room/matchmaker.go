package room

import (
	"sabacc/game"
	"sync"
	"time"
)

const (
	MatchWaitDuration = 10 * time.Second
	MinPlayersToMatch = 2
	MaxPlayersToMatch = game.MaxPlayers // 4
)

type QueueEntry struct {
	PlayerID   string
	PlayerName string
	JoinedAt   time.Time
}

type Matchmaker struct {
	hub   *Hub
	queue []QueueEntry
	mu    sync.Mutex
	tick  *time.Ticker
}

func NewMatchmaker(hub *Hub) *Matchmaker {
	return &Matchmaker{
		hub:  hub,
		tick: time.NewTicker(2 * time.Second),
	}
}

// Enqueue adds a player to the matchmaking queue.
// Returns false if already queued.
func (m *Matchmaker) Enqueue(playerID, playerName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, e := range m.queue {
		if e.PlayerID == playerID {
			return false
		}
	}

	m.queue = append(m.queue, QueueEntry{
		PlayerID:   playerID,
		PlayerName: playerName,
		JoinedAt:   time.Now(),
	})
	return true
}

// Dequeue removes a player from the queue (cancel).
func (m *Matchmaker) Dequeue(playerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, e := range m.queue {
		if e.PlayerID == playerID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return
		}
	}
}

// IsQueued reports whether a player is in the queue.
func (m *Matchmaker) IsQueued(playerID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, e := range m.queue {
		if e.PlayerID == playerID {
			return true
		}
	}
	return false
}

// Run processes the queue every 2 seconds.
// If MinPlayersToMatch players are in the queue AND the oldest has waited MatchWaitDuration,
// OR if MaxPlayersToMatch players are queued, create a room and add them all.
func (m *Matchmaker) Run() {
	for range m.tick.C {
		m.tryMatch()
	}
}

func (m *Matchmaker) tryMatch() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) < MinPlayersToMatch {
		return
	}

	oldest := m.queue[0]
	shouldMatch := len(m.queue) >= MaxPlayersToMatch ||
		time.Since(oldest.JoinedAt) >= MatchWaitDuration

	if !shouldMatch {
		return
	}

	// Take up to MaxPlayersToMatch players
	take := len(m.queue)
	if take > MaxPlayersToMatch {
		take = MaxPlayersToMatch
	}
	players := m.queue[:take]
	m.queue = m.queue[take:]

	// Create room with first player as host
	go m.hub.createMatchedRoom(players)
}
