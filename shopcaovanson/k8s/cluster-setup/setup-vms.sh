#!/usr/bin/env bash
# shopcaovanson — Create 3 KVM/libvirt VMs for Kubernetes cluster (Ubuntu 22.04 cloud-init)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLOUD_INIT_DIR="${SCRIPT_DIR}/cloud-init"
SSH_PUB_KEY="${SSH_PUB_KEY:-${HOME}/.ssh/id_rsa.pub}"
UBUNTU_IMAGE_URL="https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
UBUNTU_IMAGE="${SCRIPT_DIR}/ubuntu-22.04-server-cloudimg-amd64.img"
NETWORK_NAME="${NETWORK_NAME:-shopcaovanson-k8s}"
NETWORK_SUBNET="192.168.100.0/24"
NETWORK_GATEWAY="192.168.100.1"

declare -A VM_SPECS=(
  ["k8s-master"]="2 3072 30"
  ["k8s-worker-1"]="2 4096 50"
  ["k8s-worker-2"]="2 4096 50"
)

declare -A VM_IPS=(
  ["k8s-master"]="192.168.100.10"
  ["k8s-worker-1"]="192.168.100.11"
  ["k8s-worker-2"]="192.168.100.12"
)

log() { echo "[setup-vms] $*"; }
die() { echo "[setup-vms] ERROR: $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Required command not found: $1"
}

check_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    die "Run as root: sudo $0"
  fi
}

install_packages() {
  log "Installing KVM/libvirt packages..."
  apt-get update -qq
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
    qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils \
    virtinst cloud-image-utils genisoimage wget curl
  systemctl enable --now libvirtd
}

download_ubuntu_image() {
  if [[ ! -f "${UBUNTU_IMAGE}" ]]; then
    log "Downloading Ubuntu 22.04 cloud image..."
    wget -q -O "${UBUNTU_IMAGE}" "${UBUNTU_IMAGE_URL}"
  else
    log "Ubuntu cloud image already exists."
  fi
}

setup_network() {
  if ! virsh net-info "${NETWORK_NAME}" >/dev/null 2>&1; then
    log "Creating libvirt network ${NETWORK_NAME}..."
    cat > /tmp/shopcaovanson-net.xml <<EOF
<network>
  <name>${NETWORK_NAME}</name>
  <forward mode='nat'/>
  <bridge name='virbr-shop' stp='on' delay='0'/>
  <ip address='${NETWORK_GATEWAY}' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.100.100' end='192.168.100.200'/>
      <host mac='52:54:00:00:00:10' name='k8s-master' ip='192.168.100.10'/>
      <host mac='52:54:00:00:00:11' name='k8s-worker-1' ip='192.168.100.11'/>
      <host mac='52:54:00:00:00:12' name='k8s-worker-2' ip='192.168.100.12'/>
    </dhcp>
  </ip>
</network>
EOF
    virsh net-define /tmp/shopcaovanson-net.xml
    virsh net-autostart "${NETWORK_NAME}"
    virsh net-start "${NETWORK_NAME}"
    rm -f /tmp/shopcaovanson-net.xml
  else
    log "Network ${NETWORK_NAME} already exists."
    virsh net-start "${NETWORK_NAME}" 2>/dev/null || true
  fi
}

generate_cloud_init() {
  local vm_name="$1"
  local vm_ip="$2"
  mkdir -p "${CLOUD_INIT_DIR}/${vm_name}"

  if [[ ! -f "${SSH_PUB_KEY}" ]]; then
    die "SSH public key not found at ${SSH_PUB_KEY}. Set SSH_PUB_KEY env var."
  fi

  cat > "${CLOUD_INIT_DIR}/${vm_name}/meta-data" <<EOF
instance-id: ${vm_name}
local-hostname: ${vm_name}
EOF

  cat > "${CLOUD_INIT_DIR}/${vm_name}/user-data" <<EOF
#cloud-config
hostname: ${vm_name}
manage_etc_hosts: true
users:
  - name: ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - $(cat "${SSH_PUB_KEY}")
package_update: true
package_upgrade: true
packages:
  - curl
  - apt-transport-https
  - ca-certificates
  - gnupg
  - lsb-release
  - containerd
  - nfs-common
  - open-iscsi
write_files:
  - path: /etc/modules-load.d/k8s.conf
    content: |
      overlay
      br_netfilter
  - path: /etc/sysctl.d/k8s.conf
    content: |
      net.bridge.bridge-nf-call-iptables  = 1
      net.bridge.bridge-nf-call-ip6tables = 1
      net.ipv4.ip_forward                 = 1
  - path: /etc/netplan/60-static.yaml
    content: |
      network:
        version: 2
        ethernets:
          enp1s0:
            dhcp4: false
            addresses:
              - ${vm_ip}/24
            routes:
              - to: default
                via: ${NETWORK_GATEWAY}
            nameservers:
              addresses: [8.8.8.8, 8.8.4.4]
runcmd:
  - modprobe overlay
  - modprobe br_netfilter
  - sysctl --system
  - swapoff -a
  - sed -i '/ swap / s/^/#/' /etc/fstab
  - systemctl enable --now containerd
  - sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
  - systemctl restart containerd
  - netplan apply
EOF

  genisoimage -output "${CLOUD_INIT_DIR}/${vm_name}/seed.iso" \
    -volid cidata -joliet -rock \
    "${CLOUD_INIT_DIR}/${vm_name}/user-data" \
    "${CLOUD_INIT_DIR}/${vm_name}/meta-data"
}

