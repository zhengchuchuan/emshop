#!/usr/bin/env bash
set -euo pipefail

# 速率优先模式：constant-arrival-rate。直接通过 RATE 固定每秒请求数。
# 可通过环境变量覆盖：TARGET、FLASH_SALE_ID、RATE、PRE_VUS、MAX_VUS、DURATION、TIMEOUT、TUNE

TARGET=${TARGET:-127.0.0.1:28056}
FLASH_SALE_ID=${FLASH_SALE_ID:-1}
RATE=${RATE:-4000}         # 固定 RPS
PRE_VUS=${PRE_VUS:-3000}    # 预分配 VUs
MAX_VUS=${MAX_VUS:-3000}   # 最大 VUs（需足够大以支撑 RATE）
DURATION=${DURATION:-10s}
TIMEOUT=${TIMEOUT:-15s}
TUNE=${TUNE:-1}

if [[ "$TUNE" == "1" ]]; then
  # 基础调优：若无 sudo 或已是 root，则直接 sysctl
  run_sysctl() {
    local k="$1"; shift; local v="$*"
    if command -v sudo >/dev/null 2>&1; then
      sudo sysctl -w "$k=$v" || true
    else
      sysctl -w "$k=$v" || true
    fi
  }
  # run_sysctl net.ipv4.ip_local_port_range "40000 65535"
  run_sysctl net.ipv4.tcp_tw_reuse 1
  run_sysctl net.ipv4.tcp_fin_timeout 15
  # run_sysctl net.core.somaxconn 65535
  # run_sysctl net.core.netdev_max_backlog 262144
  # run_sysctl net.ipv4.tcp_max_syn_backlog 262144
fi

# 提高文件描述符上限（对本进程及子进程生效）
ulimit -n 200000 || true

mkdir -p ./logs/k6

echo "Running k6 RATE mode: target=$TARGET rate=$RATE pre_vus=$PRE_VUS max_vus=$MAX_VUS duration=$DURATION"

k6 run \
  -e GRPC_COUPON_TARGET="$TARGET" \
  -e FLASH_SALE_ID="$FLASH_SALE_ID" \
  -e FLASH_PARTICIPATE_RPS="$RATE" \
  -e PRE_VUS="$PRE_VUS" \
  -e MAX_VUS="$MAX_VUS" \
  -e GRPC_TIMEOUT="$TIMEOUT" \
  -e DURATION="$DURATION" \
  -e DISABLE_THRESHOLDS=1 \
  -e SUMMARY_JSON=./logs/k6/k6-summary.json \
  -e SUMMARY_HTML=./logs/k6/k6-summary.html \
  scripts/k6/grpc_coupon_flash_sale.js
