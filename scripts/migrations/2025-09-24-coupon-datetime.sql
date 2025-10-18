-- Coupon module: switch TIMESTAMP to DATETIME(0) (seconds precision)

-- 1) Coupon templates
ALTER TABLE coupon_templates
  MODIFY COLUMN valid_start_time DATETIME(0) NOT NULL COMMENT '有效期开始时间(秒)',
  MODIFY COLUMN valid_end_time   DATETIME(0) NOT NULL COMMENT '有效期结束时间(秒)';

-- 2) User coupons
ALTER TABLE user_coupons
  MODIFY COLUMN received_at DATETIME(0) NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间(秒)',
  MODIFY COLUMN used_at     DATETIME(0) NULL COMMENT '使用时间(秒)',
  MODIFY COLUMN expired_at  DATETIME(0) NOT NULL COMMENT '过期时间(秒)';

-- 3) Coupon usage logs
ALTER TABLE coupon_usage_logs
  MODIFY COLUMN created_at DATETIME(0) NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- 4) Flash sale activities
ALTER TABLE flash_sale_activities
  MODIFY COLUMN start_time DATETIME(0) NOT NULL COMMENT '秒杀开始时间(秒)',
  MODIFY COLUMN end_time   DATETIME(0) NOT NULL COMMENT '秒杀结束时间(秒)';

-- 5) Flash sale records
ALTER TABLE flash_sale_records
  MODIFY COLUMN created_at DATETIME(0) NOT NULL DEFAULT CURRENT_TIMESTAMP;

