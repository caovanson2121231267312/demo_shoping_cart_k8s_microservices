#!/usr/bin/env bash
# In URL và lệnh truy cập monitoring
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CREDS_FILE="${PROJECT_ROOT}/.monitoring-credentials"
GRAFANA_HOST="${GRAFANA_HOST:-monitor.shopcaovanson.xyz}"

section() { echo ""; echo "========== $* =========="; }

section "Monitoring — Grafana (dashboard)"
if [[ -f "${CREDS_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${CREDS_FILE}"
  echo "  URL:      ${GRAFANA_URL:-https://${GRAFANA_HOST}}"
  echo "  User:     ${GRAFANA_USER:-admin}"
  echo "  Password: ${GRAFANA_PASSWORD:-<xem file ${CREDS_FILE}>}"
else
  echo "  URL:      https://${GRAFANA_HOST}"
  echo "  User:     admin"
  echo "  Password: kubectl get secret kube-prometheus-stack-grafana -n monitoring \\"
  echo "              -o jsonpath='{.data.admin-password}' | base64 -d; echo"
  echo ""
  echo "  (Chưa cài? bash scripts/install-monitoring.sh)"
fi

section "Monitoring — Prometheus (query metrics)"
echo "  Chạy trên VPS hoặc máy có kubectl:"
echo "    kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090"
echo "  Mở trình duyệt: http://localhost:9090"
echo "  Ví dụ query: up, container_memory_working_set_bytes"

section "Monitoring cơ bản (không cần Grafana)"
echo "  kubectl top nodes"
echo "  kubectl top pods -n shop"
echo "  kubectl top pods -n infra"
echo "  bash scripts/verify-app.sh"
echo "  bash scripts/diagnose-infra.sh"

section "Trạng thái pods monitoring"
if kubectl get ns monitoring >/dev/null 2>&1; then
  kubectl get pods -n monitoring -o wide
  kubectl get ingress -n monitoring 2>/dev/null || true
  kubectl get certificate -n monitoring 2>/dev/null || true
else
  echo "  Namespace monitoring chưa tồn tại — chạy: bash scripts/install-monitoring.sh"
fi

echo ""
