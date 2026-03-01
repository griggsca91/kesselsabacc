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
// otherwise it inserts a new one.
func (r *Repository) SaveGameState(ctx context.Context, roomID string, round int, phase string, stateJSON []byte) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO games (room_id, round_number, phase, state_json)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE
		   SET round_number = EXCLUDED.round_number,
		       phase = EXCLUDED.phase,
		       state_json = EXCLUDED.state_json`,
		roomID, round, phase, stateJSON,
	)
	if err != nil {
		return fmt.Errorf("save game state: %w", err)
	}
	return nil
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
			`INSERT INTO game_players (game_id, user_id, final_chips, is_winner)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (game_id, user_id) DO UPDATE
			   SET final_chips = EXCLUDED.final_chips,
			       is_winner = EXCLUDED.is_winner`,
			gameID, pr.UserID, pr.FinalChips, pr.IsWinner,
		)
		if err != nil {
			return fmt.Errorf("insert player result: %w", err)
		}
	}

	return tx.Commit()
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
