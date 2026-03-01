package room

import (
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

// --- generateCode tests ---

func TestGenerateCodeLength(t *testing.T) {
	code := generateCode()
	if len(code) != 4 {
		t.Errorf("expected code length 4, got %d (%q)", len(code), code)
	}
}

func TestGenerateCodeCharset(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateCode()
		for _, ch := range code {
			if !strings.ContainsRune(codeChars, ch) {
				t.Errorf("code %q contains invalid character %q", code, ch)
			}
		}
	}
}

func TestGenerateCodeExcludesAmbiguous(t *testing.T) {
	// codeChars should not contain I, O, 0, 1 which are ambiguous
	for _, ch := range "IO01" {
		if strings.ContainsRune(codeChars, ch) {
			t.Errorf("codeChars should not contain ambiguous character %q", ch)
		}
	}
}

func TestGenerateCodeRandomness(t *testing.T) {
	// Generate many codes and check that we get at least some variety
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		seen[generateCode()] = true
	}
	if len(seen) < 10 {
		t.Errorf("expected at least 10 unique codes from 50 generations, got %d", len(seen))
	}
}

// --- NewRoom tests ---

func TestNewRoom(t *testing.T) {
	r := NewRoom()
	if r == nil {
		t.Fatal("NewRoom returned nil")
	}
	if r.Game == nil {
		t.Error("NewRoom should initialize Game")
	}
	if r.Clients == nil {
		t.Error("NewRoom should initialize Clients map")
	}
	if len(r.Clients) != 0 {
		t.Error("NewRoom should have empty Clients map")
	}
}

// --- AddClient / RemoveClient tests ---

func newTestClient(playerID, roomCode string) *Client {
	return &Client{
		PlayerID: playerID,
		RoomCode: roomCode,
		send:     make(chan []byte, sendBufferSize),
	}
}

func TestAddClient(t *testing.T) {
	r := NewRoom()
	c := newTestClient("p1", "ABCD")
	r.AddClient(c)

	if len(r.Clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(r.Clients))
	}
	if r.Clients["p1"] != c {
		t.Error("client not found in room")
	}
}

func TestAddMultipleClients(t *testing.T) {
	r := NewRoom()
	c1 := newTestClient("p1", "ABCD")
	c2 := newTestClient("p2", "ABCD")
	r.AddClient(c1)
	r.AddClient(c2)

	if len(r.Clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(r.Clients))
	}
}

func TestAddClientOverwritesSameID(t *testing.T) {
	r := NewRoom()
	c1 := newTestClient("p1", "ABCD")
	c2 := newTestClient("p1", "ABCD") // same playerID
	r.AddClient(c1)
	r.AddClient(c2)

	if len(r.Clients) != 1 {
		t.Errorf("expected 1 client after overwrite, got %d", len(r.Clients))
	}
	if r.Clients["p1"] != c2 {
		t.Error("client should be overwritten with new one")
	}
}

func TestRemoveClient(t *testing.T) {
	r := NewRoom()
	c := newTestClient("p1", "ABCD")
	r.AddClient(c)
	r.RemoveClient("p1")

	if len(r.Clients) != 0 {
		t.Errorf("expected 0 clients after removal, got %d", len(r.Clients))
	}
}

func TestRemoveClientNonexistent(t *testing.T) {
	r := NewRoom()
	// Should not panic when removing a client that doesn't exist
	r.RemoveClient("nonexistent")
	if len(r.Clients) != 0 {
		t.Error("room should still be empty")
	}
}

func TestAddRemoveConcurrent(t *testing.T) {
	r := NewRoom()
	var wg sync.WaitGroup

	// Concurrently add clients
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pid := "p" + string(rune('A'+id))
			c := newTestClient(pid, "ABCD")
			r.AddClient(c)
		}(i)
	}
	wg.Wait()

	// Concurrently remove clients
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pid := "p" + string(rune('A'+id))
			r.RemoveClient(pid)
		}(i)
	}
	wg.Wait()

	if len(r.Clients) != 0 {
		t.Errorf("expected 0 clients after concurrent add/remove, got %d", len(r.Clients))
	}
}

// --- Broadcast / SendTo tests ---

func TestBroadcast(t *testing.T) {
	r := NewRoom()
	c1 := newTestClient("p1", "ABCD")
	c2 := newTestClient("p2", "ABCD")
	r.AddClient(c1)
	r.AddClient(c2)

	msg := []byte("hello")
	r.Broadcast(msg)

	// Check both clients received the message
	select {
	case got := <-c1.send:
		if string(got) != "hello" {
			t.Errorf("c1 got %q, want %q", got, "hello")
		}
	default:
		t.Error("c1 did not receive broadcast")
	}

	select {
	case got := <-c2.send:
		if string(got) != "hello" {
			t.Errorf("c2 got %q, want %q", got, "hello")
		}
	default:
		t.Error("c2 did not receive broadcast")
	}
}

func TestBroadcastEmpty(t *testing.T) {
	r := NewRoom()
	// Should not panic broadcasting to empty room
	r.Broadcast([]byte("hello"))
}

func TestSendTo(t *testing.T) {
	r := NewRoom()
	c1 := newTestClient("p1", "ABCD")
	c2 := newTestClient("p2", "ABCD")
	r.AddClient(c1)
	r.AddClient(c2)

	msg := []byte("secret")
	r.SendTo("p1", msg)

	// c1 should receive the message
	select {
	case got := <-c1.send:
		if string(got) != "secret" {
			t.Errorf("c1 got %q, want %q", got, "secret")
		}
	default:
		t.Error("c1 did not receive SendTo message")
	}

	// c2 should NOT receive the message
	select {
	case <-c2.send:
		t.Error("c2 should not receive SendTo message targeted at p1")
	default:
		// expected
	}
}

func TestSendToNonexistent(t *testing.T) {
	r := NewRoom()
	// Should not panic sending to nonexistent player
	r.SendTo("nobody", []byte("hello"))
}

// --- Client.Send tests ---

func TestClientSend(t *testing.T) {
	c := newTestClient("p1", "ABCD")
	c.Send([]byte("test"))

	select {
	case got := <-c.send:
		if string(got) != "test" {
			t.Errorf("got %q, want %q", got, "test")
		}
	default:
		t.Error("message not sent")
	}
}

func TestClientSendBufferFull(t *testing.T) {
	c := &Client{
		PlayerID: "p1",
		RoomCode: "ABCD",
		send:     make(chan []byte, 1),
	}
	// Fill the buffer
	c.Send([]byte("first"))
	// This should not block (just drops the message)
	c.Send([]byte("second"))

	got := <-c.send
	if string(got) != "first" {
		t.Errorf("got %q, want %q", got, "first")
	}
}

// --- NewClient tests ---

func TestNewClient(t *testing.T) {
	hub := NewHub()
	// NewClient requires a *websocket.Conn; pass nil for unit test
	c := NewClient("p1", "ABCD", (*websocket.Conn)(nil), hub)

	if c.PlayerID != "p1" {
		t.Errorf("PlayerID = %q, want %q", c.PlayerID, "p1")
	}
	if c.RoomCode != "ABCD" {
		t.Errorf("RoomCode = %q, want %q", c.RoomCode, "ABCD")
	}
	if c.hub != hub {
		t.Error("hub not set correctly")
	}
	if c.send == nil {
		t.Error("send channel not initialized")
	}
	if cap(c.send) != sendBufferSize {
		t.Errorf("send buffer size = %d, want %d", cap(c.send), sendBufferSize)
	}
}
