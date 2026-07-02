import { verifyToken } from '../middleware/auth.js'

export function registerSocketHandlers(io, roomManager) {
  io.use((socket, next) => {
    try {
      const token =
        socket.handshake.auth?.token ||
        socket.handshake.query?.token ||
        socket.handshake.headers?.authorization?.replace(/^Bearer\s+/i, '')
      const user = verifyToken(token)
      socket.data.user = {
        userId: user.userId,
        email: user.email,
        role: user.role,
        name: user.name,
      }
      next()
    } catch (err) {
      next(new Error('unauthorized'))
    }
  })

  io.on('connection', async (socket) => {
    const user = socket.data.user
    await roomManager.attachSocket(socket, user)

    socket.emit('connected', { user_id: user.userId })

    socket.on('join_room', async ({ code }, ack) => {
      try {
        const result = await roomManager.joinRoom(socket, code)
        await roomManager.startRpsIfReady(result.room.id)
        ack?.({ ok: true, ...result })
      } catch (err) {
        ack?.({ ok: false, error: err.message })
      }
    })

    socket.on('rps', async ({ choice }, ack) => {
      try {
        const meta = roomManager.sockets.get(socket.id)
        if (!meta?.roomId) throw new Error('not in room')
        const result = await roomManager.handleRps(meta.roomId, user.userId, choice)
        ack?.({ ok: true, result })
      } catch (err) {
        ack?.({ ok: false, error: err.message })
      }
    })

    socket.on('move', async ({ row, col }, ack) => {
      try {
        const meta = roomManager.sockets.get(socket.id)
        if (!meta?.roomId) throw new Error('not in room')
        const result = await roomManager.handleMove(meta.roomId, user.userId, row, col)
        ack?.({ ok: true, result })
      } catch (err) {
        ack?.({ ok: false, error: err.message })
      }
    })

    socket.on('chat', async ({ content }, ack) => {
      try {
        const meta = roomManager.sockets.get(socket.id)
        if (!meta?.roomId) throw new Error('not in room')
        await roomManager.handleChat(meta.roomId, user.userId, user.name, content)
        ack?.({ ok: true })
      } catch (err) {
        ack?.({ ok: false, error: err.message })
      }
    })

    socket.on('webrtc_signal', (payload) => {
      const meta = roomManager.sockets.get(socket.id)
      if (!meta?.roomId) return
      roomManager.relaySignal(meta.roomId, user.userId, payload)
    })

    socket.on('disconnect', () => {
      roomManager.detachSocket(socket.id)
    })
  })
}
