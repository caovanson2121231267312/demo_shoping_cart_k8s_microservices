#!/usr/bin/env bash
# shopcaovanson — Seed fake data for all services (idempotent)
# Usage:
#   LOCAL=true bash scripts/seed-data.sh     # local dev (go run scripts/fake_data.go)
#   bash scripts/seed-data.sh                # K8s cluster (kubectl jobs)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NAMESPACE="${NAMESPACE:-shop}"
SEED_IMAGE_TAG="${SEED_IMAGE_TAG:-latest}"

log() { echo "[seed-data] $*"; }
die() { echo "[seed-data] ERROR: $*" >&2; exit 1; }

load_env() {
  local dir="$1"
  if [[ -f "${dir}/.env" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "${dir}/.env"
    set +a
  elif [[ -f "${dir}/.env.example" ]]; then
    log "WARN: $(basename "${dir}")/.env not found, using .env.example"
    set -a
    # shellcheck disable=SC1090
    source "${dir}/.env.example"
    set +a
  fi
}

run_local_seed() {
  log "Starting local idempotent seed in dependency order..."

  local services=(auth-service product-service order-service chat-service)
  for svc in "${services[@]}"; do
    local dir="${PROJECT_ROOT}/services/${svc}"
    if [[ ! -f "${dir}/scripts/fake_data.go" ]]; then
      die "fake_data.go not found for ${svc}"
    fi
    load_env "${dir}"
    log "Seeding ${svc}..."
    (cd "${dir}" && go run scripts/fake_data.go)
    log "Seed for ${svc} completed."
  done

  log "All local seed scripts completed."
  log "Dev-scale data (override via SEED_* env vars)."
  log "For load testing (3M users, 1M orders, 50M reviews):"
  log "  powershell -File scripts/seed-scale.ps1 -Profile full"
}

check_kubectl() {
  command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
  kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster. Set KUBECONFIG."
}

run_seed_job() {
  local service="$1"
  local job_name="seed-${service}-$(date +%s)"

  log "Seeding data for ${service}..."

  kubectl -n "${NAMESPACE}" delete job "${job_name}" --ignore-not-found=true

  cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${job_name}
  namespace: ${NAMESPACE}
spec:
  backoffLimit: 2
  ttlSecondsAfterFinished: 600
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: seed
        image: ghcr.io/caovanson/${service}:${SEED_IMAGE_TAG}
        command: ["/app/seed"]
        envFrom:
        - configMapRef:
            name: ${service}-config
        - secretRef:
            name: ${service}-secret
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
EOF

  kubectl -n "${NAMESPACE}" wait --for=condition=complete "job/${job_name}" --timeout=600s
  log "Seed for ${service} completed."
  kubectl -n "${NAMESPACE}" logs "job/${job_name}"
}

run_k8s_seed() {
  check_kubectl

  log "Starting idempotent seed in dependency order..."

  run_seed_job "auth-service"
  run_seed_job "product-service"
  run_seed_job "order-service"
  run_seed_job "chat-service"

  log "All seed jobs completed."
  log "Expected data:"
  log "  - 50 users (1 admin: admin@shop.com / Admin@123)"
  log "  - 10 categories, 200 products, 500 reviews"
  log "  - 300 orders"
  log "  - 20 chat rooms, ~400 messages"
}

main() {
  if [[ "${LOCAL:-false}" == "true" ]]; then
    run_local_seed
  else
    run_k8s_seed
  fi
}

main "$@"
