-- 创建物流服务数据库和表结构
-- 执行命令: docker exec emshop-mysql mysql -u root -p123456 < create_logistics_database.sql

-- 创建物流服务数据库
CREATE DATABASE IF NOT EXISTS emshop_logistics_srv DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE emshop_logistics_srv;

-- 1. 物流订单表 (logistics_orders)
CREATE TABLE logistics_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    logistics_sn VARCHAR(64) NOT NULL UNIQUE COMMENT '物流单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    logistics_company TINYINT NOT NULL COMMENT '物流公司：1-圆通，2-申通，3-中通，4-韵达，5-顺丰，6-京东，7-邮政',
    shipping_method TINYINT NOT NULL COMMENT '配送方式：1-标准配送，2-急速配送，3-经济配送，4-自提',
    tracking_number VARCHAR(64) NOT NULL COMMENT '快递单号',
    logistics_status TINYINT NOT NULL DEFAULT 1 COMMENT '物流状态：1-待发货，2-已发货，3-运输中，4-配送中，5-已签收，6-拒收，7-退货中，8-已退货',
    
    -- 发货信息
    sender_name VARCHAR(64) NOT NULL COMMENT '发货人姓名',
    sender_phone VARCHAR(32) NOT NULL COMMENT '发货人电话',
    sender_address TEXT NOT NULL COMMENT '发货地址',
    
    -- 收货信息  
    receiver_name VARCHAR(64) NOT NULL COMMENT '收货人姓名',
    receiver_phone VARCHAR(32) NOT NULL COMMENT '收货人电话',
    receiver_address TEXT NOT NULL COMMENT '收货地址',
    
    -- 时间记录
    shipped_at TIMESTAMP NULL COMMENT '发货时间',
    delivered_at TIMESTAMP NULL COMMENT '签收时间', 
    estimated_delivery_at TIMESTAMP NULL COMMENT '预计送达时间',
    
    -- 费用信息
    shipping_fee DECIMAL(8,2) DEFAULT 0 COMMENT '运费',
    insurance_fee DECIMAL(8,2) DEFAULT 0 COMMENT '保价费',
    
    -- 商品信息(JSON格式存储)
    items_info TEXT COMMENT '商品信息JSON',
    
    remark TEXT COMMENT '备注信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_logistics_sn (logistics_sn),
    INDEX idx_order_sn (order_sn),
    INDEX idx_tracking_number (tracking_number),
    INDEX idx_user_id (user_id),
    INDEX idx_status (logistics_status),
    INDEX idx_company (logistics_company)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流订单表';

-- 2. 物流轨迹表 (logistics_tracks)  
CREATE TABLE logistics_tracks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    logistics_sn VARCHAR(64) NOT NULL COMMENT '物流单号',
    tracking_number VARCHAR(64) NOT NULL COMMENT '快递单号',
    location VARCHAR(128) NOT NULL COMMENT '当前位置',
    description TEXT NOT NULL COMMENT '轨迹描述',
    track_time TIMESTAMP NOT NULL COMMENT '轨迹时间',
    operator_name VARCHAR(64) COMMENT '操作员姓名',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    
    INDEX idx_logistics_sn (logistics_sn),
    INDEX idx_tracking_number (tracking_number),
    INDEX idx_track_time (track_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流轨迹表';

-- 3. 物流配送员表 (logistics_couriers)
CREATE TABLE logistics_couriers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    courier_code VARCHAR(32) NOT NULL UNIQUE COMMENT '配送员编号',
    courier_name VARCHAR(64) NOT NULL COMMENT '配送员姓名',
    phone VARCHAR(32) NOT NULL COMMENT '联系电话',
    logistics_company TINYINT NOT NULL COMMENT '所属物流公司',
    service_area VARCHAR(128) COMMENT '服务区域',
    status TINYINT DEFAULT 1 COMMENT '状态：1-在职，0-离职',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_courier_code (courier_code),
    INDEX idx_company (logistics_company),
    INDEX idx_area (service_area)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流配送员表';

-- 4. 初始化配送员数据
INSERT INTO logistics_couriers (courier_code, courier_name, phone, logistics_company, service_area) VALUES
('YT001', '张小明', '13800138001', 1, '北京市朝阳区'),
('YT002', '李小红', '13800138002', 1, '北京市海淀区'),
('ST001', '王小强', '13800138003', 2, '上海市浦东新区'),
('ST002', '赵小芳', '13800138004', 2, '上海市徐汇区'),
('ZT001', '陈小军', '13800138005', 3, '广州市天河区'),
('ZT002', '刘小丽', '13800138006', 3, '深圳市南山区'),
('YD001', '杨小峰', '13800138007', 4, '杭州市西湖区'),
('SF001', '周小华', '13800138008', 5, '成都市锦江区'),
('JD001', '徐小东', '13800138009', 6, '武汉市洪山区'),
('EMS001', '孙小梅', '13800138010', 7, '南京市鼓楼区');

-- 创建成功提示
SELECT 'Logistics database and tables created successfully!' AS message;