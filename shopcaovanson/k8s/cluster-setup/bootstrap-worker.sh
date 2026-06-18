#!/usr/bin/env bash
# shopcaovanson — Join worker nodes to Kubernetes cluster
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INVENTORY="${SCRIPT_DIR}/inventory.env"
JOIN_FILE="${SCRIPT_DIR}/kubeadm-join.sh"
K8S_VERSION="${K8S_VERSION:-1.29}"

log() { echo "[bootstrap-worker] $*"; }
die() { echo "[bootstrap-worker] ERROR: $*" >&2; exit 1; }

load_inventory() {
  [[ -f "${INVENTORY}" ]] || die "Inventory not found. Run setup-vms.sh first."
  [[ -f "${JOIN_FILE}" ]] || die "Join file not found. Run bootstrap-master.sh first."
  # shellcheck source=/dev/null
  source "${INVENTORY}"
}

install_k8s_packages() {
  local host="$1"
  log "Installing Kubernetes ${K8S_VERSION} on ${host}..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${host}" bash -s <<REMOTE
set -euo pipefail
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v${K8S_VERSION}/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v${K8S_VERSION}/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update -qq
sudo apt-get install -y -qq kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable --now kubelet
REMOTE
}

join_worker() {
  local host="$1"
  local node_name="$2"

  log "Joining worker ${node_name} (${host})..."
  local already_joined
  already_joined=$(ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" \
    "kubectl get nodes -o name 2>/dev/null | grep -c '${node_name}' || true")

  if [[ "${already_joined}" -gt 0 ]]; then
    log "Node ${node_name} already joined, skipping."
    return 0
  fi

  local join_cmd
  join_cmd=$(grep -v '^#' "${JOIN_FILE}" | grep 'kubeadm join' | head -1)
  [[ -n "${join_cmd}" ]] || die "Could not parse join command from ${JOIN_FILE}"

  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${host}" bash -s <<REMOTE
set -euo pipefail
if [[ -f /etc/kubernetes/kubelet.conf ]]; then
  echo "Node already joined."
  exit 0
fi
sudo ${join_cmd} --node-name=${node_name}
REMOTE

  log "Waiting for node ${node_name} to become Ready..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" \
    "kubectl wait --for=condition=Ready node/${node_name} --timeout=300s"
  log "Node ${node_name} is Ready."
}

main() {
  load_inventory
  install_k8s_packages "${K8S_WORKER1_IP}"
  install_k8s_packages "${K8S_WORKER2_IP}"
  join_worker "${K8S_WORKER1_IP}" "${K8S_WORKER1_HOST}"
  join_worker "${K8S_WORKER2_IP}" "${K8S_WORKER2_HOST}"
  log "All workers joined successfully."
  log "Next: bash ${SCRIPT_DIR}/verify-cluster.sh"
}

main "$@"
