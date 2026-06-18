---
name: shopcaovanson
description: >
  Skill cho toàn bộ dự án demo e-commerce "shopcaovanson". Đọc file này trước
  KHI làm bất kỳ việc gì trong project: tạo file, sửa code, viết K8s manifest,
  thêm route, thêm component. Skill này là source of truth cho mọi convention,
  cấu trúc thư mục, tên biến, tên service, domain, tech stack, và scope tính
  năng. KHÔNG tự ý thêm tính năng ngoài scope. KHÔNG đổi tên service hay domain.
---

# shopcaovanson — Project Skill

## 1. Tổng quan dự án

Demo e-commerce đơn giản, đủ tính năng cốt lõi, chạy trên K8s multi-node.
**Không cần** payment gateway, voucher, đa ngôn ngữ, push notification mobile,
hay bất kỳ tính năng phức tạp nào khác — chỉ cần những thứ liệt kê trong §4.

| Mục          | Giá trị                                      |
|--------------|----------------------------------------------|
| Frontend     | https://shopcaovanson.xyz                    |
| Backend API  | https://shopapicaovanson.xyz                 |
| SSL          | cert-manager + Let's Encrypt (letsencrypt-prod) |
| K8s          | Multi-node: 1 master + 2 worker (KVM trên 1 VPS) |
| VPS spec     | 6 CPU · 12 GB RAM · 150 GB NVMe             |

---

## 2. Tech stack — KHÔNG thay thế, KHÔNG nâng cấp phiên bản tùy tiện

| Layer            | Technology                          | Version  |
|------------------|-------------------------------------|----------|
| Frontend         | NuxtJS 3 + Vuetify 3                | Nuxt 3.12, Vuetify 3.6 |
| State management | Pinia                               | latest   |
| Go services      | Go + Fiber v2                       | Go 1.22  |
| Python services  | Python + FastAPI + confluent-kafka  | Python 3.12 |
| Primary DB       | PostgreSQL                          | 15       |
| Document DB      | MongoDB                             | 7        |
| Cache / Pub-Sub  | Redis                               | 7        |
| Message broker   | Kafka + Zookeeper                   | Kafka 3.6 |
| Search           | Elasticsearch                       | 8        |
| Migrations       | golang-migrate/migrate v4           | —        |
| Container        | Docker (multi-stage builds)         | —        |
| Orchestration    | Kubernetes (kubeadm)                | 1.29     |
| CNI              | Calico                              | —        |
| Ingress          | nginx-ingress-controller            | —        |
| SSL              | cert-manager                        | 1.14     |
| Registry         | ghcr.io/caovanson/                  | —        |

---

## 3. Cấu trúc thư mục gốc — BẮT BUỘC giữ nguyên

```
shopcaovanson/
├── README.md
├── .gitignore
├── Makefile
│
├── services/
│   ├── api-gateway/          # Go · cổng duy nhất ra ngoài
│   ├── auth-service/         # Go · login · register · JWT
│   ├── product-service/      # Go · sản phẩm · danh mục · tìm kiếm
│   ├── order-service/        # Go · giỏ hàng · đơn hàng
│   ├── chat-service/         # Go · WebSocket real-time chat
│   ├── notification-service/ # Python · Kafka consumer · gửi email
│   └── search-service/       # Python · Kafka consumer · index Elasticsearch
│
├── frontend/
│   └── web/                  # NuxtJS 3 + Vuetify 3
│
├── k8s/
│   ├── cluster-setup/        # scripts KVM + kubeadm
│   ├── base/
│   │   ├── namespace.yaml
│   │   ├── cert-manager/
│   │   ├── ingress/
│   │   ├── infra/            # StatefulSets: postgres · mongo · redis · kafka · elastic
│   │   ├── api-gateway/
│   │   ├── auth-service/
│   │   ├── product-service/
│   │   ├── order-service/
│   │   ├── chat-service/
│   │   ├── notification-service/
│   │   ├── search-service/
│   │   └── frontend/
│   └── overlays/
│       └── prod/
│
├── scripts/
│   ├── setup-cluster.sh
│   ├── deploy-all.sh
│   ├── migrate-all.sh
│   └── seed-data.sh
│
└── docs/
    ├── architecture.md
    └── api-spec.md
```

