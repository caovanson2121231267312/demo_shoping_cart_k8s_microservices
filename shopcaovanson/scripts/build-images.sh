#!/usr/bin/env bash
# Build container images trên VPS và đưa vào containerd namespace k8s.io
#
# VPS K8s dùng containerd (không có Docker daemon):
#   sudo bash scripts/install-build-tools.sh   # một lần
#   bash scripts/build-images.sh
#
# Usage:
#   bash scripts/build-images.sh              # tất cả services (~20–40 phút)
#   bash scripts/build-images.sh api-gateway  # một service
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
REGISTRY="${REGISTRY:-ghcr.io/caovanson}"
TAG="${TAG:-latest}"
CONTAINERD_NS="${CONTAINERD_NS:-k8s.io}"

log() { echo "[build-images] $*"; }
die() { echo "[build-images] ERROR: $*" >&2; exit 1; }

detect_builder() {
  if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    echo "docker"
    return 0
  fi
  if command -v nerdctl >/dev/null 2>&1; then
    echo "nerdctl"
    return 0
  fi
  return 1
}

require_builder() {
  if ! detect_builder >/dev/null 2>&1; then
    die "Chưa có công cụ build. Chạy: sudo bash scripts/install-build-tools.sh"
  fi
}

ensure_buildkit() {
  export BUILDKIT_HOST="${BUILDKIT_HOST:-unix:///run/buildkit/buildkitd.sock}"
  if [[ -S "/run/buildkit/buildkitd.sock" ]]; then
    return 0
  fi
  if systemctl is-active buildkit >/dev/null 2>&1; then
    sleep 2
    [[ -S "/run/buildkit/buildkitd.sock" ]] && return 0
  fi
  log "WARN: buildkit socket missing — chạy: sudo systemctl start buildkit"
}

build_image() {
  local ctx="$1"
  local ref="$2"
  local builder
  builder=$(detect_builder)

  [[ -d "${ctx}" ]] || die "Missing context: ${ctx}"

  log "Building ${ref} (${builder})..."
  ensure_buildkit
  export BUILDKIT_HOST="${BUILDKIT_HOST:-unix:///run/buildkit/buildkitd.sock}"

  case "${builder}" in
    docker)
      docker build -t "${ref}" "${ctx}"
      if command -v ctr >/dev/null 2>&1; then
        docker save "${ref}" | ctr -n "${CONTAINERD_NS}" images import -
      else
        die "ctr not found — cannot import docker image into containerd"
      fi
      ;;
    nerdctl)
      nerdctl --namespace "${CONTAINERD_NS}" build -t "${ref}" "${ctx}"
      ;;
  esac
  log "✓ ${ref}"
}

ALL_SERVICES=(
  api-gateway
  auth-service
  product-service
  order-service
  chat-service
  notification-service
  search-service
  frontend-web
)

build_service() {
  local svc="$1"
  case "${svc}" in
    api-gateway|auth-service|product-service|order-service)
      build_image "${PROJECT_ROOT}/services/${svc}" "${REGISTRY}/${svc}:${TAG}"
      ;;
    chat-service|notification-service|search-service)
      build_image "${PROJECT_ROOT}/services/${svc}" "${REGISTRY}/${svc}:${TAG}"
      ;;
    frontend|frontend-web)
      build_image "${PROJECT_ROOT}/frontend/web" "${REGISTRY}/frontend-web:${TAG}"
      ;;
    *)
      die "Unknown service: ${svc} (available: ${ALL_SERVICES[*]})"
      ;;
  esac
}

restart_deployments() {
  log "Setting imagePullPolicy=IfNotPresent (dùng image local, không pull GHCR)..."
  for dep in api-gateway auth-service product-service order-service \
             chat-service notification-service search-service frontend; do
    kubectl patch deployment "${dep}" -n shop --type=json \
      -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"IfNotPresent"}]' \
      2>/dev/null || true
  done

  log "Restarting shop deployments..."
  kubectl rollout restart deployment -n shop \
    api-gateway auth-service product-service order-service \
    chat-service notification-service search-service frontend 2>/dev/null || true
}

main() {
  if [[ "${1:-}" == "--list" ]]; then
    printf '%s\n' "${ALL_SERVICES[@]}"
    exit 0
  fi

  command -v kubectl >/dev/null 2>&1 || log "WARN: kubectl not in PATH"

  require_builder
  ensure_buildkit

  local targets=()
  if [[ $# -gt 0 ]]; then
    targets=("$@")
  else
    targets=("${ALL_SERVICES[@]}")
  fi

  log "Builder: $(detect_builder) | namespace: ${CONTAINERD_NS}"
  log "Building ${#targets[@]} image(s)..."

  for svc in "${targets[@]}"; do
    build_service "${svc}"
  done

  echo ""
  log "Listing images on node:"
  if command -v nerdctl >/dev/null 2>&1; then
    nerdctl --namespace "${CONTAINERD_NS}" images | grep -E 'ghcr.io/caovanson|REPOSITORY' || true
  elif command -v crictl >/dev/null 2>&1; then
    crictl images | grep ghcr.io/caovanson || true
  fi

  restart_deployments
  echo ""
  log "Done. Kiểm tra:"
  echo "  kubectl get pods -n shop"
  echo "  bash scripts/diagnose-apps.sh"
}

main "$@"
