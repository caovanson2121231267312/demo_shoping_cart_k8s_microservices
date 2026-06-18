# Search Service

## Trách nhiệm
Đồng bộ sản phẩm từ Kafka vào Elasticsearch và cung cấp API tìm kiếm nội bộ cho product-service.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| Python | 3.12 | Runtime |
| FastAPI | 0.111 | HTTP API |
| confluent-kafka | 2.4 | Kafka consumer |
| elasticsearch-py | 8.13 | Search index & query |
| Elasticsearch | 8 | Search engine |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|------------------|
| PORT | Cổng HTTP | 8086 |
| KAFKA_BOOTSTRAP_SERVERS | Kafka brokers | localhost:9092 |
| KAFKA_TOPIC_PRODUCT_CREATED | Topic tạo sản phẩm | product.created |
| KAFKA_TOPIC_PRODUCT_UPDATED | Topic cập nhật sản phẩm | product.updated |
| KAFKA_GROUP_PRODUCT | Consumer group | search-product |
| ELASTICSEARCH_URL | Elasticsearch URL | http://localhost:9200 |

## Chạy local

```bash
docker-compose up --build
```

## Migration

Index `products` được tạo tự động khi service khởi động với mapping theo chuẩn dự án.

## API endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | /health | Không | Health check |
| GET | /search | Không | Tìm kiếm sản phẩm |

### GET /search query params

| Param | Mô tả |
|-------|-------|
| q | Từ khóa tìm kiếm |
| category | Lọc theo danh mục |
| min_price | Giá tối thiểu |
| max_price | Giá tối đa |
| page | Trang (mặc định 1) |
| limit | Số kết quả/trang (mặc định 20) |

## Kafka

| Topic | Produce/Consume | Mô tả |
|-------|-----------------|-------|
| product.created | Consume | Index sản phẩm mới |
| product.updated | Consume | Cập nhật document sản phẩm |

## Fake data

Service này không có fake data script. Dữ liệu được đồng bộ từ product-service qua Kafka topic `product.created` / `product.updated`.
