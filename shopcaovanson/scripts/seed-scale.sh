#!/usr/bin/env bash
# Seed dữ liệu quy mô lớn — load / stress test
#
# Profile load60m (1M user, 1M SP, 60M đơn):
#   bash scripts/seed-scale.sh --profile load60m --yes
#
# Dev nhanh:
#   bash scripts/seed-scale.sh --profile dev
#
# Tùy chỉnh:
#   SEED_USERS=1000000 SEED_PRODUCTS=1000000 SEED_ORDERS=60000000 bash scripts/seed-scale.sh
#
# Trên VPS (K8s jobs — cần image đã build /app/seed):
#   bash scripts/seed-scale.sh --profile load60m --yes --k8s
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NAMESPACE="${NAMESPACE:-shop}"
JOB_TIMEOUT="${JOB_TIMEOUT:-172800}"

PROFILE="custom"
USERS=1000000
PRODUCTS=1000000
ORDERS=60000000
REVIEWS=0
BATCH_SIZE=20000
BCRYPT_COST=10
SKIP_REVIEWS=true
SKIP_ORDERS=false
SKIP_USERS=false
SKIP_PRODUCTS=false
USE_K8S=false
ASSUME_YES=false

log() { echo "[seed-scale] $*"; }
die() { echo "[seed-scale] ERROR: $*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: bash scripts/seed-scale.sh [options]

Profiles:
  dev      — 50 users, 200 products, 300 orders
  load60m  — 1M users, 1M products, 60M orders (stress test)

Options:
  --profile NAME    dev | load60m | custom (default: custom)
  --yes             Bỏ qua xác nhận (bắt buộc cho scale lớn)
  --k8s             Chạy qua Kubernetes Job (VPS đã deploy)
  -h, --help

Env: SEED_USERS, SEED_PRODUCTS, SEED_ORDERS, SEED_BATCH_SIZE, SEED_BCRYPT_COST
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile) PROFILE="$2"; shift 2 ;;
    --yes|-y) ASSUME_YES=true; shift ;;
    --k8s) USE_K8S=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown option: $1" ;;
  esac
done

case "${PROFILE}" in
  dev)
    USERS=50
    PRODUCTS=200
    ORDERS=300
    REVIEWS=2000
    BATCH_SIZE=5000
    SKIP_REVIEWS=false
    ;;
  load60m)
    USERS=1000000
    PRODUCTS=1000000
    ORDERS=60000000
    REVIEWS=0
    BATCH_SIZE=20000
    SKIP_REVIEWS=true
    ;;
  custom) ;;
  *) die "Unknown profile: ${PROFILE}" ;;
esac

# Override từ env nếu set
USERS="${SEED_USERS:-${USERS}}"
PRODUCTS="${SEED_PRODUCTS:-${PRODUCTS}}"
ORDERS="${SEED_ORDERS:-${ORDERS}}"
REVIEWS="${SEED_REVIEWS:-${REVIEWS}}"
BATCH_SIZE="${SEED_BATCH_SIZE:-${BATCH_SIZE}}"
BCRYPT_COST="${SEED_BCRYPT_COST:-${BCRYPT_COST}}"

warn_scale() {
  if [[ "${PROFILE}" == "dev" ]]; then
    return
  fi
  echo ""
  log "=== Profile: ${PROFILE} ==="
  log "  Users:    ${USERS}"
  log "  Products: ${PRODUCTS}"
  log "  Orders:   ${ORDERS}"
  log "  Reviews:  ${REVIEWS}"
  log "  Batch:    ${BATCH_SIZE}"
  echo ""
  echo "CẢNH BÁO — scale lớn cần tài nguyên đáng kể:"
  echo "  • 1M users (bcrypt cost=${BCRYPT_COST}): ~30-60 phút"
  echo "  • 1M products (COPY bulk):              ~1-3 giờ"
  echo "  • 60M orders (COPY bulk):               ~15-40+ giờ"
  echo "  • Disk Postgres ước tính:               100-200 GB+"
  echo "  • VPS 16GB RAM: có thể OOM — khuyến nghị ≥32GB RAM, ≥250GB disk"
  echo "  • Bỏ qua Mongo detail + ES index để nhanh hơn"
  echo ""
  if [[ "${ASSUME_YES}" != "true" ]]; then
    read -r -p "Tiếp tục? (y/N) " confirm
    [[ "${confirm}" =~ ^[yY] ]] || { log "Đã hủy."; exit 0; }
  fi
}

load_env() {
  local dir="$1"
  [[ -f "${dir}/.env.production" ]] && set -a && source "${dir}/.env.production" && set +a
  [[ -f "${dir}/.env" ]] && set -a && source "${dir}/.env" && set +a
}

run_local_go_seed() {
  local svc="$1"
  shift
  local dir="${PROJECT_ROOT}/services/${svc}"
  load_env "${dir}"
  log "Local seed: ${svc}"
  (cd "${dir}" && go run "$@")
}

