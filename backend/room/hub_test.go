package room

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"sabacc/game"
)

// --- Helper: create a hub + start Run loop ---

func newTestHub(t *testing.T) *Hub {
	t.Helper()
	h := NewHub()
	go h.Run()
	return h
}

// readMessage reads a JSON message from the client's send channel with a timeout.
func readMessage(t *testing.T, c *Client, timeout time.Duration) Envelope {
	t.Helper()
	select {
	case data := <-c.send:
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatalf("unmarshal: %v (raw: %s)", err, data)
		}
		return env
	case <-time.After(timeout):
		t.Fatal("timed out waiting for message")
		return Envelope{}
	}
}

// drainMessages reads and discards all pending messages from the client's send channel.
func drainMessages(c *Client) {
	for {
		select {
		case <-c.send:
		default:
			return
		}
	}
}

// --- CreateRoom tests ---

func TestCreateRoomReturnsValidCode(t *testing.T) {
	h := NewHub()
	code, err := h.CreateRoom("host1", "Host")
	if err != nil {
		t.Fatalf("CreateRoom error: %v", err)
	}
	if len(code) != 4 {
		t.Errorf("expected 4-char code, got %q", code)
	}
	for _, ch := range code {
		if !strings.ContainsRune(codeChars, ch) {
			t.Errorf("code %q contains invalid char %q", code, ch)
		}
	}
}

func TestCreateRoomAddsPlayerAsHost(t *testing.T) {
	h := NewHub()
	code, err := h.CreateRoom("host1", "HostName")
	if err != nil {
		t.Fatalf("CreateRoom error: %v", err)
	}

	room, ok := h.GetRoom(code)
	if !ok {
		t.Fatal("room not found after creation")
	}

	if len(room.Game.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(room.Game.Players))
	}

	p := room.Game.Players[0]
	if p.ID != "host1" {
		t.Errorf("player ID = %q, want %q", p.ID, "host1")
	}
	if p.Name != "HostName" {
		t.Errorf("player Name = %q, want %q", p.Name, "HostName")
	}
	if !p.IsHost {
		t.Error("player should be host")
	}
	if p.Chips != game.StartingChips {
		t.Errorf("player chips = %d, want %d", p.Chips, game.StartingChips)
	}
}

func TestCreateRoomUniqueCode(t *testing.T) {
	h := NewHub()
	codes := map[string]bool{}
	for i := 0; i < 20; i++ {
		code, err := h.CreateRoom("p"+string(rune('A'+i)), "Player")
		if err != nil {
			t.Fatalf("CreateRoom error: %v", err)
		}
		if codes[code] {
			t.Errorf("duplicate code %q", code)
		}
		codes[code] = true
	}
}

// --- JoinRoom tests ---

func TestJoinRoomSuccess(t *testing.T) {
	h := NewHub()
	code, _ := h.CreateRoom("host1", "Host")

	err := h.JoinRoom(code, "p2", "Player2")
	if err != nil {
		t.Fatalf("JoinRoom error: %v", err)
	}

	room, _ := h.GetRoom(code)
	if len(room.Game.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(room.Game.Players))
	}

	p := room.Game.PlayerByID("p2")
	if p == nil {
		t.Fatal("player p2 not found")
	}
	if p.IsHost {
		t.Error("joined player should not be host")
	}
}

func TestJoinRoomNotFound(t *testing.T) {
	h := NewHub()
	err := h.JoinRoom("ZZZZ", "p1", "Player")
	if err == nil {
		t.Fatal("expected error for nonexistent room")
	}
	if err.Error() != "room not found" {
		t.Errorf("error = %q, want %q", err.Error(), "room not found")
	}
}

