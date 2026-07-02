#!/usr/bin/env bash
# Tạo ghcr-secret để cluster pull image private từ GitHub Container Registry
# Chạy trên VPS sau khi CI push image lên ghcr.io
set -euo pipefail

NAMESPACE="${NAMESPACE:-shop}"
SECRET_NAME="${SECRET_NAME:-ghcr-secret}"

log() { echo "[setup-ghcr-pull-secret] $*"; }
die() { echo "[setup-ghcr-pull-secret] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

GHCR_USERNAME="${GHCR_USERNAME:-}"
GHCR_TOKEN="${GHCR_TOKEN:-}"
GHCR_EMAIL="${GHCR_EMAIL:-${GHCR_USERNAME}@users.noreply.github.com}"

if [[ -z "${GHCR_USERNAME}" || -z "${GHCR_TOKEN}" ]]; then
  echo "Tạo PAT: GitHub → Settings → Developer settings → Personal access tokens"
  echo "  Quyền cần: read:packages"
  echo ""
  read -r -p "GitHub username: " GHCR_USERNAME
  read -r -s -p "PAT (read:packages): " GHCR_TOKEN
  echo ""
fi

[[ -n "${GHCR_USERNAME}" && -n "${GHCR_TOKEN}" ]] || die "GHCR_USERNAME và GHCR_TOKEN bắt buộc"

kubectl create secret docker-registry "${SECRET_NAME}" -n "${NAMESPACE}" \
  --docker-server=ghcr.io \
  --docker-username="${GHCR_USERNAME}" \
  --docker-password="${GHCR_TOKEN}" \
  --docker-email="${GHCR_EMAIL}" \
  --dry-run=client -o yaml | kubectl apply -f -

log "✓ secret/${SECRET_NAME} -n ${NAMESPACE}"

SERVICES="api-gateway auth-service product-service order-service chat-service \
  notification-service search-service analytics-service rasa-service frontend"

for svc in ${SERVICES}; do
  if kubectl get deployment "${svc}" -n "${NAMESPACE}" >/dev/null 2>&1; then
    kubectl patch deployment "${svc}" -n "${NAMESPACE}" -p \
      '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"'"${SECRET_NAME}"'"}]}}}}' \
      >/dev/null 2>&1 && log "  ✓ ${svc}" || log "  skip ${svc}"
  fi
done

log ""
log "Xong. Re-run GitHub Actions deploy-k8s hoặc: kubectl rollout restart deployment -n shop"
