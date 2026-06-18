# Product Service

## Trách nhiệm
Quản lý danh mục, sản phẩm, đánh giá và tìm kiếm sản phẩm cho hệ thống shopcaovanson.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Go | 1.22 | Runtime |
| Fiber | v2 | HTTP framework |
| PostgreSQL | 15 | categories, products, product_reviews |
| MongoDB | 7 | product_details (rich content) |
| Elasticsearch | 8 | Full-text search |
| Kafka | 3.6 | product.created, product.updated events |
| golang-migrate | v4 | Database migrations |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|------------------|
| APP_PORT | HTTP port | 8082 |
| APP_ENV | Môi trường | development |
| POSTGRES_HOST | PostgreSQL host | localhost |
| POSTGRES_PORT | PostgreSQL port | 5432 |
| POSTGRES_USER | PostgreSQL user | shop |
| POSTGRES_PASSWORD | PostgreSQL password | shop_secret |
| POSTGRES_DB | Database name | product_db |
| POSTGRES_SSLMODE | SSL mode | disable |
| MONGO_URI | MongoDB connection URI | mongodb://localhost:27017 |
| MONGO_DB | MongoDB database | shop |
| ELASTICSEARCH_URL | Elasticsearch URL | http://localhost:9200 |
| ELASTICSEARCH_INDEX | ES index name | products |
| KAFKA_BROKERS | Kafka brokers (comma-separated) | localhost:9092 |

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
| GET | /api/categories | No | Danh mục dạng cây |
| GET | /api/products | No | Danh sách sản phẩm (phân trang, lọc, tìm kiếm) |
| GET | /api/products/:slug | No | Chi tiết sản phẩm theo slug |
| POST | /api/products | Admin | Tạo sản phẩm |
| PUT | /api/products/:id | Admin | Cập nhật sản phẩm |
| DELETE | /api/products/:id | Admin | Xóa sản phẩm |
| GET | /api/products/:id/reviews | No | Danh sách đánh giá |
| POST | /api/products/:id/reviews | Auth | Tạo đánh giá |
| GET | /internal/products/:id | Internal | Thông tin stock/price cho order-service |

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| product.created | Produce | Khi tạo sản phẩm mới |
| product.updated | Produce | Khi cập nhật sản phẩm |

## Fake data

```bash
go run scripts/fake_data.go
```

Kết quả mong đợi: 10 categories, 200 products, 500 reviews (idempotent — chạy lại không duplicate).