func TestJoinRoomFull(t *testing.T) {
	h := NewHub()
	code, _ := h.CreateRoom("host", "Host")

	// Fill up to MaxPlayers
	for i := 1; i < game.MaxPlayers; i++ {
		pid := "p" + string(rune('0'+i))
		if err := h.JoinRoom(code, pid, "Player"); err != nil {
			t.Fatalf("JoinRoom player %d error: %v", i, err)
		}
	}

	// One more should fail
	err := h.JoinRoom(code, "extra", "Extra")
	if err == nil {
		t.Fatal("expected error for full room")
	}
	if err.Error() != "room is full" {
		t.Errorf("error = %q, want %q", err.Error(), "room is full")
	}
}

func TestJoinRoomGameInProgress(t *testing.T) {
	h := NewHub()
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)
	// Start the game so phase is no longer lobby
	room.Game.Start()

	err := h.JoinRoom(code, "p3", "Player3")
	if err == nil {
		t.Fatal("expected error for game in progress")
	}
	if err.Error() != "game already in progress" {
		t.Errorf("error = %q, want %q", err.Error(), "game already in progress")
	}
}

func TestJoinRoomDuplicateReconnect(t *testing.T) {
	h := NewHub()
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	// Joining again with same ID should succeed (reconnect)
	err := h.JoinRoom(code, "p2", "Player2")
	if err != nil {
		t.Fatalf("reconnect should not error: %v", err)
	}

	room, _ := h.GetRoom(code)
	// Should still be 2 players, not 3
	if len(room.Game.Players) != 2 {
		t.Errorf("expected 2 players after reconnect, got %d", len(room.Game.Players))
	}
}

// --- GetRoom tests ---

func TestGetRoomExists(t *testing.T) {
	h := NewHub()
	code, _ := h.CreateRoom("host", "Host")
	room, ok := h.GetRoom(code)
	if !ok || room == nil {
		t.Fatal("GetRoom should find existing room")
	}
	if room.Code != code {
		t.Errorf("room code = %q, want %q", room.Code, code)
	}
}

func TestGetRoomNotFound(t *testing.T) {
	h := NewHub()
	_, ok := h.GetRoom("ZZZZ")
	if ok {
		t.Error("GetRoom should return false for nonexistent room")
	}
}

// --- Hub Register / Unregister via channels ---

func TestHubRegisterClient(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")

	c := newTestClient("host", code)
	h.Register <- c

	// Wait for the hub to process and broadcast state
	msg := readMessage(t, c, 2*time.Second)
	if msg.Type != "game_state" {
		t.Errorf("expected game_state, got %q", msg.Type)
	}
}

func TestHubUnregisterClient(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")

	c := newTestClient("host", code)
	h.Register <- c

	// Drain the initial broadcast
	readMessage(t, c, 2*time.Second)

	h.Unregister <- c

	// After unregister, send channel should be closed
	select {
	case _, ok := <-c.send:
		if ok {
			// Could be a final state broadcast from unregister, drain it
			select {
			case _, ok2 := <-c.send:
				if ok2 {
					t.Error("send channel should eventually be closed")
				}
			case <-time.After(2 * time.Second):
				t.Error("timed out waiting for channel close")
			}
		}
		// Channel was closed; expected
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for channel close")
	}
}

// --- Message handling tests ---

func TestHandleStartGameHostOnly(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	// Drain initial state broadcasts
	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Non-host tries to start game
	action := Action{Type: "start_game"}
	data, _ := json.Marshal(action)
	h.Incoming <- IncomingMessage{Client: p2Client, Data: data}

	// p2 should receive an error
	msg := readMessage(t, p2Client, 2*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error for non-host start, got %q", msg.Type)
	}

	// Host starts game
	h.Incoming <- IncomingMessage{Client: hostClient, Data: data}

	// Both should get a game_state broadcast
	hostMsg := readMessage(t, hostClient, 2*time.Second)
	if hostMsg.Type != "game_state" {
		t.Errorf("expected game_state after start, got %q", hostMsg.Type)
	}
	p2Msg := readMessage(t, p2Client, 2*time.Second)
	if p2Msg.Type != "game_state" {
		t.Errorf("expected game_state after start, got %q", p2Msg.Type)
	}
}

