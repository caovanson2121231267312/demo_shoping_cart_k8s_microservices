Continue building the shopcaovanson project. Now create all backend services with complete, runnable code.

## SHARED CONVENTIONS FOR ALL GO SERVICES

Each Go service must follow this internal structure:
```
services/{service-name}/
├── README.md
├── Dockerfile
├── docker-compose.yml          # for local dev only
├── .env.example
├── go.mod
├── go.sum
├── main.go
├── cmd/
│   └── migrate/
│       └── main.go             # standalone migration runner
├── internal/
│   ├── config/
│   │   └── config.go           # loads from env vars
│   ├── domain/
│   │   └── {entity}.go         # pure domain structs, no framework deps
│   ├── repository/
│   │   └── {entity}_repo.go    # DB access, interfaces
│   ├── service/
│   │   └── {entity}_svc.go     # business logic
│   ├── handler/
│   │   └── {entity}_handler.go # HTTP handlers (fiber or chi)
│   ├── middleware/
│   │   └── auth.go             # JWT middleware
│   └── kafka/
│       ├── producer.go
│       └── consumer.go
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
└── scripts/
    └── fake_data.go            # generates and inserts fake data
```

## AUTH SERVICE (Go + PostgreSQL + Redis)

Tech: Go 1.22, Fiber v2, golang-migrate, sqlx, go-redis, golang-jwt/jwt v5

Domain entities:
- User: id(uuid), email, password_hash, full_name, role(admin/customer), created_at, updated_at
- RefreshToken: id, user_id, token_hash, expires_at, revoked

PostgreSQL migrations (migrations/000001_init.up.sql):
- CREATE TABLE users with all fields, indexes on email
- CREATE TABLE refresh_tokens with index on token_hash

API endpoints:
- POST /api/auth/register - validate email/password, hash with bcrypt, return JWT
- POST /api/auth/login - verify credentials, return access_token (15min) + refresh_token (7days) stored in Redis
- POST /api/auth/refresh - validate refresh token from Redis, issue new access_token
- POST /api/auth/logout - revoke refresh token in Redis
- GET /api/auth/me - requires JWT middleware, return user profile
- PUT /api/auth/me - update full_name

JWT: RS256, private key from env var, access token payload: {sub, email, role, iat, exp}

Fake data (scripts/fake_data.go): generate 50 users (1 admin + 49 customers), realistic Vietnamese names and emails

## PRODUCT SERVICE (Go + PostgreSQL + MongoDB + Elasticsearch)

Domain entities:
- Category: id, name, slug, parent_id, image_url
- Product: id, category_id, name, slug, description, price, sale_price, stock, images[]string, is_active, created_at
- ProductReview: id, product_id, user_id, rating(1-5), comment, created_at

PostgreSQL: categories, products tables with full-text search index
MongoDB: product_details collection (rich description, specifications as flexible BSON)
Elasticsearch: products index (name, description, category, price, tags for search)

API endpoints:
- GET /api/products?page=1&limit=20&category=&search=&min_price=&max_price=&sort=
- GET /api/products/:slug
- POST /api/products (admin only)
- PUT /api/products/:id (admin only)
- DELETE /api/products/:id (admin only)
- GET /api/categories (tree structure)
- POST /api/products/:id/reviews (authenticated)
- GET /api/products/:id/reviews

Search: uses Elasticsearch for /api/products?search=, falls back to Postgres ILIKE if ES unavailable

Kafka producer: publishes to "product.created" and "product.updated" topics (consumed by search-service to sync ES)

Fake data: 10 categories (Electronics, Fashion, Food, Beauty, Sports, Home, Books, Toys, Automotive, Garden), 200 products with realistic Vietnamese product names and prices in VND, 500 reviews

## ORDER SERVICE (Go + PostgreSQL + Kafka)

Domain entities:
- Cart: id, user_id, items[]{product_id, qty, price_snapshot}, updated_at
- Order: id, user_id, status(pending/confirmed/shipping/delivered/cancelled), items[], total_amount, shipping_address, created_at
- OrderItem: id, order_id, product_id, product_name_snapshot, qty, unit_price

