#!/usr/bin/env bash
# Xóa ReplicaSet cũ đang kẹt rollout (pod mới Ready nhưng RS cũ không terminate)
#
# Usage:
#   bash scripts/cleanup-stuck-rollouts.sh
#   bash scripts/cleanup-stuck-rollouts.sh notification-service search-service
set -euo pipefail

NS="${NAMESPACE:-shop}"

log() { echo "[cleanup-rollouts] $*"; }

cleanup_deployment() {
  local dep="$1"
  local target_rev
  target_rev=$(kubectl get deployment "${dep}" -n "${NS}" \
    -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' 2>/dev/null || true)
  if [[ -z "${target_rev}" ]]; then
    log "Skip ${dep} (deployment not found)"
    return 0
  fi

  local rs rev
  for rs in $(kubectl get rs -n "${NS}" -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null); do
    [[ "${rs}" == "${dep}-"* ]] || continue
    rev=$(kubectl get rs "${rs}" -n "${NS}" \
      -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}' 2>/dev/null || true)
    if [[ "${rev}" != "${target_rev}" ]]; then
      log "Deleting stale rs/${rs} (revision ${rev} ≠ current ${target_rev})"
      kubectl delete rs "${rs}" -n "${NS}" --cascade=foreground --grace-period=0 --force 2>/dev/null || true
    fi
  done

  if kubectl rollout status "deployment/${dep}" -n "${NS}" --timeout=90s 2>/dev/null; then
    log "✓ deployment/${dep} ready"
  else
    log "WARN: deployment/${dep} still not ready — kiểm tra: kubectl get pods -n ${NS} -l app=${dep}"
  fi
}

cleanup_failed_jobs() {
  log "Cleaning failed migration jobs..."
  kubectl get jobs -n "${NS}" -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null \
    | grep -E '^migrate-' || true \
    | while read -r job; do
        [[ -z "${job}" ]] && continue
        local succeeded
        succeeded=$(kubectl get job "${job}" -n "${NS}" -o jsonpath='{.status.succeeded}' 2>/dev/null || echo "")
        if [[ "${succeeded}" != "1" ]]; then
          log "Deleting job/${job}"
          kubectl delete job "${job}" -n "${NS}" --ignore-not-found=true
        fi
      done
}

main() {
  command -v kubectl >/dev/null 2>&1 || { echo "kubectl not found" >&2; exit 1; }

  local deps=("$@")
  if [[ ${#deps[@]} -eq 0 ]]; then
    deps=(
      api-gateway auth-service product-service order-service
      chat-service notification-service search-service frontend
    )
  fi

  for dep in "${deps[@]}"; do
    cleanup_deployment "${dep}"
  done
  cleanup_failed_jobs

  echo ""
  kubectl get pods -n "${NS}" -o wide
}

main "$@"
