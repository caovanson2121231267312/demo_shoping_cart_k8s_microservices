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
kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster"

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
  kubectl set image "deployment/${svc}" "${svc}=${ref}" -n "${NAMESPACE}"
  kubectl rollout status "deployment/${svc}" -n "${NAMESPACE}" --timeout=300s
  log "✓ ${svc} deployed"
}

apply_manifests() {
  local overlay="${OVERLAY:-prod}"
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