---

## 4. Scope tính năng — CHỈ làm những thứ này, không thêm

### 4.1 Auth (auth-service)
- [x] Đăng ký: email + password + full_name
- [x] Đăng nhập: trả về access_token (JWT 15 phút) + refresh_token (7 ngày, lưu Redis)
- [x] Refresh token
- [x] Đăng xuất (revoke refresh token)
- [x] Xem và sửa profile cá nhân
- [ ] ~~Quên mật khẩu~~ — ngoài scope
- [ ] ~~OAuth Google/Facebook~~ — ngoài scope
- [ ] ~~2FA~~ — ngoài scope

### 4.2 Sản phẩm (product-service)
- [x] Danh sách sản phẩm: phân trang, lọc theo danh mục, lọc theo giá, sắp xếp
- [x] Chi tiết sản phẩm
- [x] Tìm kiếm sản phẩm (Elasticsearch, fallback sang Postgres ILIKE)
- [x] Danh mục (hiển thị dạng cây, tối đa 2 cấp)
- [x] CRUD sản phẩm (admin only)
- [x] Đánh giá sản phẩm (rating 1–5 + comment)
- [ ] ~~Flash sale / voucher~~ — ngoài scope
- [ ] ~~So sánh sản phẩm~~ — ngoài scope
- [ ] ~~Upload ảnh thật~~ — dùng URL ảnh giả (picsum.photos)

### 4.3 Giỏ hàng & Đơn hàng (order-service)
- [x] Thêm sản phẩm vào giỏ
- [x] Xem giỏ hàng
- [x] Cập nhật số lượng trong giỏ
- [x] Xóa khỏi giỏ
- [x] Đặt hàng (checkout): tạo order từ giỏ hàng, lưu địa chỉ giao hàng
- [x] Xem danh sách đơn của tôi
- [x] Xem chi tiết đơn hàng
- [x] Hủy đơn (chỉ khi status = pending)
- [x] Admin: xem tất cả đơn, cập nhật trạng thái
- [ ] ~~Thanh toán online~~ — ngoài scope (COD giả định)
- [ ] ~~Theo dõi vận chuyển thật~~ — chỉ cập nhật status thủ công

### 4.4 Chat (chat-service)
- [x] WebSocket real-time
- [x] Phòng chat 1-1 giữa customer và support (admin)
- [x] Gửi / nhận tin nhắn text
- [x] Lịch sử tin nhắn (phân trang cursor-based)
- [x] Typing indicator
- [x] Redis pub/sub để sync giữa các pod
- [ ] ~~Chat nhóm~~ — ngoài scope
- [ ] ~~Gửi file / ảnh~~ — ngoài scope
- [ ] ~~Message reactions~~ — ngoài scope

### 4.5 Frontend pages (NuxtJS)
- [x] `/` — trang chủ: banner, sản phẩm nổi bật, danh mục
- [x] `/products` — danh sách + bộ lọc + tìm kiếm
- [x] `/products/[slug]` — chi tiết sản phẩm + đánh giá
- [x] `/cart` — giỏ hàng
- [x] `/checkout` — nhập địa chỉ + đặt hàng
- [x] `/orders` — danh sách đơn của tôi
- [x] `/orders/[id]` — chi tiết đơn
- [x] `/auth/login` — đăng nhập
- [x] `/auth/register` — đăng ký
- [x] `/profile` — sửa thông tin cá nhân
- [x] `/admin` — dashboard thống kê đơn giản
- [x] `/admin/products` — quản lý sản phẩm
- [x] `/admin/orders` — quản lý đơn hàng
- [x] Chat widget — nổi góc dưới phải, mở trên mọi page