create_vm() {
  local vm_name="$1"
  local cpus="$2"
  local memory_mb="$3"
  local disk_gb="$4"
  local vm_ip="${VM_IPS[$vm_name]}"
  local disk_path="/var/lib/libvirt/images/${vm_name}.qcow2"

  if virsh dominfo "${vm_name}" >/dev/null 2>&1; then
    log "VM ${vm_name} already exists, skipping."
    return 0
  fi

  log "Creating VM ${vm_name} (${cpus} CPU, ${memory_mb}MB RAM, ${disk_gb}GB disk, IP ${vm_ip})..."
  generate_cloud_init "${vm_name}" "${vm_ip}"

  qemu-img create -f qcow2 -F qcow2 -b "${UBUNTU_IMAGE}" "${disk_path}" "${disk_gb}G"

  local mac
  case "${vm_name}" in
    k8s-master)    mac="52:54:00:00:00:10" ;;
    k8s-worker-1)  mac="52:54:00:00:00:11" ;;
    k8s-worker-2)  mac="52:54:00:00:00:12" ;;
  esac

  virt-install \
    --name "${vm_name}" \
    --memory "${memory_mb}" \
    --vcpus "${cpus}" \
    --disk path="${disk_path}",format=qcow2,bus=virtio \
    --disk path="${CLOUD_INIT_DIR}/${vm_name}/seed.iso",device=cdrom \
    --os-variant ubuntu22.04 \
    --network network="${NETWORK_NAME}",model=virtio,mac="${mac}" \
    --graphics none \
    --console pty,target_type=serial \
    --import \
    --noautoconsole

  log "VM ${vm_name} created. Waiting for SSH..."
  local retries=60
  while ! ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 "ubuntu@${vm_ip}" "echo ready" 2>/dev/null; do
    retries=$((retries - 1))
    if [[ ${retries} -le 0 ]]; then
      die "SSH not available on ${vm_name} (${vm_ip})"
    fi
    sleep 10
  done
  log "VM ${vm_name} is reachable at ${vm_ip}"
}

write_inventory() {
  local inventory="${SCRIPT_DIR}/inventory.env"
  cat > "${inventory}" <<EOF
# Generated by setup-vms.sh — source before bootstrap scripts
export K8S_MASTER_IP=192.168.100.10
export K8S_WORKER1_IP=192.168.100.11
export K8S_WORKER2_IP=192.168.100.12
export K8S_MASTER_HOST=k8s-master
export K8S_WORKER1_HOST=k8s-worker-1
export K8S_WORKER2_HOST=k8s-worker-2
export K8S_SSH_USER=ubuntu
export K8S_POD_CIDR=192.168.0.0/16
export K8S_SERVICE_CIDR=10.96.0.0/12
EOF
  log "Inventory written to ${inventory}"
}

main() {
  check_root
  require_cmd virsh
  require_cmd virt-install
  require_cmd qemu-img
  require_cmd genisoimage
  require_cmd wget

  install_packages
  download_ubuntu_image
  setup_network

  for vm_name in k8s-master k8s-worker-1 k8s-worker-2; do
    read -r cpus memory_mb disk_gb <<< "${VM_SPECS[$vm_name]}"
    create_vm "${vm_name}" "${cpus}" "${memory_mb}" "${disk_gb}"
  done

  write_inventory
  log "All VMs created successfully."
  log "Next: source ${SCRIPT_DIR}/inventory.env && bash ${SCRIPT_DIR}/bootstrap-master.sh"
}

main "$@"
