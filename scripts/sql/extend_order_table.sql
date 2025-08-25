-- 扩展订单表，添加物流相关字段
-- 执行命令: docker exec emshop-mysql mysql -u root -p123456 < extend_order_table.sql

USE emshop_order_srv;

-- 添加物流相关字段到订单表（如果不存在）
-- 使用存储过程方式处理字段存在性检查

DELIMITER $$
CREATE PROCEDURE AddColumnIfNotExists()
BEGIN
    DECLARE CONTINUE HANDLER FOR 1060 BEGIN END; -- 忽略字段已存在错误
    
    ALTER TABLE orderinfo ADD COLUMN logistics_status TINYINT DEFAULT 0 COMMENT '物流子状态：0-无，1-备货中，2-已发货，3-运输中，4-配送中，5-已送达，6-拒收';
    ALTER TABLE orderinfo ADD COLUMN logistics_sn VARCHAR(64) COMMENT '物流单号';  
    ALTER TABLE orderinfo ADD COLUMN tracking_number VARCHAR(64) COMMENT '快递单号';
    ALTER TABLE orderinfo ADD COLUMN shipped_at TIMESTAMP NULL COMMENT '发货时间';
    ALTER TABLE orderinfo ADD COLUMN delivered_at TIMESTAMP NULL COMMENT '送达时间';
    ALTER TABLE orderinfo ADD COLUMN completed_at TIMESTAMP NULL COMMENT '完成时间';
    ALTER TABLE orderinfo ADD COLUMN cancelled_at TIMESTAMP NULL COMMENT '取消时间';
    ALTER TABLE orderinfo ADD COLUMN cancel_reason TEXT COMMENT '取消原因';
    ALTER TABLE orderinfo ADD COLUMN auto_complete_at TIMESTAMP NULL COMMENT '自动完成时间';
END$$
DELIMITER ;

CALL AddColumnIfNotExists();
DROP PROCEDURE AddColumnIfNotExists;

-- 添加索引优化查询（忽略已存在错误）
DELIMITER $$
CREATE PROCEDURE AddIndexIfNotExists()
BEGIN
    DECLARE CONTINUE HANDLER FOR 1061 BEGIN END; -- 忽略索引已存在错误
    
    ALTER TABLE orderinfo ADD INDEX idx_logistics_sn (logistics_sn);
    ALTER TABLE orderinfo ADD INDEX idx_tracking_number (tracking_number);  
    ALTER TABLE orderinfo ADD INDEX idx_logistics_status (logistics_status);
    ALTER TABLE orderinfo ADD INDEX idx_shipped_at (shipped_at);
    ALTER TABLE orderinfo ADD INDEX idx_delivered_at (delivered_at);
END$$
DELIMITER ;

CALL AddIndexIfNotExists();
DROP PROCEDURE AddIndexIfNotExists;

-- 创建订单状态变更日志表
CREATE TABLE IF NOT EXISTS order_status_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    status_from VARCHAR(20) COMMENT '变更前状态',
    status_to VARCHAR(20) NOT NULL COMMENT '变更后状态',
    sub_status_from TINYINT COMMENT '变更前子状态',
    sub_status_to TINYINT COMMENT '变更后子状态',
    change_type ENUM('payment', 'logistics', 'manual', 'system') NOT NULL COMMENT '变更类型',
    operator_id INT COMMENT '操作员ID',
    operator_type ENUM('user', 'admin', 'system') NOT NULL COMMENT '操作员类型',
    remark TEXT COMMENT '变更说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_order_sn (order_sn),
    INDEX idx_created_at (created_at),
    INDEX idx_change_type (change_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态变更日志表';

-- 查看扩展后的表结构
SELECT 'Order table extended successfully!' AS message;
DESCRIBE orderinfo;