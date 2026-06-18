# Order Service

## Trách nhiệm
Quản lý giỏ hàng (Redis) và đơn hàng (PostgreSQL) cho hệ thống shopcaovanson.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Go | 1.22 | Runtime |
| Fiber | v2 | HTTP framework |
| PostgreSQL | 15 | orders, order_items |
| Redis | 7 | Giỏ hàng cart:{user_id} |
| Kafka | 3.6 | order.created events |
| golang-migrate | v4 | Database migrations |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|------------------|
| APP_PORT | HTTP port | 8083 |
| APP_ENV | Môi trường | development |
| POSTGRES_HOST | PostgreSQL host | localhost |
| POSTGRES_PORT | PostgreSQL port | 5432 |
| POSTGRES_USER | PostgreSQL user | shop |
| POSTGRES_PASSWORD | PostgreSQL password | shop_secret |
| POSTGRES_DB | Database name | order_db |
| POSTGRES_SSLMODE | SSL mode | disable |
| REDIS_HOST | Redis host | localhost |
| REDIS_PORT | Redis port | 6379 |
| REDIS_PASSWORD | Redis password | (empty) |
| REDIS_DB | Redis DB index | 0 |
| CART_TTL_HOURS | TTL giỏ hàng (giờ) | 720 (30 ngày) |
| PRODUCT_SERVICE_URL | URL product-service | http://localhost:8082 |
| KAFKA_BROKERS | Kafka brokers | localhost:9092 |

## Chạy local

```bash
docker-compose up --build
```

## Migration

```bash
# Up
go run cmd/migrate/main.go -direction up

# Down
go run cmd/migrate/main.go -direction down
```

## API endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | /health | No | Health check |
| GET | /api/cart | Auth | Lấy giỏ hàng |
| POST | /api/cart/items | Auth | Thêm sản phẩm vào giỏ |
| PUT | /api/cart/items/:productId | Auth | Cập nhật số lượng |
| DELETE | /api/cart/items/:productId | Auth | Xóa sản phẩm khỏi giỏ |
| DELETE | /api/cart | Auth | Xóa toàn bộ giỏ |
| POST | /api/orders | Auth | Checkout từ giỏ hàng |
| GET | /api/orders | Auth | Danh sách đơn của tôi |
| GET | /api/orders/:id | Auth | Chi tiết đơn hàng |
| PUT | /api/orders/:id/cancel | Auth | Hủy đơn (pending only) |
| GET | /api/admin/orders | Admin | Tất cả đơn (?status=) |
| PUT | /api/admin/orders/:id/status | Admin | Cập nhật trạng thái |

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| order.created | Produce | Khi đặt hàng thành công |

## Fake data

```bash
go run scripts/fake_data.go
```

Kết quả mong đợi: 300 orders với các status khác nhau (idempotent).