---

## 5. Cấu trúc bên trong mỗi Go service

```
services/{tên-service}/
├── README.md                  # BẮT BUỘC có
├── Dockerfile
├── docker-compose.yml         # chỉ để dev local
├── .env.example
├── go.mod
├── go.sum
├── main.go
├── cmd/
│   └── migrate/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go          # đọc env vars, struct Config
│   ├── domain/
│   │   └── *.go               # struct thuần, không import framework
│   ├── repository/
│   │   └── *_repo.go          # interface + implement, SQL hoặc MongoDB
│   ├── service/
│   │   └── *_svc.go           # business logic, gọi repository
│   ├── handler/
│   │   └── *_handler.go       # HTTP handler Fiber, gọi service
│   ├── middleware/
│   │   └── auth.go            # parse JWT, inject user vào ctx
│   └── kafka/
│       ├── producer.go
│       └── consumer.go
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
└── scripts/
    └── fake_data.go
```

---

## 6. Cấu trúc bên trong Python service

```
services/{tên-service}/
├── README.md
├── Dockerfile
├── docker-compose.yml
├── requirements.txt
├── .env.example
├── main.py                    # FastAPI app, chỉ có /health endpoint
├── consumers/
│   └── *_consumer.py          # Kafka consumer, mỗi topic 1 file
├── templates/                 # Jinja2 email templates (notification-service)
│   └── *.html
└── utils/
    └── *.py
```

---

## 7. Database schema — đủ dùng, không thêm cột thừa

### PostgreSQL (auth-service)
```sql
-- users
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
email       VARCHAR(255) UNIQUE NOT NULL
password_hash VARCHAR(255) NOT NULL
full_name   VARCHAR(255) NOT NULL
role        VARCHAR(20) NOT NULL DEFAULT 'customer'  -- 'admin' | 'customer'
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- refresh_tokens
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
token_hash  VARCHAR(255) NOT NULL
expires_at  TIMESTAMPTZ NOT NULL
revoked     BOOLEAN NOT NULL DEFAULT FALSE
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

### PostgreSQL (product-service)
```sql
-- categories
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
name        VARCHAR(255) NOT NULL
slug        VARCHAR(255) UNIQUE NOT NULL
parent_id   UUID REFERENCES categories(id)
image_url   TEXT

-- products
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
category_id UUID NOT NULL REFERENCES categories(id)
name        VARCHAR(500) NOT NULL
slug        VARCHAR(500) UNIQUE NOT NULL
description TEXT
price       NUMERIC(15,2) NOT NULL
sale_price  NUMERIC(15,2)
stock       INT NOT NULL DEFAULT 0
images      TEXT[]                        -- mảng URL ảnh
is_active   BOOLEAN NOT NULL DEFAULT TRUE
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- product_reviews
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
product_id  UUID NOT NULL REFERENCES products(id)
user_id     UUID NOT NULL
rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5)
comment     TEXT
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

### PostgreSQL (order-service)
```sql
-- orders
id               UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id          UUID NOT NULL
status           VARCHAR(30) NOT NULL DEFAULT 'pending'
-- status values: pending | confirmed | shipping | delivered | cancelled
total_amount     NUMERIC(15,2) NOT NULL
shipping_name    VARCHAR(255) NOT NULL
shipping_phone   VARCHAR(20) NOT NULL
shipping_address TEXT NOT NULL
created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- order_items
id                   UUID PRIMARY KEY DEFAULT gen_random_uuid()
order_id             UUID NOT NULL REFERENCES orders(id)
product_id           UUID NOT NULL
product_name_snapshot VARCHAR(500) NOT NULL
product_image_snapshot TEXT
unit_price           NUMERIC(15,2) NOT NULL
quantity             INT NOT NULL
```

