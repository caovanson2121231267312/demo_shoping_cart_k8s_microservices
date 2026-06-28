#!/usr/bin/env bash
# Sửa REDIS_URL trong shop secrets — khi api-gateway log: ping redis [::1]:6379
set -euo pipefail

log() { echo "[fix-redis-secrets] $*"; }
die() { echo "[fix-redis-secrets] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

REDIS_PASS=$(kubectl get secret redis-secret -n infra -o jsonpath='{.data.REDIS_PASSWORD}' 2>/dev/null | base64 -d) || true
[[ -n "${REDIS_PASS}" ]] || die "Missing infra/redis-secret — chạy: bash scripts/create-secrets.sh"

# Password hex từ create-secrets — không cần urlencode
REDIS_BASE="redis://:${REDIS_PASS}@redis.infra.svc.cluster.local:6379"

patch_secret() {
  local name="$1"
  shift
  log "Updating secret/${name}..."
  kubectl create secret generic "${name}" -n shop "$@" \
    --dry-run=client -o yaml | kubectl apply -f -
}

# Giữ JWT keys nếu đã có
jwt_pub=$(kubectl get secret api-gateway-secret -n shop -o jsonpath='{.data.JWT_PUBLIC_KEY}' 2>/dev/null | base64 -d || true)
jwt_priv=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.JWT_PRIVATE_KEY}' 2>/dev/null | base64 -d || true)
jwt_pub_auth=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.JWT_PUBLIC_KEY}' 2>/dev/null | base64 -d || true)
db_pass=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d || true)

[[ -n "${jwt_pub}" ]] || die "Missing JWT_PUBLIC_KEY in api-gateway-secret — chạy: bash scripts/create-secrets.sh --force"

patch_secret api-gateway-secret \
  --from-literal=REDIS_URL="${REDIS_BASE}/0" \
  --from-literal=JWT_PUBLIC_KEY="${jwt_pub}"

if [[ -n "${jwt_priv}" && -n "${db_pass}" ]]; then
  patch_secret auth-service-secret \
    --from-literal=DB_PASSWORD="${db_pass}" \
    --from-literal=REDIS_URL="${REDIS_BASE}/0" \
    --from-literal=JWT_PRIVATE_KEY="${jwt_priv}" \
    --from-literal=JWT_PUBLIC_KEY="${jwt_pub_auth:-${jwt_pub}}"
fi

chat_jwt=$(kubectl get secret chat-service-secret -n shop -o jsonpath='{.data.JWT_PUBLIC_KEY}' 2>/dev/null | base64 -d || true)
mongo_uri=$(kubectl get secret chat-service-secret -n shop -o jsonpath='{.data.MONGODB_URI}' 2>/dev/null | base64 -d || true)
mongo_pass=$(kubectl get secret chat-service-secret -n shop -o jsonpath='{.data.MONGO_PASSWORD}' 2>/dev/null | base64 -d || true)
if [[ -n "${mongo_uri}" ]]; then
  patch_secret chat-service-secret \
    --from-literal=MONGO_PASSWORD="${mongo_pass}" \
    --from-literal=MONGODB_URI="${mongo_uri}" \
    --from-literal=REDIS_URL="${REDIS_BASE}/2" \
    --from-literal=JWT_PUBLIC_KEY="${chat_jwt:-${jwt_pub}}"
fi

order_db=$(kubectl get secret order-service-secret -n shop -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d || true)
if [[ -n "${order_db}" ]]; then
  patch_secret order-service-secret \
    --from-literal=DB_PASSWORD="${order_db}" \
    --from-literal=REDIS_PASSWORD="${REDIS_PASS}"
fi

log "Restarting deployments..."
kubectl rollout restart deployment -n shop \
  api-gateway auth-service chat-service order-service 2>/dev/null || true

echo ""
log "✓ Done. Kiểm tra:"
echo "  kubectl get secret api-gateway-secret -n shop -o jsonpath='{.data.REDIS_URL}' | base64 -d; echo"
echo "  kubectl logs -n shop -l app=api-gateway --tail=20"
echo "  kubectl get pods -n shop"
