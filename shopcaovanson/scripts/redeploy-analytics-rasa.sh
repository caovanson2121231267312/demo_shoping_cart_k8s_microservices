#!/usr/bin/env bash
# Deploy analytics-service + rasa-service (build image local + apply manifests)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
NS="${NS:-shop}"

log() { echo "[redeploy-analytics-rasa] $*"; }

if ! kubectl get secret analytics-service-secret -n "${NS}" >/dev/null 2>&1; then
  echo "Missing analytics-service-secret. Run: bash scripts/create-secrets.sh --apply-only --force"
  exit 1
fi

kubectl apply -f "${PROJECT_ROOT}/k8s/base/api-gateway/configmap.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/analytics-service/configmap.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/analytics-service/deployment.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/analytics-service/service.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/rasa-service/configmap.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/rasa-service/deployment.yaml"
kubectl apply -f "${PROJECT_ROOT}/k8s/base/rasa-service/service.yaml"

if [[ "${SKIP_BUILD:-}" != "1" ]]; then
  bash "${SCRIPT_DIR}/build-images.sh" analytics-service rasa-service
else
  for dep in analytics-service rasa-service api-gateway; do
    kubectl patch deployment "${dep}" -n "${NS}" --type=json \
      -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"IfNotPresent"}]' \
      2>/dev/null || true
  done
  kubectl rollout restart deployment/analytics-service deployment/rasa-service deployment/api-gateway -n "${NS}"
fi

for dep in analytics-service rasa-service api-gateway; do
  log "Waiting for deployment/${dep}..."
  kubectl rollout status "deployment/${dep}" -n "${NS}" --timeout=300s
done

kubectl get pods,svc -n "${NS}" | grep -E 'analytics|rasa|api-gateway' || true
log "Done."
