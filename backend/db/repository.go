package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository provides data-access methods for the sabacc database.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ---------- Users ----------

// CreateUser inserts a new user and returns the created record.
func (r *Repository) CreateUser(ctx context.Context, displayName, email, passwordHash string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (display_name, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, display_name, password_hash, created_at, updated_at`,
		displayName, toNullString(email), toNullString(passwordHash),
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

// GetUserByEmail looks up a user by their email address.
// Returns nil, nil if no user is found.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, display_name, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

// GetUserByID looks up a user by their UUID.
// Returns nil, nil if no user is found.
func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, display_name, password_hash, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

// ---------- Rooms ----------

// CreateRoom inserts a new room with the given code and host.
func (r *Repository) CreateRoom(ctx context.Context, code string, hostUserID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO rooms (code, host_user_id, status) VALUES ($1, $2, 'waiting')`,
		code, hostUserID,
	)
	if err != nil {
		return fmt.Errorf("create room: %w", err)
	}
	return nil
}

// UpdateRoomStatus sets the status of a room identified by its code.
func (r *Repository) UpdateRoomStatus(ctx context.Context, code string, status string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE rooms SET status = $1, updated_at = NOW() WHERE code = $2`,
		status, code,
	)
	if err != nil {
		return fmt.Errorf("update room status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("room %s not found", code)
	}
	return nil
}

// ---------- Game State ----------

// SaveGameState upserts the current game state for a room. If a game record
// already exists for the room (and has not finished), it updates that row;
// otherwise it inserts a new one. Returns the game ID.
func (r *Repository) SaveGameState(ctx context.Context, roomID string, round int, phase string, stateJSON []byte) (string, error) {
	var gameID string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO games (room_id, round_number, phase, state_json)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		roomID, round, phase, stateJSON,
	).Scan(&gameID)
	if err != nil {
		return "", fmt.Errorf("save game state: %w", err)
	}
	return gameID, nil
}

// LoadGameState returns the most recent unfinished game state JSON for a room.
// Returns nil, nil if no active game exists.
func (r *Repository) LoadGameState(ctx context.Context, roomID string) ([]byte, error) {
	var stateJSON []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT state_json FROM games
		 WHERE room_id = $1 AND finished_at IS NULL
		 ORDER BY started_at DESC
		 LIMIT 1`,
		roomID,
	).Scan(&stateJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load game state: %w", err)
	}
	return stateJSON, nil
}

// ---------- Game Results ----------