run_k8s_seed_job() {
  local service="$1"
  shift
  local -a extra_env=("$@")
  local job_name="seed-scale-${service}-$(date +%s)"

  command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

  log "K8s job ${service} (${job_name}) timeout=${JOB_TIMEOUT}s"

  kubectl -n "${NAMESPACE}" delete job -l "app.kubernetes.io/seed-scale=${service}" --ignore-not-found=true 2>/dev/null || true

  local env_yaml=""
  for kv in "${extra_env[@]}"; do
    key="${kv%%=*}"
    val="${kv#*=}"
    env_yaml="${env_yaml}
            - name: ${key}
              value: \"${val}\""
  done

  cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${job_name}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/seed-scale: ${service}
spec:
  backoffLimit: 1
  ttlSecondsAfterFinished: 86400
  template:
    metadata:
      labels:
        app.kubernetes.io/seed-scale: ${service}
    spec:
      restartPolicy: Never
      containers:
      - name: seed
        image: ghcr.io/caovanson/${service}:latest
        imagePullPolicy: IfNotPresent
        command: ["/app/seed"]
        env:${env_yaml}
        envFrom:
        - configMapRef:
            name: ${service}-config
        - secretRef:
            name: ${service}-secret
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "4000m"
EOF

  # shellcheck source=lib/k8s-job-wait.sh
  source "${SCRIPT_DIR}/lib/k8s-job-wait.sh"
  JOB_LOG_PREFIX="[seed-scale]" wait_for_k8s_job "${job_name}" "${NAMESPACE}" "${JOB_TIMEOUT}" \
    || die "Seed job failed: ${service}"
  kubectl -n "${NAMESPACE}" logs "job/${job_name}" 2>/dev/null | tail -40 || true
}

seed_auth() {
  local env_args=(
    "SEED_USERS=${USERS}"
    "SEED_BATCH_SIZE=${BATCH_SIZE}"
    "SEED_BCRYPT_COST=${BCRYPT_COST}"
  )
  if [[ "${USE_K8S}" == "true" ]]; then
    run_k8s_seed_job "auth-service" "${env_args[@]}"
  else
    export SEED_USERS="${USERS}" SEED_BATCH_SIZE="${BATCH_SIZE}" SEED_BCRYPT_COST="${BCRYPT_COST}"
    run_local_go_seed "auth-service" ./scripts/fake_data.go
  fi
}

seed_product() {
  local env_args=(
    "SEED_PRODUCTS=${PRODUCTS}"
    "SEED_REVIEWS=${REVIEWS}"
    "SEED_ARTICLES=120"
    "SEED_USERS=${USERS}"
    "SEED_BATCH_SIZE=${BATCH_SIZE}"
    "SEED_SKIP_MONGO_DETAILS=true"
    "SEED_BULK_PRODUCTS=true"
    "SEED_ES_INDEX=false"
    "SEED_REFRESH_TEXT=false"
  )
  if [[ "${USE_K8S}" == "true" ]]; then
    run_k8s_seed_job "product-service" "${env_args[@]}"
  else
    export SEED_PRODUCTS="${PRODUCTS}" SEED_REVIEWS="${REVIEWS}" SEED_ARTICLES=120
    export SEED_USERS="${USERS}" SEED_BATCH_SIZE="${BATCH_SIZE}"
    export SEED_SKIP_MONGO_DETAILS=true SEED_BULK_PRODUCTS=true SEED_ES_INDEX=false SEED_REFRESH_TEXT=false
    run_local_go_seed "product-service" ./scripts/fake_data.go ./scripts/es_bulk.go
  fi
}

seed_order() {
  local env_args=(
    "SEED_ORDERS=${ORDERS}"
    "SEED_USERS=${USERS}"
    "SEED_PRODUCTS=${PRODUCTS}"
    "SEED_BATCH_SIZE=${BATCH_SIZE}"
  )
  if [[ "${USE_K8S}" == "true" ]]; then
    run_k8s_seed_job "order-service" "${env_args[@]}"
  else
    export SEED_ORDERS="${ORDERS}" SEED_USERS="${USERS}" SEED_PRODUCTS="${PRODUCTS}" SEED_BATCH_SIZE="${BATCH_SIZE}"
    run_local_go_seed "order-service" ./scripts/fake_data.go
  fi
}

main() {
  warn_scale
  local start
  start=$(date +%s)

  if [[ "${SKIP_USERS}" != "true" ]]; then
    log "[1/3] Auth — ${USERS} users"
    seed_auth
  fi

  if [[ "${SKIP_PRODUCTS}" != "true" ]]; then
    log "[2/3] Product — ${PRODUCTS} products"
    seed_product
  fi

  if [[ "${SKIP_ORDERS}" != "true" ]]; then
    log "[3/3] Order — ${ORDERS} orders"
    seed_order
  fi

  local elapsed=$(( $(date +%s) - start ))
  log "=== Hoàn tất (${elapsed}s) ==="
  log "Login admin: admin@shop.com / Admin@123"
  log "Customers: user1@shop.com .. user${USERS}@shop.com / Customer@123"
}

main "$@"