### MongoDB (product-service) — collection: product_details
```json
{
  "_id": "product_uuid",
  "specifications": { "Màu sắc": "Đen", "Kích thước": "M/L/XL" },
  "rich_description": "<html nội dung chi tiết>",
  "updated_at": "ISODate"
}
```

### MongoDB (chat-service) — 2 collections
```json
// chat_rooms
{
  "_id": "ObjectId",
  "participants": ["user_uuid_1", "user_uuid_2"],
  "created_at": "ISODate"
}

// chat_messages
{
  "_id": "ObjectId",
  "room_id": "ObjectId",
  "sender_id": "user_uuid",
  "content": "nội dung tin nhắn",
  "type": "text",
  "created_at": "ISODate"
}
```

### Redis key conventions
```
# Auth service
refresh_token:{token_hash}  → user_id  (TTL = 7 ngày)
blacklist:token:{jti}       → "1"       (TTL = access token TTL còn lại)

# Order service — giỏ hàng lưu Redis
cart:{user_id}              → JSON array of cart items  (TTL = 30 ngày)

# Chat service — pub/sub
channel: chat:{room_id}     → JSON message payload
```

### Elasticsearch index: products
```json
{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "id":          { "type": "keyword" },
      "name":        { "type": "text" },
      "description": { "type": "text" },
      "category":    { "type": "keyword" },
      "price":       { "type": "float" },
      "sale_price":  { "type": "float" },
      "is_active":   { "type": "boolean" },
      "created_at":  { "type": "date" }
    }
  }
}
```

---

## 8. API endpoints — đầy đủ, không thêm bớt

### Auth (prefix: /api/auth)
```
POST   /api/auth/register         body: {email, password, full_name}
POST   /api/auth/login            body: {email, password}  → {access_token, refresh_token}
POST   /api/auth/refresh          body: {refresh_token}    → {access_token}
POST   /api/auth/logout           header: Authorization Bearer
GET    /api/auth/me               header: Authorization Bearer
PUT    /api/auth/me               body: {full_name}
```

### Products (prefix: /api/products, /api/categories)
```
GET    /api/products              ?page=1&limit=20&category=&search=&min_price=&max_price=&sort=price_asc|price_desc|newest
GET    /api/products/:slug
POST   /api/products              [admin] body: product fields
PUT    /api/products/:id          [admin]
DELETE /api/products/:id          [admin]
GET    /api/categories
GET    /api/products/:id/reviews
POST   /api/products/:id/reviews  [auth] body: {rating, comment}
```

### Orders / Cart (prefix: /api)
```
GET    /api/cart                  [auth] lấy giỏ hàng hiện tại
POST   /api/cart/items            [auth] body: {product_id, quantity}
PUT    /api/cart/items/:productId [auth] body: {quantity}
DELETE /api/cart/items/:productId [auth]
DELETE /api/cart                  [auth] xóa toàn bộ giỏ

POST   /api/orders                [auth] checkout: {shipping_name, shipping_phone, shipping_address}
GET    /api/orders                [auth] đơn của tôi
GET    /api/orders/:id            [auth]
PUT    /api/orders/:id/cancel     [auth] chỉ khi status=pending

GET    /api/admin/orders          [admin] ?status=&page=
PUT    /api/admin/orders/:id/status [admin] body: {status}
```

### Chat (prefix: /api/chat)
```
GET    /api/chat/rooms                [auth] phòng của tôi
POST   /api/chat/rooms                [auth] body: {participant_id}
GET    /api/chat/rooms/:id/messages   [auth] ?before=cursor&limit=50
WS     /ws?token={jwt}               WebSocket endpoint
```

---

## 9. Kafka topics

| Topic               | Producer         | Consumer               | Payload key fields             |
|---------------------|------------------|------------------------|--------------------------------|
| `product.created`   | product-service  | search-service         | id, name, description, price, category, is_active |
| `product.updated`   | product-service  | search-service         | id + changed fields            |
| `order.created`     | order-service    | notification-service   | order_id, user_email, items[], total_amount |
| `user.registered`   | auth-service     | notification-service   | user_id, email, full_name      |

