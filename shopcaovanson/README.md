# shopcaovanson

Demo e-commerce platform with NuxtJS 3 frontend and Go/Python microservices, deployed on a multi-node Kubernetes cluster.

| Component | URL |
|-----------|-----|
| Frontend | https://shopcaovanson.xyz |
| Backend API | https://shopapicaovanson.xyz |
| WebSocket Chat | wss://shopapicaovanson.xyz/ws |

## Architecture

```
                         Internet
                            |
                   [ VPS 6CPU / 12GB ]
                            |
              +-------------+-------------+
              |    KVM / libvirt VMs      |
              | master | worker-1 | worker-2 |
              +-------------+-------------+
                            |
                 nginx-ingress + cert-manager
              +-------------+-------------+
              |                           |
     shopcaovanson.xyz          shopapicaovanson.xyz
              |                           |
         +----+----+              +-------+--------+
         | frontend |              | /api  /ws     |
         |  NuxtJS  |              | gateway  chat |
         +----------+              +-------+--------+
                                           |
              +----------------------------+---------------------------+
              |  shop namespace: auth, product, order, chat,         |
              |  notification, search, api-gateway, frontend         |
              +----------------------------+---------------------------+
                                           |
              +----------------------------+---------------------------+
              |  infra namespace: PostgreSQL, MongoDB, Redis,        |
              |  Kafka, Zookeeper, Elasticsearch                      |
              +-------------------------------------------------------+
```

See [docs/architecture.md](docs/architecture.md) for detailed system design and [docs/api-spec.md](docs/api-spec.md) for full API reference.

## Tech Stack

| Layer | Technology | Version |
|-------|------------|---------|
| Frontend | NuxtJS 3 + Vuetify 3 | 3.12 / 3.6 |
| Go services | Go + Fiber v2 | 1.22 |
| Python services | FastAPI + confluent-kafka | 3.12 |
| Database | PostgreSQL | 15 |
| Document DB | MongoDB | 7 |
| Cache | Redis | 7 |
| Message broker | Kafka + Zookeeper | 3.6 |
| Search | Elasticsearch | 8 |
| Kubernetes | kubeadm + Calico | 1.29 |
| Ingress | nginx-ingress-controller | — |
| SSL | cert-manager + Let's Encrypt | 1.14 |

## Prerequisites

- **VPS**: 6 CPU, 12 GB RAM, 150 GB NVMe (Ubuntu 22.04 host)
- **Domain DNS**: A records pointing to VPS public IP
  - `shopcaovanson.xyz` → VPS IP
  - `www.shopcaovanson.xyz` → VPS IP
  - `shopapicaovanson.xyz` → VPS IP
- **Host tools**: KVM/libvirt, kubectl, helm, docker, git, SSH key pair
- **GitHub Container Registry**: Images at `ghcr.io/caovanson/{service}:latest`

## Quick Start

### 1. Clone repository

```bash
git clone https://github.com/caovanson/shopcaovanson.git
cd shopcaovanson
```

### 2. Setup Kubernetes cluster

```bash
# On VPS host (requires root for KVM VM creation)
sudo make setup-cluster

# Copy kubeconfig to local machine
export KUBECONFIG=~/.kube/shopcaovanson-config
```

### 3. Create secrets

Secrets are **not** committed to git. Create them before deploying:

#### Infrastructure secrets (namespace: `infra`)

```bash
kubectl create namespace infra

kubectl create secret generic postgres-secret -n infra \
  --from-literal=POSTGRES_PASSWORD='your-strong-postgres-password'

kubectl create secret generic mongodb-secret -n infra \
  --from-literal=MONGO_PASSWORD='your-strong-mongo-password'

kubectl create secret generic redis-secret -n infra \
  --from-literal=REDIS_PASSWORD='your-strong-redis-password'

kubectl create secret generic elasticsearch-secret -n infra \
  --from-literal=ELASTIC_PASSWORD='your-strong-elastic-password'
```

#### Application secrets (namespace: `shop`)

```bash
kubectl create namespace shop

# Generate RSA key pair for JWT
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

kubectl create secret generic auth-service-secret -n shop \
  --from-literal=DB_PASSWORD='your-strong-postgres-password' \
  --from-literal=REDIS_PASSWORD='your-strong-redis-password' \
  --from-file=JWT_PRIVATE_KEY=jwt-private.pem \
  --from-file=JWT_PUBLIC_KEY=jwt-public.pem

kubectl create secret generic api-gateway-secret -n shop \
  --from-literal=REDIS_PASSWORD='your-strong-redis-password' \
  --from-file=JWT_PUBLIC_KEY=jwt-public.pem

kubectl create secret generic product-service-secret -n shop \
  --from-literal=DB_PASSWORD='your-strong-postgres-password' \
  --from-literal=MONGO_PASSWORD='your-strong-mongo-password' \
  --from-literal=ELASTIC_PASSWORD='your-strong-elastic-password'

kubectl create secret generic order-service-secret -n shop \
  --from-literal=DB_PASSWORD='your-strong-postgres-password' \
  --from-literal=REDIS_PASSWORD='your-strong-redis-password'

kubectl create secret generic chat-service-secret -n shop \
  --from-literal=MONGO_PASSWORD='your-strong-mongo-password' \
  --from-literal=REDIS_PASSWORD='your-strong-redis-password' \
  --from-file=JWT_PUBLIC_KEY=jwt-public.pem

kubectl create secret generic notification-service-secret -n shop \
  --from-literal=SMTP_USER='your-smtp@gmail.com' \
  --from-literal=SMTP_PASSWORD='your-app-password'

kubectl create secret generic search-service-secret -n shop \
  --from-literal=ELASTIC_PASSWORD='your-strong-elastic-password'

rm -f jwt-private.pem jwt-public.pem
```

