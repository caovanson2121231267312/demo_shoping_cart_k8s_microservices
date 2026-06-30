#!/usr/bin/env bash
# shopcaovanson — Run database migrations for auth, product, and order services
# Usage:
#   LOCAL=true bash scripts/migrate-all.sh     # local dev (go run)
#   bash scripts/migrate-all.sh                # K8s cluster (kubectl jobs)
#   JOB_TIMEOUT=600 bash scripts/migrate-all.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NAMESPACE="${NAMESPACE:-shop}"
MIGRATION_IMAGE_TAG="${MIGRATION_IMAGE_TAG:-latest}"
JOB_TIMEOUT="${JOB_TIMEOUT:-300}"

# shellcheck source=lib/k8s-job-wait.sh
source "${SCRIPT_DIR}/lib/k8s-job-wait.sh"

log() { echo "[migrate-all] $*"; }
die() { echo "[migrate-all] ERROR: $*" >&2; exit 1; }

run_local_migrations() {
  log "Running local migrations (go run cmd/migrate/main.go up)..."

  local services=(auth-service product-service order-service)
  local i=0
  local total=${#services[@]}
  for svc in "${services[@]}"; do
    i=$((i + 1))
    local dir="${PROJECT_ROOT}/services/${svc}"
    if [[ ! -d "${dir}" ]]; then
      die "Service directory not found: ${dir}"
    fi
    if [[ -f "${dir}/.env" ]]; then
      set -a
      # shellcheck disable=SC1090
      source "${dir}/.env"
      set +a
    elif [[ -f "${dir}/.env.example" ]]; then
      log "WARN: ${svc}/.env not found, using .env.example defaults"
      set -a
      # shellcheck disable=SC1090
      source "${dir}/.env.example"
      set +a
    fi
    log "[${i}/${total}] Migrating ${svc}..."
    (cd "${dir}" && go run cmd/migrate/main.go up)
    log "[${i}/${total}] ✓ ${svc}"
  done

  log "All local migrations completed successfully."
}

check_kubectl() {
  command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
  kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster. Set KUBECONFIG."
}

wait_for_postgres() {
  log "Waiting for PostgreSQL..."
  kubectl -n infra rollout status statefulset/postgres --timeout=300s
  log "✓ PostgreSQL ready"
}

run_migration_job() {
  local service="$1"
  local db_name="$2"
  local step="$3"
  local total="$4"
  local job_name="migrate-${service}-$(date +%s)"

  log "[${step}/${total}] Migration ${service} → database ${db_name}"
  log "[${step}/${total}] Creating job/${job_name}..."

  kubectl -n "${NAMESPACE}" delete job -l "app.kubernetes.io/migrate=${service}" --ignore-not-found=true 2>/dev/null || true

  cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${job_name}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/migrate: ${service}
spec:
  backoffLimit: 2
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/migrate: ${service}
    spec:
      restartPolicy: Never
      containers:
      - name: migrate
        image: ghcr.io/caovanson/${service}:${MIGRATION_IMAGE_TAG}
        imagePullPolicy: IfNotPresent
        workingDir: /app
        command: ["/bin/sh", "-c"]
        args:
          - |
            set -e
            export MIGRATIONS_PATH=file:///app/migrations
            echo "[migrate] starting ${service} on ${db_name}..."
            /app/migrate up
            echo "[migrate] done ${service}"
        envFrom:
        - configMapRef:
            name: ${service}-config
        - secretRef:
            name: ${service}-secret
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
EOF

  JOB_LOG_PREFIX="[migrate-all]" wait_for_k8s_job "${job_name}" "${NAMESPACE}" "${JOB_TIMEOUT}" || die "Migration failed: ${service}"

  log "[${step}/${total}] --- logs ${job_name} ---"
  kubectl -n "${NAMESPACE}" logs "job/${job_name}" 2>/dev/null | tail -20 || true
  log "[${step}/${total}] ✓ ${service}"
}

create_databases() {
  log "Creating application databases if not exist..."
  local postgres_pod
  postgres_pod=$(kubectl -n infra get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}')

  kubectl -n infra exec "${postgres_pod}" -- bash -c '
    for db in auth_db product_db order_db; do
      psql -U shopcaovanson -d shopcaovanson -tc "SELECT 1 FROM pg_database WHERE datname = '\''$db'\''" | grep -q 1 \
        && echo "  OK  database $db exists" \
        || (psql -U shopcaovanson -d shopcaovanson -c "CREATE DATABASE $db" && echo "  +   created $db")
    done
  '
  log "✓ Databases ready"
}

run_k8s_migrations() {
  check_kubectl
  wait_for_postgres
  create_databases

  local services=(auth-service product-service order-service)
  local dbs=(auth_db product_db order_db)
  local total=${#services[@]}
  local i=0
  for svc in "${services[@]}"; do
    i=$((i + 1))
    run_migration_job "${svc}" "${dbs[$((i - 1))]}" "${i}" "${total}"
  done

  log "All migrations completed successfully."
}

main() {
  if [[ "${LOCAL:-false}" == "true" ]]; then
    run_local_migrations
  else
    run_k8s_migrations
  fi
}

main "$@"