---

## 10. K8s manifests — quy tắc chung

### Namespace
- Tất cả app services: namespace `shop`
- Tất cả infrastructure: namespace `infra`
- cert-manager: namespace `cert-manager`

### Mỗi service trong k8s/base/{service}/ phải có đủ 3 file:
```
deployment.yaml    # Deployment + resource limits
service.yaml       # ClusterIP (mặc định)
configmap.yaml     # env vars không nhạy cảm
```
Secrets (password, JWT key) tạo bằng `kubectl create secret generic` — KHÔNG commit vào git.

### Resource limits chuẩn cho worker node 4GB RAM:

| Service              | Memory request | Memory limit | CPU request | CPU limit |
|----------------------|----------------|--------------|-------------|-----------|
| Go services (nhỏ)    | 64Mi           | 256Mi        | 100m        | 500m      |
| Go services (lớn)    | 128Mi          | 512Mi        | 200m        | 1000m     |
| Python services      | 128Mi          | 384Mi        | 100m        | 500m      |
| Frontend (nginx)     | 32Mi           | 128Mi        | 50m         | 200m      |
| PostgreSQL           | 256Mi          | 512Mi        | 200m        | 1000m     |
| MongoDB              | 256Mi          | 512Mi        | 200m        | 500m      |
| Redis                | 64Mi           | 128Mi        | 100m        | 200m      |
| Kafka                | 512Mi          | 1Gi          | 200m        | 1000m     |
| Zookeeper            | 256Mi          | 512Mi        | 100m        | 500m      |
| Elasticsearch        | 512Mi          | 1Gi          | 300m        | 1000m     |

### Replicas
- Go services: 2 replicas mặc định
- chat-service: 2 replicas + `sessionAffinity: ClientIP` trên Service
- Python services: 1 replica (demo)
- Frontend: 2 replicas
- Infra (Postgres, Mongo, Redis, Kafka, ES): 1 replica StatefulSet

### HPA: chỉ api-gateway và product-service (min 2, max 6, trigger 70% CPU)

### Ingress annotations bắt buộc cho chat-service route:
```yaml
nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
```

---

## 11. Frontend conventions (NuxtJS + Vuetify)

### Màu sắc theme Vuetify
```typescript
// plugins/vuetify.ts
colors: {
  primary: '#1565C0',    // xanh dương đậm
  secondary: '#F57C00',  // cam
  accent: '#00897B',     // teal
  error: '#D32F2F',
  success: '#388E3C',
}
```

### Format giá tiền — BẮT BUỘC dùng hàm này
```typescript
// composables/useFormat.ts
export const formatVND = (amount: number) =>
  new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(amount)
// → "1.250.000 ₫"
```

### Auth token storage
- access_token: lưu trong Pinia store (memory only, mất khi reload)
- refresh_token: lưu trong localStorage key `shop_rt`
- Khi reload: lấy refresh_token từ localStorage, gọi /api/auth/refresh, lưu access_token mới vào Pinia
- Khi 401: tự động gọi refresh, nếu refresh cũng fail thì redirect `/auth/login`

### WebSocket trong useChat.ts
```typescript
// Reconnect với exponential backoff
const delays = [1000, 2000, 4000, 8000, 16000, 30000]
// Gửi JWT qua query param: wss://shopapicaovanson.xyz/ws?token=xxx
// Typing indicator: debounce 500ms, tự tắt sau 3s không có keystroke
```

### Fake data hiển thị
- Ảnh sản phẩm: `https://picsum.photos/seed/{product_slug}/400/400`
- Không dùng ảnh thật, không cần upload

---

## 12. Fake data — số lượng cụ thể

