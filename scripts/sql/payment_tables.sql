-- 支付服务相关数据表
-- 创建支付订单表
CREATE TABLE payment_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_sn VARCHAR(64) NOT NULL UNIQUE COMMENT '支付单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    amount DECIMAL(10,2) NOT NULL COMMENT '支付金额',
    payment_method TINYINT NOT NULL COMMENT '支付方式: 1-微信支付, 2-支付宝, 3-银联支付, 4-网银支付, 5-余额支付',
    payment_status TINYINT NOT NULL DEFAULT 1 COMMENT '支付状态: 1-待支付, 2-支付成功, 3-支付失败, 4-已取消, 5-退款中, 6-已退款',
    third_party_sn VARCHAR(128) COMMENT '第三方支付单号(模拟)',
    paid_at TIMESTAMP NULL COMMENT '支付完成时间',
    expired_at TIMESTAMP NOT NULL COMMENT '支付过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_payment_sn (payment_sn),
    INDEX idx_order_sn (order_sn),
    INDEX idx_user_id (user_id),
    INDEX idx_status (payment_status),
    INDEX idx_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单表';

-- 创建支付记录表
CREATE TABLE payment_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_sn VARCHAR(64) NOT NULL COMMENT '支付单号',
    action VARCHAR(32) NOT NULL COMMENT '操作类型: create, pay_success, pay_fail, cancel, refund_start, refund_success',
    status_from TINYINT COMMENT '状态变更前',
    status_to TINYINT COMMENT '状态变更后',
    remark TEXT COMMENT '备注信息',
    operator_type ENUM('user', 'system', 'admin') NOT NULL DEFAULT 'system' COMMENT '操作类型',
    operator_id INT COMMENT '操作人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_payment_sn (payment_sn),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付操作日志表';

-- 创建库存预留记录表（用于分布式事务）
CREATE TABLE stock_reservations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    goods_id INT NOT NULL COMMENT '商品ID',
    reserved_num INT NOT NULL COMMENT '预留数量',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-已预留, 2-已确认, 3-已释放',
    reserved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '预留时间',
    confirmed_at TIMESTAMP NULL COMMENT '确认时间',
    released_at TIMESTAMP NULL COMMENT '释放时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_order_sn (order_sn),
    INDEX idx_goods_id (goods_id),
    INDEX idx_status (status),
    INDEX idx_reserved_at (reserved_at),
    
    UNIQUE KEY uk_order_goods (order_sn, goods_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预留记录表';

-- 扩展现有订单表，添加支付相关字段
-- 注意：这些字段应该添加到现有的 order_info 表中
-- ALTER TABLE order_info ADD COLUMN payment_status TINYINT DEFAULT 0 COMMENT '支付子状态: 0-无, 1-支付订单已创建, 2-支付中, 3-支付成功, 4-支付失败, 5-支付过期';
-- ALTER TABLE order_info ADD COLUMN payment_sn VARCHAR(64) COMMENT '支付单号';
-- ALTER TABLE order_info ADD COLUMN paid_at TIMESTAMP NULL COMMENT '支付时间';

-- ALTER TABLE order_info ADD INDEX idx_payment_sn (payment_sn);
-- ALTER TABLE order_info ADD INDEX idx_payment_status (payment_status);

-- 插入测试数据
INSERT INTO payment_orders (payment_sn, order_sn, user_id, amount, payment_method, payment_status, expired_at) VALUES
('PAY202508240001', 'ORD202508240001', 1, 99.99, 1, 1, DATE_ADD(NOW(), INTERVAL 15 MINUTE)),
('PAY202508240002', 'ORD202508240002', 2, 199.50, 2, 2, DATE_ADD(NOW(), INTERVAL -5 MINUTE));

-- 插入支付日志测试数据
INSERT INTO payment_logs (payment_sn, action, status_from, status_to, remark, operator_type) VALUES
('PAY202508240001', 'create', NULL, 1, '创建支付订单', 'system'),
('PAY202508240002', 'create', NULL, 1, '创建支付订单', 'system'),
('PAY202508240002', 'pay_success', 1, 2, '支付成功', 'system');