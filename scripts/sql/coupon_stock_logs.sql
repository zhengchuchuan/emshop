-- 库存扣减日志表（Redis 扣减日志异步落库）
CREATE TABLE IF NOT EXISTS flash_sale_stock_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    activity_id BIGINT NOT NULL COMMENT '秒杀活动ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    decr INT NOT NULL DEFAULT 1 COMMENT '扣减数量',
    ts BIGINT NOT NULL COMMENT 'Lua 扣减发生的时间戳(秒)',
    request_id VARCHAR(64) DEFAULT '' COMMENT '请求ID(幂等)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_activity_user (activity_id, user_id),
    INDEX idx_ts (ts),
    UNIQUE KEY uk_request (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀库存扣减日志表';
