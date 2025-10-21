#!/usr/bin/env bash
set -euo pipefail

# 为 flash_sale_activities 增加 async_enabled 字段（1=异步，0=同步）

MYSQL_CONTAINER="${MYSQL_CONTAINER:-emshop-mysql}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-123456}"
MYSQL_DB="${MYSQL_DB:-emshop_coupon_srv}"

docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PWD" -D "$MYSQL_DB" -e "\
  ALTER TABLE flash_sale_activities\n\
    ADD COLUMN IF NOT EXISTS async_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用异步链路(1异步,0同步)';\
" >/dev/null && echo "[migrate] flash_sale_activities.async_enabled added"
