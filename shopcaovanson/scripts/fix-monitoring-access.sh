#!/usr/bin/env bash
# Chẩn đoán + sửa https://monitor.shopcaovanson.xyz không vào được
#
# Usage:
#   bash scripts/fix-monitoring-access.sh           # chỉ chẩn đoán
#   bash scripts/fix-monitoring-access.sh --fix     # chẩn đoán + sửa tự động
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MONITORING_DIR="${PROJECT_ROOT}/k8s/base/monitoring"
GRAFANA_HOST="${GRAFANA_HOST:-monitor.shopcaovanson.xyz}"
VPS_IP="${VPS_IP:-110.172.29.72}"
DO_FIX=false

log() { echo "[fix-monitoring] $*"; }
warn() { echo "[fix-monitoring] WARN: $*" >&2; }
die() { echo "[fix-monitoring] ERROR: $*" >&2; exit 1; }

[[ "${1:-}" == "--fix" ]] && DO_FIX=true

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster"

ISSUES=0
note_issue() { ISSUES=$((ISSUES + 1)); warn "$*"; }

section() { echo ""; echo "========== $* =========="; }

check_dns() {
  section "1. DNS"
  local resolved=""
  if command -v dig >/dev/null 2>&1; then
    resolved=$(dig +short "${GRAFANA_HOST}" | tail -1)
  elif command -v host >/dev/null 2>&1; then
    resolved=$(host "${GRAFANA_HOST}" 2>/dev/null | awk '/has address/ {print $4; exit}')
  else
    resolved=$(getent ahosts "${GRAFANA_HOST}" 2>/dev/null | awk '/STREAM/ {print $1; exit}')
  fi
  echo "  ${GRAFANA_HOST} → ${resolved:-<không resolve>}"
  if [[ -z "${resolved}" ]]; then
    note_issue "DNS chưa trỏ — thêm bản ghi A: ${GRAFANA_HOST} → ${VPS_IP}"
  elif [[ "${resolved}" != "${VPS_IP}" ]]; then
    note_issue "DNS trỏ sai IP (cần ${VPS_IP}, đang ${resolved})"
  else
    log "DNS OK"
  fi
}

check_installed() {
  section "2. Namespace monitoring"
  if ! kubectl get ns monitoring >/dev/null 2>&1; then
    note_issue "Chưa cài monitoring — chạy: bash scripts/install-monitoring.sh"
    return 1
  fi
  kubectl get pods -n monitoring -o wide
  local grafana_ready
  grafana_ready=$(kubectl get pods -n monitoring -l app.kubernetes.io/name=grafana \
    -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null || echo "false")
  if [[ "${grafana_ready}" != "true" ]]; then
    note_issue "Grafana pod chưa Ready — xem: kubectl describe pod -n monitoring -l app.kubernetes.io/name=grafana"
  else
    log "Grafana pod Ready"
  fi
}

check_ingress() {
  section "3. Ingress Grafana"
  if ! kubectl get ingress grafana-ingress -n monitoring >/dev/null 2>&1; then
    note_issue "Thiếu ingress grafana-ingress — sẽ apply khi --fix"
    return 0
  fi
  kubectl get ingress grafana-ingress -n monitoring -o wide
  local addr
  addr=$(kubectl get ingress grafana-ingress -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)
  if [[ -n "${addr}" && "${addr}" != "${VPS_IP}" ]]; then
    warn "Ingress ADDRESS=${addr} (kỳ vọng ${VPS_IP})"
  fi
}

check_tls() {
  section "4. Certificate TLS (cert-manager)"
  if ! kubectl get certificate grafana-tls -n monitoring >/dev/null 2>&1; then
    note_issue "Chưa có Certificate grafana-tls — cert-manager chưa tạo (DNS/port 80?)"
    kubectl get challenges -A 2>/dev/null | head -5 || true
    return 0
  fi
  kubectl get certificate grafana-tls -n monitoring
  local ready
  ready=$(kubectl get certificate grafana-tls -n monitoring -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "False")
  if [[ "${ready}" != "True" ]]; then
    note_issue "grafana-tls chưa Ready — xem: kubectl describe certificate grafana-tls -n monitoring"
    kubectl describe certificate grafana-tls -n monitoring 2>/dev/null | tail -20 || true
    kubectl get challenges -n monitoring 2>/dev/null || kubectl get challenges -A 2>/dev/null | grep -i grafana || true
  else
    log "Certificate grafana-tls Ready"
  fi
}

