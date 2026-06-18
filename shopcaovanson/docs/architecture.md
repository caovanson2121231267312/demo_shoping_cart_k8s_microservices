# shopcaovanson — System Architecture

## Overview

shopcaovanson is a demo e-commerce platform deployed on a multi-node Kubernetes cluster (1 master + 2 workers) running on KVM/libvirt VMs atop a single VPS. The system uses microservices architecture with Go and Python backends, NuxtJS frontend, and infrastructure StatefulSets for data stores and messaging.

## Domains

| Domain | Purpose |
|--------|---------|
| https://shopcaovanson.xyz | NuxtJS 3 frontend (Vuetify 3) |
| https://shopapicaovanson.xyz | Backend API + WebSocket |

SSL is provisioned automatically via cert-manager ClusterIssuer `letsencrypt-prod`.

## Architecture Diagram

```
                                    Internet
                                       |
                              [ VPS - 6CPU/12GB ]
                                       |
                    +------------------+------------------+
                    |         KVM / libvirt              |
                    |  +----------+  +--------+--------+ |
                    |  | k8s-master|  |worker-1|worker-2|
                    |  | 2CPU/3GB |  |2CPU/4GB|2CPU/4GB|
                    |  +----------+  +--------+--------+ |
                    +------------------+------------------+
                                       |
                         nginx-ingress-controller
                         (TLS: cert-manager)
                    +---------+----------+---------+
                    |         |          |         |
            shopcaovanson.xyz |   shopapicaovanson.xyz
                    |         |          |
              +-----v-----+   |    +-----v------------------+
              | frontend  |   |    | /api -> api-gateway     |
              | (NuxtJS)  |   |    | /ws  -> chat-service    |
              | 2 replicas|   |    +------------+------------+
              +-----------+   |                 |
                              |    +------------v------------+
                              |    |      api-gateway        |
                              |    |  (rate limit, JWT, proxy)|
                              |    +--+--+-----+-----+--------+
                              |       |  |     |     |
                    +---------+   +---v--v-v-----v-----v--------+
                    |  namespace: shop (app services)          |
                    |                                          |
                    |  auth-service    product-service         |
                    |  order-service   chat-service (WS)       |
                    |  notification    search-service          |
                    +------------------+-----------------------+
                                       |
                    +------------------v-----------------------+
                    |  namespace: infra (StatefulSets)         |
                    |                                          |
                    |  PostgreSQL 15   MongoDB 7   Redis 7     |
                    |  Kafka 3.6 + Zookeeper   Elasticsearch 8 |
                    +------------------------------------------+
```

## Service Responsibilities

### Go Services (namespace: shop)

| Service | Port | Replicas | Description |
|---------|------|----------|-------------|
| api-gateway | 8080 | 2 (HPA 2-6) | Single HTTP entry point, JWT validation, rate limiting, reverse proxy |
| auth-service | 8080 | 2 | Registration, login, JWT refresh, profile |
| product-service | 8080 | 2 (HPA 2-6) | Products, categories, reviews, search integration |
| order-service | 8080 | 2 | Cart (Redis), checkout, order management |
| chat-service | 8080 | 2 | WebSocket chat, MongoDB messages, Redis pub/sub |

### Python Services (namespace: shop)

| Service | Port | Replicas | Description |
|---------|------|----------|-------------|
| notification-service | 8000 | 1 | Kafka consumer, sends welcome and order confirmation emails |
| search-service | 8000 | 1 | Kafka consumer, indexes products to Elasticsearch |

### Frontend (namespace: shop)

| Service | Port | Replicas | Description |
|---------|------|----------|-------------|
| frontend | 80 | 2 | NuxtJS 3 SPA served by nginx |

### Infrastructure (namespace: infra)

| Component | Image | Storage | Purpose |
|-----------|-------|---------|---------|
| PostgreSQL 15 | postgres:15-alpine | 20Gi | auth_db, product_db, order_db |
| MongoDB 7 | mongo:7 | 15Gi | product_details, chat collections |
| Redis 7 | redis:7-alpine | 5Gi | refresh tokens, cart, chat pub/sub |
| Zookeeper | cp-zookeeper:7.6.0 | 10Gi | Kafka coordination |
| Kafka 3.6 | cp-kafka:7.6.0 | 10Gi | Event streaming |
| Elasticsearch 8 | elasticsearch:8.12.2 | 20Gi | Product search index |

## Request Flow

### User browses products

1. Browser loads `shopcaovanson.xyz` → nginx-ingress → frontend pods
2. Frontend calls `shopapicaovanson.xyz/api/products` → ingress → api-gateway
3. api-gateway validates JWT (if required), proxies to product-service
4. product-service queries PostgreSQL; search uses Elasticsearch with Postgres ILIKE fallback

### User places order

1. Frontend POST `/api/orders` → api-gateway → order-service
2. order-service reads cart from Redis, validates stock via product-service
3. order-service writes order to PostgreSQL, publishes `order.created` to Kafka
4. notification-service consumes event, sends confirmation email via SMTP

### Real-time chat

1. Frontend opens WebSocket `wss://shopapicaovanson.xyz/ws?token=JWT`
2. Ingress routes to chat-service with 3600s timeout annotations
3. chat-service uses `sessionAffinity: ClientIP` for sticky sessions
4. Messages broadcast across pods via Redis pub/sub channel `chat:{room_id}`

## Kafka Topics

| Topic | Producer | Consumer |
|-------|----------|----------|
| `product.created` | product-service | search-service |
| `product.updated` | product-service | search-service |
| `order.created` | order-service | notification-service |
| `user.registered` | auth-service | notification-service |

## Data Stores per Service

| Service | PostgreSQL | MongoDB | Redis | Elasticsearch | Kafka |
|---------|:----------:|:-------:|:-----:|:-------------:|:-----:|
| auth-service | users, refresh_tokens | — | refresh tokens | — | produce |
| product-service | products, categories, reviews | product_details | — | read | produce |
| order-service | orders, order_items | — | cart | — | produce |
| chat-service | — | chat_rooms, chat_messages | pub/sub | — | — |
| notification-service | — | — | — | — | consume |
| search-service | — | — | — | index | consume |

## Network & Security

- **Namespaces**: `shop` (apps), `infra` (data), `cert-manager`, `ingress-nginx`
- **Secrets**: Created via `kubectl create secret` — never committed to git
- **TLS**: Let's Encrypt production via HTTP-01 challenge
- **JWT**: RS256, 15-minute access tokens, 7-day refresh tokens in Redis
- **CORS**: api-gateway allows `https://shopcaovanson.xyz`

## Deployment Model

```
k8s/base/           # Base manifests (all services + infra)
k8s/overlays/dev/   # Reduced replicas for development
k8s/overlays/prod/  # Production settings (default)
```

Deploy with Kustomize:

```bash
kubectl apply -k k8s/overlays/prod
```

## Resource Allocation

Worker nodes have 4GB RAM each. Resource limits follow SKILL.md section 10:

- Go services (small): 64-256Mi RAM
- Go services (large): 128-512Mi RAM
- Python services: 128-384Mi RAM
- Frontend: 32-128Mi RAM
- Infra: sized for single-replica StatefulSets on 4GB workers

## Observability

- Health endpoints: `/health` on all services
- Kubernetes probes: liveness + readiness on every deployment
- Logs: `make logs svc=auth-service`
- Status: `make status`

## CI/CD

GitHub Actions builds changed service images on push to `main`, tags with git SHA, pushes to `ghcr.io/caovanson/{service}`, and updates deployment manifests.
