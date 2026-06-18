# Chat Service

## Trách nhiệm
Cung cấp chat real-time giữa khách hàng và support qua WebSocket, lưu lịch sử tin nhắn trên MongoDB và đồng bộ giữa các pod bằng Redis pub/sub.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Go | 1.22 | Runtime |
| Fiber | v2 | HTTP framework |
| MongoDB | 7 | Lưu phòng chat và tin nhắn |
| Redis | 7 | Pub/sub cross-pod |
| WebSocket | — | Real-time messaging |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|------------------|
| PORT | Cổng HTTP | 8084 |
| MONGODB_URI | MongoDB connection string | mongodb://localhost:27017 |
| MONGODB_DATABASE | Tên database | shop_chat |
| REDIS_URL | Redis connection string | redis://localhost:6379/0 |
| JWT_PUBLIC_KEY | RSA public key (PEM) để verify JWT | — |

## Chạy local

```bash
docker-compose up --build
```

## Migration

```bash
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
```

## API endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | /health | Không | Health check |
| GET | /api/chat/rooms | JWT / X-User-Id | Danh sách phòng của tôi |
| POST | /api/chat/rooms | JWT / X-User-Id | Tạo phòng direct với user khác |
| GET | /api/chat/rooms/:id/messages | JWT / X-User-Id | Lịch sử tin nhắn (cursor) |
| WS | /ws?token={jwt} | JWT query param | WebSocket real-time |

### WebSocket protocol

Client → server: `join`, `message`, `typing`  
Server → client: `message`, `user_joined`, `typing`

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| — | — | Chat service không dùng Kafka |

## Fake data

```bash
go run scripts/fake_data.go
```

Kết quả: 20 phòng chat và 400 tin nhắn. Script idempotent — chạy lại sẽ bỏ qua nếu đã seed.
