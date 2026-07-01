#!/usr/bin/env bash
# Chẩn đoán + sửa https://monitor.shopcaovanson.xyz không vào được
#
# Usage:
#   bash scripts/fix-monitoring-access.sh           # chẩn đoán
#   bash scripts/fix-monitoring-access.sh --fix     # sửa + renew TLS
#   bash scripts/fix-monitoring-access.sh --renew-cert
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MONITORING_DIR="${PROJECT_ROOT}/k8s/base/monitoring"
GRAFANA_HOST="${GRAFANA_HOST:-monitor.shopcaovanson.xyz}"
VPS_IP="${VPS_IP:-110.172.29.72}"
DO_FIX=false
DO_RENEW=false

log() { echo "[fix-monitoring] $*"; }
warn() { echo "[fix-monitoring] WARN: $*" >&2; }
die() { echo "[fix-monitoring] ERROR: $*" >&2; exit 1; }

for arg in "$@"; do
  case "${arg}" in
    --fix) DO_FIX=true ;;
    --renew-cert) DO_RENEW=true; DO_FIX=true ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster"

ISSUES=0
DNS_OK=false
note_issue() { ISSUES=$((ISSUES + 1)); warn "$*"; }

section() { echo ""; echo "========== $* =========="; }

resolve_dns() {
  local resolver="$1"
  if [[ -n "${resolver}" ]]; then
    dig +short "${GRAFANA_HOST}" @"${resolver}" 2>/dev/null | tail -1
  elif command -v dig >/dev/null 2>&1; then
    dig +short "${GRAFANA_HOST}" 2>/dev/null | tail -1
  elif command -v host >/dev/null 2>&1; then
    host "${GRAFANA_HOST}" 2>/dev/null | awk '/has address/ {print $4; exit}'
  else
    getent ahosts "${GRAFANA_HOST}" 2>/dev/null | awk '/STREAM/ {print $1; exit}'
  fi
}

check_dns() {
  section "1. DNS (nhiều resolver)"
  local local_ip google_ip cloudflare_ip
  local_ip=$(resolve_dns "")
  google_ip=$(resolve_dns "8.8.8.8")
  cloudflare_ip=$(resolve_dns "1.1.1.1")

  echo "  local resolver     → ${local_ip:-<rỗng>}"
  echo "  Google 8.8.8.8     → ${google_ip:-<rỗng>}"
  echo "  Cloudflare 1.1.1.1 → ${cloudflare_ip:-<rỗng>}"
  echo "  Cần trỏ về:        ${VPS_IP}"

  for ip in "${local_ip}" "${google_ip}" "${cloudflare_ip}"; do
    if [[ "${ip}" == "${VPS_IP}" ]]; then
      DNS_OK=true
      break
    fi
  done

  if [[ "${DNS_OK}" != "true" ]]; then
    note_issue "DNS chưa propagate hoặc sai — kiểm tra panel domain:"
    echo "    Type: A | Name/Host: monitor | Value: ${VPS_IP}"
    echo "    (KHÔNG dùng CNAME nếu chưa chắc; TTL 300–600)"
    echo "    Kiểm tra từ máy Windows: nslookup ${GRAFANA_HOST}"
  else
    log "DNS OK (ít nhất 1 resolver trả ${VPS_IP})"
  fi
}

check_installed() {
  section "2. Pods monitoring"
  if ! kubectl get ns monitoring >/dev/null 2>&1; then
    note_issue "Chưa cài monitoring — chạy: bash scripts/install-monitoring.sh"
    return 1
  fi
  kubectl get pods -n monitoring -o wide
  local grafana_ready
  grafana_ready=$(kubectl get pods -n monitoring -l app.kubernetes.io/name=grafana \
    -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null || echo "false")
  if [[ "${grafana_ready}" != "true" ]]; then
    note_issue "Grafana pod chưa Ready"
  else
    log "Grafana pod Ready"
  fi
}

check_ingress() {
  section "3. Ingress Grafana"
  if ! kubectl get ingress grafana-ingress -n monitoring >/dev/null 2>&1; then
    note_issue "Thiếu grafana-ingress"
    return 0
  fi
  kubectl get ingress -n monitoring
}

check_tls() {
  section "4. Certificate TLS"
  if ! kubectl get certificate grafana-tls -n monitoring >/dev/null 2>&1; then
    note_issue "Chưa có Certificate grafana-tls"
    return 0
  fi
  kubectl get certificate grafana-tls -n monitoring
  local ready msg
  ready=$(kubectl get certificate grafana-tls -n monitoring \
    -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "False")
  msg=$(kubectl get certificate grafana-tls -n monitoring \
    -o jsonpath='{.status.conditions[?(@.type=="Ready")].message}' 2>/dev/null || true)
  if [[ "${ready}" != "True" ]]; then
    note_issue "grafana-tls chưa Ready — ${msg:-xem describe bên dưới}"
    kubectl describe certificate grafana-tls -n monitoring 2>/dev/null | tail -25 || true
    echo ""
    kubectl get challenges -n monitoring 2>/dev/null || true
    kubectl describe challenge -n monitoring 2>/dev/null | tail -30 || true
  else
    log "Certificate grafana-tls Ready ✓"
  fi
}

