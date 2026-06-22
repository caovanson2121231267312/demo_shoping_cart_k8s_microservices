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
| SUPPORT_STAFF_USER_ID | UUID tài khoản support@shop.com | d0773246-879b-522b-b5f9-35e4722abc4e |

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
| POST | /api/chat/rooms/support | JWT / X-User-Id | Phòng hỗ trợ khách ↔ nhân viên |
| GET | /api/chat/admin/support-rooms | JWT (staff+) | Danh sách phòng support (admin) |
| POST | /api/chat/admin/support-rooms | JWT (staff+) | Mở phòng với customer_id |
| POST | /api/chat/upload | JWT | Upload ảnh (max 5MB) |
| GET | /api/chat/media/:filename | JWT | Xem ảnh đã upload |
| GET | /api/chat/rooms/:id/messages | JWT / X-User-Id | Lịch sử tin nhắn (cursor) |
| WS | /ws?token={jwt} | JWT query param | WebSocket real-time |

### WebSocket protocol

Client → server: `join`, `message` (msg_type: text|image), `typing`, `reaction`  
Server → client: `message`, `user_joined`, `typing`, `reaction`

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| — | — | Chat service không dùng Kafka |

## Fake data

```bash
go run scripts/fake_data.go
```

Kết quả: 20 phòng chat và 400 tin nhắn. Script idempotent — chạy lại sẽ bỏ qua nếu đã seed.

## Tài khoản test chat nhân viên

| Vai trò | Email | Mật khẩu |
|---------|-------|----------|
| Nhân viên hỗ trợ | support@shop.com | Support@123 |
| Khách hàng | user1@shop.com | Customer@123 |

- Khách: widget chat → tab **Nhân viên** (cần đăng nhập)
- Admin: `/admin/chat` → danh sách khách online + hội thoại
