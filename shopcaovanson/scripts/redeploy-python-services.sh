#!/usr/bin/env bash
# Khôi phục notification-service + search-service (probe port 8000, image local)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NS="${NAMESPACE:-shop}"

log() { echo "[redeploy-python] $*"; }

command -v kubectl >/dev/null 2>&1 || { echo "kubectl not found" >&2; exit 1; }

log "Applying manifests (probe port 8000, KAFKA_BOOTSTRAP_SERVERS)..."
kubectl apply -f "${PROJECT_ROOT}/k8s/base/notification-service/configmap.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/notification-service/deployment.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/search-service/configmap.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/search-service/deployment.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/notification-service/service.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/search-service/service.yaml"

if command -v nerdctl >/dev/null 2>&1 || (command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1); then
  log "Rebuilding images..."
  export BUILDKIT_HOST="${BUILDKIT_HOST:-unix:///run/buildkit/buildkitd.sock}"
  bash "${SCRIPT_DIR}/build-images.sh" notification-service search-service
else
  log "WARN: skip build — dùng image local hiện có"
fi

for dep in notification-service search-service; do
  log "Removing all ReplicaSets for ${dep}..."
  kubectl get rs -n "${NS}" -o name 2>/dev/null \
    | grep "/${dep}-" \
    | xargs -r kubectl delete -n "${NS}" --cascade=foreground --grace-period=0 --force 2>/dev/null || true
done

log "Restarting deployments..."
kubectl rollout restart deployment/notification-service deployment/search-service -n "${NS}"

for dep in notification-service search-service; do
  log "Waiting for ${dep}..."
  if ! kubectl rollout status "deployment/${dep}" -n "${NS}" --timeout=180s; then
    log "WARN: ${dep} chưa ready — logs:"
    pod=$(kubectl get pods -n "${NS}" -l "app=${dep}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
    [[ -n "${pod}" ]] && kubectl logs "${pod}" -n "${NS}" --tail=25 2>/dev/null || true
    [[ -n "${pod}" ]] && kubectl describe pod "${pod}" -n "${NS}" 2>/dev/null | sed -n '/Events:/,$p' | head -12 || true
  else
    log "✓ ${dep} ready"
  fi
done

echo ""
kubectl get pods -n "${NS}" -l 'app in (notification-service,search-service)' -o wide
