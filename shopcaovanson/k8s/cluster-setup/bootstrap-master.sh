#!/usr/bin/env bash
# shopcaovanson — Bootstrap Kubernetes master node (kubeadm 1.29 + Calico CNI)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INVENTORY="${SCRIPT_DIR}/inventory.env"
K8S_VERSION="${K8S_VERSION:-1.29}"
CALICO_VERSION="${CALICO_VERSION:-v3.27.0}"
JOIN_FILE="${SCRIPT_DIR}/kubeadm-join.sh"

log() { echo "[bootstrap-master] $*"; }
die() { echo "[bootstrap-master] ERROR: $*" >&2; exit 1; }

load_inventory() {
  [[ -f "${INVENTORY}" ]] || die "Inventory not found. Run setup-vms.sh first."
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

init_cluster() {
  log "Initializing cluster on ${K8S_MASTER_IP}..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" bash -s <<REMOTE
set -euo pipefail
if [[ -f /etc/kubernetes/admin.conf ]]; then
  echo "Cluster already initialized."
  exit 0
fi
sudo kubeadm init \
  --pod-network-cidr=${K8S_POD_CIDR} \
  --service-cidr=${K8S_SERVICE_CIDR} \
  --apiserver-advertise-address=${K8S_MASTER_IP} \
  --control-plane-endpoint=${K8S_MASTER_IP} \
  --node-name=${K8S_MASTER_HOST}
mkdir -p \$HOME/.kube
sudo cp -f /etc/kubernetes/admin.conf \$HOME/.kube/config
sudo chown \$(id -u):\$(id -g) \$HOME/.kube/config
REMOTE
}

install_calico() {
  log "Installing Calico CNI ${CALICO_VERSION}..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" bash -s <<REMOTE
set -euo pipefail
curl -fsSL "https://raw.githubusercontent.com/projectcalico/calico/${CALICO_VERSION}/manifests/calico.yaml" -o /tmp/calico.yaml
sed -i 's|# - name: CALICO_IPV4POOL_CIDR|- name: CALICO_IPV4POOL_CIDR|' /tmp/calico.yaml
sed -i 's|#   value: "192.168.0.0/16"|  value: "${K8S_POD_CIDR}"|' /tmp/calico.yaml
kubectl apply -f /tmp/calico.yaml
kubectl -n kube-system wait --for=condition=ready pod -l k8s-app=calico-node --timeout=300s
REMOTE
}

untaint_master() {
  log "Allowing workloads on master (single-VPS demo)..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" \
    "kubectl taint nodes --all node-role.kubernetes.io/control-plane- 2>/dev/null || true"
}

generate_join_command() {
  log "Generating worker join command..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" bash -s <<'REMOTE' > "${JOIN_FILE}"
set -euo pipefail
JOIN_CMD=$(sudo kubeadm token create --print-join-command 2>/dev/null)
if [[ -z "${JOIN_CMD}" ]]; then
  JOIN_CMD=$(sudo kubeadm token create --print-join-command)
fi
echo "#!/usr/bin/env bash"
echo "set -euo pipefail"
echo "${JOIN_CMD} --node-name=\${K8S_NODE_NAME}"
REMOTE
  chmod +x "${JOIN_FILE}"
  log "Join command saved to ${JOIN_FILE}"
  echo ""
  echo "========================================"
  echo "Worker join command:"
  cat "${JOIN_FILE}"
  echo "========================================"
}

install_helm_components() {
  log "Installing helm on master (for cert-manager and ingress)..."
  ssh -o StrictHostKeyChecking=no "${K8S_SSH_USER}@${K8S_MASTER_IP}" bash -s <<'REMOTE'
set -euo pipefail
if ! command -v helm >/dev/null 2>&1; then
  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi
REMOTE
}

main() {
  load_inventory
  install_k8s_packages "${K8S_MASTER_IP}"
  init_cluster
  install_calico
  untaint_master
  generate_join_command
  install_helm_components
  log "Master bootstrap complete."
  log "Next: bash ${SCRIPT_DIR}/bootstrap-worker.sh"
}

main "$@"
