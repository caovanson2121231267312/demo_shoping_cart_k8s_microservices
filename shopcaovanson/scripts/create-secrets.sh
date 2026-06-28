#!/usr/bin/env bash
# shopcaovanson — Tạo .env.production + Kubernetes Secrets cho production
#
# Usage:
#   bash scripts/create-secrets.sh              # auto: tạo/sinh .env.production + apply secrets
#   bash scripts/create-secrets.sh --env-only   # chỉ tạo .env.production, chưa apply K8s
#   bash scripts/create-secrets.sh --apply-only # chỉ apply K8s từ .env.production có sẵn
#   bash scripts/create-secrets.sh --force      # ghi đè secrets đã tồn tại
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"
ENV_EXAMPLE="${PROJECT_ROOT}/.env.production.example"
JWT_DIR="${TMPDIR:-/tmp}/shopcaovanson-jwt-$$"

ENV_ONLY=false
APPLY_ONLY=false
FORCE=false

log() { echo "[create-secrets] $*"; }
die() { echo "[create-secrets] ERROR: $*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Tạo secrets production cho shopcaovanson trên Kubernetes.

  bash scripts/create-secrets.sh              # đầy đủ (khuyến nghị lần đầu)
  bash scripts/create-secrets.sh --env-only   # chỉ tạo file .env.production
  bash scripts/create-secrets.sh --apply-only # apply secrets từ .env.production
  bash scripts/create-secrets.sh --force      # cập nhật secrets đã có

File output: .env.production (chmod 600) — LƯU BACKUP, không commit git!
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --env-only)   ENV_ONLY=true; shift ;;
      --apply-only) APPLY_ONLY=true; shift ;;
      --force)      FORCE=true; shift ;;
      -h|--help)    usage; exit 0 ;;
      *) die "Unknown option: $1 (use --help)" ;;
    esac
  done
}

rand_pass() {
  openssl rand -hex 24
}

urlencode() {
  python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$1" 2>/dev/null \
    || printf '%s' "$1"
}

load_or_create_env() {
  if [[ "${APPLY_ONLY}" == "true" ]]; then
    [[ -f "${ENV_FILE}" ]] || die "Missing ${ENV_FILE}. Run without --apply-only first."
    return 0
  fi

  if [[ ! -f "${ENV_FILE}" ]]; then
    if [[ -f "${ENV_EXAMPLE}" ]]; then
      log "Creating ${ENV_FILE} from .env.production.example"
      cp "${ENV_EXAMPLE}" "${ENV_FILE}"
    else
      log "Creating empty ${ENV_FILE}"
      touch "${ENV_FILE}"
    fi
  fi

  # shellcheck source=/dev/null
  set -a
  source "${ENV_FILE}"
  set +a

  local changed=false
  ensure_var POSTGRES_PASSWORD && changed=true
  ensure_var MONGO_PASSWORD && changed=true
  ensure_var REDIS_PASSWORD && changed=true
  ensure_var ELASTIC_PASSWORD && changed=true

  : "${SMTP_HOST:=smtp.gmail.com}"
  : "${SMTP_PORT:=587}"
  : "${SMTP_FROM:=noreply@shopcaovanson.xyz}"
  : "${SMTP_USE_TLS:=true}"
  : "${SMTP_USERNAME:=}"
  : "${SMTP_PASSWORD:=}"
  : "${GHCR_USERNAME:=}"
  : "${GHCR_TOKEN:=}"
  : "${GHCR_EMAIL:=admin@shopcaovanson.xyz}"
  : "${API_DOMAIN:=vocabee.cloud}"
  : "${FRONTEND_DOMAIN:=shopcaovanson.xyz}"
  : "${POSTGRES_USER:=shopcaovanson}"
  : "${MONGO_USER:=shopcaovanson}"

  write_env_file
  chmod 600 "${ENV_FILE}"
  log "Saved: ${ENV_FILE}"
}

ensure_var() {
  local name="$1"
  local val="${!name:-}"
  if [[ -z "${val}" ]]; then
    printf -v "$name" '%s' "$(rand_pass)"
    return 0
  fi
  return 1
}

write_env_file() {
  cat > "${ENV_FILE}" <<EOF
# shopcaovanson production — generated $(date -Iseconds 2>/dev/null || date)
# KHÔNG commit file này vào git!

POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
MONGO_PASSWORD=${MONGO_PASSWORD}
REDIS_PASSWORD=${REDIS_PASSWORD}
ELASTIC_PASSWORD=${ELASTIC_PASSWORD}

SMTP_HOST=${SMTP_HOST}
SMTP_PORT=${SMTP_PORT}
SMTP_USERNAME=${SMTP_USERNAME}
SMTP_PASSWORD=${SMTP_PASSWORD}
SMTP_FROM=${SMTP_FROM}
SMTP_USE_TLS=${SMTP_USE_TLS}

GHCR_USERNAME=${GHCR_USERNAME}
GHCR_TOKEN=${GHCR_TOKEN}
GHCR_EMAIL=${GHCR_EMAIL}

API_DOMAIN=${API_DOMAIN}
FRONTEND_DOMAIN=${FRONTEND_DOMAIN}
POSTGRES_USER=${POSTGRES_USER}
MONGO_USER=${MONGO_USER}
EOF
}

