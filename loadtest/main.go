// loadtest simulates concurrent WebSocket connections and room activity
// against a running Sabacc backend to measure concurrency limits and latency.
//
// Usage:
//
//	go run ./loadtest [flags]
//
// Flags:
//
//	-url      Backend base URL (default: http://localhost:8080)
//	-rooms    Number of rooms to create (default: 10)
//	-players  Players per room (default: 2, max 8)
//	-duration Test duration in seconds (default: 30)
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ── CLI flags ─────────────────────────────────────────────────────────────────

var (
	baseURL  = flag.String("url", "http://localhost:8080", "backend base URL")
	numRooms = flag.Int("rooms", 10, "number of rooms to create")
	perRoom  = flag.Int("players", 2, "players per room (2-8)")
	duration = flag.Int("duration", 30, "test duration in seconds")
)

// ── Counters ──────────────────────────────────────────────────────────────────

var (
	connOK     atomic.Int64
	connFail   atomic.Int64
	msgsRecv   atomic.Int64
	createOK   atomic.Int64
	createFail atomic.Int64
	joinOK     atomic.Int64
	joinFail   atomic.Int64
	latSum     atomic.Int64 // microseconds
	latCount   atomic.Int64
)

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func post(path string, body any) (map[string]any, error) {
	b, _ := json.Marshal(body)
	resp, err := http.Post(*baseURL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return result, nil
}

// ── Room scenario ─────────────────────────────────────────────────────────────

func runRoom(roomIdx int, stopAt time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	hostID := fmt.Sprintf("lt-host-%d", roomIdx)
	hostName := fmt.Sprintf("Host%d", roomIdx)

	// Create room
	t0 := time.Now()
	resp, err := post("/rooms", map[string]any{
		"playerId":   hostID,
		"playerName": hostName,
		"isPublic":   false,
	})
	if err != nil {
		createFail.Add(1)
		slog.Warn("create room failed", "room", roomIdx, "error", err)
		return
	}
	createOK.Add(1)
	latSum.Add(int64(time.Since(t0).Microseconds()))
	latCount.Add(1)

	code, _ := resp["code"].(string)
	if code == "" {
		createFail.Add(1)
		return
	}

	// Join remaining players
	players := make([]string, 0, *perRoom)
	players = append(players, hostID)

	for i := 1; i < *perRoom; i++ {
		pid := fmt.Sprintf("lt-p%d-%d", roomIdx, i)
		pname := fmt.Sprintf("P%d_%d", roomIdx, i)
		t1 := time.Now()
		_, err := post(fmt.Sprintf("/rooms/%s/join", code), map[string]any{
			"playerId":   pid,
			"playerName": pname,
		})
		if err != nil {
			joinFail.Add(1)
		} else {
			joinOK.Add(1)
			latSum.Add(int64(time.Since(t1).Microseconds()))
			latCount.Add(1)
		}
		players = append(players, pid)
	}

	// Connect all players via WebSocket
	wsBase := strings.Replace(*baseURL, "http://", "ws://", 1)
	wsBase = strings.Replace(wsBase, "https://", "wss://", 1)

	var wsWg sync.WaitGroup
	for _, pid := range players {
		wsWg.Add(1)
		go func(playerID string) {
			defer wsWg.Done()
			connectAndListen(wsBase, code, playerID, stopAt)
		}(pid)
	}
	wsWg.Wait()
}

func connectAndListen(wsBase, code, playerID string, stopAt time.Time) {
	u := url.Values{}
	u.Set("playerId", playerID)
	u.Set("roomCode", code)
	wsURL := wsBase + "/ws?" + u.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		connFail.Add(1)
		return
	}
	defer conn.Close()
	connOK.Add(1)

	conn.SetReadDeadline(stopAt)
	for time.Now().Before(stopAt) {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		msgsRecv.Add(1)
	}
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	flag.Parse()
	slog.Info("load test starting",
		"url", *baseURL,
		"rooms", *numRooms,
		"playersPerRoom", *perRoom,
		"duration", fmt.Sprintf("%ds", *duration),
	)

	// Verify server is up
	resp, err := http.Get(*baseURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		slog.Error("server not reachable — is the backend running?", "url", *baseURL)
		return
	}
	resp.Body.Close()

	stopAt := time.Now().Add(time.Duration(*duration) * time.Second)

	var wg sync.WaitGroup
	for i := 0; i < *numRooms; i++ {
		wg.Add(1)
		go runRoom(i, stopAt, &wg)
	}
	wg.Wait()

	// Print results
	totalReqs := latCount.Load()
	var avgLatency float64
	if totalReqs > 0 {
		avgLatency = float64(latSum.Load()) / float64(totalReqs) / 1000.0 // ms
	}

	fmt.Println("\n── Load Test Results ──────────────────────────────")
	fmt.Printf("Rooms created:       %d ok / %d fail\n", createOK.Load(), createFail.Load())
	fmt.Printf("Player joins:        %d ok / %d fail\n", joinOK.Load(), joinFail.Load())
	fmt.Printf("WS connections:      %d ok / %d fail\n", connOK.Load(), connFail.Load())
	fmt.Printf("Messages received:   %d\n", msgsRecv.Load())
	fmt.Printf("Avg HTTP latency:    %.2f ms\n", avgLatency)
	fmt.Println("───────────────────────────────────────────────────")
}