check_external_http() {
  section "5. Test từ VPS ra internet"
  if ! command -v curl >/dev/null 2>&1; then
    warn "curl không có — bỏ qua test external"
    return 0
  fi

  echo "  HTTP (port 80 — Let's Encrypt cần path này):"
  local http_code https_code
  http_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 15 \
    "http://${GRAFANA_HOST}/login" 2>/dev/null || echo "000")
  # VPS thường không resolve subdomain local — test bằng --resolve
  if [[ "${http_code}" == "000" ]]; then
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 15 \
      --resolve "${GRAFANA_HOST}:80:${VPS_IP}" \
      "http://${GRAFANA_HOST}/login" 2>/dev/null || echo "000")
    echo "    http://${GRAFANA_HOST}/login → HTTP ${http_code} (qua --resolve ${VPS_IP})"
  else
    echo "    http://${GRAFANA_HOST}/login → HTTP ${http_code}"
  fi

  echo "  HTTPS:"
  https_code=$(curl -sk -o /dev/null -w "%{http_code}" --max-time 15 \
    "https://${GRAFANA_HOST}/login" 2>/dev/null || echo "000")
  if [[ "${https_code}" == "000" ]]; then
    https_code=$(curl -sk -o /dev/null -w "%{http_code}" --max-time 15 \
      --resolve "${GRAFANA_HOST}:443:${VPS_IP}" \
      "https://${GRAFANA_HOST}/login" 2>/dev/null || echo "000")
    echo "    https://${GRAFANA_HOST}/login → HTTP ${https_code} (qua --resolve ${VPS_IP})"
  else
    echo "    https://${GRAFANA_HOST}/login → HTTP ${https_code}"
  fi

  if [[ "${http_code}" == "000" && "${https_code}" == "000" ]]; then
    note_issue "Không kết nối được — kiểm tra ingress: kubectl get ingress -n monitoring"
  fi
  if [[ "${https_code}" == "200" || "${https_code}" == "302" ]]; then
    log "HTTPS truy cập OK"
  elif [[ "${https_code}" == "000" ]]; then
    :
  else
    warn "HTTPS code ${https_code}"
  fi
}

renew_grafana_cert() {
  section "6. Renew certificate Let's Encrypt"
  log "Xóa challenge/order/certificate cũ (stale)..."
  kubectl delete challenge -n monitoring --all --ignore-not-found --wait=false
  kubectl delete order -n monitoring --all --ignore-not-found --wait=false
  kubectl delete certificaterequest -n monitoring --all --ignore-not-found --wait=false
  kubectl delete certificate grafana-tls -n monitoring --ignore-not-found --wait=false
  kubectl delete secret grafana-tls -n monitoring --ignore-not-found --wait=false
  kubectl delete ingress -n monitoring -l acme.cert-manager.io/http01-solver=true --ignore-not-found 2>/dev/null || true
  for ing in $(kubectl get ingress -n monitoring -o name 2>/dev/null | grep acme-http-solver || true); do
    kubectl delete -n monitoring "${ing}" --ignore-not-found 2>/dev/null || true
  done

  sleep 3
  kubectl apply -f "${MONITORING_DIR}/grafana-ingress.yaml"
  kubectl apply -f "${MONITORING_DIR}/allow-grafana-ingress.yaml" 2>/dev/null || true

  log "Đợi cert-manager cấp cert (tối đa 5 phút)..."
  local ready="False"
  for _ in $(seq 1 30); do
    ready=$(kubectl get certificate grafana-tls -n monitoring \
      -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "False")
    if [[ "${ready}" == "True" ]]; then
      log "grafana-tls Ready ✓"
      return 0
    fi
    local state
    state=$(kubectl get challenges -n monitoring -o jsonpath='{.items[0].status.state}' 2>/dev/null || echo "?")
    echo "  ... chờ TLS (challenge state: ${state})"
    sleep 10
  done
  warn "TLS vẫn chưa Ready sau 5 phút"
  kubectl describe challenge -n monitoring 2>/dev/null | tail -20 || true
  return 1
}

apply_fixes() {
  if ! kubectl get ns monitoring >/dev/null 2>&1; then
    bash "${SCRIPT_DIR}/install-monitoring.sh"
    return 0
  fi

  kubectl apply -f "${MONITORING_DIR}/grafana-ingress.yaml"
  kubectl apply -f "${MONITORING_DIR}/allow-grafana-ingress.yaml" 2>/dev/null || true
  kubectl apply -f "${MONITORING_DIR}/allow-monitoring-scrape.yaml" 2>/dev/null || true

  renew_grafana_cert || true

  kubectl rollout restart deployment -n ingress-nginx \
    -l app.kubernetes.io/component=controller 2>/dev/null || true
}

main() {
  log "Grafana: https://${GRAFANA_HOST}"
  check_dns
  check_installed || true
  check_ingress
  check_tls
  check_external_http

  if [[ "${DO_FIX}" == "true" ]]; then
    if [[ "${DNS_OK}" != "true" ]]; then
      die "DNS chưa trỏ ${VPS_IP} — sửa DNS trước. Test: dig +short ${GRAFANA_HOST} @8.8.8.8"
    fi
    apply_fixes
    echo ""
    check_tls
    check_external_http
    bash "${SCRIPT_DIR}/monitoring-access.sh" 2>/dev/null || true
  else
    echo ""
    if [[ "${DNS_OK}" == "true" ]]; then
      log "DNS đã OK — nếu vẫn không vào được, chạy:"
      echo "  bash scripts/fix-monitoring-access.sh --renew-cert"
    else
      log "Sửa DNS trước, sau đó: bash scripts/fix-monitoring-access.sh --fix"
    fi
    echo ""
    log "Vào tạm không cần domain:"
    echo "  kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80"
    echo "  → http://localhost:3000 (SSH tunnel từ Windows nếu cần)"
  fi
}

main "$@"
