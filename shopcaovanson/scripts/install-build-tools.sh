#!/usr/bin/env bash
# Cài nerdctl + buildkit — build image trên VPS dùng containerd (KHÔNG cài docker.io)
#
# Ubuntu 22.04 thường KHÔNG có apt package "buildkit" → tải binary từ GitHub.
#
# Usage: sudo bash scripts/install-build-tools.sh
set -euo pipefail

NERDCTL_VERSION="${NERDCTL_VERSION:-1.7.6}"
BUILDKIT_VERSION="${BUILDKIT_VERSION:-0.13.2}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BUILDKIT_SOCKET="${BUILDKIT_SOCKET:-/run/buildkit/buildkitd.sock}"

log() { echo "[install-build-tools] $*"; }
die() { echo "[install-build-tools] ERROR: $*" >&2; exit 1; }

[[ "${EUID}" -eq 0 ]] || die "Chạy với sudo: sudo bash scripts/install-build-tools.sh"

ARCH=$(uname -m)
case "${ARCH}" in
  x86_64) BK_ARCH=amd64; NERD_ARCH=amd64 ;;
  aarch64|arm64) BK_ARCH=arm64; NERD_ARCH=arm64 ;;
  *) die "Unsupported arch: ${ARCH}" ;;
esac

if ! command -v containerd >/dev/null 2>&1; then
  die "containerd not found — cài K8s/containerd trước (Phase 2–3)"
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y wget ca-certificates tar

install_buildkit() {
  if command -v buildkitd >/dev/null 2>&1 && command -v buildctl >/dev/null 2>&1; then
    log "buildkit already installed: $(buildkitd --version 2>&1 | head -1)"
    return 0
  fi

  log "Installing buildkit v${BUILDKIT_VERSION} from GitHub..."
  tmp=$(mktemp -d)
  trap 'rm -rf "${tmp}"' RETURN

  wget -q "https://github.com/moby/buildkit/releases/download/v${BUILDKIT_VERSION}/buildkit-v${BUILDKIT_VERSION}.linux-${BK_ARCH}.tar.gz" \
    -O "${tmp}/buildkit.tgz"
  tar -xzf "${tmp}/buildkit.tgz" -C "${tmp}"

  install -m 755 "${tmp}/bin/buildkitd" "${INSTALL_DIR}/buildkitd"
  install -m 755 "${tmp}/bin/buildctl" "${INSTALL_DIR}/buildctl"
  log "✓ buildkitd + buildctl → ${INSTALL_DIR}"
}

install_buildkit_service() {
  mkdir -p /run/buildkit
  cat >/etc/systemd/system/buildkit.service <<EOF
[Unit]
Description=Moby BuildKit (container image builder)
Documentation=https://github.com/moby/buildkit

[Service]
Type=notify
ExecStart=${INSTALL_DIR}/buildkitd --addr unix://${BUILDKIT_SOCKET}
Restart=always
RestartSec=3
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable buildkit
  systemctl restart buildkit

  for i in $(seq 1 30); do
    if [[ -S "${BUILDKIT_SOCKET}" ]]; then
      log "✓ buildkitd running (${BUILDKIT_SOCKET})"
      return 0
    fi
    sleep 1
  done
  die "buildkitd socket not ready — check: journalctl -u buildkit -n 30"
}

install_nerdctl() {
  if command -v nerdctl >/dev/null 2>&1; then
    log "nerdctl already installed"
    return 0
  fi

  log "Installing nerdctl v${NERDCTL_VERSION}..."
  tmp=$(mktemp -d)
  wget -q "https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${NERD_ARCH}.tar.gz" \
    -O "${tmp}/nerdctl.tgz"
  tar -xzf "${tmp}/nerdctl.tgz" -C "${INSTALL_DIR}" nerdctl
  chmod +x "${INSTALL_DIR}/nerdctl"
  rm -rf "${tmp}"
  log "✓ nerdctl → ${INSTALL_DIR}/nerdctl"
}

configure_nerdctl() {
  mkdir -p /etc/nerdctl
  cat >/etc/nerdctl/nerdctl.toml <<'EOF'
namespace = "k8s.io"
snapshotter = "overlayfs"
EOF
}

install_buildkit
install_buildkit_service
install_nerdctl
configure_nerdctl

export BUILDKIT_HOST="unix://${BUILDKIT_SOCKET}"

log "Verifying..."
nerdctl version
buildctl debug workers 2>/dev/null | head -5 || true

cat >/etc/profile.d/buildkit.sh <<EOF
export BUILDKIT_HOST=unix://${BUILDKIT_SOCKET}
EOF

echo ""
log "=========================================="
log " ✓ Cài đặt hoàn tất"
log "=========================================="
echo ""
echo "  buildkitd:  systemctl status buildkit"
echo "  nerdctl:    nerdctl --namespace k8s.io images"
echo ""
echo "  Tiếp theo (user thường, không cần sudo):"
echo "    cd /home/demo_shoping_cart_k8s_microservices/shopcaovanson"
echo "    bash scripts/build-images.sh api-gateway    # thử 1 service"
echo "    bash scripts/build-images.sh                # tất cả"
echo ""
log "KHÔNG chạy: apt install docker.io (conflict containerd.io của K8s)"