func TestHandleDrawAction(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start the game
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Get current turn player
	currentID := room.Game.PlayerOrder[room.Game.CurrentTurn]
	var currentClient *Client
	if currentID == "host" {
		currentClient = hostClient
	} else {
		currentClient = p2Client
	}

	// Draw a sand card
	drawPayload, _ := json.Marshal(DrawPayload{Suit: game.SuitSand})
	drawAction, _ := json.Marshal(Action{
		Type:    "draw",
		Payload: drawPayload,
	})
	h.Incoming <- IncomingMessage{Client: currentClient, Data: drawAction}

	// Should get a state broadcast (both players)
	msg := readMessage(t, currentClient, 2*time.Second)
	if msg.Type != "game_state" {
		t.Errorf("expected game_state after draw, got %q", msg.Type)
	}
}

func TestHandleStandAction(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start the game
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Current player stands
	currentID := room.Game.PlayerOrder[room.Game.CurrentTurn]
	var currentClient *Client
	if currentID == "host" {
		currentClient = hostClient
	} else {
		currentClient = p2Client
	}

	standPayload, _ := json.Marshal(StandPayload{})
	standAction, _ := json.Marshal(Action{
		Type:    "stand",
		Payload: standPayload,
	})
	h.Incoming <- IncomingMessage{Client: currentClient, Data: standAction}

	msg := readMessage(t, currentClient, 2*time.Second)
	if msg.Type != "game_state" {
		t.Errorf("expected game_state after stand, got %q", msg.Type)
	}
}

func TestHandleNextRound(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start and force the game to round_end
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Manually advance game to round_end for testing
	room.mu.Lock()
	room.Game.StartReveal()
	room.Game.Reveal()
	room.mu.Unlock()

	// Now send next_round
	nextRoundAction, _ := json.Marshal(Action{Type: "next_round"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: nextRoundAction}

	msg := readMessage(t, hostClient, 2*time.Second)
	if msg.Type != "game_state" {
		t.Errorf("expected game_state after next_round, got %q", msg.Type)
	}

	// Verify game is back in turn phase
	room.mu.RLock()
	phase := room.Game.Phase
	round := room.Game.Round
	room.mu.RUnlock()

	if phase != game.PhaseTurn {
		t.Errorf("expected phase %q after next_round, got %q", game.PhaseTurn, phase)
	}
	if round != 2 {
		t.Errorf("expected round 2 after next_round, got %d", round)
	}
}

func TestHandleUnknownAction(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")

	c := newTestClient("host", code)
	h.Register <- c

	time.Sleep(100 * time.Millisecond)
	drainMessages(c)

	action, _ := json.Marshal(Action{Type: "bogus_action"})
	h.Incoming <- IncomingMessage{Client: c, Data: action}

	msg := readMessage(t, c, 2*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error for unknown action, got %q", msg.Type)
	}
}

func TestHandleInvalidJSON(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")

	c := newTestClient("host", code)
	h.Register <- c

	time.Sleep(100 * time.Millisecond)
	drainMessages(c)

	h.Incoming <- IncomingMessage{Client: c, Data: []byte("not json")}

	msg := readMessage(t, c, 2*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error for invalid JSON, got %q", msg.Type)
	}
}

func TestHandleMessageRoomNotFound(t *testing.T) {
	h := newTestHub(t)

	c := newTestClient("host", "ZZZZ") // room doesn't exist
	// Don't register via hub, just send a message
	action, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: c, Data: action}

	msg := readMessage(t, c, 2*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error for missing room, got %q", msg.Type)
	}
}

// --- State broadcasting: per-player card visibility ---

