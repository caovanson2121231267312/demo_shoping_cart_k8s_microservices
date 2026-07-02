#!/usr/bin/env bash
# Kiểm tra DNS shopcaovanson.xyz vs subdomain monitor
set -euo pipefail

MAIN="${MAIN_DOMAIN:-shopcaovanson.xyz}"
MONITOR="${MONITOR_HOST:-monitor.shopcaovanson.xyz}"
VPS_IP="${VPS_IP:-110.172.29.72}"

echo "=== Kiểm tra DNS ==="
echo ""

check() {
  local host="$1"
  local g cf local
  g=$(dig +short "${host}" @8.8.8.8 2>/dev/null | tail -1)
  cf=$(dig +short "${host}" @1.1.1.1 2>/dev/null | tail -1)
  local=$(dig +short "${host}" 2>/dev/null | tail -1)
  echo "${host}"
  echo "  Google 8.8.8.8:     ${g:-<KHÔNG CÓ — NXDOMAIN hoặc chưa tạo>}"
  echo "  Cloudflare 1.1.1.1: ${cf:-<KHÔNG CÓ>}"
  echo "  Local:              ${local:-<KHÔNG CÓ>}"
  if [[ "${g}" == "${VPS_IP}" || "${cf}" == "${VPS_IP}" ]]; then
    echo "  → OK"
  else
    echo "  → Cần A record → ${VPS_IP}"
  fi
  echo ""
}

echo "Nameserver của ${MAIN}:"
dig +short NS "${MAIN}" @8.8.8.8 | sed 's/^/  /'
echo ""

check "${MAIN}"
check "www.${MAIN}"
check "${MONITOR}"

echo "=== Lưu ý panel DNS ==="
echo "  Chỉ cần 1 bản ghi:  Name = monitor  (KHÔNG tạo thêm monitor.shopcaovanson.xyz)"
echo "  Nhiều panel tự thêm domain → bản ghi 'monitor.shopcaovanson.xyz' tạo hostname SAI."
echo ""

echo "=== Kết luận ==="
if dig +short "${MONITOR}" @8.8.8.8 | grep -q "${VPS_IP}"; then
  echo "DNS monitor OK — chạy: bash scripts/fix-monitoring-access.sh --renew-cert"
else
  echo "DNS monitor CHƯA TỒN TẠI trên internet."
  echo ""
  echo "Bạn phải thêm bản ghi tại nơi quản lý DNS của ${MAIN}"
  echo "(cùng panel đã tạo bản ghi A cho shopcaovanson.xyz → ${VPS_IP}):"
  echo ""
  echo "  Type:  A"
  echo "  Name:  monitor          (chỉ 'monitor', không gõ full domain)"
  echo "  Value: ${VPS_IP}"
  echo "  TTL:   300"
  echo ""
  echo "Sau 5–30 phút: dig +short ${MONITOR} @8.8.8.8"
  echo ""
  echo "Truy cập tạm (không cần DNS public):"
  echo "  bash scripts/grafana-tls-selfsigned.sh"
  echo "  + thêm hosts file trên máy: ${VPS_IP} ${MONITOR}"
fi
