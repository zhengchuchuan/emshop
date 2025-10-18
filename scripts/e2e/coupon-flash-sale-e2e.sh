#!/usr/bin/env bash
set -euo pipefail

# End-to-end: Admin login -> ensure/create template -> create flash sale
# -> User login -> participate -> verify stock and record

# Requirements: curl, jq

ADMIN_BASE=${ADMIN_BASE:-"http://127.0.0.1:8050"}
API_BASE=${API_BASE:-"http://127.0.0.1:8051"}

# Admin credentials (override via env)
ADMIN_MOBILE=${ADMIN_MOBILE:-"13800138000"}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-"Admin@123"}

# User credentials (override via env)
USER_MOBILE=${USER_MOBILE:-"13800138000"}
USER_PASSWORD=${USER_PASSWORD:-"admin123"}

JSON=(-H "Content-Type: application/json" -H "Accept: application/json")

require() { command -v "$1" >/dev/null 2>&1 || { echo "[ERROR] $1 not found" >&2; exit 1; }; }
require curl
require jq

echo "[1/8] Admin captcha..."
ACAP=$(curl -sS "$ADMIN_BASE/v1/admin/auth/captcha")
ACID=$(echo "$ACAP" | jq -r .captchaId)
AANS=$(echo "$ACAP" | jq -r .answer)

echo "[2/8] Admin login..."
ALOGIN=$(curl -sS -X POST "$ADMIN_BASE/v1/admin/auth/pwd_login" \
  -H "Content-Type: application/json" -H "Accept: application/json" \
  -d "{\"mobile\":\"$ADMIN_MOBILE\",\"password\":\"$ADMIN_PASSWORD\",\"captcha\":\"$AANS\",\"captcha_id\":\"$ACID\"}")
ATOKEN=$(echo "$ALOGIN" | jq -r .data.token)
[ -n "$ATOKEN" ] || { echo "[ERROR] admin login failed: $ALOGIN" >&2; exit 1; }
AAUTH=(-H "Authorization: Bearer $ATOKEN")

echo "[3/8] Ensure default template..."
ENSURE=$(curl -sS -X POST "$ADMIN_BASE/v1/admin/coupons/templates/ensure-default" "${AAUTH[@]}" -H "Accept: application/json")
TPL_ID=$(echo "$ENSURE" | jq -r .data.id)
[ -n "$TPL_ID" ] || { echo "[ERROR] ensure default template failed: $ENSURE" >&2; exit 1; }
echo "Template ID: $TPL_ID"

echo "[4/8] Create flash sale activity..."
NOW=$(date +%s)
START=$((NOW-60))
END=$((NOW+3600))
CREATE_FS=$(jq -n \
  --argjson tpl "$TPL_ID" \
  --arg name "E2E 秒杀-$(date +%H%M%S)" \
  --argjson st "$START" --argjson et "$END" \
  '{coupon_template_id:$tpl|tonumber,name:$name,start_time:$st,end_time:$et,flash_sale_count:5,per_user_limit:1}' )
FS_RESP=$(echo "$CREATE_FS" | curl -sS -X POST "$ADMIN_BASE/v1/admin/coupons/flash-sale" "${AAUTH[@]}" -H "Content-Type: application/json" -d @-)
FS_ID=$(echo "$FS_RESP" | jq -r .data.id)
[ -n "$FS_ID" ] || { echo "[ERROR] create flash sale failed: $FS_RESP" >&2; exit 1; }
echo "FlashSale ID: $FS_ID"

echo "[5/8] User captcha..."
UCAP=$(curl -sS "$API_BASE/v1/base/captcha")
UCID=$(echo "$UCAP" | jq -r .captchaId)
UANS=$(echo "$UCAP" | jq -r .answer)

echo "[6/8] User login..."
ULOGIN=$(curl -sS -X POST "$API_BASE/v1/user/pwd_login" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "mobile=$USER_MOBILE&password=$USER_PASSWORD&captcha=$UANS&captchaId=$UCID")
UTOKEN=$(echo "$ULOGIN" | jq -r .data.token)
[ -n "$UTOKEN" ] || { echo "[ERROR] user login failed: $ULOGIN" >&2; exit 1; }
UAUTH=(-H "Authorization: Bearer $UTOKEN")

echo "[7/8] Participate flash sale..."
PART=$(curl -sS -X POST "$API_BASE/v1/coupons/flash-sale/$FS_ID/participate" "${UAUTH[@]}" -H "Accept: application/json")
echo "Participate response: $PART"

echo "[8/8] Verify stock and my record..."
STOCK=$(curl -sS "$API_BASE/v1/coupons/flash-sale/$FS_ID/stock" -H "Accept: application/json")
RECORD=$(curl -sS "$API_BASE/v1/coupons/flash-sale/$FS_ID/record" "${UAUTH[@]}" -H "Accept: application/json")

echo "Stock: $STOCK"
echo "Record: $RECORD"

echo "[OK] E2E finished."

