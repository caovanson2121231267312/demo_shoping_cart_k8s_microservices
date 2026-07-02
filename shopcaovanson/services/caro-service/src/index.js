import http from 'http'
import express from 'express'
import cors from 'cors'
import { Server } from 'socket.io'
import { config } from './config.js'
import apiRouter from './routes/api.js'
import { RoomManager } from './socket/roomManager.js'
import { registerSocketHandlers } from './socket/handlers.js'

const app = express()
app.use(cors({ origin: config.corsOrigins, credentials: true }))
app.use(express.json())

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', service: 'caro-service' })
})

app.use('/api/caro', apiRouter)

const server = http.createServer(app)
const io = new Server(server, {
  path: '/ws/caro/socket.io',
  cors: {
    origin: config.corsOrigins,
    credentials: true,
  },
})

const roomManager = new RoomManager(io)
registerSocketHandlers(io, roomManager)

server.listen(config.port, () => {
  console.log(`caro-service listening on :${config.port}`)
})
