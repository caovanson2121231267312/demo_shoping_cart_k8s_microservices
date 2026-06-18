#!/usr/bin/env bash
# shopcaovanson — Verify Kubernetes cluster health
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INVENTORY="${SCRIPT_DIR}/inventory.env"
EXPECTED_NODES=3
TIMEOUT=300

log() { echo "[verify-cluster] $*"; }
die() { echo "[verify-cluster] ERROR: $*" >&2; exit 1; }
pass() { echo "[verify-cluster] PASS: $*"; }
fail() { echo "[verify-cluster] FAIL: $*" >&2; }

load_inventory() {
  [[ -f "${INVENTORY}" ]] || die "Inventory not found. Run setup-vms.sh first."
  # shellcheck source=/dev/null
  source "${INVENTORY}"
}

run_kubectl() {
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" "$@"
}

check_nodes() {
  log "Checking nodes..."
  local ready_count
  ready_count=$(run_kubectl "kubectl get nodes --no-headers 2>/dev/null | grep -c ' Ready' || true")
  local total_count
  total_count=$(run_kubectl "kubectl get nodes --no-headers 2>/dev/null | wc -l")

  run_kubectl "kubectl get nodes -o wide"

  if [[ "${total_count}" -lt "${EXPECTED_NODES}" ]]; then
    fail "Expected ${EXPECTED_NODES} nodes, found ${total_count}"
    return 1
  fi

  if [[ "${ready_count}" -lt "${EXPECTED_NODES}" ]]; then
    fail "Expected ${EXPECTED_NODES} Ready nodes, found ${ready_count}"
    return 1
  fi

  pass "All ${EXPECTED_NODES} nodes are Ready"
}

check_system_pods() {
  log "Checking system pods in kube-system..."
  run_kubectl "kubectl get pods -n kube-system -o wide"

  local not_running
  not_running=$(run_kubectl "kubectl get pods -n kube-system --no-headers 2>/dev/null | grep -vE 'Running|Completed' | wc -l")

  if [[ "${not_running}" -gt 0 ]]; then
    run_kubectl "kubectl get pods -n kube-system --no-headers | grep -vE 'Running|Completed' || true"
    fail "${not_running} system pod(s) not Running/Completed"
    return 1
  fi

  pass "All kube-system pods are Running or Completed"
}

check_cni() {
  log "Checking Calico pods..."
  run_kubectl "kubectl -n kube-system wait --for=condition=ready pod -l k8s-app=calico-node --timeout=${TIMEOUT}s"
  pass "Calico CNI pods are ready"
}

check_cluster_info() {
  log "Cluster info:"
  run_kubectl "kubectl cluster-info"
  run_kubectl "kubectl get componentstatuses 2>/dev/null || kubectl get --raw='/readyz?verbose' 2>/dev/null || true"
}

main() {
  load_inventory
  local errors=0

  check_nodes || errors=$((errors + 1))
  check_system_pods || errors=$((errors + 1))
  check_cni || errors=$((errors + 1))
  check_cluster_info

  if [[ "${errors}" -gt 0 ]]; then
    die "Cluster verification failed with ${errors} error(s)"
  fi

  pass "Cluster verification complete — cluster is healthy"
}

main "$@"
