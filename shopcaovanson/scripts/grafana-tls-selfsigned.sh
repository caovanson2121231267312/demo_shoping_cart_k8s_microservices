#!/usr/bin/env bash
# TLS tự ký cho Grafana — dùng TẠM khi DNS monitor.* chưa có (Let's Encrypt không cấp được)
# Truy cập: thêm vào file hosts máy bạn:  110.172.29.72 monitor.shopcaovanson.xyz
# Trình duyệt sẽ cảnh báo cert — chọn Advanced → Proceed
#
# Khi DNS public đã OK, chạy: bash scripts/fix-monitoring-access.sh --renew-cert
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MONITORING_DIR="${PROJECT_ROOT}/k8s/base/monitoring"
GRAFANA_HOST="${GRAFANA_HOST:-monitor.shopcaovanson.xyz}"
TMPDIR="${TMPDIR:-/tmp}"
KEY="${TMPDIR}/grafana-selfsigned.key"
CRT="${TMPDIR}/grafana-selfsigned.crt"

log() { echo "[grafana-selfsigned] $*"; }

command -v kubectl >/dev/null 2>&1 || { echo "kubectl not found"; exit 1; }
kubectl get ns monitoring >/dev/null 2>&1 || { echo "Chưa có namespace monitoring — chạy install-monitoring.sh trước"; exit 1; }

log "Dừng cert-manager cho grafana (xóa Certificate/Challenge cũ)..."
kubectl delete certificate grafana-tls -n monitoring --ignore-not-found
kubectl delete challenge -n monitoring --all --ignore-not-found
kubectl delete order -n monitoring --all --ignore-not-found
kubectl delete certificaterequest -n monitoring --all --ignore-not-found

log "Tạo chứng chỉ tự ký cho ${GRAFANA_HOST}..."
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "${KEY}" -out "${CRT}" \
  -subj "/CN=${GRAFANA_HOST}" \
  -addext "subjectAltName=DNS:${GRAFANA_HOST}" 2>/dev/null || \
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "${KEY}" -out "${CRT}" \
  -subj "/CN=${GRAFANA_HOST}"

kubectl create secret tls grafana-tls -n monitoring \
  --cert="${CRT}" --key="${KEY}" \
  --dry-run=client -o yaml | kubectl apply -f -

log "Apply Ingress (Let's Encrypt qua cert-manager)..."
kubectl apply -f "${MONITORING_DIR}/grafana-ingress.yaml"

kubectl apply -f "${MONITORING_DIR}/allow-grafana-ingress.yaml" 2>/dev/null || true

rm -f "${KEY}" "${CRT}"

echo ""
log "══════════════════════════════════════════════════════════"
log " Grafana TLS tự ký — TRUY CẬP TẠM"
log "══════════════════════════════════════════════════════════"
echo ""
echo "  1. Trên máy Windows (Admin), sửa C:\\Windows\\System32\\drivers\\etc\\hosts:"
echo "     110.172.29.72  monitor.shopcaovanson.xyz"
echo ""
echo "  2. Mở: https://monitor.shopcaovanson.xyz"
echo "     (Chấp nhận cảnh báo certificate không tin cậy)"
echo ""
echo "  3. Đăng nhập: bash scripts/monitoring-access.sh"
echo ""
echo "  Khi DNS public đã OK, dùng Let's Encrypt (khuyến nghị — không cảnh báo trình duyệt):"
echo "    bash scripts/fix-monitoring-access.sh --renew-cert"
echo ""
