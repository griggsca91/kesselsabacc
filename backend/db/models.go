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
