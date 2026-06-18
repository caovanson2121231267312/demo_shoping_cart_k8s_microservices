#!/usr/bin/env bash
# shopcaovanson — Deploy all infrastructure, cert-manager, ingress, and application services
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OVERLAY="${OVERLAY:-prod}"
K8S_DIR="${PROJECT_ROOT}/k8s/overlays/${OVERLAY}"
CERT_MANAGER_VERSION="${CERT_MANAGER_VERSION:-v1.14.5}"
INGRESS_VERSION="${INGRESS_VERSION:-4.10.0}"

log() { echo "[deploy-all] $*"; }
die() { echo "[deploy-all] ERROR: $*" >&2; exit 1; }

check_kubectl() {
  command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
  kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster. Set KUBECONFIG."
}

check_helm() {
  command -v helm >/dev/null 2>&1 || die "helm not found"
}

deploy_cert_manager() {
  log "Installing cert-manager ${CERT_MANAGER_VERSION}..."
  helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
  helm repo update jetstack
  helm upgrade --install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --create-namespace \
    --version "${CERT_MANAGER_VERSION}" \
    --set installCRDs=true \
    --wait --timeout 5m
  kubectl apply -f "${PROJECT_ROOT}/k8s/base/cert-manager/cluster-issuer.yaml"
  log "cert-manager deployed."
}

deploy_ingress() {
  log "Installing nginx-ingress ${INGRESS_VERSION}..."
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || true
  helm repo update ingress-nginx
  helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace ingress-nginx \
    --create-namespace \
    --version "${INGRESS_VERSION}" \
    --set controller.service.type=LoadBalancer \
    --set controller.admissionWebhooks.enabled=false \
    --wait --timeout 5m
  log "nginx-ingress deployed."
}

create_infra_secrets() {
  log "Checking infra secrets in namespace infra..."
  kubectl create namespace infra --dry-run=client -o yaml | kubectl apply -f -

  if ! kubectl get secret postgres-secret -n infra >/dev/null 2>&1; then
    die "Missing secret postgres-secret in infra. See README.md for kubectl create secret commands."
  fi
  if ! kubectl get secret mongodb-secret -n infra >/dev/null 2>&1; then
    die "Missing secret mongodb-secret in infra. See README.md for kubectl create secret commands."
  fi
  if ! kubectl get secret redis-secret -n infra >/dev/null 2>&1; then
    die "Missing secret redis-secret in infra. See README.md for kubectl create secret commands."
  fi
  if ! kubectl get secret elasticsearch-secret -n infra >/dev/null 2>&1; then
    die "Missing secret elasticsearch-secret in infra. See README.md for kubectl create secret commands."
  fi
}

create_app_secrets() {
  log "Checking application secrets in namespace shop..."
  kubectl create namespace shop --dry-run=client -o yaml | kubectl apply -f -

  local required_secrets=(
    api-gateway-secret
    auth-service-secret
    product-service-secret
    order-service-secret
    chat-service-secret
    notification-service-secret
    search-service-secret
  )

  for secret in "${required_secrets[@]}"; do
    if ! kubectl get secret "${secret}" -n shop >/dev/null 2>&1; then
      die "Missing secret ${secret} in shop. See README.md for kubectl create secret commands."
    fi
  done
}

deploy_manifests() {
  log "Deploying manifests from overlay: ${OVERLAY}..."
  kubectl apply -k "${K8S_DIR}"
}

wait_for_infra() {
  log "Waiting for infrastructure pods..."
  kubectl -n infra rollout status statefulset/postgres --timeout=300s
  kubectl -n infra rollout status statefulset/mongodb --timeout=300s
  kubectl -n infra rollout status statefulset/redis --timeout=300s
  kubectl -n infra rollout status statefulset/zookeeper --timeout=300s
  kubectl -n infra rollout status statefulset/kafka --timeout=300s
  kubectl -n infra rollout status statefulset/elasticsearch --timeout=300s
}

wait_for_apps() {
  log "Waiting for application pods..."
  local deployments=(
    api-gateway auth-service product-service order-service
    chat-service notification-service search-service frontend
  )
  for dep in "${deployments[@]}"; do
    kubectl -n shop rollout status "deployment/${dep}" --timeout=300s
  done
}

show_status() {
  log "Deployment status:"
  kubectl get pods,svc,ingress -n shop
  kubectl get pods,svc -n infra
  kubectl get clusterissuer
}

main() {
  check_kubectl
  check_helm
  deploy_cert_manager
  deploy_ingress
  create_infra_secrets
  create_app_secrets
  deploy_manifests
  wait_for_infra
  wait_for_apps
  show_status
  log "Deploy complete."
  log "Run migrations: bash scripts/migrate-all.sh"
  log "Seed data:       bash scripts/seed-data.sh"
}

main "$@"
