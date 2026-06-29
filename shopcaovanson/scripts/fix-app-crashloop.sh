#!/usr/bin/env bash
# Sửa CrashLoopBackOff phổ biến sau build-images:
#   - postgres/mongo/elastic password ≠ shop secrets (PVC init cũ)
#   - redis WRONGPASS
#   - notification/search: probe port + KAFKA_BOOTSTRAP_SERVERS
#
# Usage:
#   bash scripts/fix-app-crashloop.sh              # kiểm tra + hướng dẫn
#   bash scripts/fix-app-crashloop.sh --reset-infra --yes   # reset PVC + sync secrets
#   bash scripts/fix-app-crashloop.sh --sync-only           # chỉ sync secrets (password đã đúng)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"

RESET_INFRA=false
YES=false
SYNC_ONLY=false

log() { echo "[fix-app-crashloop] $*"; }
die() { echo "[fix-app-crashloop] ERROR: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --reset-infra) RESET_INFRA=true; shift ;;
    --sync-only)   SYNC_ONLY=true; shift ;;
    --yes|-y)      YES=true; shift ;;
    -h|--help)
      cat <<'EOF'
Sửa app CrashLoopBackOff (password auth, kafka, probe port).

  bash scripts/fix-app-crashloop.sh --reset-infra --yes
    → Reset postgres/mongo/elastic/redis PVC, đồng bộ secrets, migrate, restart

  bash scripts/fix-app-crashloop.sh --sync-only
    → Chỉ cập nhật shop secrets từ .env.production (infra password đã khớp)
EOF
      exit 0
      ;;
    *) die "Unknown option: $1" ;;
  esac
done

command -v kubectl >/dev/null 2>&1 || die "kubectl not found"

verify_postgres() {
  local pass="$1"
  [[ -n "${pass}" ]] || return 1
  kubectl get pod postgres-0 -n infra >/dev/null 2>&1 || return 1
  kubectl exec postgres-0 -n infra -- env PGPASSWORD="${pass}" \
    psql -h 127.0.0.1 -U shopcaovanson -d shopcaovanson -tAc 'SELECT 1' 2>/dev/null | grep -q 1
}

verify_mongo() {
  local pass="$1"
  [[ -n "${pass}" ]] || return 1
  kubectl get pod mongodb-0 -n infra >/dev/null 2>&1 || return 1
  kubectl exec mongodb-0 -n infra -- mongosh \
    -u shopcaovanson -p "${pass}" --authenticationDatabase admin \
    --eval "db.adminCommand('ping')" --quiet >/dev/null 2>&1
}

verify_elastic() {
  local pass="$1"
  [[ -n "${pass}" ]] || return 1
  kubectl get pod elasticsearch-0 -n infra >/dev/null 2>&1 || return 1
  kubectl exec elasticsearch-0 -n infra -- \
    curl -sf -u "elastic:${pass}" "http://127.0.0.1:9200/_cluster/health" >/dev/null 2>&1
}

check_infra_auth() {
  local pg_pass mongo_pass elastic_pass
  pg_pass=$(kubectl get secret postgres-secret -n infra -o jsonpath='{.data.POSTGRES_PASSWORD}' 2>/dev/null | base64 -d || true)
  mongo_pass=$(kubectl get secret mongodb-secret -n infra -o jsonpath='{.data.MONGO_PASSWORD}' 2>/dev/null | base64 -d || true)
  elastic_pass=$(kubectl get secret elasticsearch-secret -n infra -o jsonpath='{.data.ELASTIC_PASSWORD}' 2>/dev/null | base64 -d || true)

  local ok=true
  if verify_postgres "${pg_pass}"; then
    log "✓ PostgreSQL password OK (infra secret)"
  else
    log "✗ PostgreSQL: password authentication failed (secret ≠ PVC)"
    ok=false
  fi
  if verify_mongo "${mongo_pass}"; then
    log "✓ MongoDB password OK"
  else
    log "✗ MongoDB: authentication failed (secret ≠ PVC)"
    ok=false
  fi
  if verify_elastic "${elastic_pass}"; then
    log "✓ Elasticsearch password OK"
  else
    log "✗ Elasticsearch: auth failed (secret ≠ PVC)"
    ok=false
  fi
  [[ "${ok}" == "true" ]]
}

