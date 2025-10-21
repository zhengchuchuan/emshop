#!/usr/bin/env bash
set -euo pipefail

# 一键压测启动器
# 功能：
# - 可选：更新 DB 活动时间/库存/限购
# - 可选：清理本活动相关的 Redis 键
# - 必选：调用管理接口启动活动（预热 coupon:* 键并置为进行中）
# - 运行 k6 脚本（支持单机多进程，自动分片 USER_ID_BASE）
#
# 依赖：k6、curl；可选：mysql、redis-cli
#
# 用法示例（单进程直连 gRPC，RPS=800）：
#   ACT_ID=1 TARGET=127.0.0.1:28056 \
#   UPDATE_ACTIVITY=1 FS_COUNT=100000 PER_USER_LIMIT=0 \
#   CLEAR_REDIS=1 HTTP_PORT=8056 \
#   RPS=800 DURATION=120s PRE_VUS=60 MAX_VUS=800 USER_ID_MODE=PER_ITER \
#   scripts/k6/launcher.sh
#
# 多进程分片（共4个进程，总RPS=20000，每进程5000）：
#   ACT_ID=1 TARGET=127.0.0.1:28056 INSTANCES=4 TOTAL_RPS=20000 \
#   DURATION=300s PRE_VUS=1200 MAX_VUS=3000 USER_ID_MODE=PER_ITER \
#   scripts/k6/launcher.sh

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
K6_SCRIPT="$ROOT_DIR/scripts/k6/grpc_coupon.js"

# 基本参数
ACT_ID="${ACT_ID:-}"
TARGET="${TARGET:-}"             # 直连 gRPC，如 127.0.0.1:28056；若留空则需设置 Consul 环境变量
DURATION="${DURATION:-120s}"
RPS="${RPS:-800}"
PRE_VUS="${PRE_VUS:-60}"
MAX_VUS="${MAX_VUS:-800}"
USER_ID_MODE="${USER_ID_MODE:-PER_ITER}"
USER_ID_BASE="${USER_ID_BASE:-100000}"
GRPC_TIMEOUT="${GRPC_TIMEOUT:-3s}"
ENABLE_STOCK="${ENABLE_STOCK:-0}"
ENABLE_CALC="${ENABLE_CALC:-0}"
SLEEP_INTERVAL="${SLEEP:-0}"

# 多进程分片
INSTANCES="${INSTANCES:-1}"
TOTAL_RPS="${TOTAL_RPS:-}"

# 管理接口与清理
HTTP_PORT="${HTTP_PORT:-8056}"
CLEAR_REDIS="${CLEAR_REDIS:-0}"
FULL_REDIS_FLUSH="${FULL_REDIS_FLUSH:-0}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-16379}"
REDIS_DB="${REDIS_DB:-0}"

# DB 更新（可选）
UPDATE_ACTIVITY="${UPDATE_ACTIVITY:-0}"
FS_COUNT="${FS_COUNT:-100000}"
PER_USER_LIMIT="${PER_USER_LIMIT:-0}"
MYSQL_HOST="${MYSQL_HOST:-localhost}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-root}"
MYSQL_DB="${MYSQL_DB:-emshop_coupon_srv}"

