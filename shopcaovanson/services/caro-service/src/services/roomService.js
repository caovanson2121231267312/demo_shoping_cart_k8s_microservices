import { randomBytes } from 'crypto'
import { v4 as uuidv4 } from 'uuid'
import { query } from '../db/pool.js'

function genCode() {
  return randomBytes(3).toString('hex').toUpperCase()
}

export async function createRoom({ hostUserId, hostName, visibility, boardSize }) {
  let code = genCode()
  for (let i = 0; i < 5; i++) {
    try {
      const { rows } = await query(
        `INSERT INTO caro_rooms (code, host_user_id, host_name, visibility, board_size, status)
         VALUES ($1, $2, $3, $4, $5, 'waiting')
         RETURNING *`,
        [code, hostUserId, hostName, visibility, boardSize],
      )
      return rows[0]
    } catch (err) {
      if (err.code !== '23505') throw err
      code = genCode()
    }
  }
  throw new Error('could not generate room code')
}

export async function getRoomByCode(code) {
  const { rows } = await query('SELECT * FROM caro_rooms WHERE code = $1', [code.toUpperCase()])
  return rows[0] || null
}

export async function getRoomById(id) {
  const { rows } = await query('SELECT * FROM caro_rooms WHERE id = $1', [id])
  return rows[0] || null
}

export async function listPublicRooms() {
  const { rows } = await query(
    `SELECT id, code, host_name, board_size, status, player_x_name, player_o_name, created_at
     FROM caro_rooms
     WHERE visibility = 'public' AND status IN ('waiting', 'rps', 'playing')
     ORDER BY created_at DESC
     LIMIT 50`,
  )
  return rows
}

export async function listActiveRoomsAdmin() {
  const { rows } = await query(
    `SELECT id, code, host_name, board_size, status, visibility,
            player_x_name, player_o_name, player_x_id, player_o_id, created_at, updated_at
     FROM caro_rooms
     WHERE status IN ('waiting', 'rps', 'playing')
     ORDER BY updated_at DESC
     LIMIT 100`,
  )
  return rows
}

export async function updateRoom(id, fields) {
  const keys = Object.keys(fields)
  const sets = keys.map((k, i) => `${k} = $${i + 2}`)
  const values = keys.map((k) => fields[k])
  const { rows } = await query(
    `UPDATE caro_rooms SET ${sets.join(', ')}, updated_at = NOW() WHERE id = $1 RETURNING *`,
    [id, ...values],
  )
  return rows[0]
}

export async function createMatch(room, playerXId, playerOId) {
  const matchId = uuidv4()
  await query(
    `INSERT INTO caro_matches (id, room_id, board_size, player_x_id, player_o_id)
     VALUES ($1, $2, $3, $4, $5)`,
    [matchId, room.id, room.board_size, playerXId, playerOId],
  )
  return matchId
}

export async function finishMatch(matchId, winnerId, winReason, moves) {
  await query(
    `UPDATE caro_matches
     SET winner_id = $2, win_reason = $3, moves_json = $4::jsonb, ended_at = NOW()
     WHERE id = $1`,
    [matchId, winnerId, winReason, JSON.stringify(moves)],
  )
}

export async function saveChatMessage(roomId, userId, userName, content) {
  const { rows } = await query(
    `INSERT INTO caro_chat_messages (room_id, user_id, user_name, content)
     VALUES ($1, $2, $3, $4)
     RETURNING *`,
    [roomId, userId, userName, content.trim().slice(0, 500)],
  )
  return rows[0]
}

export async function getChatHistory(roomId, limit = 50) {
  const { rows } = await query(
    `SELECT id, user_id, user_name, content, created_at
     FROM caro_chat_messages WHERE room_id = $1
     ORDER BY created_at DESC LIMIT $2`,
    [roomId, limit],
  )
  return rows.reverse()
}

export async function getMatchHistory(userId, limit = 20) {
  const { rows } = await query(
    `SELECT m.*, r.code AS room_code
     FROM caro_matches m
     JOIN caro_rooms r ON r.id = m.room_id
     WHERE m.player_x_id = $1 OR m.player_o_id = $1
     ORDER BY m.started_at DESC LIMIT $2`,
    [userId, limit],
  )
  return rows
}
