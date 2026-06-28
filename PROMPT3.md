Continue building shopcaovanson. Create the NuxtJS frontend and all remaining K8s manifests to make everything deployable end-to-end.

## FRONTEND: frontend/web/

Tech: NuxtJS 3.12, Vuetify 3.6, Pinia, @vueuse/core, socket.io-client (for WebSocket)

### Directory structure:
```
frontend/web/
├── README.md
├── Dockerfile                  # multi-stage: build then nginx
├── docker-compose.yml
├── .env.example
├── nuxt.config.ts
├── app.vue
├── assets/
│   └── css/
│       └── main.css
├── components/
│   ├── layout/
│   │   ├── AppHeader.vue       # navbar with cart count badge, user avatar, search bar
│   │   ├── AppFooter.vue
│   │   └── AppSidebar.vue      # mobile drawer
│   ├── product/
│   │   ├── ProductCard.vue     # image, name, price, sale badge, add-to-cart btn
│   │   ├── ProductGrid.vue     # responsive grid of ProductCard
│   │   ├── ProductFilter.vue   # sidebar filters: category, price range, rating
│   │   └── ProductSearch.vue   # search input with debounce
│   ├── cart/
│   │   ├── CartDrawer.vue      # slide-in cart panel
│   │   └── CartItem.vue
│   ├── chat/
│   │   ├── ChatWidget.vue      # floating chat button + chat panel
│   │   ├── ChatRoom.vue        # messages list + input
│   │   └── ChatMessage.vue     # single message bubble
│   └── common/
│       ├── LoadingSpinner.vue
│       ├── EmptyState.vue
│       └── ConfirmDialog.vue
├── composables/
│   ├── useAuth.ts              # login, logout, register, getCurrentUser
│   ├── useCart.ts              # add, remove, update, total
│   ├── useChat.ts              # WebSocket connection, send message, rooms
│   └── useProducts.ts          # fetch products, search, filter
├── layouts/
│   ├── default.vue             # header + main + footer
│   └── auth.vue                # centered card layout for login/register
├── middleware/
│   └── auth.ts                 # redirect to /login if no token
├── pages/
│   ├── index.vue               # homepage: hero banner, featured products, categories
│   ├── products/
│   │   ├── index.vue           # product listing with filters + pagination
│   │   └── [slug].vue          # product detail: images, desc, add to cart, reviews
│   ├── cart.vue                # cart page
│   ├── checkout.vue            # address form + order summary + place order
│   ├── orders/
│   │   ├── index.vue           # my orders list
│   │   └── [id].vue            # order detail + status timeline
│   ├── auth/
│   │   ├── login.vue           # email + password form
│   │   └── register.vue        # name + email + password
│   ├── profile.vue             # edit profile (auth required)
│   └── admin/
│       ├── index.vue           # dashboard: stats cards (total orders, revenue, users)
│       ├── products.vue        # CRUD table with DataTable
│       └── orders.vue          # orders management table
├── plugins/
│   └── vuetify.ts
├── server/                     # Nuxt server routes (proxy)
│   └── api/
│       └── [...].ts            # proxy all /api/* to vocabee.cloud
├── stores/
│   ├── auth.ts                 # Pinia: user, token, isLoggedIn
│   ├── cart.ts                 # Pinia: items, count, total
│   └── chat.ts                 # Pinia: rooms, messages, unread count
└── types/
    └── index.ts                # TypeScript interfaces: User, Product, Order, Message
```

### Key implementation details:

**nuxt.config.ts**: configure Vuetify module, runtimeConfig with NUXT_PUBLIC_API_URL=https://vocabee.cloud, NUXT_PUBLIC_WS_URL=wss://vocabee.cloud

**useAuth.ts composable**: 
- stores JWT in httpOnly cookie via server route (not localStorage)
- auto-refresh token 1 minute before expiry using setInterval
- on 401 response, redirect to /login

**useChat.ts composable**:
- connects WebSocket to wss://vocabee.cloud/ws?token={jwt}
- auto-reconnect with exponential backoff (1s, 2s, 4s, max 30s)
- stores messages in Pinia chat store
- shows typing indicator with 3s timeout

**ChatWidget.vue**: floating button bottom-right, unread badge count, opens chat panel sliding from right. For logged-in users shows "Support Chat" room by default.

**ProductCard.vue**: Vuetify v-card, product image with aspect-ratio 1:1, sale badge if sale_price < price, price in VND format (toLocaleString('vi-VN')), add to cart button triggers useCart.addItem()

**Admin pages**: protected by middleware checking role=admin from Pinia auth store

**Dockerfile** (multi-stage):
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/.output/public /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

**nginx.conf**: try_files with fallback to index.html for SPA routing, gzip enabled

## K8s DEPLOYMENTS - complete all files in k8s/base/

For EACH service create deployment.yaml + service.yaml + configmap.yaml:

### Pattern for Go services (example: auth-service):
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: shop
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: ghcr.io/caovanson/auth-service:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: auth-service-config
        - secretRef:
            name: auth-service-secret
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 15
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### chat-service deployment: add sessionAffinity: ClientIP to Service

### frontend deployment: replicas: 2, image: ghcr.io/caovanson/frontend-web:latest, containerPort: 80

### HPA for product-service and api-gateway:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: product-service-hpa
  namespace: shop
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: product-service
  minReplicas: 2
  maxReplicas: 6
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## FINAL INGRESS with SSL

```yaml
# k8s/base/ingress/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: shop-ingress
  namespace: shop
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-body-size: "20m"
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - shopcaovanson.xyz
    - www.shopcaovanson.xyz
    secretName: frontend-tls
  - hosts:
    - vocabee.cloud
    secretName: backend-tls
  rules:
  - host: shopcaovanson.xyz
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
  - host: vocabee.cloud
    http:
      paths:
      - path: /ws
        pathType: Prefix
        backend:
          service:
            name: chat-service
            port:
              number: 80
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 80
```

## GITHUB ACTIONS CI/CD

Create .github/workflows/deploy.yaml that:
1. On push to main: build Docker images for changed services only (using paths filter)
2. Push to ghcr.io/caovanson/{service-name}:{git-sha}
3. Update image tag in k8s/base/{service}/deployment.yaml
4. Commit back to repo (ArgoCD or Flux will auto-sync)

## ROOT README.md

Complete README covering:
1. System architecture diagram (ASCII art)
2. Prerequisites (VPS 6CPU/12GB, domain DNS setup)
3. Quick start: clone -> setup cluster -> deploy infra -> migrate -> seed -> deploy services -> verify
4. Domain DNS records to configure (A records for shopcaovanson.xyz and vocabee.cloud pointing to VPS IP)
5. How to access: frontend URL, API docs URL, Grafana URL, how to login as admin
6. Development: how to run any single service locally with docker-compose
7. Troubleshooting common issues

Generate ALL files completely. Every .vue file, every .go file, every .yaml file must be complete with real working code. No placeholder comments like "// implement this".