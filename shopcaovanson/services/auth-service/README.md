# Auth Service

## Trách nhiệm
Service xác thực người dùng: đăng ký, đăng nhập, phát hành JWT RS256, quản lý refresh token qua Redis, và cung cấp API profile cá nhân.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Go | 1.22 | Ngôn ngữ chính |
| Fiber | v2 | HTTP framework |
| PostgreSQL | 15 | Lưu users và refresh tokens |
| Redis | 7 | Cache refresh token (TTL 7 ngày) |
| Kafka | 3.6 | Publish sự kiện user.registered |
| golang-migrate | v4 | Database migrations |
| sqlx | — | Truy vấn PostgreSQL |
| golang-jwt | v5 | JWT RS256 signing |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|-------------------|
| PORT | Cổng HTTP | 8081 |
| DATABASE_URL | PostgreSQL connection string | — (bắt buộc) |
| REDIS_URL | Redis connection string | redis://localhost:6379/0 |
| KAFKA_BROKERS | Kafka broker addresses | localhost:9092 |
| JWT_PRIVATE_KEY | RSA private key PEM (PKCS#8) | — (bắt buộc) |
| ACCESS_TOKEN_TTL_MINUTES | Thời hạn access token | 15 |
| REFRESH_TOKEN_TTL_DAYS | Thời hạn refresh token | 7 |
| BCRYPT_COST | Chi phí bcrypt hash | 12 |
| MIGRATIONS_PATH | Đường dẫn migrations | file://migrations |

## Chạy local

```bash
cp .env.example .env
docker-compose up --build
```

## Migration

```bash
# Up
go run cmd/migrate/main.go up

# Down
go run cmd/migrate/main.go down
```

## API endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| POST | /api/auth/register | Không | Đăng ký tài khoản mới |
| POST | /api/auth/login | Không | Đăng nhập, trả access + refresh token |
| POST | /api/auth/refresh | Không | Đổi refresh token lấy access token mới |
| POST | /api/auth/logout | Bearer JWT | Thu hồi refresh token |
| GET | /api/auth/me | Bearer JWT | Xem profile |
| PUT | /api/auth/me | Bearer JWT | Cập nhật full_name |
| GET | /health | Không | Health check |

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| user.registered | Produce | Gửi khi user đăng ký mới (user_id, email, full_name) |

## Fake data

```bash
export DATABASE_URL=postgres://auth:authpass@localhost:5432/authdb?sslmode=disable
go run scripts/fake_data.go
```

Kết quả: 50 users (1 admin `admin@shop.com` / `Admin@123` + 49 customers), tên tiếng Việt, idempotent (chạy lại không tạo duplicate).
