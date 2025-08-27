-- 优惠券服务相关数据表
-- 创建优惠券模板表
CREATE TABLE coupon_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '优惠券名称',
    type TINYINT NOT NULL COMMENT '优惠券类型: 1-满减券, 2-折扣券, 3-立减券, 4-包邮券',
    discount_type TINYINT NOT NULL COMMENT '折扣类型: 1-固定金额, 2-百分比',
    discount_value DECIMAL(10,2) NOT NULL COMMENT '折扣值(金额或百分比)',
    min_order_amount DECIMAL(10,2) DEFAULT 0.00 COMMENT '最小订单金额',
    max_discount_amount DECIMAL(10,2) DEFAULT 0.00 COMMENT '最大折扣金额(折扣券专用)',
    total_count INT NOT NULL DEFAULT 0 COMMENT '总发放数量',
    used_count INT NOT NULL DEFAULT 0 COMMENT '已使用数量',
    per_user_limit INT NOT NULL DEFAULT 1 COMMENT '每用户限领数量',
    valid_start_time TIMESTAMP NOT NULL COMMENT '有效期开始时间',
    valid_end_time TIMESTAMP NOT NULL COMMENT '有效期结束时间',
    valid_days INT DEFAULT 0 COMMENT '有效天数(从领取时间开始计算, 0表示使用固定时间)',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-活跃, 2-暂停, 3-结束',
    description TEXT COMMENT '使用说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_valid_time (valid_start_time, valid_end_time),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表';

-- 创建用户优惠券表
CREATE TABLE user_coupons (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    coupon_template_id BIGINT NOT NULL COMMENT '优惠券模板ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    coupon_code VARCHAR(32) NOT NULL UNIQUE COMMENT '优惠券码',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-未使用, 2-已使用, 3-已过期, 4-已冻结',
    order_sn VARCHAR(64) COMMENT '使用的订单号',
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
    used_at TIMESTAMP NULL COMMENT '使用时间',
    expired_at TIMESTAMP NOT NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_coupon_template_id (coupon_template_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_order_sn (order_sn),
    INDEX idx_expired_at (expired_at),
    INDEX idx_user_status (user_id, status),
    
    FOREIGN KEY (coupon_template_id) REFERENCES coupon_templates(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';

-- 创建优惠券使用记录表
CREATE TABLE coupon_usage_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_coupon_id BIGINT NOT NULL COMMENT '用户优惠券ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    original_amount DECIMAL(10,2) NOT NULL COMMENT '原始订单金额',
    discount_amount DECIMAL(10,2) NOT NULL COMMENT '优惠金额',
    final_amount DECIMAL(10,2) NOT NULL COMMENT '最终订单金额',
    action VARCHAR(32) NOT NULL COMMENT '操作类型: use, rollback',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_coupon_id (user_coupon_id),
    INDEX idx_user_id (user_id),
    INDEX idx_order_sn (order_sn),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (user_coupon_id) REFERENCES user_coupons(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券使用记录表';

-- 创建秒杀活动表
CREATE TABLE flash_sale_activities (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    coupon_template_id BIGINT NOT NULL COMMENT '关联的优惠券模板ID',
    name VARCHAR(100) NOT NULL COMMENT '秒杀活动名称',
    start_time TIMESTAMP NOT NULL COMMENT '秒杀开始时间',
    end_time TIMESTAMP NOT NULL COMMENT '秒杀结束时间',
    flash_sale_count INT NOT NULL COMMENT '秒杀数量',
    sold_count INT NOT NULL DEFAULT 0 COMMENT '已售数量',
    per_user_limit INT NOT NULL DEFAULT 1 COMMENT '每用户限抢数量',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-待开始, 2-进行中, 3-已结束, 4-已暂停',
    sort_order INT DEFAULT 0 COMMENT '排序权重',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_coupon_template_id (coupon_template_id),
    INDEX idx_status (status),
    INDEX idx_flash_sale_time (start_time, end_time),
    INDEX idx_sort_order (sort_order),
    
    FOREIGN KEY (coupon_template_id) REFERENCES coupon_templates(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';

-- 创建秒杀参与记录表
CREATE TABLE flash_sale_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    flash_sale_id BIGINT NOT NULL COMMENT '秒杀活动ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    user_coupon_id BIGINT COMMENT '生成的用户优惠券ID',
    status TINYINT NOT NULL COMMENT '状态: 1-成功, 2-失败, 3-超时',
    fail_reason VARCHAR(200) COMMENT '失败原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_flash_sale_id (flash_sale_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    
    UNIQUE KEY uk_flash_sale_user (flash_sale_id, user_id),
    FOREIGN KEY (flash_sale_id) REFERENCES flash_sale_activities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_coupon_id) REFERENCES user_coupons(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀参与记录表';

-- 创建优惠券配置表
CREATE TABLE coupon_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(64) NOT NULL UNIQUE COMMENT '配置键',
    config_value TEXT NOT NULL COMMENT '配置值(JSON格式)',
    description VARCHAR(255) COMMENT '配置说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券配置表';

-- 插入基础配置数据
INSERT INTO coupon_configs (config_key, config_value, description) VALUES
('max_stack_coupons', '3', '最大叠加优惠券数量'),
('flash_sale_timeout', '300', '秒杀超时时间(秒)'),
('coupon_code_prefix', 'CPN', '优惠券码前缀'),
('cache_ttl_seconds', '3600', '缓存TTL时间(秒)');

-- 插入测试优惠券模板数据
INSERT INTO coupon_templates (name, type, discount_type, discount_value, min_order_amount, total_count, per_user_limit, valid_start_time, valid_end_time, description) VALUES
('新用户专享20元优惠券', 1, 1, 20.00, 100.00, 1000, 1, '2024-01-01 00:00:00', '2024-12-31 23:59:59', '仅限新用户首次下单使用'),
('全场8折优惠券', 2, 2, 80.00, 200.00, 500, 2, '2024-08-01 00:00:00', '2024-08-31 23:59:59', '全场商品8折优惠，最高优惠50元'),
('立减10元券', 3, 1, 10.00, 50.00, 2000, 5, '2024-08-01 00:00:00', '2024-09-30 23:59:59', '无门槛立减10元');

-- 插入测试秒杀活动数据
INSERT INTO flash_sale_activities (coupon_template_id, name, start_time, end_time, flash_sale_count, per_user_limit, status) VALUES
(1, '新用户专享券秒杀', '2024-08-26 10:00:00', '2024-08-26 12:00:00', 100, 1, 1),
(3, '立减10元券限时抢', '2024-08-26 14:00:00', '2024-08-26 16:00:00', 500, 2, 1);

-- 插入测试用户优惠券数据
INSERT INTO user_coupons (coupon_template_id, user_id, coupon_code, expired_at) VALUES
(1, 1, 'CPN202408260001', '2024-12-31 23:59:59'),
(2, 1, 'CPN202408260002', '2024-08-31 23:59:59'),
(3, 2, 'CPN202408260003', '2024-09-30 23:59:59');