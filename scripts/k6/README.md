一键压测启动器使用说明（scripts/k6/launcher.sh）

前置条件
- 已启动 emshop-coupon-srv（configs/coupon/srv.yaml 默认 gRPC:28056，HTTP:8056）
- Redis/（可选）MySQL/RocketMQ 已按需就绪
- 已安装 k6、curl；可选安装 mysql、redis-cli 以支持自动更新/清理

常用示例
1) 单进程直连 gRPC，先更新DB活动并清理Redis，再预热并压测：

   ACT_ID=1 TARGET=127.0.0.1:28056 \
   UPDATE_ACTIVITY=1 FS_COUNT=100000 PER_USER_LIMIT=0 \
   CLEAR_REDIS=1 HTTP_PORT=8056 \
   RPS=800 DURATION=120s PRE_VUS=60 MAX_VUS=800 USER_ID_MODE=PER_ITER \
   scripts/k6/launcher.sh

2) 多进程分片（合计 20k RPS，4 个进程，每进程 5k）：

   ACT_ID=1 TARGET=127.0.0.1:28056 \
   INSTANCES=4 TOTAL_RPS=20000 \
   DURATION=300s PRE_VUS=1200 MAX_VUS=3000 USER_ID_MODE=PER_ITER \
   scripts/k6/launcher.sh

关键环境变量
- ACT_ID：活动ID（必填）
- TARGET：gRPC 直连地址（优先使用直连；若留空将使用 Consul 环境变量解析）
- UPDATE_ACTIVITY：是否更新 DB 活动窗口/库存/限购（默认0）
- CLEAR_REDIS：是否清理活动相关 Redis（默认0）
- RPS/PRE_VUS/MAX_VUS/DURATION/USER_ID_MODE/USER_ID_BASE：k6 场景参数
- INSTANCES/TOTAL_RPS：多进程分片参数

管理接口（用于活动预热/停止）
- POST /manage/flashsale/:id/start
- POST /manage/flashsale/:id/stop

备注
- 若使用异步链路压测，请确保 RocketMQ 正常；否则 Redis 成功也不会入库。
- 不限购压测请先将 flash_sale_activities.per_user_limit 设置为 0（UPDATE_ACTIVITY=1 + PER_USER_LIMIT=0）。
- 清理 Redis 时仅清理本活动相关键；也可设置 FULL_REDIS_FLUSH=1 执行 FLUSHDB ASYNC（谨慎）。库存键需要能查询到 coupon_template_id 才会清理。
