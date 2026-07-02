#!/usr/bin/env bash
# Chạy migration SQL trước deploy — gọi từ GitHub Actions (cicd-deployer)
# Usage: SHA=abc123 IMAGE_OWNER=user bash scripts/ci-migrate.sh
set -euo pipefail

REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_OWNER="${IMAGE_OWNER:-caovanson}"
NAMESPACE="${NAMESPACE:-shop}"
SHA="${SHA:?Set SHA=git-commit-sha}"
SERVICES="${SERVICES:-}"
JOB_TIMEOUT="${JOB_TIMEOUT:-600}"

SQL_SERVICES=(auth-service product-service order-service)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/k8s-job-wait.sh
source "${SCRIPT_DIR}/lib/k8s-job-wait.sh"

log() { echo "[ci-migrate] $*"; }
die() { echo "[ci-migrate] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
kubectl get deployments -n "${NAMESPACE}" >/dev/null 2>&1 || die "Cannot connect to cluster"

image_tag_for() {
  local service="$1"
  if [[ " ${SERVICES} " == *" ${service} "* ]]; then
    echo "${SHA}"
  else
    echo "latest"
  fi
}

run_migration_job() {
  local service="$1"
  local tag
  tag="$(image_tag_for "${service}")"
  local job_name="ci-migrate-${service}-$(date +%s | tail -c 8)"
  local image="${REGISTRY}/${IMAGE_OWNER}/${service}:${tag}"

  log "Migration ${service} → ${image}"

  kubectl -n "${NAMESPACE}" delete job -l "app.kubernetes.io/ci-migrate=${service}" --ignore-not-found=true 2>/dev/null || true

  cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${job_name}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/ci-migrate: ${service}
spec:
  backoffLimit: 2
  ttlSecondsAfterFinished: 900
  template:
    metadata:
      labels:
        app.kubernetes.io/ci-migrate: ${service}
    spec:
      restartPolicy: Never
      imagePullSecrets:
      - name: ghcr-secret
      containers:
      - name: migrate
        image: ${image}
        imagePullPolicy: Always
        workingDir: /app
        command: ["/app/migrate", "up"]
        env:
        - name: MIGRATIONS_PATH
          value: file:///app/migrations
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

  JOB_LOG_PREFIX="[ci-migrate]" wait_for_k8s_job "${job_name}" "${NAMESPACE}" "${JOB_TIMEOUT}" \
    || die "Migration failed: ${service}"

  kubectl -n "${NAMESPACE}" logs "job/${job_name}" 2>/dev/null | tail -25 || true
  log "✓ ${service} migrated"
}

main() {
  log "Running SQL migrations before deploy (SHA=${SHA}, services built: ${SERVICES:-none})..."

  for svc in "${SQL_SERVICES[@]}"; do
    run_migration_job "${svc}"
  done

  log "All migrations complete."
}

main "$@"