check_http() {
  section "5. Test HTTP từ trong cluster"
  if ! kubectl get svc kube-prometheus-stack-grafana -n monitoring >/dev/null 2>&1; then
    note_issue "Service kube-prometheus-stack-grafana không tồn tại"
    return 0
  fi
  local code
  code=$(kubectl run mon-curl-test --rm -i --restart=Never -n monitoring --image=curlimages/curl -- \
    curl -s -o /dev/null -w "%{http_code}" --max-time 10 \
    "http://kube-prometheus-stack-grafana.monitoring.svc.cluster.local/login" 2>/dev/null || echo "000")
  echo "  Grafana service HTTP: ${code}"
  if [[ "${code}" != "200" && "${code}" != "302" ]]; then
    note_issue "Grafana service không phản hồi 200/302 (đang ${code})"
  fi
}

apply_fixes() {
  section "6. Áp dụng sửa (--fix)"
  if ! kubectl get ns monitoring >/dev/null 2>&1; then
    log "Cài monitoring..."
    bash "${SCRIPT_DIR}/install-monitoring.sh"
    return 0
  fi

  log "Apply Grafana Ingress + NetworkPolicy..."
  kubectl apply -f "${MONITORING_DIR}/grafana-ingress.yaml"
  kubectl apply -f "${MONITORING_DIR}/allow-grafana-ingress.yaml" 2>/dev/null || true
  kubectl apply -f "${MONITORING_DIR}/allow-monitoring-scrape.yaml" 2>/dev/null || true

  if ! kubectl get certificate grafana-tls -n monitoring >/dev/null 2>&1; then
    log "Đợi cert-manager tạo Certificate (30s)..."
    sleep 30
  fi

  local ready
  ready=$(kubectl get certificate grafana-tls -n monitoring -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "False")
  if [[ "${ready}" != "True" ]]; then
    log "Xóa Certificate/Secret cũ để cert-manager cấp lại..."
    kubectl delete certificate grafana-tls -n monitoring --ignore-not-found
    kubectl delete secret grafana-tls -n monitoring --ignore-not-found
    kubectl apply -f "${MONITORING_DIR}/grafana-ingress.yaml"
    log "Đợi TLS (tối đa 3 phút)..."
    for i in $(seq 1 18); do
      ready=$(kubectl get certificate grafana-tls -n monitoring -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "False")
      [[ "${ready}" == "True" ]] && break
      sleep 10
    done
  fi

  kubectl rollout restart deployment -n ingress-nginx -l app.kubernetes.io/component=controller 2>/dev/null || true
  log "Hoàn tất fix. Kiểm tra lại sau 1–2 phút."
}

main() {
  log "Grafana URL: https://${GRAFANA_HOST}"
  check_dns
  check_installed || true
  check_ingress
  check_tls
  check_http

  if [[ "${DO_FIX}" == "true" ]]; then
    apply_fixes
    echo ""
    check_tls
    bash "${SCRIPT_DIR}/monitoring-access.sh" 2>/dev/null || true
  else
    echo ""
    if [[ "${ISSUES}" -gt 0 ]]; then
      log "Phát hiện ${ISSUES} vấn đề. Chạy sửa tự động:"
      echo "  bash scripts/fix-monitoring-access.sh --fix"
    else
      log "Cấu hình cluster có vẻ OK — thử mở https://${GRAFANA_HOST} (Ctrl+F5)."
      echo "  Mật khẩu: bash scripts/monitoring-access.sh"
    fi
  fi
}

main "$@"
