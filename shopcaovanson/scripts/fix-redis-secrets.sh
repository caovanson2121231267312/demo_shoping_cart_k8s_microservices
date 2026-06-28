#!/usr/bin/env bash
# Sửa REDIS_URL — đồng bộ password với Redis pod đang chạy (tránh WRONGPASS)
#
# Nguyên nhân WRONGPASS: đổi redis-secret / .env.production nhưng redis-0 chưa restart
# → pod vẫn dùng password cũ, shop secrets có password mới.
#
# Usage: bash scripts/fix-redis-secrets.sh
set -euo pipefail

log() { echo "[fix-redis-secrets] $*"; }
die() { echo "[fix-redis-secrets] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

urlencode() {
  python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$1" 2>/dev/null \
    || die "python3 required to URL-encode Redis password (có ký tự / + = ...)"
}

build_redis_url() {
  local db="$1"
  local enc
  enc=$(urlencode "${REDIS_PASS}")
  echo "redis://:${enc}@redis.infra.svc.cluster.local:6379/${db}"
}

get_live_redis_password() {
  local pass=""
  if kubectl get pod redis-0 -n infra >/dev/null 2>&1; then
    pass=$(kubectl exec redis-0 -n infra -- printenv REDIS_PASSWORD 2>/dev/null || true)
    if [[ -n "${pass}" ]]; then
      if kubectl exec redis-0 -n infra -- sh -c 'redis-cli -a "$REDIS_PASSWORD" ping' 2>/dev/null | grep -q PONG; then
        echo "${pass}"
        return 0
      fi
    fi
  fi
  return 1
}

REDIS_PASS=""
if REDIS_PASS=$(get_live_redis_password); then
  log "Dùng password từ redis-0 đang chạy (đã verify PONG)"
else
  log "Không lấy được từ pod — thử infra/redis-secret..."
  REDIS_PASS=$(kubectl get secret redis-secret -n infra -o jsonpath='{.data.REDIS_PASSWORD}' 2>/dev/null | base64 -d) || true
  if [[ -n "${REDIS_PASS}" ]] && kubectl get pod redis-0 -n infra >/dev/null 2>&1; then
    if ! kubectl exec redis-0 -n infra -- redis-cli -a "${REDIS_PASS}" ping 2>/dev/null | grep -q PONG; then
      die "Password trong infra/redis-secret KHÔNG khớp redis-0. Chạy: bash scripts/reset-redis.sh --yes"
    fi
  fi
fi

[[ -n "${REDIS_PASS}" ]] || die "Không xác định được REDIS_PASSWORD"

# Đồng bộ infra secret với password thực tế trên pod
kubectl create secret generic redis-secret -n infra \
  --from-literal=REDIS_PASSWORD="${REDIS_PASS}" \
  --dry-run=client -o yaml | kubectl apply -f -
log "✓ infra/redis-secret synced"

REDIS_URL_0=$(build_redis_url 0)
REDIS_URL_2=$(build_redis_url 2)
log "REDIS_URL (encoded): redis://:***@redis.infra.svc.cluster.local:6379/0"

patch_secret() {
  local name="$1"
  shift
  log "Updating shop/${name}..."
  kubectl create secret generic "${name}" -n shop "$@" \
    --dry-run=client -o yaml | kubectl apply -f -
}

jwt_pub=$(kubectl get secret api-gateway-secret -n shop -o jsonpath='{.data.JWT_PUBLIC_KEY}' 2>/dev/null | base64 -d || true)
jwt_priv=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.JWT_PRIVATE_KEY}' 2>/dev/null | base64 -d || true)
jwt_pub_auth=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.JWT_PUBLIC_KEY}' 2>/dev/null | base64 -d || true)
db_pass=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d || true)

[[ -n "${jwt_pub}" ]] || die "Missing JWT_PUBLIC_KEY — chạy: bash scripts/create-secrets.sh --force"

patch_secret api-gateway-secret \
  --from-literal=REDIS_URL="${REDIS_URL_0}" \
  --from-literal=JWT_PUBLIC_KEY="${jwt_pub}"

if [[ -n "${jwt_priv}" && -n "${db_pass}" ]]; then
  patch_secret auth-service-secret \
    --from-literal=DB_PASSWORD="${db_pass}" \
    --from-literal=REDIS_URL="${REDIS_URL_0}" \
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
    --from-literal=REDIS_URL="${REDIS_URL_2}" \
    --from-literal=JWT_PUBLIC_KEY="${chat_jwt:-${jwt_pub}}"
fi

order_db=$(kubectl get secret order-service-secret -n shop -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d || true)
if [[ -n "${order_db}" ]]; then
  patch_secret order-service-secret \
    --from-literal=DB_PASSWORD="${order_db}" \
    --from-literal=REDIS_PASSWORD="${REDIS_PASS}"
fi

# Cập nhật .env.production nếu có (tránh sed lỗi khi password có /)
ENV_FILE="${ENV_FILE:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.env.production}"
if [[ -f "${ENV_FILE}" ]] && command -v python3 >/dev/null 2>&1; then
  python3 - "${ENV_FILE}" "${REDIS_PASS}" <<'PY'
import pathlib, re, sys
path = pathlib.Path(sys.argv[1])
val = sys.argv[2]
text = path.read_text(encoding="utf-8")
if re.search(r"^REDIS_PASSWORD=", text, re.M):
    text = re.sub(r"^REDIS_PASSWORD=.*$", f"REDIS_PASSWORD={val}", text, count=1, flags=re.M)
    path.write_text(text, encoding="utf-8")
PY
  log "✓ .env.production REDIS_PASSWORD synced"
fi

log "Restarting app deployments..."
kubectl rollout restart deployment -n shop \
  api-gateway auth-service chat-service order-service 2>/dev/null || true

echo ""
log "Verify:"
echo "  kubectl exec redis-0 -n infra -- sh -c 'redis-cli -a \"\$REDIS_PASSWORD\" ping'"
echo "  kubectl rollout status deployment/api-gateway -n shop --timeout=120s"
echo "  kubectl logs -n shop -l app=api-gateway --tail=10"
