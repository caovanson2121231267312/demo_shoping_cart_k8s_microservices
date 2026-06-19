# Rasa Chatbot Service

Trợ lý ảo hỏi đáp sản phẩm, đơn hàng, bảo hành, giờ làm việc cho Shop Cao Văn Sơn.

## Chạy local (khuyến nghị)

```powershell
cd services/rasa-service
pip install -r requirements.txt
python -m uvicorn main:app --host 0.0.0.0 --port 8090
```

Hoặc dùng `.\scripts\start-local-services.ps1` (port **8090**).

Mặc định dùng **bot engine tích hợp** (không cần Docker Rasa). Bật Rasa đầy đủ:

```env
USE_RASA=true
RASA_URL=http://localhost:5005
```

## API

| Method | Path | Mô tả |
|--------|------|-------|
| GET | `/health` | Health check |
| POST | `/api/chat` | Chat nội bộ (chat-service gọi) |
| POST | `/api/chatbot/message` | Chat công khai qua API Gateway |

Body:
```json
{
  "sender_id": "user-uuid-hoac-guest-id",
  "message": "tìm laptop",
  "metadata": { "access_token": "..." }
}
```

## Docker (Rasa + Actions + Bridge)

```bash
cd services/rasa-service
docker compose up -d
# Train model lần đầu:
docker compose run --rm rasa train
docker compose restart rasa
```

## Tích hợp

- **chat-service** gọi `RASA_SERVICE_URL` sau mỗi tin nhắn phòng `support`
- **api-gateway** proxy `/api/chatbot/*` → rasa-service (public)
- **Frontend** ChatWidget: khách dùng REST, user đăng nhập dùng WebSocket + bot
