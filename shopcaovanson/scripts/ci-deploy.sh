#!/usr/bin/env bash
# Deploy image mới lên K8s — gọi từ GitHub Actions hoặc thủ công
# Usage: SHA=abc123 SERVICES="auth-service product-service" bash scripts/ci-deploy.sh
set -euo pipefail

REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_OWNER="${IMAGE_OWNER:-caovanson}"
NAMESPACE="${NAMESPACE:-shop}"
SHA="${SHA:?Set SHA=git-commit-sha}"
SERVICES="${SERVICES:-}"
CONTINUE_ON_ERROR="${CONTINUE_ON_ERROR:-true}"
ROLLBACK_ON_FAILURE="${ROLLBACK_ON_FAILURE:-true}"

log() { echo "[ci-deploy] $*"; }
die() { echo "[ci-deploy] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
kubectl get deployments -n "${NAMESPACE}" >/dev/null 2>&1 || die "Cannot connect to cluster (check KUBECONFIG_DATA and RBAC)"

image_for_service() {
  local svc="$1"
  case "${svc}" in
    frontend) echo "frontend-web" ;;
    *) echo "${svc}" ;;
  esac
}

dump_pod_logs() {
  local svc="$1"
  local pod
  pod=$(kubectl get pods -n "${NAMESPACE}" -l "app=${svc}" \
    --sort-by=.metadata.creationTimestamp \
    -o jsonpath='{.items[-1].metadata.name}' 2>/dev/null || true)
  if [[ -z "${pod}" ]]; then
    log "Không lấy được tên pod cho ${svc}"
    return
  fi
  log "Logs pod mới nhất: ${pod}"
  kubectl logs "${pod}" -n "${NAMESPACE}" --tail=60 2>/dev/null || true
  kubectl logs "${pod}" -n "${NAMESPACE}" --previous --tail=60 2>/dev/null || true
  log "Events:"
  kubectl describe pod "${pod}" -n "${NAMESPACE}" 2>/dev/null | tail -20 || true
}

deploy_service() {
  local svc="$1"
  local img ref reason
  img=$(image_for_service "${svc}")
  ref="${REGISTRY}/${IMAGE_OWNER}/${img}:${SHA}"

  log "Rolling out ${svc} → ${ref}"

  if kubectl get secret ghcr-secret -n "${NAMESPACE}" >/dev/null 2>&1; then
    kubectl patch deployment "${svc}" -n "${NAMESPACE}" -p \
      '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"ghcr-secret"}]}}}}' \
      >/dev/null 2>&1 || true
  else
    log "WARN: secret/ghcr-secret không có — chạy: bash scripts/setup-ghcr-pull-secret.sh trên VPS"
  fi

  kubectl set image "deployment/${svc}" "${svc}=${ref}" -n "${NAMESPACE}"
  kubectl patch deployment "${svc}" -n "${NAMESPACE}" --type=json \
    -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Always"}]' \
    2>/dev/null || true

  if kubectl rollout status "deployment/${svc}" -n "${NAMESPACE}" --timeout=300s; then
    log "✓ ${svc} deployed"
    return 0
  fi

  log "Rollout failed — pod status:"
  kubectl get pods -n "${NAMESPACE}" -l "app=${svc}" -o wide 2>/dev/null || true
  dump_pod_logs "${svc}"

  reason=$(kubectl get pods -n "${NAMESPACE}" -l "app=${svc}" \
    --sort-by=.metadata.creationTimestamp \
    -o jsonpath='{.items[-1].status.containerStatuses[0].state.waiting.reason}' 2>/dev/null || true)

  if [[ "${ROLLBACK_ON_FAILURE}" == "true" ]]; then
    log "Rollback ${svc} về revision trước..."
    kubectl rollout undo "deployment/${svc}" -n "${NAMESPACE}" 2>/dev/null || true
    kubectl rollout status "deployment/${svc}" -n "${NAMESPACE}" --timeout=120s 2>/dev/null || true
  fi

  case "${reason}" in
    ImagePullBackOff|ErrImagePull)
      log "ERROR: ${svc} — ImagePullBackOff. Chạy trên VPS: bash scripts/setup-ghcr-pull-secret.sh"
      ;;
    CrashLoopBackOff)
      log "ERROR: ${svc} — CrashLoopBackOff. SSH VPS: kubectl logs -n shop -l app=${svc} --tail=80"
      ;;
    *)
      log "ERROR: ${svc} — rollout timeout (${reason:-unknown})"
      ;;
  esac
  return 1
}

apply_manifests() {
  local overlay="${OVERLAY:-dev}"
  local dir
  dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/k8s/overlays/${overlay}"
  [[ -d "${dir}" ]] || die "Overlay not found: ${dir}"
  log "Applying manifests: k8s/overlays/${overlay}"
  kubectl apply -k "${dir}"
}

main() {
  local svc failed=()

  if [[ "${APPLY_MANIFESTS:-false}" == "true" ]]; then
    apply_manifests
  fi

  if [[ -z "${SERVICES}" ]]; then
    log "No SERVICES specified — skip image rollout"
    exit 0
  fi

  for svc in ${SERVICES}; do
    if deploy_service "${svc}"; then
      continue
    fi
    failed+=("${svc}")
    if [[ "${CONTINUE_ON_ERROR}" != "true" ]]; then
      die "Dừng deploy vì ${svc} lỗi"
    fi
    log "Tiếp tục deploy service khác (CONTINUE_ON_ERROR=true)..."
  done

  kubectl get pods -n "${NAMESPACE}" 2>/dev/null || true

  if [[ ${#failed[@]} -gt 0 ]]; then
    die "Deploy xong nhưng thất bại: ${failed[*]}. Đã rollback các service lỗi (nếu ROLLBACK_ON_FAILURE=true)."
  fi

  log "Deploy complete. SHA=${SHA}"
}

main "$@"
