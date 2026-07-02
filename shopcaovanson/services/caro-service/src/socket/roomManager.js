import { createBoard, validateMove, checkWin, isBoardFull, resolveRps } from '../game/engine.js'
import * as db from '../services/roomService.js'

export class RoomManager {
  constructor(io) {
    this.io = io
    /** @type {Map<string, LiveRoom>} */
    this.rooms = new Map()
    /** socketId -> { roomId, userId } */
    this.sockets = new Map()
  }

  getLive(roomId) {
    return this.rooms.get(roomId)
  }

  async attachSocket(socket, user) {
    this.sockets.set(socket.id, { userId: user.userId, roomId: null })
    socket.data.user = user
  }

  detachSocket(socketId) {
    const meta = this.sockets.get(socketId)
    if (meta?.roomId) {
      const live = this.rooms.get(meta.roomId)
      if (live) {
        live.players.delete(meta.userId)
        this.io.to(meta.roomId).emit('player_left', { user_id: meta.userId })
        if (live.players.size === 0) {
          this.rooms.delete(meta.roomId)
        }
      }
    }
    this.sockets.delete(socketId)
  }

  async joinRoom(socket, roomCode) {
    const room = await db.getRoomByCode(roomCode)
    if (!room) throw new Error('room not found')
    if (room.status === 'finished') throw new Error('room finished')

    const user = socket.data.user
    const isPlayer = [room.player_x_id, room.player_o_id].includes(user.userId)

    if (!isPlayer) {
      const filled = [room.player_x_id, room.player_o_id].filter(Boolean).length
      if (filled >= 2) throw new Error('room full')
      if (!room.player_x_id) {
        await db.updateRoom(room.id, { player_x_id: user.userId, player_x_name: user.name })
        room.player_x_id = user.userId
        room.player_x_name = user.name
      } else if (!room.player_o_id && user.userId !== room.player_x_id) {
        await db.updateRoom(room.id, { player_o_id: user.userId, player_o_name: user.name })
        room.player_o_id = user.userId
        room.player_o_name = user.name
      }
    }

    let live = this.rooms.get(room.id)
    const updated = await db.getRoomById(room.id)

    if (!live) {
      live = {
        id: room.id,
        code: room.code,
        boardSize: room.board_size,
        status: updated.status,
        board: updated.status === 'playing' ? createBoard(room.board_size) : null,
        players: new Map(),
        rpsChoices: new Map(),
        matchId: updated.match_id,
        playerXId: updated.player_x_id,
        playerOId: updated.player_o_id,
        currentTurn: updated.current_turn || 'X',
        moves: [],
      }
      this.rooms.set(room.id, live)
    }

    live.playerXId = updated.player_x_id
    live.playerOId = updated.player_o_id
    live.status = updated.status

    live.players.set(user.userId, {
      socketId: socket.id,
      name: user.name || user.email,
      role: user.role,
    })

    const meta = this.sockets.get(socket.id)
    if (meta) meta.roomId = room.id

    socket.join(room.id)

    const chat = await db.getChatHistory(room.id)

    this.broadcastRoomState(room.id)

    return { room: this.serializeRoom(updated, user.userId), chat, live }
  }

  async startRpsIfReady(roomId) {
    const live = this.rooms.get(roomId)
    if (!live) return
    const room = await db.getRoomById(roomId)
    if (!room.player_x_id || !room.player_o_id) return
    if (room.status !== 'waiting') return

    live.status = 'rps'
    live.rpsChoices.clear()
    await db.updateRoom(roomId, { status: 'rps' })
    this.io.to(roomId).emit('rps_start', { message: 'Oẳn tù xì để chọn X/O' })
    this.broadcastRoomState(roomId)
  }

