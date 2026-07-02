import { Router } from 'express'
import { authMiddleware, staffMiddleware } from '../middleware/auth.js'
import { config } from '../config.js'
import * as roomService from '../services/roomService.js'

const router = Router()

router.get('/health', (_req, res) => {
  res.json({ status: 'ok', service: 'caro-service' })
})

router.use(authMiddleware)

router.get('/rooms/public', async (_req, res) => {
  try {
    const rooms = await roomService.listPublicRooms()
    res.json({ items: rooms })
  } catch (err) {
    res.status(500).json({ error: err.message })
  }
})

router.get('/rooms/:code', async (req, res) => {
  try {
    const room = await roomService.getRoomByCode(req.params.code)
    if (!room) return res.status(404).json({ error: 'room not found' })
    res.json(serializeRoomPublic(room))
  } catch (err) {
    res.status(500).json({ error: err.message })
  }
})

router.post('/rooms', async (req, res) => {
  try {
    const visibility = req.body.visibility === 'private' ? 'private' : 'public'
    const boardSize = [15, 20].includes(Number(req.body.board_size)) ? Number(req.body.board_size) : 15
    const room = await roomService.createRoom({
      hostUserId: req.user.userId,
      hostName: req.body.host_name || req.user.name || req.user.email,
      visibility,
      boardSize,
    })
    res.status(201).json(room)
  } catch (err) {
    res.status(500).json({ error: err.message })
  }
})

router.get('/matches/history', async (req, res) => {
  try {
    const items = await roomService.getMatchHistory(req.user.userId)
    res.json({ items })
  } catch (err) {
    res.status(500).json({ error: err.message })
  }
})

router.get('/webrtc/config', (_req, res) => {
  res.json({
    turn_url: config.turnUrl,
    turn_username: config.turnUsername,
    turn_credential: config.turnCredential,
    stun_urls: ['stun:stun.l.google.com:19302'],
  })
})

router.get('/admin/rooms', staffMiddleware, async (_req, res) => {
  try {
    const items = await roomService.listActiveRoomsAdmin()
    res.json({ items })
  } catch (err) {
    res.status(500).json({ error: err.message })
  }
})

function serializeRoomPublic(room) {
  return {
    id: room.id,
    code: room.code,
    board_size: room.board_size,
    status: room.status,
    visibility: room.visibility,
    host_name: room.host_name,
    player_x_name: room.player_x_name,
    player_o_name: room.player_o_name,
    current_turn: room.current_turn,
    winner_id: room.winner_id,
    created_at: room.created_at,
  }
}

export default router