function usage() {
  cat <<EOF
一键压测启动器

必填环境变量：
  ACT_ID               活动ID

常用环境变量：
  TARGET               gRPC直连地址，如 127.0.0.1:28056（推荐直连）
  DURATION             压测时长，默认120s
  RPS                  单进程RPS（当 INSTANCES=1 时生效），默认800
  PRE_VUS              预分配VUs，默认60
  MAX_VUS              最大VUs，默认800
  USER_ID_MODE         用户ID模式：PER_ITER|PER_VU|CONST，默认PER_ITER
  USER_ID_BASE         用户ID起始基数，默认100000
  GRPC_TIMEOUT         gRPC请求超时，默认3s
  ENABLE_STOCK         是否压库存查询场景(0/1)，默认0
  ENABLE_CALC          是否压优惠计算场景(0/1)，默认0
  SLEEP                每次迭代sleep秒，默认0

多进程分片：
  INSTANCES            k6 进程数，默认1
  TOTAL_RPS            总RPS（当INSTANCES>1时使用，总RPS会均分至各进程）

管理与清理：
  HTTP_PORT            管理HTTP端口，默认8056
  CLEAR_REDIS          压测前清理本活动相关Redis键(0/1)，默认0
  REDIS_HOST/PORT/DB   Redis连接信息，默认127.0.0.1:16379/0

数据库更新（可选）：
  UPDATE_ACTIVITY      更新DB活动窗口、库存与限购(0/1)，默认0
  FS_COUNT             更新库存为该值，默认100000
  PER_USER_LIMIT       限购数量，默认0(不限购)
  MYSQL_HOST/PORT/USER/PWD/DB  MySQL连接信息

示例：见脚本顶部注释
EOF
}

if [[ -z "${ACT_ID}" ]]; then
  echo "[ERR] 需要设置 ACT_ID" >&2
  usage
  exit 1
fi

if [[ ! -f "$K6_SCRIPT" ]]; then
  echo "[ERR] 未找到 k6 脚本: $K6_SCRIPT" >&2
  exit 1
fi

if ! command -v k6 >/dev/null 2>&1; then
  echo "[ERR] 未找到 k6，请先安装 k6" >&2
  exit 1
fi

function mysql_exec() {
  if ! command -v mysql >/dev/null 2>&1; then
    echo "[WARN] 未安装 mysql 客户端，跳过 DB 更新" >&2
    return 0
  fi
  MYSQL_PWD="$MYSQL_PWD" mysql -N -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER" "$MYSQL_DB" -e "$1"
}

function redis_cmd() {
  if ! command -v redis-cli >/dev/null 2>&1; then
    echo "[WARN] 未安装 redis-cli，跳过 Redis 操作: $*" >&2
    return 0
  fi
  redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -n "$REDIS_DB" "$@"
}

echo "[1/5] 可选：更新DB活动时间/库存/限购 UPDATE_ACTIVITY=$UPDATE_ACTIVITY"
if [[ "$UPDATE_ACTIVITY" == "1" ]]; then
  now_sql="NOW()"
  sql="UPDATE flash_sale_activities \n\
       SET start_time = $now_sql - INTERVAL 10 MINUTE, \n\
           end_time   = $now_sql + INTERVAL 1 DAY, \n\
           flash_sale_count = $FS_COUNT, \n\
           per_user_limit   = $PER_USER_LIMIT \n\
       WHERE id = $ACT_ID;"
  mysql_exec "$sql" || true
fi

echo "[2/5] 可选：清理Redis CLEAR_REDIS=$CLEAR_REDIS FULL_REDIS_FLUSH=$FULL_REDIS_FLUSH"
if [[ "$FULL_REDIS_FLUSH" == "1" ]]; then
  echo "[WARN] 将执行 FLUSHDB ASYNC（db=$REDIS_DB），请谨慎！" >&2
  redis_cmd FLUSHDB ASYNC >/dev/null || true
fi
if [[ "$CLEAR_REDIS" == "1" ]]; then
  # 清理活动信息与日志
  redis_cmd DEL "coupon:activity:$ACT_ID" >/dev/null || true
  redis_cmd DEL "coupon:log:$ACT_ID" >/dev/null || true
  # 清理用户参与记录（扫描避免阻塞）
  if command -v redis-cli >/dev/null 2>&1; then
    # shellcheck disable=SC2046
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -n "$REDIS_DB" --scan --pattern "coupon:user:$ACT_ID:*" | xargs -r -n 1000 redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -n "$REDIS_DB" DEL >/dev/null || true
  fi
  # 尝试清理库存键：优先从DB查 coupon_id；失败则跳过
  COUPON_ID=""
  if command -v mysql >/dev/null 2>&1; then
    COUPON_ID="$(MYSQL_PWD="$MYSQL_PWD" mysql -N -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER" "$MYSQL_DB" -e "SELECT coupon_template_id FROM flash_sale_activities WHERE id=$ACT_ID;" | tr -d '\n' || true)"
  fi
  if [[ -n "$COUPON_ID" ]]; then
    redis_cmd DEL "coupon:stock:$COUPON_ID" >/dev/null || true
  else
    echo "[WARN] 未能确定 coupon_id，跳过 coupon:stock:* 清理" >&2
  fi