  async handleRps(roomId, userId, choice) {
    const valid = ['rock', 'paper', 'scissors']
    if (!valid.includes(choice)) throw new Error('invalid rps choice')

    const live = this.rooms.get(roomId)
    if (!live || live.status !== 'rps') throw new Error('not in rps phase')

    live.rpsChoices.set(userId, choice)
    this.io.to(roomId).emit('rps_choice', { user_id: userId, ready: true })

    if (live.rpsChoices.size < 2) return null

    const room = await db.getRoomById(roomId)
    const [pX, pO] = [room.player_x_id, room.player_o_id]
    const cX = live.rpsChoices.get(pX)
    const cO = live.rpsChoices.get(pO)
    const result = resolveRps(cX, cO)

    if (result === 'tie') {
      live.rpsChoices.clear()
      this.io.to(roomId).emit('rps_tie', { message: 'Hòa — chơi lại oẳn tù xì' })
      return null
    }

    let xId = pX
    let oId = pO
    let xName = room.player_x_name
    let oName = room.player_o_name

    if (result === 'b') {
      ;[xId, oId] = [pO, pX]
      ;[xName, oName] = [room.player_o_name, room.player_x_name]
    }

    const matchId = await db.createMatch(room, xId, oId)
    live.matchId = matchId
    live.playerXId = xId
    live.playerOId = oId
    live.board = createBoard(room.board_size)
    live.status = 'playing'
    live.currentTurn = 'X'
    live.moves = []
    live.rpsChoices.clear()

    await db.updateRoom(roomId, {
      status: 'playing',
      player_x_id: xId,
      player_o_id: oId,
      player_x_name: xName,
      player_o_name: oName,
      current_turn: 'X',
      match_id: matchId,
    })

    const payload = {
      player_x_id: xId,
      player_o_id: oId,
      player_x_name: xName,
      player_o_name: oName,
      rps: { [pX]: cX, [pO]: cO },
      current_turn: 'X',
    }
    this.io.to(roomId).emit('game_start', payload)
    this.broadcastRoomState(roomId)
    return payload
  }

  async handleMove(roomId, userId, row, col) {
    const live = this.rooms.get(roomId)
    if (!live || live.status !== 'playing') throw new Error('game not in progress')

    const symbol = live.playerXId === userId ? 'X' : live.playerOId === userId ? 'O' : null
    if (!symbol) throw new Error('not a player')
    if (live.currentTurn !== symbol) throw new Error('not your turn')

    const err = validateMove(live.board, live.boardSize, row, col, symbol, true)
    if (err) throw new Error(err)

    live.board[row][col] = symbol
    live.moves.push({ row, col, symbol, user_id: userId })

    const won = checkWin(live.board, live.boardSize, row, col, symbol)
    const draw = !won && isBoardFull(live.board)

    if (won || draw) {
      const winnerId = won ? userId : null
      const winReason = won ? 'five_in_row' : 'draw'
      live.status = 'finished'
      await db.finishMatch(live.matchId, winnerId, winReason, live.moves)
      await db.updateRoom(roomId, {
        status: 'finished',
        winner_id: winnerId,
        current_turn: null,
      })
      this.io.to(roomId).emit('game_over', {
        winner_id: winnerId,
        reason: winReason,
        board: live.board,
        moves: live.moves,
      })
    } else {
      live.currentTurn = symbol === 'X' ? 'O' : 'X'
      await db.updateRoom(roomId, { current_turn: live.currentTurn })
      this.io.to(roomId).emit('move', {
        row,
        col,
        symbol,
        user_id: userId,
        current_turn: live.currentTurn,
        board: live.board,
      })
    }

    this.broadcastRoomState(roomId)
    return { won, draw }
  }

  async handleChat(roomId, userId, userName, content) {
    const msg = await db.saveChatMessage(roomId, userId, userName, content)
    this.io.to(roomId).emit('chat', {
      id: msg.id,
      user_id: userId,
      user_name: userName,
      content: msg.content,
      created_at: msg.created_at,
    })
    return msg
  }

  /** WebRTC signaling relay */
  relaySignal(roomId, fromUserId, payload) {
    const live = this.rooms.get(roomId)
    if (!live) return
    for (const [uid, p] of live.players) {
      if (uid !== fromUserId) {
        this.io.to(p.socketId).emit('webrtc_signal', { from: fromUserId, ...payload })
      }
    }
  }

  broadcastRoomState(roomId) {
    const live = this.rooms.get(roomId)
    if (!live) return
    this.io.to(roomId).emit('room_state', {
      id: roomId,
      status: live.status,
      board: live.board,
      current_turn: live.currentTurn,
      moves: live.moves,
      players: [...live.players.entries()].map(([id, p]) => ({ id, name: p.name })),
    })
  }

  serializeRoom(room, viewerId) {
    return {
      id: room.id,
      code: room.code,
      board_size: room.board_size,
      status: room.status,
      visibility: room.visibility,
      host_user_id: room.host_user_id,
      host_name: room.host_name,
      player_x_id: room.player_x_id,
      player_o_id: room.player_o_id,
      player_x_name: room.player_x_name,
      player_o_name: room.player_o_name,
      current_turn: room.current_turn,
      winner_id: room.winner_id,
      you_are: room.player_x_id === viewerId ? 'X' : room.player_o_id === viewerId ? 'O' : null,
      is_host: room.host_user_id === viewerId,
      created_at: room.created_at,
    }
  }
}