| Entity          | Số lượng | Ghi chú                                              |
|-----------------|----------|------------------------------------------------------|
| Users           | 50       | 1 admin (admin@shop.com / Admin@123), 49 customers   |
| Categories      | 10       | Electronics, Fashion, Food, Beauty, Sports, Home, Books, Toys, Automotive, Garden |
| Products        | 200      | 20 sản phẩm mỗi category, giá VND thực tế           |
| Product reviews | 500      | ngẫu nhiên, rating 1–5                              |
| Orders          | 300      | trải đều các status                                  |
| Chat rooms      | 20       | mỗi room có 10–30 messages                          |
| Chat messages   | 400      | text ngắn, thực tế                                  |

Tất cả fake data script phải **idempotent** — chạy nhiều lần không bị duplicate.

---

## 13. README.md bắt buộc có trong mỗi service

Mỗi file README.md phải có đủ các section sau (không thêm không bớt):

```markdown
# {Tên Service}

## Trách nhiệm
(1–2 câu mô tả service này làm gì)

## Tech stack
(bảng: Technology | Version | Mục đích)

## Environment variables
(bảng: Biến | Mô tả | Giá trị mặc định)

## Chạy local
(lệnh docker-compose up --build)

## Migration
(lệnh chạy migration up và down)

## API endpoints
(bảng: Method | Path | Auth | Mô tả) — chỉ Go services

## Kafka
(bảng: Topic | Produce/Consume | Mô tả) — nếu có

## Fake data
(lệnh chạy và kết quả mong đợi)
```

---

## 14. Makefile — các target ở root

```makefile
# Cluster
setup-cluster       # chạy k8s/cluster-setup/setup-vms.sh + bootstrap
verify-cluster      # kubectl get nodes

# Deploy
deploy-infra        # kubectl apply -f k8s/base/infra/ -n infra
deploy-cert-manager # helm install cert-manager
deploy-ingress      # helm install nginx-ingress
deploy-services     # kubectl apply -f k8s/base/*/  -n shop
deploy-frontend     # kubectl apply -f k8s/base/frontend/ -n shop

# Data
migrate             # chạy scripts/migrate-all.sh
seed                # chạy scripts/seed-data.sh

# Dev
logs svc=auth-service   # kubectl logs -f -l app=$(svc) -n shop
restart svc=product-service  # kubectl rollout restart deployment/$(svc) -n shop
status              # kubectl get pods,svc,ingress -n shop && kubectl get pods -n infra

# CI
build svc=auth-service  # docker build + push ghcr.io
```

---

## 15. .gitignore — áp dụng toàn project

```
# Go
services/*/tmp/
services/*/*.exe

# Python
**/__pycache__/
**/*.pyc
services/*/.venv/

# Node
frontend/web/node_modules/
frontend/web/.nuxt/
frontend/web/.output/

# Env files — KHÔNG commit
**/.env
# .env.example thì ĐƯỢC commit

# K8s secrets — KHÔNG commit
k8s/**/*-secret.yaml
k8s/**/secret*.yaml

# Build artifacts
**/dist/
**/build/
```

---

## 16. Quy tắc Cursor phải tuân theo khi làm việc với project này

1. **Đọc SKILL.md trước** — trước bất kỳ task nào
2. **Không tự thêm tính năng** ngoài §4. Nếu user yêu cầu tính năng mới, hỏi lại trước khi làm
3. **Giữ nguyên tên** service, domain, namespace, topic Kafka đúng như §2–§9
4. **Mỗi file phải hoàn chỉnh** — không viết `// TODO` hay `// implement later`
5. **Fake data phải idempotent** — kiểm tra tồn tại trước khi insert
6. **Secrets không commit** — chỉ tạo `.env.example` với giá trị placeholder
7. **Resource limits** phải theo bảng §10 — không để unlimited
8. **Format tiền VND** phải dùng `Intl.NumberFormat('vi-VN')` — không format thủ công
9. **Ảnh sản phẩm** dùng picsum.photos — không cần upload feature
10. **Mỗi service có README.md** theo template §13 — không bỏ qua