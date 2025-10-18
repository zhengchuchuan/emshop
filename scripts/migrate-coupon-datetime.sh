#!/usr/bin/env bash
set -euo pipefail

# Migrate coupon-related TIMESTAMP columns to DATETIME(0) (seconds precision)
# Usage (env):
#   MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_USER=root MYSQL_PASS=root MYSQL_DB=emshop_coupon_srv \
#   bash scripts/migrate-coupon-datetime.sh

MYSQL_HOST=${MYSQL_HOST:-127.0.0.1}
MYSQL_PORT=${MYSQL_PORT:-3306}
MYSQL_USER=${MYSQL_USER:-root}
MYSQL_PASS=${MYSQL_PASS:-root}
MYSQL_DB=${MYSQL_DB:-emshop_coupon_srv}

SQL_FILE="scripts/migrations/2025-09-24-coupon-datetime.sql"

if ! command -v mysql >/dev/null 2>&1; then
  echo "[ERROR] mysql client not found in PATH" >&2
  exit 1
fi

if [ ! -f "$SQL_FILE" ]; then
  echo "[ERROR] SQL file not found: $SQL_FILE" >&2
  exit 1
fi

echo "[INFO] Applying migration to MySQL: ${MYSQL_USER}@${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB}"
export MYSQL_PWD="$MYSQL_PASS"

# Print a short preview of SQL (first 10 lines)
echo "[INFO] SQL preview:"; head -n 10 "$SQL_FILE" | sed 's/^/  /'

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" "$MYSQL_DB" < "$SQL_FILE"

echo "[OK] Migration completed. Verify with: SHOW CREATE TABLE coupon_templates;"

