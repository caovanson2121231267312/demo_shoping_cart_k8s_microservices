#!/usr/bin/env bash
# Chẩn đoán infra pods khi deploy timeout — chạy trên VPS
set -euo pipefail

NS="${NAMESPACE:-infra}"

section() { echo ""; echo "========== $* =========="; }

section "Pods infra (tất cả trạng thái)"
kubectl get pods -n "${NS}" -o wide

section "Pods KHÔNG Running"
kubectl get pods -n "${NS}" --field-selector=status.phase!=Running 2>/dev/null || true

section "PVC (Pending = thiếu StorageClass / disk)"
kubectl get pvc -n "${NS}"

section "StorageClass"
kubectl get storageclass

section "Events infra (20 mới nhất)"
kubectl get events -n "${NS}" --sort-by='.lastTimestamp' | tail -20

for sts in postgres mongodb redis zookeeper kafka elasticsearch; do
  if kubectl get pod "${sts}-0" -n "${NS}" >/dev/null 2>&1; then
    phase=$(kubectl get pod "${sts}-0" -n "${NS}" -o jsonpath='{.status.phase}')
  else
    phase="NOT_FOUND"
  fi
  echo ""
  echo "--- ${sts}-0 phase=${phase} ---"
  kubectl describe pod "${sts}-0" -n "${NS}" 2>/dev/null | tail -30 || echo "(pod chưa tạo)"
  kubectl logs "${sts}-0" -n "${NS}" --tail=25 2>/dev/null || true
done

section "Node resources"
kubectl top nodes 2>/dev/null || echo "metrics-server chưa sẵn sàng"
free -h 2>/dev/null || true

echo ""
echo "Gợi ý:"
echo "  PVC Pending     → kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.26/deploy/local-path-storage.yaml"
echo "  OOM / Pending   → OVERLAY=dev bash scripts/deploy-all.sh"
echo "  MongoDB probe   → git pull && kubectl apply -f k8s/base/infra/mongodb.yaml && kubectl delete pod mongodb-0 -n infra"
