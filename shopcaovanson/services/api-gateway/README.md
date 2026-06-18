# API Gateway

## Trách nhiệm
Cổng HTTP duy nhất ra ngoài: reverse proxy tới các microservice, xác thực JWT RS256, inject header `X-User-Id` và `X-User-Role`, rate limiting và CORS.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Go | 1.22 | Ngôn ngữ chính |
| Fiber | v2 | HTTP framework + reverse proxy |
| Redis | 7 | Rate limiting 100 req/phút theo IP |
| golang-jwt | v5 | JWT RS256 validation |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|-------------------|
| PORT | Cổng HTTP | 8080 |
| REDIS_URL | Redis connection string | redis://localhost:6379/0 |
| JWT_PUBLIC_KEY | RSA public key PEM | — (bắt buộc) |
| AUTH_SERVICE_URL | URL auth-service | http://localhost:8081 |
| PRODUCT_SERVICE_URL | URL product-service | http://localhost:8082 |
| ORDER_SERVICE_URL | URL order-service | http://localhost:8083 |
| CHAT_SERVICE_URL | URL chat-service | http://localhost:8084 |
| CORS_ALLOWED_ORIGINS | Origins được phép (phân cách bởi dấu phẩy) | https://shopcaovanson.xyz |
| RATE_LIMIT_PER_MINUTE | Giới hạn request theo IP | 100 |

## Chạy local

```bash
cp .env.example .env
docker-compose up --build
```

## Migration

Không áp dụng — service không dùng database.

## API endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| * | /api/auth/* | Tùy route | Proxy tới auth-service |
| * | /api/products/* | GET công khai, còn lại JWT | Proxy tới product-service |
| * | /api/categories | GET công khai | Proxy tới product-service |
| * | /api/cart/* | JWT | Proxy tới order-service |
| * | /api/orders/* | JWT | Proxy tới order-service |
| * | /api/admin/* | JWT | Proxy tới order-service |
| * | /api/chat/* | JWT | Proxy tới chat-service |
| * | /ws | JWT (query token) | Proxy WebSocket tới chat-service |
| GET | /health | Không | Health check |

## Kafka

Không áp dụng.

## Fake data

Không áp dụng.
