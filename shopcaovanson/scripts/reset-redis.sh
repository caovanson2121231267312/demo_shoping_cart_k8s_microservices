#!/usr/bin/env bash
# Reset Redis data + password — khi WRONGPASS không sync được
set -euo pipefail

NS="${NAMESPACE:-infra}"
STS="redis"
PVC="redis-data-redis-0"
YES=false

log() { echo "[reset-redis] $*"; }
die() { echo "[reset-redis] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes|-y) YES=true; shift ;;
    -h|--help)
      echo "Usage: bash scripts/reset-redis.sh [--yes]"
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

if [[ "${YES}" != "true" ]]; then
  echo ""
  echo "⚠️  XÓA dữ liệu Redis (cache/session). Sau đó password lấy từ infra/redis-secret."
  read -r -p "Tiếp tục? [y/N] " ans
  [[ "${ans}" =~ ^[Yy]$ ]] || { log "Cancelled."; exit 0; }
fi

REDIS_PASS=$(kubectl get secret redis-secret -n "${NS}" -o jsonpath='{.data.REDIS_PASSWORD}' | base64 -d) || true
[[ -n "${REDIS_PASS}" ]] || die "Missing redis-secret — bash scripts/create-secrets.sh"

log "Scaling ${STS} to 0..."
kubectl scale statefulset "${STS}" -n "${NS}" --replicas=0
kubectl wait --for=delete "pod/${STS}-0" -n "${NS}" --timeout=120s 2>/dev/null || true

if kubectl get pvc "${PVC}" -n "${NS}" >/dev/null 2>&1; then
  kubectl delete pvc "${PVC}" -n "${NS}" --wait=true
fi

kubectl scale statefulset "${STS}" -n "${NS}" --replicas=1
kubectl rollout status statefulset/"${STS}" -n "${NS}" --timeout=300s

kubectl exec "${STS}-0" -n "${NS}" -- sh -c 'redis-cli -a "$REDIS_PASSWORD" ping' | grep -q PONG
log "✓ Redis ready — chạy: bash scripts/fix-redis-secrets.sh"
