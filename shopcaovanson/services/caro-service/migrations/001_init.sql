CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS caro_rooms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code VARCHAR(8) UNIQUE NOT NULL,
  host_user_id UUID NOT NULL,
  host_name VARCHAR(255) NOT NULL DEFAULT '',
  visibility VARCHAR(10) NOT NULL DEFAULT 'public',
  board_size INT NOT NULL DEFAULT 15 CHECK (board_size IN (15, 20)),
  status VARCHAR(20) NOT NULL DEFAULT 'waiting',
  player_x_id UUID,
  player_o_id UUID,
  player_x_name VARCHAR(255),
  player_o_name VARCHAR(255),
  current_turn CHAR(1),
  winner_id UUID,
  match_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_caro_rooms_status ON caro_rooms (status);
CREATE INDEX IF NOT EXISTS idx_caro_rooms_visibility ON caro_rooms (visibility) WHERE status != 'finished';

CREATE TABLE IF NOT EXISTS caro_matches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL REFERENCES caro_rooms(id) ON DELETE CASCADE,
  board_size INT NOT NULL,
  player_x_id UUID NOT NULL,
  player_o_id UUID NOT NULL,
  winner_id UUID,
  win_reason VARCHAR(50),
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ended_at TIMESTAMPTZ,
  moves_json JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE TABLE IF NOT EXISTS caro_chat_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL REFERENCES caro_rooms(id) ON DELETE CASCADE,
  user_id UUID NOT NULL,
  user_name VARCHAR(255) NOT NULL DEFAULT '',
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_caro_chat_room ON caro_chat_messages (room_id, created_at);