func TestBroadcastStateHidesCardsDuringTurn(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start game
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	// Read host's state
	hostMsg := readMessage(t, hostClient, 2*time.Second)
	if hostMsg.Type != "game_state" {
		t.Fatalf("expected game_state, got %q", hostMsg.Type)
	}

	// Parse the payload
	payloadBytes, _ := json.Marshal(hostMsg.Payload)
	var hostState GameStateView
	if err := json.Unmarshal(payloadBytes, &hostState); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}

	// Phase should be turn
	if hostState.Phase != game.PhaseTurn {
		t.Errorf("expected phase %q, got %q", game.PhaseTurn, hostState.Phase)
	}

	// Host should see their own hand
	if hostState.YourHand == nil {
		t.Error("host should see their own hand during turn phase")
	}

	// Other players' cards should be hidden (nil) during turn phase
	for _, pv := range hostState.Players {
		if pv.SandCard != nil {
			t.Errorf("player %s sandCard should be nil during turn phase, got %+v", pv.ID, pv.SandCard)
		}
		if pv.BloodCard != nil {
			t.Errorf("player %s bloodCard should be nil during turn phase, got %+v", pv.ID, pv.BloodCard)
		}
	}
}

func TestBroadcastStateShowsCardsDuringReveal(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start and advance to reveal
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Manually set to reveal phase for testing card visibility
	room.mu.Lock()
	room.Game.StartReveal()
	room.mu.Unlock()

	// Trigger a state broadcast by sending a message that gets handled
	// We'll use a next_round that will fail (not in round_end) but the reveal
	// should still show cards before the error. Instead, let's register a
	// new client to trigger broadcastState
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Re-register to trigger broadcast
	h.broadcastState(room)

	hostMsg := readMessage(t, hostClient, 2*time.Second)
	if hostMsg.Type != "game_state" {
		t.Fatalf("expected game_state, got %q", hostMsg.Type)
	}

	payloadBytes, _ := json.Marshal(hostMsg.Payload)
	var hostState GameStateView
	if err := json.Unmarshal(payloadBytes, &hostState); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}

	if hostState.Phase != game.PhaseReveal {
		t.Errorf("expected phase %q, got %q", game.PhaseReveal, hostState.Phase)
	}

	// During reveal, all player cards should be visible
	for _, pv := range hostState.Players {
		if pv.SandCard == nil {
			t.Errorf("player %s sandCard should be visible during reveal", pv.ID)
		}
		if pv.BloodCard == nil {
			t.Errorf("player %s bloodCard should be visible during reveal", pv.ID)
		}
	}
}

func TestBroadcastStateShowsCardsDuringRoundEnd(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	h.Register <- hostClient

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)

	// Start, reveal, and advance to round_end
	room.mu.Lock()
	room.Game.Start()
	room.Game.StartReveal()
	room.Game.Reveal()
	room.mu.Unlock()

	h.broadcastState(room)

	msg := readMessage(t, hostClient, 2*time.Second)
	payloadBytes, _ := json.Marshal(msg.Payload)
	var state GameStateView
	json.Unmarshal(payloadBytes, &state)

	if state.Phase != game.PhaseRoundEnd {
		t.Errorf("expected phase %q, got %q", game.PhaseRoundEnd, state.Phase)
	}

	// Cards should still be visible during round_end
	for _, pv := range state.Players {
		if pv.SandCard == nil {
			t.Errorf("player %s sandCard should be visible during round_end", pv.ID)
		}
		if pv.BloodCard == nil {
			t.Errorf("player %s bloodCard should be visible during round_end", pv.ID)
		}
	}
}

func TestBroadcastStateYourHandPerPlayer(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start game
	room.mu.Lock()
	room.Game.Start()
	room.mu.Unlock()

	h.broadcastState(room)

	// Read each player's state
	hostMsg := readMessage(t, hostClient, 2*time.Second)
	p2Msg := readMessage(t, p2Client, 2*time.Second)

	// Parse host state
	hostPayloadBytes, _ := json.Marshal(hostMsg.Payload)
	var hostState GameStateView
	json.Unmarshal(hostPayloadBytes, &hostState)

	// Parse p2 state
	p2PayloadBytes, _ := json.Marshal(p2Msg.Payload)
	var p2State GameStateView
	json.Unmarshal(p2PayloadBytes, &p2State)

	// Each player should see their own hand
	if hostState.YourHand == nil {
		t.Error("host should see their own hand")
	}
	if p2State.YourHand == nil {
		t.Error("p2 should see their own hand")
	}

	// The hands should be different (different players)
	if hostState.YourHand.SandCard.ID == p2State.YourHand.SandCard.ID &&
		hostState.YourHand.BloodCard.ID == p2State.YourHand.BloodCard.ID {
		t.Error("host and p2 should have different hands (or at least different card IDs)")
	}
}

