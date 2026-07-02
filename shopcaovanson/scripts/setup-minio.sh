#!/usr/bin/env bash
# Bật MinIO avatar storage trên cluster đã chạy (không cần tạo lại toàn bộ secrets).
#
# Usage:
#   bash scripts/setup-minio.sh
#   bash scripts/setup-minio.sh --force   # ghi đè minio-secret + auth-service MINIO keys
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"
FORCE=false

log() { echo "[setup-minio] $*"; }
die() { echo "[setup-minio] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --force) FORCE=true; shift ;;
    -h|--help)
      echo "Usage: bash scripts/setup-minio.sh [--force]"
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

rand_pass() { openssl rand -hex 24; }

MINIO_ROOT_USER="${MINIO_ROOT_USER:-shopminio}"
MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD:-}"

if [[ -f "${ENV_FILE}" ]]; then
  # shellcheck source=/dev/null
  set -a
  source "${ENV_FILE}"
  set +a
  MINIO_ROOT_USER="${MINIO_ROOT_USER:-shopminio}"
  MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD:-}"
fi

if [[ -z "${MINIO_ROOT_PASSWORD}" ]]; then
  if kubectl get secret minio-secret -n infra >/dev/null 2>&1; then
    MINIO_ROOT_USER=$(kubectl get secret minio-secret -n infra -o jsonpath='{.data.MINIO_ROOT_USER}' | base64 -d)
    MINIO_ROOT_PASSWORD=$(kubectl get secret minio-secret -n infra -o jsonpath='{.data.MINIO_ROOT_PASSWORD}' | base64 -d)
    log "Dùng credentials từ infra/minio-secret hiện có"
  else
    MINIO_ROOT_PASSWORD="$(rand_pass)"
    log "Sinh MINIO_ROOT_PASSWORD mới"
  fi
fi

apply_secret() {
  local ns="$1" name="$2"
  shift 2
  if [[ "${FORCE}" == "true" ]] || ! kubectl get secret "${name}" -n "${ns}" >/dev/null 2>&1; then
    kubectl create secret generic "${name}" -n "${ns}" "$@" \
      --dry-run=client -o yaml | kubectl apply -f -
    log "✓ secret/${name} -n ${ns}"
  else
    log "Skip ${ns}/${name} (đã tồn tại — dùng --force để ghi đè)"
  fi
}

log "1/4 Tạo namespace + secret MinIO..."
kubectl create namespace infra --dry-run=client -o yaml | kubectl apply -f -
apply_secret infra minio-secret \
  --from-literal=MINIO_ROOT_USER="${MINIO_ROOT_USER}" \
  --from-literal=MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD}"

log "2/4 Deploy MinIO StatefulSet..."
kubectl apply -f "${PROJECT_ROOT}/k8s/base/infra/minio.yaml"
kubectl rollout status statefulset/minio -n infra --timeout=180s

log "3/4 Cập nhật auth-service-secret (MINIO_ACCESS_KEY / MINIO_SECRET_KEY)..."
kubectl create namespace shop --dry-run=client -o yaml | kubectl apply -f -

db_pass=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d || true)
redis_url=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.REDIS_URL}' 2>/dev/null | base64 -d || true)
jwt_priv=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.JWT_PRIVATE_KEY}' 2>/dev/null | base64 -d || true)
jwt_pub=$(kubectl get secret auth-service-secret -n shop -o jsonpath='{.data.JWT_PUBLIC_KEY}' 2>/dev/null | base64 -d || true)

[[ -n "${db_pass}" && -n "${redis_url}" && -n "${jwt_priv}" ]] \
  || die "Thiếu auth-service-secret — chạy: bash scripts/create-secrets.sh --apply-only"

kubectl create secret generic auth-service-secret -n shop \
  --from-literal=DB_PASSWORD="${db_pass}" \
  --from-literal=REDIS_URL="${redis_url}" \
  --from-literal=JWT_PRIVATE_KEY="${jwt_priv}" \
  --from-literal=JWT_PUBLIC_KEY="${jwt_pub:-}" \
  --from-literal=MINIO_ACCESS_KEY="${MINIO_ROOT_USER}" \
  --from-literal=MINIO_SECRET_KEY="${MINIO_ROOT_PASSWORD}" \
  --dry-run=client -o yaml | kubectl apply -f -
log "✓ shop/auth-service-secret (đã thêm MINIO keys)"

log "4/4 Bật MINIO_ENABLED và restart auth-service..."
kubectl create configmap auth-service-config -n shop \
  --from-literal=APP_ENV=production \
  --from-literal=PORT=8080 \
  --from-literal=LOG_LEVEL=info \
  --from-literal=DB_HOST=postgres.infra.svc.cluster.local \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_NAME=auth_db \
  --from-literal=DB_USER=shopcaovanson \
  --from-literal=DB_SSLMODE=disable \
  --from-literal=REDIS_HOST=redis.infra.svc.cluster.local \
  --from-literal=REDIS_PORT=6379 \
  --from-literal=REDIS_DB=0 \
  --from-literal=KAFKA_BROKERS=kafka.infra.svc.cluster.local:9092 \
  --from-literal=JWT_ACCESS_TTL_MINUTES=15 \
  --from-literal=JWT_REFRESH_TTL_DAYS=7 \
  --from-literal=KAFKA_TOPIC_USER_REGISTERED=user.registered \
  --from-literal=MINIO_ENABLED=true \
  --from-literal=MINIO_ENDPOINT=minio.infra.svc.cluster.local:9000 \
  --from-literal=MINIO_BUCKET=shopcaovanson \
  --from-literal=MINIO_USE_SSL=false \
  --from-literal=MINIO_REGION=us-east-1 \
  --from-literal=MAX_AVATAR_MB=2 \
  --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

# Prefer kustomize configmap if user deploys via overlay — patch only MINIO keys on existing configmap
kubectl patch configmap auth-service-config -n shop --type merge -p \
  '{"data":{"MINIO_ENABLED":"true","MINIO_ENDPOINT":"minio.infra.svc.cluster.local:9000","MINIO_BUCKET":"shopcaovanson","MINIO_USE_SSL":"false","MINIO_REGION":"us-east-1","MAX_AVATAR_MB":"2"}}' \
  2>/dev/null || log "ConfigMap auth-service-config chưa có — deploy lại auth-service từ k8s/base"

kubectl rollout restart deployment/auth-service -n shop
kubectl rollout status deployment/auth-service -n shop --timeout=180s

echo ""
log "=========================================="
log " MinIO avatar storage đã sẵn sàng"
log "=========================================="
echo ""
echo "  MinIO user:     ${MINIO_ROOT_USER}"
echo "  MinIO password: (lưu trong infra/minio-secret)"
echo ""
echo "  Kiểm tra:"
echo "    kubectl get pods -n infra -l app=minio"
echo "    kubectl logs -n shop -l app=auth-service --tail=20 | grep -i minio"
echo ""
