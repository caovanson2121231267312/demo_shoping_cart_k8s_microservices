#!/usr/bin/env bash
# Reset Elasticsearch data — khi CrashLoop/OOM hoặc đổi ELASTIC_PASSWORD sau lần init đầu
set -euo pipefail

NS="${NAMESPACE:-infra}"
STS="elasticsearch"
PVC="elasticsearch-data-elasticsearch-0"
YES=false

log() { echo "[reset-elasticsearch] $*"; }
die() { echo "[reset-elasticsearch] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes|-y) YES=true; shift ;;
    -h|--help)
      echo "Usage: bash scripts/reset-elasticsearch.sh [--yes]"
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

if [[ "${YES}" != "true" ]]; then
  echo ""
  echo "⚠️  XÓA toàn bộ index Elasticsearch (search sản phẩm)."
  read -r -p "Tiếp tục? [y/N] " ans
  [[ "${ans}" =~ ^[Yy]$ ]] || { log "Cancelled."; exit 0; }
fi

log "Scaling ${STS} to 0..."
kubectl scale statefulset "${STS}" -n "${NS}" --replicas=0 2>/dev/null || true
kubectl wait --for=delete "pod/${STS}-0" -n "${NS}" --timeout=180s 2>/dev/null || true

if kubectl get pvc "${PVC}" -n "${NS}" >/dev/null 2>&1; then
  log "Deleting PVC ${PVC}..."
  kubectl delete pvc "${PVC}" -n "${NS}" --wait=true
fi

log "Scaling ${STS} back to 1..."
kubectl scale statefulset "${STS}" -n "${NS}" --replicas=1

log "Waiting for ${STS}-0 (up to 900s)..."
kubectl rollout status statefulset/"${STS}" -n "${NS}" --timeout=900s

log "Verifying cluster health..."
kubectl exec "${STS}-0" -n "${NS}" -- sh -c \
  'curl -sf -u "elastic:${ELASTIC_PASSWORD}" "http://127.0.0.1:9200/_cluster/health?wait_for_status=yellow&timeout=30s"'

log "✓ Elasticsearch ready"
