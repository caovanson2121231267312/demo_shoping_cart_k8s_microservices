#!/usr/bin/env bash
# Kiểm tra toàn bộ container và health endpoints ứng dụng
set -euo pipefail

API_URL="${API_URL:-https://shopapicaovanson.xyz}"
FRONTEND_URL="${FRONTEND_URL:-https://shopcaovanson.xyz}"
NAMESPACE_SHOP="${NAMESPACE_SHOP:-shop}"
NAMESPACE_INFRA="${NAMESPACE_INFRA:-infra}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $*"; }
fail() { echo -e "${RED}✗${NC} $*"; FAILED=1; }
warn() { echo -e "${YELLOW}!${NC} $*"; }

FAILED=0

section() { echo ""; echo "========== $* =========="; }

check_kubectl() {
  section "Kubernetes cluster"
  if ! kubectl cluster-info >/dev/null 2>&1; then
    fail "kubectl không kết nối được cluster"
    return
  fi
  ok "Cluster reachable"
  kubectl get nodes -o wide
  echo ""
  kubectl top nodes 2>/dev/null || warn "metrics-server chưa sẵn sàng (kubectl top)"
}

check_pods() {
  local ns="$1"
  section "Pods namespace: ${ns}"
  kubectl get pods -n "${ns}" -o wide
  echo ""
  local not_running
  not_running=$(kubectl get pods -n "${ns}" --field-selector=status.phase!=Running --no-headers 2>/dev/null | wc -l)
  if [[ "${not_running}" -gt 0 ]]; then
    fail "${not_running} pod(s) không Running trong ${ns}"
    kubectl get pods -n "${ns}" --field-selector=status.phase!=Running
  else
    ok "Tất cả pods Running trong ${ns}"
  fi
}

check_deployments() {
  section "Deployments namespace: ${NAMESPACE_SHOP}"
  kubectl get deployments -n "${NAMESPACE_SHOP}"
  echo ""
  for dep in $(kubectl get deployments -n "${NAMESPACE_SHOP}" -o jsonpath='{.items[*].metadata.name}'); do
    local ready desired
    ready=$(kubectl get deployment "${dep}" -n "${NAMESPACE_SHOP}" -o jsonpath='{.status.readyReplicas}')
    desired=$(kubectl get deployment "${dep}" -n "${NAMESPACE_SHOP}" -o jsonpath='{.spec.replicas}')
    ready=${ready:-0}
    if [[ "${ready}" == "${desired}" ]]; then
      ok "${dep}: ${ready}/${desired} ready"
    else
      fail "${dep}: ${ready}/${desired} ready"
    fi
  done
}

check_containers_in_pod() {
  section "Container status (shop)"
  kubectl get pods -n "${NAMESPACE_SHOP}" -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .status.containerStatuses[*]}{.name}{": "}{.ready}{" ("}{.state}{") "}{end}{"\n"}{end}' 2>/dev/null || \
  kubectl describe pods -n "${NAMESPACE_SHOP}" | grep -E "^(Name:|    Ready:|    State:|    Image:)" | head -60
}

check_ingress() {
  section "Ingress & TLS"
  kubectl get ingress -n "${NAMESPACE_SHOP}"
  echo ""
  kubectl get certificates -n "${NAMESPACE_SHOP}" 2>/dev/null || warn "cert-manager certificates không tìm thấy"
}

check_http() {
  section "HTTP health checks"
  if command -v curl >/dev/null 2>&1; then
    local api_status frontend_status
    api_status=$(curl -s -o /dev/null -w "%{http_code}" "${API_URL}/api/health" --max-time 10 || echo "000")
    frontend_status=$(curl -s -o /dev/null -w "%{http_code}" "${FRONTEND_URL}/" --max-time 10 || echo "000")

    if [[ "${api_status}" == "200" ]]; then
      ok "API ${API_URL}/api/health → HTTP ${api_status}"
      curl -s "${API_URL}/api/health" | head -1
    else
      fail "API ${API_URL}/api/health → HTTP ${api_status}"
    fi

    if [[ "${frontend_status}" == "200" ]]; then
      ok "Frontend ${FRONTEND_URL} → HTTP ${frontend_status}"
    else
      fail "Frontend ${FRONTEND_URL} → HTTP ${frontend_status}"
    fi
  else
    warn "curl không có — bỏ qua HTTP checks"
  fi
}

check_infra_connectivity() {
  section "Infra pods"
  for sts in postgres mongodb redis kafka elasticsearch zookeeper; do
    if kubectl get pod "${sts}-0" -n "${NAMESPACE_INFRA}" >/dev/null 2>&1; then
      local phase
      phase=$(kubectl get pod "${sts}-0" -n "${NAMESPACE_INFRA}" -o jsonpath='{.status.phase}')
      if [[ "${phase}" == "Running" ]]; then
        ok "${sts}-0: Running"
      else
        fail "${sts}-0: ${phase}"
      fi
    fi
  done
}

main() {
  echo "shopcaovanson — verify-app $(date '+%Y-%m-%d %H:%M:%S')"
  check_kubectl
  check_pods "${NAMESPACE_INFRA}"
  check_pods "${NAMESPACE_SHOP}"
  check_deployments
  check_containers_in_pod
  check_ingress
  check_infra_connectivity
  check_http

  echo ""
  if [[ "${FAILED}" -eq 0 ]]; then
    ok "Tất cả kiểm tra PASSED"
    exit 0
  else
    fail "Một số kiểm tra FAILED — xem log phía trên"
    exit 1
  fi
}

main "$@"
