#!/usr/bin/env bash
# shopcaovanson — Áp dụng hardening bảo mật K8s (NetworkPolicy, PSS, RBAC, ingress headers)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SECURITY_DIR="${PROJECT_ROOT}/k8s/base/security"

log() { echo "[apply-security] $*"; }
die() { echo "[apply-security] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"
kubectl cluster-info >/dev/null 2>&1 || die "Cannot connect to cluster"

apply_manifests() {
  log "Applying Pod Security Standards + NetworkPolicy + RBAC..."
  kubectl apply -f "${SECURITY_DIR}/pod-security.yaml"
  kubectl apply -f "${SECURITY_DIR}/network-policies.yaml"
  kubectl apply -f "${SECURITY_DIR}/rbac-ci.yaml"
}

patch_ingress_security() {
  log "Patching shop-ingress security annotations..."
  kubectl annotate ingress shop-ingress -n shop --overwrite \
    nginx.ingress.kubernetes.io/force-ssl-redirect="true" \
    nginx.ingress.kubernetes.io/ssl-protocols="TLSv1.2 TLSv1.3" \
    nginx.ingress.kubernetes.io/limit-rps="50" \
    nginx.ingress.kubernetes.io/limit-burst-multiplier="5" \
    nginx.ingress.kubernetes.io/proxy-hide-headers="Server" \
    nginx.ingress.kubernetes.io/enable-cors="false"
}

verify() {
  log "Network policies:"
  kubectl get networkpolicy -n shop
  kubectl get networkpolicy -n infra
  log "Namespace PSS labels:"
  kubectl get ns shop infra --show-labels | grep -E 'NAME|pod-security' || kubectl get ns shop infra --show-labels
  log "Done. Run smoke test: curl https://vocabee.cloud/api/health"
}

main() {
  apply_manifests
  patch_ingress_security
  verify
}

main "$@"
