package room

import (
	"log/slog"
	"sabacc/game"
	"time"
)

const (
	// cleanupInterval is how often the cleanup goroutine runs.
	cleanupInterval = 2 * time.Minute
	// emptyRoomTimeout: remove a room if all clients disconnected this long ago.
	emptyRoomTimeout = 5 * time.Minute
	// idleLobbyTimeout: remove a lobby room with no activity for this long.
	idleLobbyTimeout = 30 * time.Minute
)

// StartCleanup starts a background goroutine that periodically removes stale rooms.
func (h *Hub) StartCleanup() {
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			h.cleanupStaleRooms()
		}
	}()
}

func (h *Hub) cleanupStaleRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	removed := 0

	for code, r := range h.rooms {
		r.mu.RLock()
		clientCount := len(r.Clients) + len(r.Spectators)
		lastActive := r.LastActive
		phase := r.Game.Phase
		r.mu.RUnlock()

		shouldRemove := false

		if clientCount == 0 && now.Sub(lastActive) > emptyRoomTimeout {
			// All clients gone for too long
			shouldRemove = true
		} else if phase == game.PhaseLobby && now.Sub(lastActive) > idleLobbyTimeout {
			// Lobby with no activity for 30 minutes
			shouldRemove = true
		}

		if shouldRemove {
			delete(h.rooms, code)
			removed++
			slog.Info("cleanup: removed stale room",
				"code", code,
				"phase", string(phase),
				"clients", clientCount,
				"idle_minutes", int(now.Sub(lastActive).Minutes()),
			)
		}
	}

	if removed > 0 {
		slog.Info("cleanup: finished", "rooms_removed", removed, "rooms_remaining", len(h.rooms))
	}
}
