#!/usr/bin/env bash
# Deploy image mới lên K8s — gọi từ GitHub Actions hoặc thủ công
# Usage: SHA=abc123 SERVICES="auth-service product-service" bash scripts/ci-deploy.sh
set -euo pipefail

REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_OWNER="${IMAGE_OWNER:-caovanson}"
NAMESPACE="${NAMESPACE:-shop}"
SHA="${SHA:?Set SHA=git-commit-sha}"
SERVICES="${SERVICES:-}"

log() { echo "[ci-deploy] $*"; }
die() { echo "[ci-deploy] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
# cicd-deployer chỉ có quyền namespace shop — cluster-info cần list services ở kube-system
kubectl get deployments -n "${NAMESPACE}" >/dev/null 2>&1 || die "Cannot connect to cluster (check KUBECONFIG_DATA and RBAC)"

image_for_service() {
  local svc="$1"
  case "${svc}" in
    frontend) echo "frontend-web" ;;
    *) echo "${svc}" ;;
  esac
}

deploy_service() {
  local svc="$1"
  local img
  img=$(image_for_service "${svc}")
  local ref="${REGISTRY}/${IMAGE_OWNER}/${img}:${SHA}"

  log "Rolling out ${svc} → ${ref}"

  if kubectl get secret ghcr-secret -n "${NAMESPACE}" >/dev/null 2>&1; then
    kubectl patch deployment "${svc}" -n "${NAMESPACE}" -p \
      '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"ghcr-secret"}]}}}}' \
      >/dev/null 2>&1 || true
  else
    log "WARN: secret/ghcr-secret không có — cần public GHCR hoặc chạy scripts/setup-ghcr-pull-secret.sh trên VPS"
  fi

  kubectl set image "deployment/${svc}" "${svc}=${ref}" -n "${NAMESPACE}"
  kubectl patch deployment "${svc}" -n "${NAMESPACE}" --type=json \
    -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Always"}]' \
    2>/dev/null || true

  if ! kubectl rollout status "deployment/${svc}" -n "${NAMESPACE}" --timeout=300s; then
    log "Rollout failed — pod status:"
    kubectl get pods -n "${NAMESPACE}" -l "app=${svc}" -o wide 2>/dev/null || true
    local reason
    reason=$(kubectl get pods -n "${NAMESPACE}" -l "app=${svc}" \
      -o jsonpath='{range .items[*]}{.status.containerStatuses[0].state.waiting.reason}{"\n"}{end}' 2>/dev/null | head -1)
    if [[ "${reason}" == "ImagePullBackOff" || "${reason}" == "ErrImagePull" ]]; then
      die "${svc}: không pull được image từ GHCR. Trên VPS chạy: bash scripts/setup-ghcr-pull-secret.sh HOẶC public packages trên GitHub."
    fi
    die "${svc}: rollout timeout. SSH VPS: kubectl describe pod -n shop -l app=${svc}"
  fi
  log "✓ ${svc} deployed"
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
  if [[ "${APPLY_MANIFESTS:-false}" == "true" ]]; then
    apply_manifests
  fi

  if [[ -z "${SERVICES}" ]]; then
    log "No SERVICES specified — skip image rollout"
    exit 0
  fi

  for svc in ${SERVICES}; do
    deploy_service "${svc}"
  done

  log "Deploy complete. SHA=${SHA}"
  kubectl get pods -n "${NAMESPACE}" -l "app in ($(echo ${SERVICES} | tr ' ' ','))" 2>/dev/null || kubectl get pods -n "${NAMESPACE}"
}

main "$@"
