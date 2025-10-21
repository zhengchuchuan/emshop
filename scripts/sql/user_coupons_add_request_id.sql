ALTER TABLE user_coupons 
  ADD COLUMN IF NOT EXISTS request_id VARCHAR(64) DEFAULT '' COMMENT '请求幂等ID',
  ADD INDEX IF NOT EXISTS idx_request_id (request_id),
  ADD UNIQUE KEY IF NOT EXISTS uk_usercoupon_req (request_id);