// RecordGameResult marks a game as finished and writes player outcomes.
func (r *Repository) RecordGameResult(ctx context.Context, gameID string, playerResults []PlayerResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Mark the game as finished
	_, err = tx.ExecContext(ctx,
		`UPDATE games SET finished_at = NOW() WHERE id = $1`, gameID,
	)
	if err != nil {
		return fmt.Errorf("finish game: %w", err)
	}

	// Insert player results
	for _, pr := range playerResults {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO game_players (game_id, user_id, final_chips, is_winner, hand_rank)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (game_id, user_id) DO UPDATE
			   SET final_chips = EXCLUDED.final_chips,
			       is_winner = EXCLUDED.is_winner,
			       hand_rank = EXCLUDED.hand_rank`,
			gameID, pr.UserID, pr.FinalChips, pr.IsWinner, pr.HandRank,
		)
		if err != nil {
			return fmt.Errorf("insert player result: %w", err)
		}
	}

	return tx.Commit()
}

// ---------- Game History ----------

// GetGameHistory returns the most recent completed games for a player, limited to 50.
func (r *Repository) GetGameHistory(ctx context.Context, playerID string) ([]GameHistoryEntry, error) {
	// Step 1: Get the game IDs and basic info for games this player participated in.
	rows, err := r.db.QueryContext(ctx,
		`SELECT g.id, r.code, g.finished_at, g.round_number
		 FROM games g
		 JOIN rooms r ON r.id = g.room_id
		 JOIN game_players gp ON gp.game_id = g.id
		 WHERE gp.user_id = $1 AND g.finished_at IS NOT NULL
		 ORDER BY g.finished_at DESC
		 LIMIT 50`,
		playerID,
	)
	if err != nil {
		return nil, fmt.Errorf("get game history: %w", err)
	}
	defer rows.Close()

	type gameInfo struct {
		entry GameHistoryEntry
	}
	var games []gameInfo
	gameIDs := []string{}

	for rows.Next() {
		var gi gameInfo
		if err := rows.Scan(&gi.entry.GameID, &gi.entry.RoomCode, &gi.entry.FinishedAt, &gi.entry.Rounds); err != nil {
			return nil, fmt.Errorf("scan game history: %w", err)
		}
		games = append(games, gi)
		gameIDs = append(gameIDs, gi.entry.GameID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game history: %w", err)
	}
	if len(games) == 0 {
		return []GameHistoryEntry{}, nil
	}

	// Step 2: Get all player results for these games.
	// Build a parameterized query for the game IDs.
	placeholders := ""
	args := make([]interface{}, len(gameIDs))
	for i, id := range gameIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	playerRows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(
			`SELECT gp.game_id, gp.user_id, u.display_name, gp.final_chips, gp.is_winner
			 FROM game_players gp
			 JOIN users u ON u.id = gp.user_id
			 WHERE gp.game_id IN (%s)
			 ORDER BY gp.is_winner DESC, gp.final_chips DESC`,
			placeholders,
		),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("get game players: %w", err)
	}
	defer playerRows.Close()

	// Map game_id -> []GamePlayerSummary
	playersByGame := map[string][]GamePlayerSummary{}
	winnerByGame := map[string]string{}
	for playerRows.Next() {
		var gameID, userID, displayName string
		var finalChips int
		var isWinner bool
		if err := playerRows.Scan(&gameID, &userID, &displayName, &finalChips, &isWinner); err != nil {
			return nil, fmt.Errorf("scan game player: %w", err)
		}
		playersByGame[gameID] = append(playersByGame[gameID], GamePlayerSummary{
			UserID:      userID,
			DisplayName: displayName,
			FinalChips:  finalChips,
			IsWinner:    isWinner,
		})
		if isWinner {
			winnerByGame[gameID] = displayName
		}
	}
	if err := playerRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game players: %w", err)
	}

	// Step 3: Assemble the results.
	entries := make([]GameHistoryEntry, len(games))
	for i, gi := range games {
		entry := gi.entry
		entry.Players = playersByGame[entry.GameID]
		if entry.Players == nil {
			entry.Players = []GamePlayerSummary{}
		}
		entry.WinnerName = winnerByGame[entry.GameID]
		entries[i] = entry
	}

	return entries, nil
}

// ---------- Player Stats ----------

// GetPlayerStats returns aggregated statistics for a player.
func (r *Repository) GetPlayerStats(ctx context.Context, userID string) (*PlayerStats, error) {
	var stats PlayerStats
	var bestHandRank sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT
			COUNT(*) AS games_played,
			COUNT(*) FILTER (WHERE is_winner) AS wins,
			MIN(CASE hand_rank
				WHEN 'pure_sabacc' THEN 1
				WHEN 'sabacc' THEN 2
				WHEN 'no_sabacc' THEN 3
			END) AS best_hand_rank
		 FROM game_players
		 WHERE user_id = $1`,
		userID,
	).Scan(&stats.GamesPlayed, &stats.Wins, &bestHandRank)
	if err != nil {
		return nil, fmt.Errorf("get player stats: %w", err)
	}

	stats.Losses = stats.GamesPlayed - stats.Wins
	if stats.GamesPlayed > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.GamesPlayed)
	}

	if bestHandRank.Valid {
		var name string
		switch bestHandRank.Int64 {
		case 1:
			name = "pure_sabacc"
		case 2:
			name = "sabacc"
		case 3:
			name = "no_sabacc"
		}
		if name != "" {
			stats.BestHand = &name
		}
	}

	return &stats, nil
}

// GetPlayerProfile returns the full profile for a player: user info, stats, and game history.
// If includeEmail is false, the email field is omitted from the response.
func (r *Repository) GetPlayerProfile(ctx context.Context, userID string, includeEmail bool) (*PlayerProfile, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	stats, err := r.GetPlayerStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	games, err := r.GetGameHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	profileUser := ProfileUser{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		MemberSince: user.CreatedAt,
	}
	if includeEmail {
		profileUser.Email = user.Email
	}

	return &PlayerProfile{
		User:  profileUser,
		Stats: *stats,
		Games: games,
	}, nil
}

// ---------- Game Events ----------

// RecordEvent appends an event to the game_events log.
func (r *Repository) RecordEvent(ctx context.Context, gameID string, userID *string, eventType string, eventData []byte) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO game_events (game_id, user_id, event_type, event_data)
		 VALUES ($1, $2, $3, $4)`,
		gameID, userID, eventType, eventData,
	)
	if err != nil {
		return fmt.Errorf("record event: %w", err)
	}
	return nil
}

// ---------- Helpers ----------

// toNullString converts an empty string to a sql.NullString with Valid=false.
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
