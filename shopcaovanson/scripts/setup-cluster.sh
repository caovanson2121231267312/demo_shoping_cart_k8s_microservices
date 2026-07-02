#!/usr/bin/env bash
# shopcaovanson — Full cluster bootstrap: KVM VMs + kubeadm + verification
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CLUSTER_SETUP="${PROJECT_ROOT}/k8s/cluster-setup"

log() { echo "[setup-cluster] $*"; }
die() { echo "[setup-cluster] ERROR: $*" >&2; exit 1; }

require_root_for_vms() {
  if [[ "${SKIP_VMS:-false}" != "true" && "${EUID}" -ne 0 ]]; then
    die "VM setup requires root. Run: sudo $0  OR  set SKIP_VMS=true if cluster already exists."
  fi
}

setup_vms() {
  if [[ "${SKIP_VMS:-false}" == "true" ]]; then
    log "Skipping VM creation (SKIP_VMS=true)"
    return 0
  fi
  log "Creating KVM VMs..."
  bash "${CLUSTER_SETUP}/setup-vms.sh"
}

bootstrap_master() {
  if [[ "${SKIP_BOOTSTRAP:-false}" == "true" ]]; then
    log "Skipping master bootstrap (SKIP_BOOTSTRAP=true)"
    return 0
  fi
  log "Bootstrapping master node..."
  bash "${CLUSTER_SETUP}/bootstrap-master.sh"
}

bootstrap_workers() {
  if [[ "${SKIP_BOOTSTRAP:-false}" == "true" ]]; then
    log "Skipping worker bootstrap (SKIP_BOOTSTRAP=true)"
    return 0
  fi
  log "Joining worker nodes..."
  bash "${CLUSTER_SETUP}/bootstrap-worker.sh"
}

verify_cluster() {
  log "Verifying cluster health..."
  bash "${CLUSTER_SETUP}/verify-cluster.sh"
}

copy_kubeconfig() {
  local inventory="${CLUSTER_SETUP}/inventory.env"
  [[ -f "${inventory}" ]] || return 0
  # shellcheck source=/dev/null
  source "${inventory}"

  local kubeconfig_local="${HOME}/.kube/shopcaovanson-config"
  mkdir -p "${HOME}/.kube"
  scp -o StrictHostKeyChecking=no \
    "${K8S_SSH_USER}@${K8S_MASTER_IP}:.kube/config" \
    "${kubeconfig_local}" 2>/dev/null || true

  if [[ -f "${kubeconfig_local}" ]]; then
    log "Kubeconfig saved to ${kubeconfig_local}"
    log "Use: export KUBECONFIG=${kubeconfig_local}"
  fi
}

main() {
  require_root_for_vms
  setup_vms
  bootstrap_master
  bootstrap_workers
  verify_cluster
  copy_kubeconfig
  log "Cluster setup complete."
  log "Next steps:"
  log "  1. export KUBECONFIG=~/.kube/shopcaovanson-config"
  log "  2. bash scripts/deploy-all.sh"
}

main "$@"