reset_all_infra() {
  local args=()
  [[ "${YES}" == "true" ]] && args+=(--yes)
  bash "${SCRIPT_DIR}/reset-postgres.sh" "${args[@]}"
  bash "${SCRIPT_DIR}/reset-mongodb.sh" "${args[@]}"
  bash "${SCRIPT_DIR}/reset-elasticsearch.sh" "${args[@]}"
  bash "${SCRIPT_DIR}/reset-redis.sh" "${args[@]}"
}

sync_secrets() {
  [[ -f "${ENV_FILE}" ]] || die "Missing ${ENV_FILE} — chạy: bash scripts/create-secrets.sh"
  bash "${SCRIPT_DIR}/create-secrets.sh" --apply-only --force
  bash "${SCRIPT_DIR}/fix-redis-secrets.sh"
}

apply_config_patches() {
  log "Patching ConfigMaps (KAFKA_BOOTSTRAP_SERVERS)..."
  kubectl apply -f "${PROJECT_ROOT}/k8s/base/notification-service/configmap.yaml" 2>/dev/null || true
  kubectl apply -f "${PROJECT_ROOT}/k8s/base/search-service/configmap.yaml" 2>/dev/null || true
  kubectl apply -f "${PROJECT_ROOT}/k8s/base/chat-service/configmap.yaml" 2>/dev/null || true

  # Hotfix probe port nếu image cũ listen 8085/8086 thay vì 8000
  log "Patching Python service probe ports (8085/8086)..."
  kubectl patch deployment notification-service -n shop --type=json -p='[
    {"op":"replace","path":"/spec/template/spec/containers/0/livenessProbe/httpGet/port","value":8085},
    {"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/port","value":8085}
  ]' 2>/dev/null || true
  kubectl patch deployment search-service -n shop --type=json -p='[
    {"op":"replace","path":"/spec/template/spec/containers/0/livenessProbe/httpGet/port","value":8086},
    {"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/port","value":8086}
  ]' 2>/dev/null || true
  kubectl patch service notification-service -n shop --type=json \
    -p='[{"op":"replace","path":"/spec/ports/0/targetPort","value":8085}]' 2>/dev/null || true
  kubectl patch service search-service -n shop --type=json \
    -p='[{"op":"replace","path":"/spec/ports/0/targetPort","value":8086}]' 2>/dev/null || true
}

restart_shop() {
  log "Restarting shop deployments..."
  kubectl rollout restart deployment -n shop \
    api-gateway auth-service product-service order-service chat-service \
    notification-service search-service 2>/dev/null || true
}

cleanup_old_pods() {
  log "Deleting failed pods..."
  kubectl delete pod -n shop --field-selector=status.phase=Failed 2>/dev/null || true
  for dep in auth-service product-service order-service chat-service \
             notification-service search-service; do
    kubectl get pods -n shop -l "app=${dep}" --no-headers 2>/dev/null \
      | awk '$3 ~ /CrashLoop|Error|ImagePull/ {print $1}' \
      | xargs -r kubectl delete pod -n shop 2>/dev/null || true
  done
}

main() {
  log "Checking infra auth..."
  if check_infra_auth; then
    log "Infra passwords match secrets."
    if [[ "${RESET_INFRA}" == "true" ]]; then
      log "Infra OK — bỏ qua reset."
    fi
  else
    if [[ "${SYNC_ONLY}" == "true" ]]; then
      die "Infra password mismatch — không thể --sync-only. Dùng: bash scripts/fix-app-crashloop.sh --reset-infra --yes"
    fi
    if [[ "${RESET_INFRA}" != "true" ]]; then
      echo ""
      log "CẦN reset infra PVC để password khớp secrets."
      log "Chạy:"
      echo "  bash scripts/fix-app-crashloop.sh --reset-infra --yes"
      echo ""
      exit 1
    fi
    log "Resetting infra data stores..."
    reset_all_infra
    check_infra_auth || die "Infra vẫn lỗi sau reset — kiểm tra: kubectl get pods -n infra"
  fi

  if [[ "${RESET_INFRA}" == "true" || "${SYNC_ONLY}" == "true" || "${YES}" == "true" ]]; then
    sync_secrets
    apply_config_patches
    log "Running migrations..."
    bash "${SCRIPT_DIR}/migrate-all.sh"
    restart_shop
    cleanup_old_pods
    echo ""
    log "Done. Kiểm tra:"
    echo "  kubectl get pods -n shop"
    echo "  bash scripts/seed-data.sh"
    echo "  bash scripts/verify-app.sh"
  fi
}

main "$@"