source_env() {
  # shellcheck source=/dev/null
  set -a
  source "${ENV_FILE}"
  set +a

  : "${POSTGRES_USER:=shopcaovanson}"
  : "${MONGO_USER:=shopcaovanson}"
  : "${API_DOMAIN:=vocabee.cloud}"
  : "${SMTP_HOST:=smtp.gmail.com}"
  : "${SMTP_PORT:=587}"
  : "${SMTP_FROM:=noreply@shopcaovanson.xyz}"
  : "${SMTP_USE_TLS:=true}"
  : "${SMTP_USERNAME:=}"
  : "${SMTP_PASSWORD:=}"

  local enc_redis enc_mongo enc_elastic
  enc_redis=$(urlencode "${REDIS_PASSWORD}")
  enc_mongo=$(urlencode "${MONGO_PASSWORD}")
  enc_elastic=$(urlencode "${ELASTIC_PASSWORD}")

  REDIS_URL="redis://:${enc_redis}@redis.infra.svc.cluster.local:6379"
  MONGO_URI_PRODUCT="mongodb://${MONGO_USER}:${enc_mongo}@mongodb.infra.svc.cluster.local:27017/product_db?authSource=admin"
  MONGO_URI_CHAT="mongodb://${MONGO_USER}:${enc_mongo}@mongodb.infra.svc.cluster.local:27017/chat_db?authSource=admin"
  ELASTICSEARCH_URL="http://elastic:${enc_elastic}@elasticsearch.infra.svc.cluster.local:9200"
}

apply_secret() {
  local ns="$1"
  local name="$2"
  shift 2
  if [[ "${FORCE}" == "true" ]]; then
    kubectl create secret generic "${name}" -n "${ns}" "$@" \
      --dry-run=client -o yaml | kubectl apply -f -
  else
    if kubectl get secret "${name}" -n "${ns}" >/dev/null 2>&1; then
      log "Skip ${ns}/${name} (exists — dùng --force để cập nhật)"
      return 0
    fi
    kubectl create secret generic "${name}" -n "${ns}" "$@"
  fi
  log "✓ secret/${name} -n ${ns}"
}

generate_jwt_keys() {
  mkdir -p "${JWT_DIR}"
  chmod 700 "${JWT_DIR}"
  if [[ ! -f "${JWT_DIR}/jwt-private.pem" ]]; then
    openssl genrsa -out "${JWT_DIR}/jwt-private.pem" 2048 2>/dev/null
    openssl rsa -in "${JWT_DIR}/jwt-private.pem" -pubout -out "${JWT_DIR}/jwt-public.pem" 2>/dev/null
  fi
}

apply_infra_secrets() {
  log "Creating infra secrets (namespace: infra)..."
  kubectl create namespace infra --dry-run=client -o yaml | kubectl apply -f -

  apply_secret infra postgres-secret \
    --from-literal=POSTGRES_PASSWORD="${POSTGRES_PASSWORD}"

  apply_secret infra mongodb-secret \
    --from-literal=MONGO_PASSWORD="${MONGO_PASSWORD}"

  apply_secret infra redis-secret \
    --from-literal=REDIS_PASSWORD="${REDIS_PASSWORD}"

  apply_secret infra elasticsearch-secret \
    --from-literal=ELASTIC_PASSWORD="${ELASTIC_PASSWORD}"
}

