package db

import "time"

// User represents a player account in the database.
type User struct {
	ID           string    `json:"id"`
	Email        *string   `json:"email,omitempty"`
	DisplayName  string    `json:"displayName"`
	PasswordHash *string   `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// PlayerResult holds the outcome of a game for a single player,
// used when recording game results.
type PlayerResult struct {
	UserID     string `json:"userId"`
	FinalChips int    `json:"finalChips"`
	IsWinner   bool   `json:"isWinner"`
}

// GameHistoryEntry represents a completed game in a player's history.
type GameHistoryEntry struct {
	GameID     string              `json:"gameId"`
	RoomCode   string              `json:"roomCode"`
	FinishedAt time.Time           `json:"finishedAt"`
	Rounds     int                 `json:"rounds"`
	Players    []GamePlayerSummary `json:"players"`
	WinnerName string              `json:"winnerName"`
}

// GamePlayerSummary holds a player's summary within a completed game.
type GamePlayerSummary struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	FinalChips  int    `json:"finalChips"`
	IsWinner    bool   `json:"isWinner"`
}
