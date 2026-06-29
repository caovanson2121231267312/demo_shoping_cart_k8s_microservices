#!/usr/bin/env bash
# Xóa dữ liệu PostgreSQL và khởi động lại — dùng khi password authentication failed
#
# POSTGRES_PASSWORD chỉ áp dụng lần đầu khi PVC trống. Đổi postgres-secret sau deploy
# → shop secrets ≠ password thực tế trên PVC.
#
# Usage:
#   bash scripts/reset-postgres.sh
#   bash scripts/reset-postgres.sh --yes
set -euo pipefail

NS="${NAMESPACE:-infra}"
STS="postgres"
PVC="postgres-data-postgres-0"
YES=false

log() { echo "[reset-postgres] $*"; }
die() { echo "[reset-postgres] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes|-y) YES=true; shift ;;
    -h|--help)
      echo "Usage: bash scripts/reset-postgres.sh [--yes]"
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

if [[ "${YES}" != "true" ]]; then
  echo ""
  echo "⚠️  Script này XÓA toàn bộ dữ liệu PostgreSQL (auth_db, product_db, order_db)."
  echo "   Chạy khi log: password authentication failed for user shopcaovanson"
  echo ""
  read -r -p "Tiếp tục? [y/N] " ans
  [[ "${ans}" =~ ^[Yy]$ ]] || { log "Cancelled."; exit 0; }
fi

PG_PASS=$(kubectl get secret postgres-secret -n "${NS}" -o jsonpath='{.data.POSTGRES_PASSWORD}' 2>/dev/null | base64 -d) || true
[[ -n "${PG_PASS}" ]] || die "Missing infra/postgres-secret — chạy: bash scripts/create-secrets.sh"

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

log "Verifying PostgreSQL auth..."
kubectl exec "${STS}-0" -n "${NS}" -- env PGPASSWORD="${PG_PASS}" \
  psql -h 127.0.0.1 -U shopcaovanson -d shopcaovanson -tAc 'SELECT 1' | grep -q 1

log "✓ PostgreSQL ready — chạy: bash scripts/migrate-all.sh"
