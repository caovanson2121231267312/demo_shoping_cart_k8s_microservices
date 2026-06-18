# Notification Service

## Trách nhiệm
Lắng nghe sự kiện Kafka và gửi email thông báo (xác nhận đơn hàng, chào mừng người dùng mới) qua SMTP.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Python | 3.12 | Runtime |
| FastAPI | 0.111 | Health endpoint |
| confluent-kafka | 2.4 | Kafka consumer |
| Jinja2 | 3.1 | Email templates |
| SMTP | — | Gửi email |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|------------------|
| PORT | Cổng HTTP | 8085 |
| KAFKA_BOOTSTRAP_SERVERS | Kafka brokers | localhost:9092 |
| KAFKA_TOPIC_ORDER_CREATED | Topic đơn hàng | order.created |
| KAFKA_TOPIC_USER_REGISTERED | Topic đăng ký | user.registered |
| KAFKA_GROUP_ORDER | Consumer group đơn hàng | notification-order |
| KAFKA_GROUP_USER | Consumer group user | notification-user |
| SMTP_HOST | SMTP server | localhost |
| SMTP_PORT | SMTP port | 587 |
| SMTP_USERNAME | SMTP user | — |
| SMTP_PASSWORD | SMTP password | — |
| SMTP_FROM | Địa chỉ gửi | noreply@shopcaovanson.xyz |
| SMTP_USE_TLS | Bật TLS | true |

## Chạy local

```bash
docker-compose up --build
```

## Migration

Service này không dùng database — không cần migration.

## API endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | /health | Không | Health check |

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| order.created | Consume | Gửi email xác nhận đơn hàng |
| user.registered | Consume | Gửi email chào mừng |

## Fake data

Service này không có fake data script. Test bằng cách publish message vào Kafka topic tương ứng.