fi

echo "[3/5] 启动/预热活动 via 管理接口: http://127.0.0.1:$HTTP_PORT/manage/flashsale/$ACT_ID/start"
if ! command -v curl >/dev/null 2>&1; then
  echo "[ERR] 需要 curl 以调用管理接口" >&2
  exit 1
fi
curl -fsS -X POST "http://127.0.0.1:$HTTP_PORT/manage/flashsale/$ACT_ID/start" -H 'Content-Type: application/json' -d '{}' >/dev/null || {
  echo "[WARN] 启动活动失败，可能服务未启动或管理端口不同；继续尝试运行 k6" >&2
}

echo "[4/5] 计算并发配置"
CMD_BASE=(
  k6 run \
    -e FLASH_SALE_ID="$ACT_ID" \
    -e USER_ID_MODE="$USER_ID_MODE" \
    -e USER_ID_BASE="$USER_ID_BASE" \
    -e GRPC_TIMEOUT="$GRPC_TIMEOUT" \
    -e ENABLE_FLASH_STOCK="$ENABLE_STOCK" \
    -e ENABLE_COUPON_CALC="$ENABLE_CALC" \
    -e SLEEP="$SLEEP_INTERVAL" \
    -e DURATION="$DURATION"
)

if [[ -n "${TARGET}" ]]; then
  CMD_BASE+=( -e GRPC_COUPON_TARGET="$TARGET" )
else
  # 若未提供TARGET，则使用 Consul 解析（需外部设置 CONSUL_HTTP_ADDR/CONSUL_SERVICE）
  :
fi

function run_one_instance() {
  local idx="$1" rps="$2" uid_base_offset="$3"
  local pre_vus="$4" max_vus="$5"
  local user_base=$(( USER_ID_BASE + uid_base_offset ))
  echo "[k6] instance=$idx rps=$rps pre_vus=$pre_vus max_vus=$max_vus USER_ID_BASE=$user_base"
  "${CMD_BASE[@]}" \
    -e FLASH_PARTICIPATE_RPS="$rps" \
    -e PRE_VUS="$pre_vus" \
    -e MAX_VUS="$max_vus" \
    -e USER_ID_BASE="$user_base" \
    "$K6_SCRIPT"
}

echo "[5/5] 启动 k6"
if [[ "$INSTANCES" -le 1 ]]; then
  run_one_instance 0 "$RPS" 0 "$PRE_VUS" "$MAX_VUS"
else
  if [[ -z "${TOTAL_RPS}" ]]; then
    echo "[ERR] INSTANCES>1 时需要设置 TOTAL_RPS" >&2
    exit 1
  fi
  per_rps=$(( TOTAL_RPS / INSTANCES ))
  pids=()
  for (( i=0; i<INSTANCES; i++ )); do
    # 为避免 userId 冲突，每个实例的基数相隔 1e9
    uid_offset=$(( i * 1000000000 ))
    run_one_instance "$i" "$per_rps" "$uid_offset" "$PRE_VUS" "$MAX_VUS" &
    pids+=("$!")
    sleep 1
  done
  # 等待所有子进程
  code=0
  for pid in "${pids[@]}"; do
    if ! wait "$pid"; then
      code=1
    fi
  done
  exit "$code"
fi
