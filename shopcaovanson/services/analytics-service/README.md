# Analytics Service

Python FastAPI — báo cáo admin, biểu đồ, user online, xuất Excel, Kafka consumer.

## Chạy local

```powershell
cd services/analytics-service
copy .env.example .env
pip install -r requirements.txt
# Kafka (Linux/Docker): pip install confluent-kafka
python -m uvicorn main:app --host 0.0.0.0 --port 8086
```

Thêm vào `api-gateway/.env`:
```
ANALYTICS_SERVICE_URL=http://localhost:8086
```

## API (role >= manager)

| Endpoint | Mô tả |
|----------|--------|
| GET /api/admin/analytics/overview?period=day\|week\|month\|quarter | Tổng quan KPI |
| GET /api/admin/analytics/revenue | Biểu đồ doanh thu |
| GET /api/admin/analytics/users | Khách đăng ký mới |
| GET /api/admin/analytics/reviews | Review & điểm TB |
| GET /api/admin/analytics/products/top | Top sản phẩm |
| GET /api/admin/analytics/online | User đang online |
| GET /api/admin/analytics/export/excel | Tải file .xlsx |
| POST /api/analytics/presence | Heartbeat online (frontend) |

## Kafka topics

- `order.created` — cập nhật tương tác realtime
- `user.registered` — đăng ký mới
- `analytics.event` — page view (tùy chọn)