API endpoints:
- GET /api/cart (get my cart)
- POST /api/cart/items (add item)
- PUT /api/cart/items/:productId (update qty)
- DELETE /api/cart/items/:productId
- POST /api/orders (checkout from cart, validates stock via product-service gRPC or HTTP)
- GET /api/orders (my orders)
- GET /api/orders/:id
- PUT /api/orders/:id/cancel
- GET /api/admin/orders (admin, all orders with filter by status)
- PUT /api/admin/orders/:id/status (admin update status)

Kafka producer: publish to "order.created" topic when order placed (notification-service consumes this)

Fake data: 300 orders in various statuses

## CHAT SERVICE (Go + MongoDB + Redis pub/sub + WebSocket)

Domain entities (MongoDB):
- ChatRoom: id, participants[user_id], room_type(direct/support), created_at
- ChatMessage: id, room_id, sender_id, content, type(text/image), read_by[], created_at

WebSocket protocol (JSON messages):
```json
// client -> server
{"type": "join", "room_id": "xxx"}
{"type": "message", "room_id": "xxx", "content": "hello"}
{"type": "typing", "room_id": "xxx"}

// server -> client  
{"type": "message", "room_id": "xxx", "sender_id": "yyy", "content": "hello", "created_at": "..."}
{"type": "user_joined", "user_id": "xxx"}
{"type": "typing", "room_id": "xxx", "user_id": "yyy"}
```

Redis pub/sub: channel "chat:{room_id}" for cross-pod message broadcast
JWT auth via query param: ws://host/ws?token=xxx

API endpoints:
- GET /api/chat/rooms (my rooms)
- POST /api/chat/rooms (create direct room with another user)
- GET /api/chat/rooms/:id/messages?before=cursor&limit=50
- WS /ws (upgrade to WebSocket)

## API GATEWAY (Go + Fiber)

Simple reverse proxy + JWT validation:
- Routes: /api/auth/* -> auth-service, /api/products/* -> product-service, /api/orders/* -> order-service, /api/chat/* -> chat-service
- Middleware: rate limiting (100 req/min per IP using Redis), CORS (allow shopcaovanson.xyz), request logging, recover from panic
- Injects X-User-Id and X-User-Role headers after JWT validation so downstream services don't need to verify JWT again

## NOTIFICATION SERVICE (Python + Kafka)

Tech: Python 3.12, confluent-kafka, FastAPI (health endpoint only), Jinja2 for email templates

Kafka consumers:
- "order.created" topic: send order confirmation email (HTML template with order details)
- "user.registered" topic: send welcome email

Structure:
```
services/notification-service/
├── README.md
├── Dockerfile
├── requirements.txt
├── main.py                     # FastAPI app with /health endpoint
├── consumers/
│   ├── order_consumer.py
│   └── user_consumer.py
├── templates/
│   ├── order_confirmation.html
│   └── welcome.html
└── utils/
    └── email.py               # SMTP via env vars
```

## SEARCH SERVICE (Python + Kafka + Elasticsearch)

Tech: Python 3.12, confluent-kafka, elasticsearch-py 8.x, FastAPI

Kafka consumers:
- "product.created" topic: index new product to Elasticsearch
- "product.updated" topic: update ES document

Elasticsearch index mapping (products):
```json
{
  "mappings": {
    "properties": {
      "name": {"type": "text", "analyzer": "standard"},
      "description": {"type": "text"},
      "category": {"type": "keyword"},
      "price": {"type": "float"},
      "tags": {"type": "keyword"},
      "is_active": {"type": "boolean"}
    }
  }
}
```

API: GET /search?q=&category=&min_price=&max_price=&page=1 (called internally by product-service)

## MIGRATIONS - ALL SERVICES

Each Go service: use golang-migrate/migrate v4. Migration files in migrations/ directory.
Run via: `go run cmd/migrate/main.go up`

Create a root-level `scripts/migrate-all.sh` that runs migrations for all services in order:
1. auth-service migrations
2. product-service migrations  
3. order-service migrations

## FAKE DATA - ROOT SCRIPTS

Create `scripts/seed-data.sh` that runs seeders in dependency order:
1. auth-service fake users
2. product-service fake categories + products + reviews
3. order-service fake orders (using real user IDs and product IDs)
4. chat-service fake chat rooms and messages

All fake data scripts must be idempotent (safe to run multiple times).

Generate ALL files completely with real, working Go and Python code. No TODOs, no placeholder functions.