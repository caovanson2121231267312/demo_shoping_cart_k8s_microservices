.PHONY: help setup-cluster verify-cluster deploy-infra deploy-cert-manager deploy-ingress deploy-services deploy-frontend deploy-all migrate seed logs restart status build

PROJECT_ROOT := $(shell pwd)
K8S_BASE := $(PROJECT_ROOT)/k8s/base
K8S_OVERLAY ?= prod
KUBECONFIG ?= $(HOME)/.kube/shopcaovanson-config

export KUBECONFIG

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# Cluster
setup-cluster: ## Setup KVM multi-node K8s cluster
	bash scripts/setup-cluster.sh

verify-cluster: ## Verify cluster nodes and system pods
	bash k8s/cluster-setup/verify-cluster.sh

# Deploy
deploy-cert-manager: ## Install cert-manager via Helm
	helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
	helm repo update jetstack
	helm upgrade --install cert-manager jetstack/cert-manager \
		--namespace cert-manager --create-namespace \
		--version v1.14.5 --set installCRDs=true --wait
	kubectl apply -f $(K8S_BASE)/cert-manager/cluster-issuer.yaml

deploy-ingress: ## Install nginx-ingress via Helm
	helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
	helm repo update ingress-nginx
	helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
		--namespace ingress-nginx --create-namespace \
		--version 4.10.0 \
		--set controller.service.type=LoadBalancer \
		--set controller.admissionWebhooks.enabled=false --wait

deploy-infra: ## Deploy Kafka, Postgres, Mongo, Redis, Elastic
	kubectl apply -f $(K8S_BASE)/namespace.yaml
	kubectl apply -f $(K8S_BASE)/infra/ -n infra

deploy-services: ## Deploy all microservices
	kubectl apply -f $(K8S_BASE)/namespace.yaml
	kubectl apply -k $(PROJECT_ROOT)/k8s/overlays/$(K8S_OVERLAY)

deploy-frontend: ## Deploy NuxtJS frontend
	kubectl apply -f $(K8S_BASE)/frontend/ -n shop

deploy-all: ## Full deploy: cert-manager + ingress + all manifests
	bash scripts/deploy-all.sh

# Data
migrate: ## Run all DB migrations
	bash scripts/migrate-all.sh

seed: ## Run fake data seeders
	bash scripts/seed-data.sh

seed-data: seed ## Alias for seed

# Dev
logs: ## Tail logs for a service (make logs svc=auth-service)
	@if [ -z "$(svc)" ]; then echo "Usage: make logs svc=auth-service"; exit 1; fi
	kubectl logs -f -l app=$(svc) -n shop --tail=100

restart: ## Restart a deployment (make restart svc=product-service)
	@if [ -z "$(svc)" ]; then echo "Usage: make restart svc=product-service"; exit 1; fi
	kubectl rollout restart deployment/$(svc) -n shop

status: ## Show pods, services, ingress in shop and infra
	@echo "=== shop namespace ==="
	kubectl get pods,svc,ingress -n shop
	@echo "=== infra namespace ==="
	kubectl get pods,svc -n infra

verify-app: ## Full health check — pods, deployments, HTTP endpoints
	bash scripts/verify-app.sh

apply-security: ## Apply NetworkPolicy, PSS, RBAC hardening
	bash scripts/apply-security.sh

cicd-kubeconfig: ## Generate kubeconfig for GitHub Actions (run on VPS)
	bash scripts/create-cicd-kubeconfig.sh

create-secrets: ## Generate .env.production + K8s secrets (run before deploy)
	bash scripts/create-secrets.sh

create-secrets-env: ## Only create .env.production file (no kubectl)
	bash scripts/create-secrets.sh --env-only

# CI
build: ## Build and push Docker image (make build svc=auth-service)
	@if [ -z "$(svc)" ]; then echo "Usage: make build svc=auth-service"; exit 1; fi
	docker build -t ghcr.io/caovanson/$(svc):latest services/$(svc)/
	docker push ghcr.io/caovanson/$(svc):latest

build-frontend: ## Build and push frontend image
	docker build -t ghcr.io/caovanson/frontend-web:latest frontend/web/
	docker push ghcr.io/caovanson/frontend-web:latest