apply_shop_secrets() {
  log "Creating shop secrets (namespace: shop)..."
  kubectl create namespace shop --dry-run=client -o yaml | kubectl apply -f -

  generate_jwt_keys

  apply_secret shop auth-service-secret \
    --from-literal=DB_PASSWORD="${POSTGRES_PASSWORD}" \
    --from-literal=REDIS_URL="${REDIS_URL}/0" \
    --from-file=JWT_PRIVATE_KEY="${JWT_DIR}/jwt-private.pem" \
    --from-file=JWT_PUBLIC_KEY="${JWT_DIR}/jwt-public.pem"

  apply_secret shop api-gateway-secret \
    --from-literal=REDIS_URL="${REDIS_URL}/0" \
    --from-file=JWT_PUBLIC_KEY="${JWT_DIR}/jwt-public.pem"

  apply_secret shop product-service-secret \
    --from-literal=DB_PASSWORD="${POSTGRES_PASSWORD}" \
    --from-literal=MONGO_PASSWORD="${MONGO_PASSWORD}" \
    --from-literal=MONGO_URI="${MONGO_URI_PRODUCT}" \
    --from-literal=ELASTIC_PASSWORD="${ELASTIC_PASSWORD}" \
    --from-literal=ELASTICSEARCH_URL="${ELASTICSEARCH_URL}"

  apply_secret shop order-service-secret \
    --from-literal=DB_PASSWORD="${POSTGRES_PASSWORD}" \
    --from-literal=REDIS_PASSWORD="${REDIS_PASSWORD}"

  apply_secret shop chat-service-secret \
    --from-literal=MONGO_PASSWORD="${MONGO_PASSWORD}" \
    --from-literal=MONGODB_URI="${MONGO_URI_CHAT}" \
    --from-literal=REDIS_URL="${REDIS_URL}/2" \
    --from-file=JWT_PUBLIC_KEY="${JWT_DIR}/jwt-public.pem"

  apply_secret shop notification-service-secret \
    --from-literal=SMTP_HOST="${SMTP_HOST}" \
    --from-literal=SMTP_PORT="${SMTP_PORT}" \
    --from-literal=SMTP_USERNAME="${SMTP_USERNAME}" \
    --from-literal=SMTP_PASSWORD="${SMTP_PASSWORD}" \
    --from-literal=SMTP_FROM="${SMTP_FROM}" \
    --from-literal=SMTP_USE_TLS="${SMTP_USE_TLS}"

  apply_secret shop search-service-secret \
    --from-literal=ELASTIC_PASSWORD="${ELASTIC_PASSWORD}" \
    --from-literal=ELASTICSEARCH_URL="${ELASTICSEARCH_URL}"

  if [[ -n "${GHCR_USERNAME}" && -n "${GHCR_TOKEN}" ]]; then
    if [[ "${FORCE}" == "true" ]]; then
      kubectl create secret docker-registry ghcr-secret -n shop \
        --docker-server=ghcr.io \
        --docker-username="${GHCR_USERNAME}" \
        --docker-password="${GHCR_TOKEN}" \
        --docker-email="${GHCR_EMAIL}" \
        --dry-run=client -o yaml | kubectl apply -f -
    elif ! kubectl get secret ghcr-secret -n shop >/dev/null 2>&1; then
      kubectl create secret docker-registry ghcr-secret -n shop \
        --docker-server=ghcr.io \
        --docker-username="${GHCR_USERNAME}" \
        --docker-password="${GHCR_TOKEN}" \
        --docker-email="${GHCR_EMAIL}"
    fi
    log "✓ secret/ghcr-secret -n shop"
  else
    log "Skip ghcr-secret (GHCR_USERNAME/GHCR_TOKEN trống — OK nếu image public)"
  fi
}

cleanup() {
  rm -rf "${JWT_DIR}" 2>/dev/null || true
}
trap cleanup EXIT

verify_secrets() {
  log "Verifying secrets..."
  local missing=0
  for s in postgres-secret mongodb-secret redis-secret elasticsearch-secret; do
    kubectl get secret "$s" -n infra >/dev/null 2>&1 || { log "MISSING infra/$s"; missing=1; }
  done
  for s in auth-service-secret api-gateway-secret product-service-secret order-service-secret \
           chat-service-secret notification-service-secret search-service-secret; do
    kubectl get secret "$s" -n shop >/dev/null 2>&1 || { log "MISSING shop/$s"; missing=1; }
  done
  [[ "${missing}" -eq 0 ]] || die "Some secrets missing"
  log "All required secrets OK"
}

print_summary() {
  echo ""
  log "=========================================="
  log " HOÀN TẤT — Production secrets"
  log "=========================================="
  echo ""
  echo "  File lưu mật khẩu:  ${ENV_FILE}"
  echo "  (chmod 600 — backup vào password manager!)"
  echo ""
  echo "  Infra secrets:  postgres, mongodb, redis, elasticsearch"
  echo "  Shop secrets:   auth, api-gateway, product, order, chat,"
  echo "                  notification, search"
  echo ""
  echo "  Bước tiếp theo:"
  echo "    OVERLAY=dev bash scripts/deploy-all.sh"
  echo "    bash scripts/migrate-all.sh"
  echo "    bash scripts/seed-data.sh"
  echo ""
  if [[ -z "${SMTP_USERNAME}" ]]; then
    echo "  ⚠️  SMTP chưa cấu hình — email notification sẽ không gửi được."
    echo "      Sửa SMTP_* trong ${ENV_FILE} rồi chạy lại với --force"
    echo ""
  fi
}

main() {
  parse_args "$@"

  command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
  command -v openssl >/dev/null 2>&1 || die "openssl not found"

  if [[ "${APPLY_ONLY}" != "true" ]]; then
    load_or_create_env
    if [[ "${ENV_ONLY}" == "true" ]]; then
      log "Done (--env-only). Edit ${ENV_FILE} then run:"
      log "  bash scripts/create-secrets.sh --apply-only"
      exit 0
    fi
  fi

  [[ -f "${ENV_FILE}" ]] || die "Missing ${ENV_FILE}"
  source_env

  kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to Kubernetes cluster"

  apply_infra_secrets
  apply_shop_secrets
  verify_secrets
  print_summary
}

main "$@"
