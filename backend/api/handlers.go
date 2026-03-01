package api

import (
	"encoding/json"
	"net/http"
	"sabacc/db"
	"sabacc/room"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins for local dev
}

type Handler struct {
	Hub  *room.Hub
	Repo *db.Repository // nil when running without a database
}

func NewHandler(hub *room.Hub, repo *db.Repository) *Handler {
	return &Handler{Hub: hub, Repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /rooms", h.CreateRoom)
	mux.HandleFunc("POST /rooms/{code}/join", h.JoinRoom)
	mux.HandleFunc("GET /api/games", h.GetGameHistory)
	mux.HandleFunc("GET /ws", h.WebSocket)
}

type CreateRoomRequest struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

type CreateRoomResponse struct {
	Code string `json:"code"`
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerID == "" || req.PlayerName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	code, err := h.Hub.CreateRoom(req.PlayerID, req.PlayerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateRoomResponse{Code: code})
}

type JoinRoomRequest struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerID == "" || req.PlayerName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.Hub.JoinRoom(code, req.PlayerID, req.PlayerName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetGameHistory(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		http.Error(w, "playerId query parameter is required", http.StatusBadRequest)
		return
	}

	if h.Repo == nil {
		http.Error(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	games, err := h.Repo.GetGameHistory(r.Context(), playerID)
	if err != nil {
		http.Error(w, "failed to fetch game history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("playerId")
	roomCode := r.URL.Query().Get("roomCode")
	if playerID == "" || roomCode == "" {
		http.Error(w, "playerId and roomCode are required", http.StatusBadRequest)
		return
	}

	rm, ok := h.Hub.GetRoom(roomCode)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	if rm.Game.PlayerByID(playerID) == nil {
		http.Error(w, "player not in room", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := room.NewClient(playerID, roomCode, conn, h.Hub)
	h.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
