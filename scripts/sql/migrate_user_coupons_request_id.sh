#!/usr/bin/env bash
set -euo pipefail

# 为 user_coupons 表添加 request_id 字段及索引（适配同步/异步秒杀幂等）
# 注意：默认通过 docker exec 进入 MySQL 容器执行；请根据实际容器名/密码调整环境变量。

MYSQL_CONTAINER="${MYSQL_CONTAINER:-emshop-mysql}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-123456}"
MYSQL_DB="${MYSQL_DB:-emshop_coupon_srv}"

echo "[migrate] container=$MYSQL_CONTAINER db=$MYSQL_DB user=$MYSQL_USER"

exec_mysql() {
  docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PWD" -D "$MYSQL_DB" -e "$1" >/dev/null
}

# 1) 加列（若不存在）
exec_mysql "ALTER TABLE user_coupons ADD COLUMN IF NOT EXISTS request_id VARCHAR(64) DEFAULT '' COMMENT '请求幂等ID';" || true

# 2) 索引（幂等处理：若已存在则先删除再创建，避免语法不支持 IF NOT EXISTS 的情况）
exec_mysql "DROP INDEX IF EXISTS idx_request_id ON user_coupons;" || true
exec_mysql "CREATE INDEX idx_request_id ON user_coupons(request_id);" || true

exec_mysql "DROP INDEX IF EXISTS uk_usercoupon_req ON user_coupons;" || true
exec_mysql "CREATE UNIQUE INDEX uk_usercoupon_req ON user_coupons(request_id);" || true

echo "[migrate] done"
