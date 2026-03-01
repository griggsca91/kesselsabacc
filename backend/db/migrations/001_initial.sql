CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE,
    display_name TEXT NOT NULL,
    password_hash TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code CHAR(4) UNIQUE NOT NULL,
    host_user_id UUID REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'waiting', -- waiting, playing, finished
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES rooms(id) NOT NULL,
    round_number INT NOT NULL DEFAULT 1,
    phase TEXT NOT NULL DEFAULT 'lobby',
    state_json JSONB, -- serialized game state for recovery
    started_at TIMESTAMPTZ DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS game_players (
    game_id UUID REFERENCES games(id) NOT NULL,
    user_id UUID REFERENCES users(id) NOT NULL,
    final_chips INT,
    is_winner BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (game_id, user_id)
);

CREATE TABLE IF NOT EXISTS game_events (
    id BIGSERIAL PRIMARY KEY,
    game_id UUID REFERENCES games(id) NOT NULL,
    user_id UUID REFERENCES users(id),
    event_type TEXT NOT NULL,
    event_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rooms_code ON rooms(code);
CREATE INDEX idx_games_room ON games(room_id);
CREATE INDEX idx_game_events_game ON game_events(game_id);
CREATE INDEX idx_game_players_user ON game_players(user_id);
