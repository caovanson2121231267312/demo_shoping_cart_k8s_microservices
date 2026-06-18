# Frontend Web

## Trách nhiệm

Ứng dụng web NuxtJS 3 + Vuetify 3 cho cửa hàng trực tuyến Shop Cao Van Son. Cung cấp giao diện mua sắm, quản trị, chat hỗ trợ và tích hợp với backend API tại `shopapicaovanson.xyz`.

## Tech stack

| Technology | Version | Mục đích |
|------------|---------|----------|
| NuxtJS | 3.12 | Framework frontend SPA |
| Vuetify | 3.6 | UI component library |
| Pinia | latest | State management (auth, cart, chat) |
| @vueuse/core | latest | Utilities (debounce, composables) |
| nginx | alpine | Phục vụ static build trong production |

## Environment variables

| Biến | Mô tả | Giá trị mặc định |
|------|-------|------------------|
| `NUXT_PUBLIC_API_URL` | Base URL backend API | `https://shopapicaovanson.xyz` |
| `NUXT_PUBLIC_WS_URL` | WebSocket URL (chat) | `wss://shopapicaovanson.xyz` |
| `NUXT_API_PROXY_TARGET` | Target proxy cho dev server `/api/*` | `https://shopapicaovanson.xyz` |

## Chạy local

```bash
cp .env.example .env
npm install
npm run dev
```

Truy cập: http://localhost:3000

## Build & Docker

```bash
docker-compose up --build
```

Production image dùng multi-stage build (Node 20 → nginx alpine), phục vụ static files từ `.output/public`.

## Tính năng chính

- Trang chủ, danh sách/chi tiết sản phẩm, tìm kiếm & lọc
- Giỏ hàng, checkout (COD), quản lý đơn hàng
- Đăng nhập/đăng ký với JWT (access token trong Pinia, refresh token trong `localStorage` key `shop_rt`)
- Chat widget hỗ trợ real-time qua native WebSocket
- Trang admin: dashboard, CRUD sản phẩm, quản lý đơn hàng

## Tài khoản admin mặc định

| Email | Password |
|-------|----------|
| admin@shop.com | Admin@123 |

## Migration

Không áp dụng — frontend không có database riêng.

## API endpoints

Frontend gọi các endpoint backend qua `NUXT_PUBLIC_API_URL`. Xem chi tiết tại `docs/api-spec.md`.

## Kafka

Không áp dụng.

## Fake data

Ảnh sản phẩm dùng `https://picsum.photos/seed/{slug}/400/400`. Dữ liệu sản phẩm/đơn hàng lấy từ backend seed script.
