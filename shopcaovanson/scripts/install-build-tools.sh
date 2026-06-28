#!/usr/bin/env bash
# Cài nerdctl + buildkit — build image trên VPS dùng containerd (KHÔNG cài docker.io)
#
# VPS K8s dùng containerd.io → apt install docker.io bị conflict.
# Script này cài nerdctl (CLI containerd) + buildkit để build image.
#
# Usage: sudo bash scripts/install-build-tools.sh
set -euo pipefail

NERDCTL_VERSION="${NERDCTL_VERSION:-1.7.6}"

log() { echo "[install-build-tools] $*"; }
die() { echo "[install-build-tools] ERROR: $*" >&2; exit 1; }

[[ "${EUID}" -eq 0 ]] || die "Chạy với sudo: sudo bash scripts/install-build-tools.sh"

ARCH=$(uname -m)
case "${ARCH}" in
  x86_64) NERD_ARCH=amd64 ;;
  aarch64|arm64) NERD_ARCH=arm64 ;;
  *) die "Unsupported arch: ${ARCH}" ;;
esac

if ! command -v containerd >/dev/null 2>&1; then
  die "containerd not found — cài K8s/containerd trước (Phase 2–3)"
fi

log "Installing buildkit..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y buildkit wget ca-certificates

systemctl enable buildkit 2>/dev/null || true
systemctl start buildkit 2>/dev/null || true

if ! command -v nerdctl >/dev/null 2>&1; then
  log "Installing nerdctl v${NERDCTL_VERSION}..."
  tmp=$(mktemp -d)
  wget -q "https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${NERD_ARCH}.tar.gz" \
    -O "${tmp}/nerdctl.tgz"
  tar -xzf "${tmp}/nerdctl.tgz" -C /usr/local/bin nerdctl
  chmod +x /usr/local/bin/nerdctl
  rm -rf "${tmp}"
fi

mkdir -p /etc/nerdctl
cat >/etc/nerdctl/nerdctl.toml <<'EOF'
# Image build vào namespace K8s — kubelet dùng ngay, không cần push registry
namespace = "k8s.io"
snapshotter = "overlayfs"
EOF

export BUILDKIT_HOST="${BUILDKIT_HOST:-unix:///run/buildkit/buildkitd.sock}"

log "Verifying..."
nerdctl version
if command -v buildctl >/dev/null 2>&1; then
  buildctl debug workers 2>/dev/null | head -5 || log "buildkit socket OK (buildctl)"
fi

echo ""
log "✓ Sẵn sàng build. Tiếp theo:"
echo "  cd /home/demo_shoping_cart_k8s_microservices/shopcaovanson"
echo "  bash scripts/build-images.sh"
echo ""
log "Lưu ý: KHÔNG chạy apt install docker.io — conflict với containerd.io của K8s"
