package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sabacc/auth"
	"sabacc/db"
	"sabacc/room"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins for local dev
}

type Handler struct {
	Hub  *room.Hub
	Repo *db.Repository    // nil when running without a database
	Auth *auth.AuthService // nil when running without a database
}

func NewHandler(hub *room.Hub, repo *db.Repository, authSvc *auth.AuthService) *Handler {
	return &Handler{Hub: hub, Repo: repo, Auth: authSvc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /rooms", h.CreateRoom)
	mux.HandleFunc("POST /rooms/{code}/join", h.JoinRoom)
	mux.HandleFunc("GET /api/games", h.GetGameHistory)
	mux.HandleFunc("GET /api/profile/{id}", h.GetProfileByID)
	mux.HandleFunc("GET /api/profile", h.GetProfile)
	mux.HandleFunc("GET /ws", h.WebSocket)

	// Auth routes — only registered when auth is available
	if h.Auth != nil {
		mux.HandleFunc("POST /auth/signup", h.Signup)
		mux.HandleFunc("POST /auth/login", h.Login)
		mux.HandleFunc("GET /auth/me", h.Me)
	}
}

// ── Room handlers ──

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

	// If an auth token is present, use the authenticated user's ID as the player ID.
	if userID := h.authenticatedUserID(r); userID != "" {
		req.PlayerID = userID
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

	// If an auth token is present, use the authenticated user's ID as the player ID.
	if userID := h.authenticatedUserID(r); userID != "" {
		req.PlayerID = userID
	}

	if err := h.Hub.JoinRoom(code, req.PlayerID, req.PlayerName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetGameHistory(w http.ResponseWriter, r *http.Request) {
	// Prefer the authenticated user's ID; fall back to the query param for guests.
	playerID := h.authenticatedUserID(r)
	if playerID == "" {
		playerID = r.URL.Query().Get("playerId")
	}
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

// GetProfile returns the authenticated user's own profile with stats and game history.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := h.authenticatedUserID(r)
	if userID == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if h.Repo == nil {
		http.Error(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	profile, err := h.Repo.GetPlayerProfile(r.Context(), userID, true)
	if err != nil {
		http.Error(w, "failed to fetch profile", http.StatusInternalServerError)
		return
	}
	if profile == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// GetProfileByID returns a public profile for any player by their user ID.
func (h *Handler) GetProfileByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	if h.Repo == nil {
		http.Error(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	// includeEmail only if the requester is viewing their own profile
	requesterID := h.authenticatedUserID(r)
	includeEmail := requesterID == userID

	profile, err := h.Repo.GetPlayerProfile(r.Context(), userID, includeEmail)
	if err != nil {
		http.Error(w, "failed to fetch profile", http.StatusInternalServerError)
		return
	}
	if profile == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// authenticatedUserID extracts and validates a Bearer token from the Authorization
// header. Returns the authenticated user ID or "" if not present/invalid.
func (h *Handler) authenticatedUserID(r *http.Request) string {
	if h.Auth == nil {
		return ""
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	tokenStr := strings.TrimSpace(parts[1])
	if tokenStr == "" {
		return ""
	}
	userID, err := h.Auth.ValidateToken(tokenStr)
	if err != nil {
		return ""
	}
	return userID
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("playerId")
	roomCode := r.URL.Query().Get("roomCode")
	if playerID == "" || roomCode == "" {
		http.Error(w, "playerId and roomCode are required", http.StatusBadRequest)
		return
	}

	// If a token query param is provided and auth is available, validate it
	// and override the playerID with the authenticated user's ID.
	if token := r.URL.Query().Get("token"); token != "" && h.Auth != nil {
		if userID, err := h.Auth.ValidateToken(token); err == nil {
			playerID = userID
		}
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

// ── Auth handlers ──

type SignupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	// Validate
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if req.DisplayName == "" {
		http.Error(w, "displayName is required", http.StatusBadRequest)
		return
	}

	user, token, err := h.Auth.Signup(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "signup failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, token, err := h.Auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	// Extract and validate token manually (this endpoint is self-protecting
	// rather than relying on middleware, since other routes are not protected yet).
	tokenStr := ""
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenStr = strings.TrimSpace(parts[1])
		}
	}
	if tokenStr == "" {
		http.Error(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	userID, err := h.Auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	user, err := h.Auth.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
