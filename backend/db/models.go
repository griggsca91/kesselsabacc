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
	UserID     string  `json:"userId"`
	FinalChips int     `json:"finalChips"`
	IsWinner   bool    `json:"isWinner"`
	HandRank   *string `json:"handRank,omitempty"` // "pure_sabacc", "sabacc", "no_sabacc"
}

// PlayerStats holds aggregated statistics for a player.
type PlayerStats struct {
	GamesPlayed int     `json:"gamesPlayed"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	WinRate     float64 `json:"winRate"`
	BestHand    *string `json:"bestHand"` // "pure_sabacc", "sabacc", or "no_sabacc"; nil if no data
}

// PlayerProfile combines user info, stats, and game history for a profile page.
type PlayerProfile struct {
	User  ProfileUser        `json:"user"`
	Stats PlayerStats        `json:"stats"`
	Games []GameHistoryEntry `json:"games"`
}

// ProfileUser is a subset of user data exposed on profiles.
type ProfileUser struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	Email       *string   `json:"email,omitempty"` // only shown on own profile
	MemberSince time.Time `json:"memberSince"`
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