### 4. Deploy everything

```bash
make deploy-all
# or with dev overlay (reduced replicas):
# OVERLAY=dev bash scripts/deploy-all.sh
```

### 5. Run migrations and seed data

```bash
make migrate
make seed
```

### 6. Verify deployment

```bash
make status
kubectl get ingress -n shop
kubectl get certificates -n shop
```

## Access

| Resource | URL / Credentials |
|----------|-------------------|
| Frontend | https://shopcaovanson.xyz |
| API base | https://shopapicaovanson.xyz/api |
| WebSocket | wss://shopapicaovanson.xyz/ws |
| API docs | [docs/api-spec.md](docs/api-spec.md) |
| Admin login | `admin@shop.com` / `Admin@123` |
| Monitoring | Grafana không nằm trong scope demo này |

Wait 2–5 minutes after deploy for Let's Encrypt certificates to be issued. Check with:

```bash
kubectl describe certificate -n shop
```

## Project Structure

```
shopcaovanson/
├── README.md
├── Makefile
├── .gitignore
├── services/           # Go + Python microservices
├── frontend/web/       # NuxtJS 3 frontend
├── k8s/
│   ├── cluster-setup/  # KVM + kubeadm scripts
│   ├── base/           # K8s manifests
│   └── overlays/     # dev / prod kustomize overlays
├── scripts/            # Deploy, migrate, seed scripts
└── docs/               # Architecture + API docs
```

## Makefile Commands

```bash
make setup-cluster      # Create KVM VMs + bootstrap K8s
make verify-cluster     # Check nodes and system pods
make deploy-all         # Full deployment
make deploy-infra       # Deploy infra StatefulSets only
make deploy-services    # Deploy app services
make migrate            # Run DB migrations
make seed               # Seed fake data
make status             # Show cluster status
make logs svc=auth-service   # Tail service logs
make restart svc=product-service  # Restart deployment
make build svc=auth-service      # Build + push Docker image
```

## Development

### Run a single service locally

```bash
cd services/auth-service
cp .env.example .env
docker-compose up --build
```

### Run frontend locally

```bash
cd frontend/web
cp .env.example .env
# Local dev — trỏ API về gateway hoặc auth-service:
# NUXT_PUBLIC_API_URL=http://localhost:8080
# NUXT_PUBLIC_WS_URL=ws://localhost:8084
npm install
npm run dev
# → http://localhost:3000
```

### Local migrations & seed (không cần K8s)

```bash
LOCAL=true bash scripts/migrate-all.sh
LOCAL=true bash scripts/seed-data.sh
```

### Kustomize overlays

```bash
# Production (default)
kubectl apply -k k8s/overlays/prod

# Development (reduced replicas)
kubectl apply -k k8s/overlays/dev
```

## DNS Configuration

Point these A records to your VPS public IP:

| Record | Type | Value |
|--------|------|-------|
| shopcaovanson.xyz | A | `<VPS_IP>` |
| www.shopcaovanson.xyz | A | `<VPS_IP>` |
| shopapicaovanson.xyz | A | `<VPS_IP>` |

## Troubleshooting

### Pods stuck in Pending

```bash
kubectl describe pod <pod-name> -n shop
kubectl get pvc -n infra
```

Usually caused by missing PVC storage class or insufficient node resources.

### Certificate not issued

```bash
kubectl describe certificate -n shop
kubectl logs -n cert-manager -l app=cert-manager
```

Ensure DNS A records resolve to the VPS IP and port 80 is reachable from the internet.

### Infrastructure pods not starting

```bash
kubectl logs -n infra postgres-0
kubectl logs -n infra elasticsearch-0
```

Check that secrets exist: `kubectl get secrets -n infra`

### WebSocket connection fails

Verify ingress annotations and chat-service session affinity:

```bash
kubectl get ingress shop-ingress -n shop -o yaml
kubectl get svc chat-service -n shop -o yaml
```

### Migration job failed

```bash
kubectl get jobs -n shop
kubectl logs job/migrate-auth-service-<timestamp> -n shop
```

Ensure PostgreSQL is running and `auth-service-secret` contains `DB_PASSWORD`.

### Cannot join worker nodes

```bash
cat k8s/cluster-setup/kubeadm-join.sh
bash k8s/cluster-setup/bootstrap-worker.sh
```

Join tokens expire after 24 hours. Re-run `bootstrap-master.sh` to generate a new token.

## Fake Data

After `make seed`, the database contains:

| Entity | Count |
|--------|-------|
| Users | 50 (1 admin + 49 customers) |
| Categories | 10 |
| Products | 200 |
| Reviews | 500 |
| Orders | 300 |
| Chat rooms | 20 |
| Chat messages | ~400 |

All seed scripts are idempotent — safe to run multiple times.

## License

MIT
