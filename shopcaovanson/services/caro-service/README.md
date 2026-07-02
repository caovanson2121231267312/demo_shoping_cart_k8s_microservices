# caro-service

Game cờ Caro realtime — Node.js + Socket.io + PostgreSQL.

## Chạy local

```bash
cp .env.example .env
# Cấu hình DATABASE_URL, JWT_PUBLIC_KEY (cùng public key auth-service)
npm install
npm run migrate
npm run dev
```

- REST: `http://localhost:8087/api/caro`
- Socket.io: `http://localhost:8087` path `/ws/caro/socket.io`

## API

| Method | Path | Mô tả |
|--------|------|-------|
| GET | /api/caro/rooms/public | Phòng công khai |
| POST | /api/caro/rooms | Tạo phòng |
| GET | /api/caro/rooms/:code | Thông tin phòng |
| GET | /api/caro/admin/rooms | Admin — phòng đang chơi |
| GET | /api/caro/webrtc/config | ICE/TURN config |

## Socket events

Client → server: `join_room`, `rps`, `move`, `chat`, `webrtc_signal`  
Server → client: `game_start`, `move`, `game_over`, `rps_start`, `chat`, `room_state`
