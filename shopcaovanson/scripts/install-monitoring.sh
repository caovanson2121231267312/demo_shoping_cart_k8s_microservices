#!/usr/bin/env bash
# Cài Prometheus + Grafana (kube-prometheus-stack) — tùy chọn sau khi deploy app
#
# Usage:
#   bash scripts/install-monitoring.sh
#   bash scripts/monitoring-access.sh   # xem URL + mật khẩu
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MONITORING_DIR="${PROJECT_ROOT}/k8s/base/monitoring"
CREDS_FILE="${PROJECT_ROOT}/.monitoring-credentials"
CHART_VERSION="${MONITORING_CHART_VERSION:-65.1.1}"
GRAFANA_HOST="${GRAFANA_HOST:-monitor.shopcaovanson.xyz}"

log() { echo "[install-monitoring] $*"; }
die() { echo "[install-monitoring] ERROR: $*" >&2; exit 1; }

command -v helm >/dev/null 2>&1 || die "helm not found"
command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to Kubernetes cluster"

if ! kubectl get ingressclass nginx >/dev/null 2>&1; then
  die "nginx IngressClass not found — chạy deploy-all.sh trước"
fi

GRAFANA_PASS="${GRAFANA_ADMIN_PASSWORD:-}"
if [[ -z "${GRAFANA_PASS}" ]]; then
  GRAFANA_PASS=$(openssl rand -base64 18 | tr -d '/+=' | head -c 20)
fi

log "Adding helm repo prometheus-community..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
helm repo update prometheus-community

log "Installing kube-prometheus-stack (chart ${CHART_VERSION})..."
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --version "${CHART_VERSION}" \
  -f "${MONITORING_DIR}/values-kube-prometheus.yaml" \
  --set "grafana.adminPassword=${GRAFANA_PASS}" \
  --set "grafana.grafana.ini.server.root_url=https://${GRAFANA_HOST}/" \
  --wait --timeout 15m

log "Waiting for Grafana pod..."
kubectl wait --for=condition=ready pod \
  -l app.kubernetes.io/name=grafana -n monitoring \
  --timeout=300s

log "Applying Grafana Ingress + NetworkPolicy scrape..."
kubectl apply -f "${MONITORING_DIR}/grafana-ingress.yaml"
kubectl apply -f "${MONITORING_DIR}/allow-monitoring-scrape.yaml"

cat > "${CREDS_FILE}" <<EOF
# shopcaovanson monitoring — KHÔNG commit git
# Generated: $(date -Iseconds 2>/dev/null || date)

GRAFANA_URL=https://${GRAFANA_HOST}
GRAFANA_USER=admin
GRAFANA_PASSWORD=${GRAFANA_PASS}

PROMETHEUS_SVC=kube-prometheus-stack-prometheus
PROMETHEUS_NS=monitoring
EOF
chmod 600 "${CREDS_FILE}"

echo ""
log "=========================================="
log " MONITORING ĐÃ CÀI"
log "=========================================="
echo ""
echo "  Grafana (trình duyệt):"
echo "    URL:      https://${GRAFANA_HOST}"
echo "    User:     admin"
echo "    Password: ${GRAFANA_PASS}"
echo ""
echo "  Lưu mật khẩu: ${CREDS_FILE}"
echo ""
echo "  DNS cần có bản ghi A:"
echo "    ${GRAFANA_HOST} → 110.172.29.72"
echo ""
echo "  Prometheus (chỉ admin — port-forward):"
echo "    kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090"
echo "    → http://localhost:9090"
echo ""
echo "  Xem lại: bash scripts/monitoring-access.sh"
echo ""
