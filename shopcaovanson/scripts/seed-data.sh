#!/usr/bin/env bash
# shopcaovanson — Seed fake data for all services (idempotent)
# Usage:
#   LOCAL=true bash scripts/seed-data.sh     # local dev (go run scripts/fake_data.go)
#   bash scripts/seed-data.sh                # K8s cluster (kubectl jobs)
#   JOB_TIMEOUT=900 bash scripts/seed-data.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NAMESPACE="${NAMESPACE:-shop}"
SEED_IMAGE_TAG="${SEED_IMAGE_TAG:-latest}"
JOB_TIMEOUT="${JOB_TIMEOUT:-600}"

# shellcheck source=lib/k8s-job-wait.sh
source "${SCRIPT_DIR}/lib/k8s-job-wait.sh"

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
  local i=0
  local total=${#services[@]}
  for svc in "${services[@]}"; do
    i=$((i + 1))
    local dir="${PROJECT_ROOT}/services/${svc}"
    if [[ ! -f "${dir}/scripts/fake_data.go" ]]; then
      die "fake_data.go not found for ${svc}"
    fi
    load_env "${dir}"
    log "[${i}/${total}] Seeding ${svc}..."
    (cd "${dir}" && go run ./scripts/)
    log "[${i}/${total}] ✓ ${svc}"
  done

  log "All local seed scripts completed."
}

check_kubectl() {
  command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
  kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster. Set KUBECONFIG."
}

run_seed_job() {
  local service="$1"
  local step="$2"
  local total="$3"
  local job_name="seed-${service}-$(date +%s)"

  log "[${step}/${total}] Seed ${service} — job/${job_name}"

  kubectl -n "${NAMESPACE}" delete job -l "app.kubernetes.io/seed=${service}" --ignore-not-found=true 2>/dev/null || true

  cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${job_name}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/seed: ${service}
spec:
  backoffLimit: 2
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/seed: ${service}
    spec:
      restartPolicy: Never
      containers:
      - name: seed
        image: ghcr.io/caovanson/${service}:${SEED_IMAGE_TAG}
        imagePullPolicy: IfNotPresent
        command: ["/bin/sh", "-c"]
        args:
          - |
            set -e
            echo "[seed] starting ${service}..."
            /app/seed
            echo "[seed] done ${service}"
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

  JOB_LOG_PREFIX="[seed-data]" wait_for_k8s_job "${job_name}" "${NAMESPACE}" "${JOB_TIMEOUT}" || die "Seed failed: ${service}"

  log "[${step}/${total}] --- logs ${job_name} ---"
  kubectl -n "${NAMESPACE}" logs "job/${job_name}" 2>/dev/null | tail -30 || true
  log "[${step}/${total}] ✓ ${service}"
}

run_k8s_seed() {
  check_kubectl

  log "Starting idempotent seed in dependency order..."

  local services=(auth-service product-service order-service chat-service)
  local total=${#services[@]}
  local i=0
  for svc in "${services[@]}"; do
    i=$((i + 1))
    run_seed_job "${svc}" "${i}" "${total}"
  done

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
