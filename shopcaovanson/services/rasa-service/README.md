# Rasa Chatbot Service

Trợ lý ảo hỏi đáp sản phẩm, đơn hàng, bảo hành cho Shop Cao Văn Sơn.

## Kiến trúc

| Thành phần | Port | Vai trò |
|------------|------|---------|
| **rasa** (Docker) | 5005 | NLU + dialogue, **load model** `chatbot.tar.gz` |
| **rasa-actions** (Docker) | 5055 | Custom actions (tìm sản phẩm, v.v.) |
| **rasa-service** (Python) | 8090 | Bridge → gọi Rasa, fallback `bot_engine.py` |

Mặc định `USE_RASA=true` — ưu tiên Rasa ML, không có model thì fallback bot engine.

## Train + chạy model (khuyến nghị)

```powershell
# Từ thư mục gốc repo
.\scripts\train-rasa.ps1 -Start

# Train lại sau khi sửa nlu.yml / domain.yml
.\scripts\train-rasa.ps1 -Force -Start
```

Model lưu tại: `services/rasa-service/rasa/models/chatbot.tar.gz`

## Chạy local đầy đủ

```powershell
.\scripts\start-local-services.ps1
```

Script sẽ tự:
1. Train model nếu chưa có
2. Start Docker Rasa + actions
3. Start `rasa-service` bridge (USE_RASA=true)

Hoặc chỉ bridge (cần Rasa đang chạy):

```powershell
cd services/rasa-service
pip install -r requirements.txt
copy .env.example .env
python -m uvicorn main:app --host 0.0.0.0 --port 8090
```

## Docker Compose (toàn bộ stack)

```bash
cd services/rasa-service
docker compose up -d --build
```

- Lần đầu: entrypoint tự **train** rồi **load** `models/chatbot.tar.gz`
- Train lại: `FORCE_RASA_TRAIN=true docker compose up -d --build rasa`

## API

| Method | Path | Mô tả |
|--------|------|-------|
| GET | `/health` | `{ use_rasa: true/false }` |
| POST | `/api/chatbot/message` | Chat công khai |
| POST | `/api/chat` | Chat nội bộ (chat-service) |

## Tích hợp

- **api-gateway** proxy `/api/chatbot/*` → rasa-service
- **chat-service** gọi `RASA_SERVICE_URL` cho auto-reply bot
