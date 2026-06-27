#!/usr/bin/env bash
# Tạo kubeconfig cho GitHub Actions CI/CD (ServiceAccount cicd-deployer)
# Chạy trên VPS sau khi apply k8s/base/security/rbac-ci.yaml
set -euo pipefail

NAMESPACE="${NAMESPACE:-shop}"
SA_NAME="${SA_NAME:-cicd-deployer}"
DURATION="${DURATION:-87600h}"   # ~10 năm
OUT_FILE="${OUT_FILE:-./cicd-kubeconfig.yaml}"
CLUSTER_NAME="${CLUSTER_NAME:-shopcaovanson}"
API_SERVER="${API_SERVER:-https://110.172.29.72:6443}"

log() { echo "[create-cicd-kubeconfig] $*"; }
die() { echo "[create-cicd-kubeconfig] ERROR: $*" >&2; exit 1; }

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

if ! kubectl get sa "${SA_NAME}" -n "${NAMESPACE}" >/dev/null 2>&1; then
  die "ServiceAccount ${SA_NAME} not found in ${NAMESPACE}. Run: bash scripts/apply-security.sh"
fi

# Lấy CA từ kubeconfig hiện tại hoặc cluster
CA_DATA=$(kubectl config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')
if [[ -z "${CA_DATA}" ]]; then
  CA_DATA=$(sudo cat /etc/kubernetes/pki/ca.crt | base64 -w0)
fi

TOKEN=$(kubectl create token "${SA_NAME}" -n "${NAMESPACE}" --duration="${DURATION}")

cat > "${OUT_FILE}" <<EOF
apiVersion: v1
kind: Config
clusters:
- name: ${CLUSTER_NAME}
  cluster:
    server: ${API_SERVER}
    certificate-authority-data: ${CA_DATA}
contexts:
- name: cicd@${CLUSTER_NAME}
  context:
    cluster: ${CLUSTER_NAME}
    namespace: ${NAMESPACE}
    user: ${SA_NAME}
current-context: cicd@${CLUSTER_NAME}
users:
- name: ${SA_NAME}
  user:
    token: ${TOKEN}
EOF

chmod 600 "${OUT_FILE}"
log "Kubeconfig saved: ${OUT_FILE}"
log ""
log "=== Thêm vào GitHub Secrets ==="
log "Secret name: KUBECONFIG_DATA"
log "Secret value (copy dòng dưới):"
echo ""
base64 -w0 "${OUT_FILE}" 2>/dev/null || base64 "${OUT_FILE}"
echo ""
log ""
log "Test local: KUBECONFIG=${OUT_FILE} kubectl get deployments -n shop"
