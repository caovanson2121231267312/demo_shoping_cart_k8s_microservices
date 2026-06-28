#!/usr/bin/env bash
# Xóa dữ liệu MongoDB và khởi động lại — dùng khi CrashLoopBackOff do password cũ trên PVC
#
# MONGO_INITDB_* chỉ chạy lần đầu khi /data/db trống. Nếu đổi mongodb-secret sau khi
# đã deploy, password trong DB ≠ secret → app không kết nối được.
#
# Usage:
#   bash scripts/reset-mongodb.sh          # hỏi xác nhận
#   bash scripts/reset-mongodb.sh --yes      # không hỏi (CI / đã chắc chắn)
set -euo pipefail

NS="${NAMESPACE:-infra}"
STS="mongodb"
PVC="mongodb-data-mongodb-0"
YES=false

log() { echo "[reset-mongodb] $*"; }
die() { echo "[reset-mongodb] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes|-y) YES=true; shift ;;
    -h|--help)
      echo "Usage: bash scripts/reset-mongodb.sh [--yes]"
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

if [[ "${YES}" != "true" ]]; then
  echo ""
  echo "⚠️  Script này XÓA toàn bộ dữ liệu MongoDB (product_db, chat_db)."
  echo "   Chỉ chạy khi: CrashLoopBackOff, đổi password secret, hoặc deploy lần đầu bị lỗi."
  echo ""
  read -r -p "Tiếp tục? [y/N] " ans
  [[ "${ans}" =~ ^[Yy]$ ]] || { log "Cancelled."; exit 0; }
fi

log "Scaling ${STS} to 0..."
kubectl scale statefulset "${STS}" -n "${NS}" --replicas=0 2>/dev/null || true
kubectl wait --for=delete "pod/${STS}-0" -n "${NS}" --timeout=120s 2>/dev/null || true

if kubectl get pvc "${PVC}" -n "${NS}" >/dev/null 2>&1; then
  log "Deleting PVC ${PVC}..."
  kubectl delete pvc "${PVC}" -n "${NS}" --wait=true
else
  log "PVC ${PVC} not found (OK)"
fi

log "Scaling ${STS} back to 1..."
kubectl scale statefulset "${STS}" -n "${NS}" --replicas=1

log "Waiting for ${STS}-0 (up to 600s)..."
kubectl rollout status statefulset/"${STS}" -n "${NS}" --timeout=600s

log "Verifying MongoDB auth..."
kubectl exec "${STS}-0" -n "${NS}" -- mongosh \
  -u "${MONGO_USER:-shopcaovanson}" \
  -p "$(kubectl get secret mongodb-secret -n "${NS}" -o jsonpath='{.data.MONGO_PASSWORD}' | base64 -d)" \
  --authenticationDatabase admin \
  --eval "db.adminCommand('ping')" --quiet

log "✓ MongoDB ready"
