#!/usr/bin/env bash
# Build Docker images trên VPS và import vào containerd (K8s)
# Dùng khi chưa push image lên GHCR hoặc ImagePullBackOff
#
# Usage:
#   bash scripts/build-images.sh              # build tất cả
#   bash scripts/build-images.sh api-gateway  # build 1 service
#   bash scripts/build-images.sh --list
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
REGISTRY="${REGISTRY:-ghcr.io/caovanson}"
TAG="${TAG:-latest}"

log() { echo "[build-images] $*"; }
die() { echo "[build-images] ERROR: $*" >&2; exit 1; }

import_image() {
  local ref="$1"
  if command -v ctr >/dev/null 2>&1; then
    docker save "${ref}" | ctr -n k8s.io images import -
  elif command -v nerdctl >/dev/null 2>&1; then
    nerdctl -n k8s.io load -i <(docker save "${ref}")
  else
    die "Need ctr or nerdctl to import images into containerd"
  fi
  log "✓ imported ${ref}"
}

build_go() {
  local svc="$1"
  local ctx="${PROJECT_ROOT}/services/${svc}"
  [[ -d "${ctx}" ]] || die "Missing ${ctx}"
  local ref="${REGISTRY}/${svc}:${TAG}"
  log "Building ${ref}..."
  docker build -t "${ref}" "${ctx}"
  import_image "${ref}"
}

build_python() {
  local svc="$1"
  local ctx="${PROJECT_ROOT}/services/${svc}"
  [[ -d "${ctx}" ]] || die "Missing ${ctx}"
  local ref="${REGISTRY}/${svc}:${TAG}"
  log "Building ${ref}..."
  docker build -t "${ref}" "${ctx}"
  import_image "${ref}"
}

build_frontend() {
  local ref="${REGISTRY}/frontend-web:${TAG}"
  log "Building ${ref}..."
  docker build -t "${ref}" "${PROJECT_ROOT}/frontend/web"
  import_image "${ref}"
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

list_services() {
  printf '%s\n' "${ALL_SERVICES[@]}"
}

main() {
  command -v docker >/dev/null 2>&1 || die "docker not found — cài: apt install docker.io"

  if [[ "${1:-}" == "--list" ]]; then
    list_services
    exit 0
  fi

  local targets=()
  if [[ $# -gt 0 ]]; then
    targets=("$@")
  else
    targets=("${ALL_SERVICES[@]}")
  fi

  for svc in "${targets[@]}"; do
    case "${svc}" in
      api-gateway|auth-service|product-service|order-service)
        build_go "${svc}"
        ;;
      chat-service|notification-service|search-service)
        build_python "${svc}"
        ;;
      frontend|frontend-web)
        build_frontend
        ;;
      *)
        die "Unknown service: ${svc} (use --list)"
        ;;
    esac
  done

  echo ""
  log "Done. Restart deployments:"
  echo "  kubectl rollout restart deployment -n shop api-gateway auth-service product-service order-service chat-service notification-service search-service frontend"
  echo "  OVERLAY=dev bash scripts/deploy-all.sh   # hoặc chờ rollout"
}

main "$@"