// --- sendError tests ---

func TestSendError(t *testing.T) {
	h := NewHub()
	c := newTestClient("p1", "ABCD")

	h.sendError(c, "something went wrong")

	msg := readMessage(t, c, 1*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error type, got %q", msg.Type)
	}
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("payload is not a map: %T", msg.Payload)
	}
	if payload["message"] != "something went wrong" {
		t.Errorf("error message = %q, want %q", payload["message"], "something went wrong")
	}
}

// --- Full game flow test ---

func TestFullGameFlowTwoPlayers(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start game
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// All players stand to trigger reveal
	for i := 0; i < 2; i++ {
		room.mu.RLock()
		currentID := room.Game.PlayerOrder[room.Game.CurrentTurn]
		room.mu.RUnlock()

		var currentClient *Client
		if currentID == "host" {
			currentClient = hostClient
		} else {
			currentClient = p2Client
		}

		standPayload, _ := json.Marshal(StandPayload{})
		standAction, _ := json.Marshal(Action{
			Type:    "stand",
			Payload: standPayload,
		})
		h.Incoming <- IncomingMessage{Client: currentClient, Data: standAction}

		time.Sleep(100 * time.Millisecond)
		drainMessages(hostClient)
		drainMessages(p2Client)
	}

	// After all stand, game should be in round_end (reveal was auto-triggered)
	room.mu.RLock()
	phase := room.Game.Phase
	room.mu.RUnlock()

	if phase != game.PhaseRoundEnd {
		t.Errorf("expected phase %q after all stand, got %q", game.PhaseRoundEnd, phase)
	}

	// Next round
	nextAction, _ := json.Marshal(Action{Type: "next_round"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: nextAction}

	time.Sleep(100 * time.Millisecond)

	room.mu.RLock()
	phase = room.Game.Phase
	round := room.Game.Round
	room.mu.RUnlock()

	if phase != game.PhaseTurn {
		t.Errorf("expected phase %q for round 2, got %q", game.PhaseTurn, phase)
	}
	if round != 2 {
		t.Errorf("expected round 2, got %d", round)
	}
}

// --- WebSocket integration test ---

// readWSGameState reads messages from a WebSocket connection until it gets a
// game_state envelope or times out. It skips non-game_state messages.
func readWSGameState(t *testing.T, conn *websocket.Conn, timeout time.Duration) GameStateView {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("readWSGameState: %v", err)
		}
		var env Envelope
		if err := json.Unmarshal(msgBytes, &env); err != nil {
			continue
		}
		if env.Type == "game_state" {
			payloadBytes, _ := json.Marshal(env.Payload)
			var state GameStateView
			json.Unmarshal(payloadBytes, &state)
			return state
		}
	}
}

