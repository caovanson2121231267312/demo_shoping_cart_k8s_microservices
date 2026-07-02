#!/usr/bin/env bash
# Sửa kafka-0 CrashLoopBackOff trên VPS (OOM, PVC corrupt, DNS auth-service fail)
set -euo pipefail

NS="${NAMESPACE:-infra}"
YES=false

log() { echo "[fix-kafka] $*"; }
die() { echo "[fix-kafka] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes|-y) YES=true; shift ;;
    -h|--help)
      cat <<'EOF'
Usage:
  bash scripts/fix-kafka.sh           # chẩn đoán
  bash scripts/fix-kafka.sh --yes     # reset PVC kafka + apply manifest mới

Reset xóa topic Kafka (OK cho dev) — postgres/mongo data giữ nguyên.
EOF
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log "=== Kafka status ==="
kubectl get pods -n "${NS}" -l app=kafka -o wide 2>/dev/null || true
kubectl get pods -n "${NS}" -l app=zookeeper -o wide 2>/dev/null || true

if kubectl get pod kafka-0 -n "${NS}" >/dev/null 2>&1; then
  log "=== kafka-0 logs (tail) ==="
  kubectl logs kafka-0 -n "${NS}" --tail=40 2>/dev/null || true
  kubectl logs kafka-0 -n "${NS}" --previous --tail=40 2>/dev/null || true
  log "=== kafka-0 events ==="
  kubectl describe pod kafka-0 -n "${NS}" 2>/dev/null | tail -25 || true
fi

if [[ "${YES}" != "true" ]]; then
  echo ""
  log "Nếu thấy OOMKilled hoặc InconsistentClusterId / log corrupt, chạy:"
  echo "  bash scripts/fix-kafka.sh --yes"
  echo ""
  log "Sau kafka Running:"
  echo "  kubectl rollout restart deployment/auth-service -n shop"
  echo "  # hoặc Re-run GitHub Actions deploy-k8s"
  exit 0
fi

log "Scaling kafka down..."
kubectl scale statefulset kafka -n "${NS}" --replicas=0
kubectl wait --for=delete pod/kafka-0 -n "${NS}" --timeout=120s 2>/dev/null || true

PVC="kafka-data-kafka-0"
if kubectl get pvc "${PVC}" -n "${NS}" >/dev/null 2>&1; then
  log "Deleting PVC ${PVC} (reset Kafka data)..."
  kubectl delete pvc "${PVC}" -n "${NS}" --wait=true
fi

log "Applying kafka manifest (heap + startupProbe)..."
kubectl apply -f "${PROJECT_ROOT}/k8s/base/infra/kafka.yaml"

log "Scaling kafka up..."
kubectl scale statefulset kafka -n "${NS}" --replicas=1

log "Waiting for kafka-0 (tối đa 10 phút)..."
kubectl rollout status statefulset/kafka -n "${NS}" --timeout=600s

log "✓ kafka-0 Running"
kubectl get pods -n "${NS}" -l app=kafka

log ""
log "Test DNS từ shop:"
kubectl run kafka-dns-test --rm -i --restart=Never --image=busybox:1.36 -n shop -- \
  nslookup kafka.infra.svc.cluster.local 2>/dev/null || true

log "Restart auth-service (cần Kafka lúc startup)..."
kubectl rollout restart deployment/auth-service -n shop 2>/dev/null || true
kubectl rollout status deployment/auth-service -n shop --timeout=300s 2>/dev/null || true

log "Done. Kiểm tra: kubectl get pods -n infra && kubectl get pods -n shop -l app=auth-service"
