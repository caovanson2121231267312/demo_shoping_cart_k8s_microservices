#!/usr/bin/env bash
# shopcaovanson — Run database migrations for auth, product, and order services
# Usage:
#   LOCAL=true bash scripts/migrate-all.sh     # local dev (go run)
#   bash scripts/migrate-all.sh                # K8s cluster (kubectl jobs)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NAMESPACE="${NAMESPACE:-shop}"
MIGRATION_IMAGE_TAG="${MIGRATION_IMAGE_TAG:-latest}"

log() { echo "[migrate-all] $*"; }
die() { echo "[migrate-all] ERROR: $*" >&2; exit 1; }

run_local_migrations() {
  log "Running local migrations (go run cmd/migrate/main.go up)..."

  local services=(auth-service product-service order-service)
  for svc in "${services[@]}"; do
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
    log "Migrating ${svc}..."
    (cd "${dir}" && go run cmd/migrate/main.go up)
    log "Migration for ${svc} completed."
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
}

run_migration_job() {
  local service="$1"
  local db_name="$2"
  local job_name="migrate-${service}-$(date +%s)"
  local migrate_cmd="/app/migrate up"
  if [[ "${service}" != "auth-service" ]]; then
    migrate_cmd="/app/migrate -direction up"
  fi

  log "Running migration for ${service} (database: ${db_name})..."

  kubectl -n "${NAMESPACE}" delete job "${job_name}" --ignore-not-found=true

  cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${job_name}
  namespace: ${NAMESPACE}
spec:
  backoffLimit: 3
  ttlSecondsAfterFinished: 600
  template:
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
            export MIGRATIONS_PATH=file:///app/migrations
            ${migrate_cmd}
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

  kubectl -n "${NAMESPACE}" wait --for=condition=complete "job/${job_name}" --timeout=300s
  log "Migration for ${service} completed."
  kubectl -n "${NAMESPACE}" logs "job/${job_name}"
}

create_databases() {
  log "Creating application databases if not exist..."
  local postgres_pod
  postgres_pod=$(kubectl -n infra get pod -l app=postgres -o jsonpath='{.items[0].metadata.name}')

  kubectl -n infra exec "${postgres_pod}" -- bash -c '
    for db in auth_db product_db order_db; do
      psql -U shopcaovanson -d shopcaovanson -tc "SELECT 1 FROM pg_database WHERE datname = '\''$db'\''" | grep -q 1 \
        || psql -U shopcaovanson -d shopcaovanson -c "CREATE DATABASE $db"
    done
  '
}

run_k8s_migrations() {
  check_kubectl
  wait_for_postgres
  create_databases
  run_migration_job "auth-service" "auth_db"
  run_migration_job "product-service" "product_db"
  run_migration_job "order-service" "order_db"
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