func TestWebSocketIntegration(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	// Create a test server that acts as a WebSocket relay
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		playerID := r.URL.Query().Get("playerId")
		roomCode := r.URL.Query().Get("roomCode")

		client := NewClient(playerID, roomCode, conn, h)
		h.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}))
	defer srv.Close()

	// Connect both players first
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "?playerId=host&roomCode=" + code
	hostConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial host: %v", err)
	}
	defer hostConn.Close()

	// Host receives initial game_state (lobby)
	state := readWSGameState(t, hostConn, 2*time.Second)
	if state.Phase != game.PhaseLobby {
		t.Errorf("expected initial phase %q, got %q", game.PhaseLobby, state.Phase)
	}

	wsURL2 := "ws" + strings.TrimPrefix(srv.URL, "http") + "?playerId=p2&roomCode=" + code
	p2Conn, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("dial p2: %v", err)
	}
	defer p2Conn.Close()

	// p2 receives initial game_state
	_ = readWSGameState(t, p2Conn, 2*time.Second)

	// Host also gets a state broadcast when p2's client registers
	// (may or may not arrive depending on timing, drain via goroutine)
	// Give hub time to process p2 registration
	time.Sleep(200 * time.Millisecond)

	// Host sends start_game via WebSocket
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	if err := hostConn.WriteMessage(websocket.TextMessage, startAction); err != nil {
		t.Fatalf("write start_game: %v", err)
	}

	// Read messages from host until we get a game_state with turn phase
	// (there might be intermediate lobby-phase broadcasts from p2 joining)
	hostConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var finalState GameStateView
	for {
		_, msgBytes, readErr := hostConn.ReadMessage()
		if readErr != nil {
			t.Fatalf("read host after start: %v", readErr)
		}
		var env Envelope
		json.Unmarshal(msgBytes, &env)
		if env.Type == "game_state" {
			payloadBytes, _ := json.Marshal(env.Payload)
			json.Unmarshal(payloadBytes, &finalState)
			if finalState.Phase == game.PhaseTurn {
				break
			}
		}
	}

	if finalState.Phase != game.PhaseTurn {
		t.Errorf("expected phase %q after start, got %q", game.PhaseTurn, finalState.Phase)
	}

	// Verify each player sees their own hand
	if finalState.YourHand == nil {
		t.Error("host should receive their hand in game_state")
	}
}

func TestBroadcastStateCurrentTurnPlayerID(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	c := newTestClient("host", code)
	h.Register <- c

	time.Sleep(100 * time.Millisecond)
	drainMessages(c)

	room.mu.Lock()
	room.Game.Start()
	room.mu.Unlock()

	h.broadcastState(room)

	msg := readMessage(t, c, 2*time.Second)
	payloadBytes, _ := json.Marshal(msg.Payload)
	var state GameStateView
	json.Unmarshal(payloadBytes, &state)

	if state.CurrentTurnPlayerID == "" {
		t.Error("currentTurnPlayerId should be set during turn phase")
	}

	// It should be one of the two player IDs
	if state.CurrentTurnPlayerID != "host" && state.CurrentTurnPlayerID != "p2" {
		t.Errorf("unexpected currentTurnPlayerId: %q", state.CurrentTurnPlayerID)
	}
}

func TestHandleDrawWrongTurn(t *testing.T) {
	h := newTestHub(t)
	code, _ := h.CreateRoom("host", "Host")
	h.JoinRoom(code, "p2", "Player2")

	room, _ := h.GetRoom(code)

	hostClient := newTestClient("host", code)
	p2Client := newTestClient("p2", code)
	h.Register <- hostClient
	h.Register <- p2Client

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Start game
	startAction, _ := json.Marshal(Action{Type: "start_game"})
	h.Incoming <- IncomingMessage{Client: hostClient, Data: startAction}

	time.Sleep(100 * time.Millisecond)
	drainMessages(hostClient)
	drainMessages(p2Client)

	// Find who is NOT the current player
	room.mu.RLock()
	currentID := room.Game.PlayerOrder[room.Game.CurrentTurn]
	room.mu.RUnlock()

	var wrongClient *Client
	if currentID == "host" {
		wrongClient = p2Client
	} else {
		wrongClient = hostClient
	}

	drawPayload, _ := json.Marshal(DrawPayload{Suit: game.SuitSand})
	drawAction, _ := json.Marshal(Action{
		Type:    "draw",
		Payload: drawPayload,
	})
	h.Incoming <- IncomingMessage{Client: wrongClient, Data: drawAction}

	msg := readMessage(t, wrongClient, 2*time.Second)
	if msg.Type != "error" {
		t.Errorf("expected error for wrong turn draw, got %q", msg.Type)
	}
}
