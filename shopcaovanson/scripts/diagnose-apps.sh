#!/usr/bin/env bash
# Chẩn đoán app pods (api-gateway, auth, ...) khi deploy timeout
set -euo pipefail

NS="${NAMESPACE:-shop}"

section() { echo ""; echo "========== $* =========="; }

section "Pods shop"
kubectl get pods -n "${NS}" -o wide

section "Pods KHÔNG Ready"
kubectl get pods -n "${NS}" --field-selector=status.phase!=Running 2>/dev/null || true
for p in $(kubectl get pods -n "${NS}" -o jsonpath='{range .items[?(@.status.phase=="Running")]}{.metadata.name}{" "}{end}' 2>/dev/null); do
  ready=$(kubectl get pod "$p" -n "${NS}" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "Unknown")
  [[ "${ready}" == "True" ]] || echo "  NOT READY: ${p}"
done

section "Events shop (15 mới nhất)"
kubectl get events -n "${NS}" --sort-by='.lastTimestamp' | tail -15

for dep in api-gateway auth-service product-service order-service chat-service notification-service search-service frontend; do
  section "deployment/${dep}"
  kubectl get deployment "${dep}" -n "${NS}" 2>/dev/null || { echo "(không tồn tại)"; continue; }
  kubectl get pods -n "${NS}" -l "app=${dep}" -o wide 2>/dev/null || true
  pod=$(kubectl get pods -n "${NS}" -l "app=${dep}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  if [[ -n "${pod}" ]]; then
    echo ""
    echo "--- describe ${pod} (Events) ---"
    kubectl describe pod "${pod}" -n "${NS}" 2>/dev/null | sed -n '/Events:/,$p' | head -20
    echo ""
    echo "--- logs ${pod} ---"
    kubectl logs "${pod}" -n "${NS}" --tail=30 2>/dev/null || true
    echo ""
    echo "--- previous logs (nếu restart) ---"
    kubectl logs "${pod}" -n "${NS}" --previous --tail=20 2>/dev/null || echo "(none)"
  fi
done

section "Image pull secrets"
kubectl get secret ghcr-secret -n "${NS}" 2>/dev/null || echo "ghcr-secret: không có (OK nếu image public hoặc đã import local)"

section "Kiểm tra image trên node"
if command -v ctr >/dev/null 2>&1; then
  ctr -n k8s.io images ls 2>/dev/null | grep -E 'ghcr.io/caovanson|frontend-web' || echo "(chưa có image local — chạy: bash scripts/build-images.sh)"
elif command -v crictl >/dev/null 2>&1; then
  crictl images 2>/dev/null | grep -E 'ghcr.io/caovanson|frontend-web' || echo "(chưa có image local)"
else
  echo "ctr/crictl not found"
fi

echo ""
echo "Gợi ý nhanh:"
echo "  ImagePullBackOff     → bash scripts/build-images.sh  (hoặc push GHCR + ghcr-secret)"
echo "  CrashLoop + redis    → kiểm tra REDIS_URL trong secret, redis-0 Running"
echo "  JWT_PUBLIC_KEY       → bash scripts/create-secrets.sh --force"
echo "  progress deadline    → kubectl describe deployment api-gateway -n shop"
