# shopcaovanson — API Specification

Base URL: `https://vocabee.cloud`

All authenticated endpoints require header: `Authorization: Bearer <access_token>`

Admin endpoints require JWT with `role: admin`.

---

## Auth Service

Prefix: `/api/auth`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/register` | — | Register new user. Body: `{email, password, full_name}` |
| POST | `/api/auth/login` | — | Login. Body: `{email, password}`. Returns `{access_token, refresh_token}` |
| POST | `/api/auth/refresh` | — | Refresh access token. Body: `{refresh_token}`. Returns `{access_token}` |
| POST | `/api/auth/logout` | Bearer | Revoke refresh token |
| GET | `/api/auth/me` | Bearer | Get current user profile |
| PUT | `/api/auth/me` | Bearer | Update profile. Body: `{full_name}` |

### Register Request

```json
{
  "email": "user@example.com",
  "password": "SecurePass123",
  "full_name": "Nguyen Van A"
}
```

### Login Response

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

---

## Product Service

Prefix: `/api/products`, `/api/categories`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/products` | — | List products with pagination and filters |
| GET | `/api/products/:slug` | — | Get product detail by slug |
| POST | `/api/products` | Admin | Create product |
| PUT | `/api/products/:id` | Admin | Update product |
| DELETE | `/api/products/:id` | Admin | Delete product |
| GET | `/api/categories` | — | Get category tree (max 2 levels) |
| GET | `/api/products/:id/reviews` | — | List product reviews |
| POST | `/api/products/:id/reviews` | Bearer | Add review. Body: `{rating, comment}` |

### List Products Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 20) |
| `category` | string | Filter by category slug or ID |
| `search` | string | Full-text search (Elasticsearch, Postgres fallback) |
| `min_price` | number | Minimum price filter |
| `max_price` | number | Maximum price filter |
| `sort` | string | `price_asc`, `price_desc`, or `newest` |

### Create Product Request (Admin)

```json
{
  "category_id": "uuid",
  "name": "Tai nghe Bluetooth",
  "slug": "tai-nghe-bluetooth",
  "description": "Tai nghe chống ồn",
  "price": 1250000,
  "sale_price": 990000,
  "stock": 50,
  "images": ["https://picsum.photos/seed/tai-nghe-bluetooth/400/400"],
  "is_active": true
}
```

### Add Review Request

```json
{
  "rating": 5,
  "comment": "Sản phẩm rất tốt, giao hàng nhanh"
}
```

---

## Order Service

Prefix: `/api/cart`, `/api/orders`, `/api/admin/orders`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/cart` | Bearer | Get current user's cart |
| POST | `/api/cart/items` | Bearer | Add item. Body: `{product_id, quantity}` |
| PUT | `/api/cart/items/:productId` | Bearer | Update quantity. Body: `{quantity}` |
| DELETE | `/api/cart/items/:productId` | Bearer | Remove item from cart |
| DELETE | `/api/cart` | Bearer | Clear entire cart |
| POST | `/api/orders` | Bearer | Checkout. Body: shipping info |
| GET | `/api/orders` | Bearer | List my orders |
| GET | `/api/orders/:id` | Bearer | Get order detail |
| PUT | `/api/orders/:id/cancel` | Bearer | Cancel order (only if status=pending) |
| GET | `/api/admin/orders` | Admin | List all orders. Query: `?status=&page=` |
| PUT | `/api/admin/orders/:id/status` | Admin | Update order status. Body: `{status}` |

### Add Cart Item Request

```json
{
  "product_id": "uuid",
  "quantity": 2
}
```

### Checkout Request

```json
{
  "shipping_name": "Nguyen Van A",
  "shipping_phone": "0901234567",
  "shipping_address": "123 Nguyen Hue, Q1, TP.HCM"
}
```

### Order Status Values

`pending` → `confirmed` → `shipping` → `delivered`

Or: `pending` → `cancelled`

---

## Chat Service

Prefix: `/api/chat`, WebSocket `/ws`

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/chat/rooms` | Bearer | List my chat rooms |
| POST | `/api/chat/rooms` | Bearer | Create room. Body: `{participant_id}` |
| GET | `/api/chat/rooms/:id/messages` | Bearer | Message history. Query: `?before=cursor&limit=50` |
| WS | `/ws?token={jwt}` | JWT query | WebSocket connection |

### WebSocket Protocol

**Client → Server:**

```json
{"type": "join", "room_id": "xxx"}
{"type": "message", "room_id": "xxx", "content": "hello"}
{"type": "typing", "room_id": "xxx"}
```

**Server → Client:**

```json
{"type": "message", "room_id": "xxx", "sender_id": "yyy", "content": "hello", "created_at": "2026-01-01T00:00:00Z"}
{"type": "user_joined", "user_id": "xxx"}
{"type": "typing", "room_id": "xxx", "user_id": "yyy"}
```

---

## Health Endpoints

| Service | Path |
|---------|------|
| api-gateway | `GET /health` |
| auth-service | `GET /health` |
| product-service | `GET /health` |
| order-service | `GET /health` |
| chat-service | `GET /health` |
| notification-service | `GET /health` |
| search-service | `GET /health` |

---

## Error Response Format

```json
{
  "error": "error_code",
  "message": "Human readable description"
}
```

| HTTP Status | Meaning |
|-------------|---------|
| 400 | Bad request / validation error |
| 401 | Unauthorized / invalid token |
| 403 | Forbidden / insufficient role |
| 404 | Resource not found |
| 409 | Conflict (e.g. duplicate email) |
| 429 | Rate limit exceeded (api-gateway: 100 req/min per IP) |
| 500 | Internal server error |

---

## Kafka Events (Internal)

Not exposed via HTTP. Documented for reference.

| Topic | Producer | Payload Key Fields |
|-------|----------|-------------------|
| `product.created` | product-service | id, name, description, price, category, is_active |
| `product.updated` | product-service | id + changed fields |
| `order.created` | order-service | order_id, user_email, items[], total_amount |
| `user.registered` | auth-service | user_id, email, full_name |

---

## Rate Limiting

api-gateway enforces **100 requests per minute per IP** using Redis.

## CORS

Allowed origin: `https://shopcaovanson.xyz`

## Authentication Flow

1. `POST /api/auth/login` → receive `access_token` (15 min) + `refresh_token` (7 days)
2. Include `Authorization: Bearer <access_token>` on protected routes
3. On 401, call `POST /api/auth/refresh` with `refresh_token`
4. `POST /api/auth/logout` revokes refresh token

## Default Admin Account (seed data)

| Field | Value |
|-------|-------|
| Email | admin@shop.com |
| Password | Admin@123 |
| Role | admin |
