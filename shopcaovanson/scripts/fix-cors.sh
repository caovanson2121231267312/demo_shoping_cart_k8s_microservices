#!/usr/bin/env bash
# Sửa CORS: shopcaovanson.xyz → vocabee.cloud/api
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log() { echo "[fix-cors] $*"; }

kubectl apply -f "${PROJECT_ROOT}/k8s/base/api-gateway/configmap.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/ingress/ingress.yaml"

bash "${SCRIPT_DIR}/apply-security.sh" 2>/dev/null || {
  kubectl annotate ingress shop-ingress -n shop --overwrite \
    nginx.ingress.kubernetes.io/enable-cors="true" \
    nginx.ingress.kubernetes.io/cors-allow-origin="https://shopcaovanson.xyz, https://www.shopcaovanson.xyz"
}

if [[ "${SKIP_BUILD:-}" != "1" ]]; then
  export BUILDKIT_HOST="${BUILDKIT_HOST:-unix:///run/buildkit/buildkitd.sock}"
  bash "${SCRIPT_DIR}/build-images.sh" api-gateway
else
  kubectl rollout restart deployment/api-gateway -n shop
fi

kubectl rollout status deployment/api-gateway -n shop --timeout=180s

log "Test CORS (từ VPS):"
curl -sI -X OPTIONS "https://vocabee.cloud/api/categories" \
  -H "Origin: https://shopcaovanson.xyz" \
  -H "Access-Control-Request-Method: GET" | grep -i access-control || true

log "Done. Hard refresh trình duyệt (Ctrl+Shift+R)."
