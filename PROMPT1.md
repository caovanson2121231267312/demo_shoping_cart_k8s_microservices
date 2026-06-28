You are a senior DevOps + backend architect. Create a complete e-commerce project called "shopcaovanson" with the following exact specifications.

## PROJECT OVERVIEW
- Frontend: shopcaovanson.xyz (NuxtJS 3 + Vuetify 3)
- Backend API: vocabee.cloud (Go microservices)
- SSL: cert-manager with Let's Encrypt (production ClusterIssuer)
- K8s: multi-node setup (1 master + 2 workers via KVM on single VPS)
- Chat feature: real-time WebSocket

## ROOT DIRECTORY STRUCTURE
Create exactly this structure at project root:
```
shopcaovanson/
├── README.md                          # root overview + quick start
├── .gitignore
├── Makefile                           # top-level make commands
│
├── services/
│   ├── api-gateway/                   # Go - routes all HTTP traffic
│   ├── auth-service/                  # Go - login, register, JWT, refresh token
│   ├── product-service/               # Go - products, categories, search (Elastic)
│   ├── order-service/                 # Go - orders, cart, checkout
│   ├── chat-service/                  # Go - WebSocket real-time chat
│   ├── notification-service/          # Python - email/push via Kafka consumer
│   └── search-service/                # Python - Elasticsearch indexing via Kafka
│
├── frontend/
│   └── web/                           # NuxtJS 3 + Vuetify 3
│
├── k8s/
│   ├── cluster-setup/                 # KVM multi-node scripts
│   ├── base/                          # K8s manifests (namespace, deployments, services)
│   │   ├── namespace.yaml
│   │   ├── cert-manager/
│   │   ├── ingress/
│   │   ├── infra/                     # kafka, postgres, mongo, redis, elastic
│   │   ├── api-gateway/
│   │   ├── auth-service/
│   │   ├── product-service/
│   │   ├── order-service/
│   │   ├── chat-service/
│   │   ├── notification-service/
│   │   ├── search-service/
│   │   └── frontend/
│   └── overlays/
│       ├── dev/
│       └── prod/
│
├── scripts/
│   ├── setup-cluster.sh               # bootstrap KVM VMs + kubeadm
│   ├── deploy-all.sh
│   └── seed-data.sh
│
└── docs/
    ├── architecture.md
    └── api-spec.md
```

## EACH SERVICE MUST HAVE ITS OWN README.md containing:
- Service description and responsibility
- Tech stack used
- Environment variables table
- How to run locally (docker-compose)
- How to run migrations
- API endpoints (for Go services)
- Kafka topics consumed/produced

## K8s MULTI-NODE CLUSTER SETUP
Create `k8s/cluster-setup/` with:

1. `setup-vms.sh` - Uses KVM/libvirt to create 3 VMs:
   - master: 2 CPU, 3GB RAM, 30GB disk
   - worker-1: 2 CPU, 4GB RAM, 50GB disk  
   - worker-2: 2 CPU, 4GB RAM, 50GB disk
   Uses cloud-init for Ubuntu 22.04

2. `bootstrap-master.sh` - kubeadm init, Calico CNI, prints join command

3. `bootstrap-worker.sh` - kubeadm join (reads token from master)

4. `verify-cluster.sh` - checks all nodes Ready, all system pods Running

## K8s MANIFESTS - create all files in k8s/base/

### namespace.yaml
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: shop
---
apiVersion: v1
kind: Namespace
metadata:
  name: infra
```

### cert-manager/cluster-issuer.yaml
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@shopcaovanson.xyz
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### ingress/ingress.yaml
Single ingress covering both domains:
- shopcaovanson.xyz -> frontend service
- vocabee.cloud/api/* -> api-gateway service  
- vocabee.cloud/ws/* -> chat-service (with WebSocket annotations)
- TLS for both domains using cert-manager annotation

### infra/ - create StatefulSets for:
- PostgreSQL 15 (StatefulSet, PVC 20Gi, secret for password)
- MongoDB 7 (StatefulSet, PVC 15Gi)
- Redis 7 (StatefulSet, PVC 5Gi)
- Kafka + Zookeeper (StatefulSet, PVC 10Gi)
- Elasticsearch 8 (StatefulSet, PVC 20Gi, init container to set vm.max_map_count)

All infra services must:
- Have resource requests/limits appropriate for a 4GB worker node
- Use Secrets for credentials (not hardcoded)
- Have readinessProbe and livenessProbe

## Makefile at root level:
```makefile
setup-cluster:    ## Setup KVM multi-node K8s cluster
deploy-infra:     ## Deploy Kafka, Postgres, Mongo, Redis, Elastic
deploy-services:  ## Deploy all microservices
deploy-frontend:  ## Deploy NuxtJS frontend
seed-data:        ## Run fake data seeders
migrate:          ## Run all DB migrations
logs:             ## Tail logs for all services
status:           ## kubectl get all -n shop
```

Generate ALL files completely. Do not use placeholders. Every yaml, every script must be complete and runnable.