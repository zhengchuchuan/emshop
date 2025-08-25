# 数据库设计文档

## 1. 概述

本文档定义了电商系统支付服务和物流服务的完整数据库设计，包括表结构、索引、约束和数据迁移方案。

## 2. 数据库架构

### 2.1 数据库分库策略

```
emshop (主库)
├── order_db      # 订单相关表
├── payment_db    # 支付相关表
├── logistics_db  # 物流相关表
├── user_db       # 用户相关表
├── goods_db      # 商品相关表
└── system_db     # 系统配置表
```

### 2.2 分表策略

- 按用户ID分表：`payment_orders_0` ~ `payment_orders_99`
- 按时间分表：`logistics_tracks_202501` ~ `logistics_tracks_202512`
- 按订单号分表：`order_status_logs_0` ~ `order_status_logs_15`

## 3. 支付服务数据表设计

### 3.1 支付订单表 (payment_orders)

```sql
CREATE TABLE payment_orders (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    payment_sn VARCHAR(64) NOT NULL COMMENT '支付单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    amount DECIMAL(10,2) NOT NULL COMMENT '支付金额',
    payment_method TINYINT NOT NULL COMMENT '支付方式：1-微信，2-支付宝，3-银联，4-网银，5-余额',
    payment_status TINYINT NOT NULL DEFAULT 1 COMMENT '支付状态：1-待支付，2-支付成功，3-支付失败，4-已取消，5-退款中，6-已退款',
    third_party_sn VARCHAR(128) COMMENT '第三方支付单号',
    third_party_data TEXT COMMENT '第三方支付数据（JSON格式）',
    
    -- 金额相关
    original_amount DECIMAL(10,2) NOT NULL COMMENT '原始金额',
    discount_amount DECIMAL(10,2) DEFAULT 0 COMMENT '优惠金额',
    actual_amount DECIMAL(10,2) NOT NULL COMMENT '实际支付金额',
    
    -- 时间相关
    paid_at TIMESTAMP NULL COMMENT '支付完成时间',
    expired_at TIMESTAMP NOT NULL COMMENT '支付过期时间',
    
    -- 回调信息
    return_url VARCHAR(512) COMMENT '支付完成跳转URL',
    notify_url VARCHAR(512) COMMENT '支付结果通知URL',
    notify_status TINYINT DEFAULT 0 COMMENT '通知状态：0-未通知，1-已通知，2-通知失败',
    notify_times INT DEFAULT 0 COMMENT '通知次数',
    notify_at TIMESTAMP NULL COMMENT '最后通知时间',
    
    -- 其他信息
    client_ip VARCHAR(45) COMMENT '客户端IP',
    user_agent TEXT COMMENT '用户代理',
    remark TEXT COMMENT '备注信息',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_payment_sn (payment_sn),
    KEY idx_order_sn (order_sn),
    KEY idx_user_id (user_id),
    KEY idx_payment_status (payment_status),
    KEY idx_payment_method (payment_method),
    KEY idx_third_party_sn (third_party_sn),
    KEY idx_created_at (created_at),
    KEY idx_paid_at (paid_at),
    KEY idx_expired_at (expired_at),
    KEY idx_notify_status (notify_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单表';
```

### 3.2 退款订单表 (refund_orders)

```sql
CREATE TABLE refund_orders (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    refund_sn VARCHAR(64) NOT NULL COMMENT '退款单号',
    payment_sn VARCHAR(64) NOT NULL COMMENT '支付单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    
    -- 金额相关
    payment_amount DECIMAL(10,2) NOT NULL COMMENT '原支付金额',
    refund_amount DECIMAL(10,2) NOT NULL COMMENT '退款金额',
    refund_fee DECIMAL(10,2) DEFAULT 0 COMMENT '退款手续费',
    actual_refund_amount DECIMAL(10,2) NOT NULL COMMENT '实际退款金额',
    
    -- 状态相关
    refund_status TINYINT NOT NULL DEFAULT 1 COMMENT '退款状态：1-申请中，2-处理中，3-退款成功，4-退款失败，5-已取消',
    refund_type TINYINT NOT NULL COMMENT '退款类型：1-全额退款，2-部分退款',
    refund_reason VARCHAR(256) NOT NULL COMMENT '退款原因',
    
    -- 第三方信息
    third_party_refund_sn VARCHAR(128) COMMENT '第三方退款单号',
    third_party_data TEXT COMMENT '第三方退款数据（JSON格式）',
    
    -- 时间相关
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    processed_at TIMESTAMP NULL COMMENT '处理时间',
    refunded_at TIMESTAMP NULL COMMENT '退款完成时间',
    expected_at TIMESTAMP NULL COMMENT '预计到账时间',
    
    -- 操作相关
    operator_id INT COMMENT '操作员ID',
    operator_type ENUM('user', 'admin', 'system') NOT NULL COMMENT '操作员类型',
    operator_name VARCHAR(64) COMMENT '操作员姓名',
    
    -- 审计字段
    remark TEXT COMMENT '备注信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_refund_sn (refund_sn),
    KEY idx_payment_sn (payment_sn),
    KEY idx_order_sn (order_sn),
    KEY idx_user_id (user_id),
    KEY idx_refund_status (refund_status),
    KEY idx_refund_type (refund_type),
    KEY idx_applied_at (applied_at),
    KEY idx_refunded_at (refunded_at),
    KEY idx_operator_id (operator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退款订单表';
```

### 3.3 支付日志表 (payment_logs)

```sql
CREATE TABLE payment_logs (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    payment_sn VARCHAR(64) NOT NULL COMMENT '支付单号',
    log_type ENUM('create', 'pay', 'cancel', 'refund', 'notify', 'callback') NOT NULL COMMENT '日志类型',
    action VARCHAR(64) NOT NULL COMMENT '操作类型',
    status_from TINYINT COMMENT '状态变更前',
    status_to TINYINT COMMENT '状态变更后',
    
    -- 操作相关
    operator_id INT COMMENT '操作员ID',
    operator_type ENUM('user', 'admin', 'system', 'third_party') NOT NULL COMMENT '操作员类型',
    operator_name VARCHAR(64) COMMENT '操作员姓名',
    
    -- 请求响应数据
    request_data TEXT COMMENT '请求数据（JSON格式）',
    response_data TEXT COMMENT '响应数据（JSON格式）',
    error_code VARCHAR(32) COMMENT '错误代码',
    error_message TEXT COMMENT '错误信息',
    
    -- 其他信息
    client_ip VARCHAR(45) COMMENT '客户端IP',
    user_agent TEXT COMMENT '用户代理',
    execution_time INT COMMENT '执行时间（毫秒）',
    remark TEXT COMMENT '备注信息',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    KEY idx_payment_sn (payment_sn),
    KEY idx_log_type (log_type),
    KEY idx_action (action),
    KEY idx_operator_id (operator_id),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付日志表';
```

## 4. 物流服务数据表设计

### 4.1 物流订单表 (logistics_orders)

```sql
CREATE TABLE logistics_orders (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    logistics_sn VARCHAR(64) NOT NULL COMMENT '物流单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    
    -- 物流信息
    logistics_company TINYINT NOT NULL COMMENT '物流公司：1-圆通，2-申通，3-中通，4-韵达，5-顺丰，6-京东，7-邮政',
    shipping_method TINYINT NOT NULL COMMENT '配送方式：1-标准配送，2-急速配送，3-经济配送，4-自提',
    tracking_number VARCHAR(64) NOT NULL COMMENT '快递单号',
    logistics_status TINYINT NOT NULL DEFAULT 1 COMMENT '物流状态：1-待发货，2-已发货，3-运输中，4-配送中，5-已签收，6-拒收，7-退货中，8-已退货',
    
    -- 发货信息
    sender_name VARCHAR(64) NOT NULL COMMENT '发货人姓名',
    sender_phone VARCHAR(32) NOT NULL COMMENT '发货人电话',
    sender_province VARCHAR(32) NOT NULL COMMENT '发货省份',
    sender_city VARCHAR(32) NOT NULL COMMENT '发货城市',
    sender_district VARCHAR(32) COMMENT '发货区县',
    sender_address TEXT NOT NULL COMMENT '发货详细地址',
    sender_postcode VARCHAR(10) COMMENT '发货邮编',
    
    -- 收货信息
    receiver_name VARCHAR(64) NOT NULL COMMENT '收货人姓名',
    receiver_phone VARCHAR(32) NOT NULL COMMENT '收货人电话',
    receiver_province VARCHAR(32) NOT NULL COMMENT '收货省份',
    receiver_city VARCHAR(32) NOT NULL COMMENT '收货城市',
    receiver_district VARCHAR(32) COMMENT '收货区县',
    receiver_address TEXT NOT NULL COMMENT '收货详细地址',
    receiver_postcode VARCHAR(10) COMMENT '收货邮编',
    
    -- 商品信息
    total_weight DECIMAL(8,3) DEFAULT 0 COMMENT '总重量（kg）',
    total_volume DECIMAL(10,3) DEFAULT 0 COMMENT '总体积（cm³）',
    total_quantity INT DEFAULT 0 COMMENT '总件数',
    goods_value DECIMAL(12,2) DEFAULT 0 COMMENT '商品价值',
    goods_description TEXT COMMENT '商品描述',
    
    -- 费用信息
    shipping_fee DECIMAL(8,2) DEFAULT 0 COMMENT '运费',
    insurance_fee DECIMAL(8,2) DEFAULT 0 COMMENT '保价费',
    other_fee DECIMAL(8,2) DEFAULT 0 COMMENT '其他费用',
    total_fee DECIMAL(8,2) DEFAULT 0 COMMENT '总费用',
    is_paid TINYINT DEFAULT 0 COMMENT '是否已付费：0-未付费，1-已付费',
    
    -- 配送信息
    courier_code VARCHAR(32) COMMENT '配送员编号',
    courier_name VARCHAR(64) COMMENT '配送员姓名',
    courier_phone VARCHAR(32) COMMENT '配送员电话',
    
    -- 时间信息
    shipped_at TIMESTAMP NULL COMMENT '发货时间',
    delivered_at TIMESTAMP NULL COMMENT '签收时间',
    estimated_delivery_at TIMESTAMP NULL COMMENT '预计送达时间',
    
    -- 特殊标记
    is_fragile TINYINT DEFAULT 0 COMMENT '是否易碎品',
    is_liquid TINYINT DEFAULT 0 COMMENT '是否液体',
    is_dangerous TINYINT DEFAULT 0 COMMENT '是否危险品',
    need_signature TINYINT DEFAULT 1 COMMENT '是否需要签收',
    need_inspection TINYINT DEFAULT 0 COMMENT '是否需要验货',
    
    -- 其他信息
    pickup_code VARCHAR(16) COMMENT '自提码',
    delivery_instructions TEXT COMMENT '配送说明',
    special_requirements TEXT COMMENT '特殊要求',
    remark TEXT COMMENT '备注信息',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_logistics_sn (logistics_sn),
    UNIQUE KEY uk_tracking_number (tracking_number),
    KEY idx_order_sn (order_sn),
    KEY idx_user_id (user_id),
    KEY idx_logistics_company (logistics_company),
    KEY idx_logistics_status (logistics_status),
    KEY idx_shipping_method (shipping_method),
    KEY idx_courier_code (courier_code),
    KEY idx_created_at (created_at),
    KEY idx_shipped_at (shipped_at),
    KEY idx_delivered_at (delivered_at),
    KEY idx_receiver_city (receiver_city),
    KEY idx_sender_city (sender_city)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流订单表';
```

### 4.2 物流轨迹表 (logistics_tracks)

```sql
CREATE TABLE logistics_tracks (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    logistics_sn VARCHAR(64) NOT NULL COMMENT '物流单号',
    tracking_number VARCHAR(64) NOT NULL COMMENT '快递单号',
    
    -- 轨迹信息
    track_sequence INT NOT NULL COMMENT '轨迹序号',
    status_code TINYINT NOT NULL COMMENT '状态编码',
    location VARCHAR(128) NOT NULL COMMENT '当前位置',
    description TEXT NOT NULL COMMENT '轨迹描述',
    track_time TIMESTAMP NOT NULL COMMENT '轨迹时间',
    
    -- 操作信息
    operator_name VARCHAR(64) COMMENT '操作员姓名',
    operator_phone VARCHAR(32) COMMENT '操作员电话',
    operation_type ENUM('pickup', 'transit', 'delivery', 'return', 'other') COMMENT '操作类型',
    
    -- 位置信息
    province VARCHAR(32) COMMENT '省份',
    city VARCHAR(32) COMMENT '城市',
    district VARCHAR(32) COMMENT '区县',
    longitude DECIMAL(10,6) COMMENT '经度',
    latitude DECIMAL(10,6) COMMENT '纬度',
    
    -- 附加信息
    next_location VARCHAR(128) COMMENT '下一站',
    estimated_time TIMESTAMP NULL COMMENT '预计时间',
    contact_info VARCHAR(128) COMMENT '联系方式',
    
    -- 数据来源
    data_source ENUM('manual', 'api', 'webhook', 'simulate') DEFAULT 'manual' COMMENT '数据来源',
    source_data TEXT COMMENT '原始数据（JSON格式）',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    KEY idx_logistics_sn (logistics_sn),
    KEY idx_tracking_number (tracking_number),
    KEY idx_track_time (track_time),
    KEY idx_status_code (status_code),
    KEY idx_city (city),
    KEY idx_track_sequence (logistics_sn, track_sequence),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流轨迹表';
```

### 4.3 物流公司表 (logistics_companies)

```sql
CREATE TABLE logistics_companies (
    id TINYINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    company_code VARCHAR(32) NOT NULL COMMENT '公司编码',
    company_name VARCHAR(64) NOT NULL COMMENT '公司名称',
    company_name_en VARCHAR(128) COMMENT '英文名称',
    
    -- 基本信息
    logo_url VARCHAR(256) COMMENT 'Logo地址',
    website_url VARCHAR(256) COMMENT '官网地址',
    customer_service VARCHAR(32) COMMENT '客服电话',
    
    -- 服务信息
    service_areas TEXT COMMENT '服务区域（JSON数组）',
    support_methods TEXT COMMENT '支持的配送方式（JSON数组）',
    max_weight DECIMAL(8,3) DEFAULT 0 COMMENT '最大重量限制（kg）',
    max_volume DECIMAL(10,3) DEFAULT 0 COMMENT '最大体积限制（cm³）',
    
    -- 费用配置
    base_fee DECIMAL(8,2) DEFAULT 0 COMMENT '起步费用',
    weight_fee DECIMAL(8,2) DEFAULT 0 COMMENT '重量费率（元/kg）',
    volume_fee DECIMAL(8,2) DEFAULT 0 COMMENT '体积费率（元/cm³）',
    insurance_rate DECIMAL(6,4) DEFAULT 0 COMMENT '保价费率',
    
    -- 时效配置
    standard_delivery_days INT DEFAULT 3 COMMENT '标准配送天数',
    express_delivery_days INT DEFAULT 1 COMMENT '急速配送天数',
    economy_delivery_days INT DEFAULT 5 COMMENT '经济配送天数',
    
    -- 状态配置
    is_available TINYINT DEFAULT 1 COMMENT '是否可用：0-不可用，1-可用',
    priority INT DEFAULT 100 COMMENT '优先级（数字越小优先级越高）',
    
    -- API配置
    api_enabled TINYINT DEFAULT 0 COMMENT '是否启用API：0-未启用，1-已启用',
    api_url VARCHAR(256) COMMENT 'API地址',
    api_key VARCHAR(128) COMMENT 'API密钥',
    api_config TEXT COMMENT 'API配置（JSON格式）',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_company_code (company_code),
    KEY idx_is_available (is_available),
    KEY idx_priority (priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流公司表';
```

### 4.4 配送员表 (logistics_couriers)

```sql
CREATE TABLE logistics_couriers (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    courier_code VARCHAR(32) NOT NULL COMMENT '配送员编号',
    courier_name VARCHAR(64) NOT NULL COMMENT '配送员姓名',
    phone VARCHAR(32) NOT NULL COMMENT '联系电话',
    id_card VARCHAR(18) COMMENT '身份证号',
    
    -- 所属信息
    logistics_company TINYINT NOT NULL COMMENT '所属物流公司',
    company_name VARCHAR(64) COMMENT '公司名称',
    station_code VARCHAR(32) COMMENT '所属站点编码',
    station_name VARCHAR(64) COMMENT '所属站点名称',
    
    -- 服务区域
    service_province VARCHAR(32) COMMENT '服务省份',
    service_city VARCHAR(32) COMMENT '服务城市',
    service_district VARCHAR(32) COMMENT '服务区县',
    service_area TEXT COMMENT '详细服务区域',
    
    -- 工作信息
    work_status TINYINT DEFAULT 1 COMMENT '工作状态：1-在岗，2-休假，3-离职',
    work_time VARCHAR(32) COMMENT '工作时间',
    max_capacity INT DEFAULT 100 COMMENT '最大配送能力（件/天）',
    current_load INT DEFAULT 0 COMMENT '当前负载（件）',
    
    -- 评价信息
    total_deliveries INT DEFAULT 0 COMMENT '总配送次数',
    success_deliveries INT DEFAULT 0 COMMENT '成功配送次数',
    rating_score DECIMAL(3,2) DEFAULT 5.00 COMMENT '评分（1-5分）',
    rating_count INT DEFAULT 0 COMMENT '评价次数',
    
    -- 联系信息
    email VARCHAR(128) COMMENT '邮箱',
    wechat VARCHAR(64) COMMENT '微信号',
    qq VARCHAR(16) COMMENT 'QQ号',
    
    -- 其他信息
    avatar_url VARCHAR(256) COMMENT '头像地址',
    description TEXT COMMENT '描述信息',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_courier_code (courier_code),
    KEY idx_logistics_company (logistics_company),
    KEY idx_service_city (service_city),
    KEY idx_work_status (work_status),
    KEY idx_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送员表';
```

## 5. 订单扩展表设计

### 5.1 订单状态日志表 (order_status_logs)

```sql
CREATE TABLE order_status_logs (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    
    -- 状态变更
    status_from TINYINT COMMENT '变更前主状态',
    status_to TINYINT NOT NULL COMMENT '变更后主状态',
    sub_status_from TINYINT COMMENT '变更前子状态',
    sub_status_to TINYINT COMMENT '变更后子状态',
    
    -- 变更类型
    change_type ENUM('payment', 'logistics', 'manual', 'system', 'user') NOT NULL COMMENT '变更类型',
    change_reason VARCHAR(256) COMMENT '变更原因',
    
    -- 操作信息
    operator_id INT COMMENT '操作员ID',
    operator_type ENUM('user', 'admin', 'system') NOT NULL COMMENT '操作员类型',
    operator_name VARCHAR(64) COMMENT '操作员姓名',
    
    -- 关联信息
    payment_sn VARCHAR(64) COMMENT '支付单号',
    logistics_sn VARCHAR(64) COMMENT '物流单号',
    related_data TEXT COMMENT '相关数据（JSON格式）',
    
    -- 其他信息
    client_ip VARCHAR(45) COMMENT '客户端IP',
    user_agent TEXT COMMENT '用户代理',
    remark TEXT COMMENT '备注信息',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    KEY idx_order_sn (order_sn),
    KEY idx_change_type (change_type),
    KEY idx_operator_id (operator_id),
    KEY idx_created_at (created_at),
    KEY idx_status_from_to (status_from, status_to)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态日志表';
```

## 6. 系统配置表设计

### 6.1 支付配置表 (payment_config)

```sql
CREATE TABLE payment_config (
    id INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    config_key VARCHAR(64) NOT NULL COMMENT '配置键',
    config_value TEXT NOT NULL COMMENT '配置值',
    config_type ENUM('string', 'int', 'float', 'bool', 'json') DEFAULT 'string' COMMENT '配置类型',
    config_group VARCHAR(32) NOT NULL COMMENT '配置分组',
    description TEXT COMMENT '配置描述',
    is_encrypted TINYINT DEFAULT 0 COMMENT '是否加密：0-否，1-是',
    is_active TINYINT DEFAULT 1 COMMENT '是否激活：0-否，1-是',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_config_key (config_key),
    KEY idx_config_group (config_group),
    KEY idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付配置表';
```

### 6.2 物流配置表 (logistics_config)

```sql
CREATE TABLE logistics_config (
    id INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    config_key VARCHAR(64) NOT NULL COMMENT '配置键',
    config_value TEXT NOT NULL COMMENT '配置值',
    config_type ENUM('string', 'int', 'float', 'bool', 'json') DEFAULT 'string' COMMENT '配置类型',
    config_group VARCHAR(32) NOT NULL COMMENT '配置分组',
    description TEXT COMMENT '配置描述',
    is_active TINYINT DEFAULT 1 COMMENT '是否激活：0-否，1-是',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_config_key (config_key),
    KEY idx_config_group (config_group),
    KEY idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='物流配置表';
```

## 7. 定时任务表设计

### 7.1 定时任务表 (scheduled_tasks)

```sql
CREATE TABLE scheduled_tasks (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    task_type ENUM('payment_expire', 'auto_ship', 'auto_complete', 'track_update') NOT NULL COMMENT '任务类型',
    task_name VARCHAR(128) NOT NULL COMMENT '任务名称',
    related_id VARCHAR(64) NOT NULL COMMENT '关联ID（订单号、支付单号等）',
    
    -- 执行信息
    scheduled_at TIMESTAMP NOT NULL COMMENT '计划执行时间',
    executed_at TIMESTAMP NULL COMMENT '实际执行时间',
    execution_status ENUM('pending', 'running', 'success', 'failed', 'cancelled') DEFAULT 'pending' COMMENT '执行状态',
    
    -- 任务数据
    task_data TEXT COMMENT '任务数据（JSON格式）',
    result_data TEXT COMMENT '执行结果（JSON格式）',
    error_message TEXT COMMENT '错误信息',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    max_retries INT DEFAULT 3 COMMENT '最大重试次数',
    
    -- 优先级
    priority INT DEFAULT 100 COMMENT '优先级（数字越小优先级越高）',
    
    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    KEY idx_task_type (task_type),
    KEY idx_related_id (related_id),
    KEY idx_scheduled_at (scheduled_at),
    KEY idx_execution_status (execution_status),
    KEY idx_priority (priority),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';
```

## 8. 数据迁移脚本

### 8.1 现有订单表扩展

```sql
-- 为现有订单表添加支付和物流字段
ALTER TABLE order_info 
ADD COLUMN payment_status TINYINT DEFAULT 0 COMMENT '支付子状态' AFTER status,
ADD COLUMN logistics_status TINYINT DEFAULT 0 COMMENT '物流子状态' AFTER payment_status,
ADD COLUMN payment_sn VARCHAR(64) COMMENT '支付单号' AFTER logistics_status,
ADD COLUMN logistics_sn VARCHAR(64) COMMENT '物流单号' AFTER payment_sn,
ADD COLUMN tracking_number VARCHAR(64) COMMENT '快递单号' AFTER logistics_sn,
ADD COLUMN paid_at TIMESTAMP NULL COMMENT '支付时间' AFTER tracking_number,
ADD COLUMN shipped_at TIMESTAMP NULL COMMENT '发货时间' AFTER paid_at,
ADD COLUMN delivered_at TIMESTAMP NULL COMMENT '送达时间' AFTER shipped_at,
ADD COLUMN completed_at TIMESTAMP NULL COMMENT '完成时间' AFTER delivered_at,
ADD COLUMN cancelled_at TIMESTAMP NULL COMMENT '取消时间' AFTER completed_at,
ADD COLUMN cancel_reason TEXT COMMENT '取消原因' AFTER cancelled_at,
ADD COLUMN auto_complete_at TIMESTAMP NULL COMMENT '自动完成时间' AFTER cancel_reason;

-- 添加索引
ALTER TABLE order_info 
ADD INDEX idx_payment_sn (payment_sn),
ADD INDEX idx_logistics_sn (logistics_sn),
ADD INDEX idx_tracking_number (tracking_number),
ADD INDEX idx_payment_status (payment_status),
ADD INDEX idx_logistics_status (logistics_status);
```

### 8.2 初始化物流公司数据

```sql
INSERT INTO logistics_companies (
    company_code, company_name, logo_url, customer_service, 
    is_available, priority, standard_delivery_days, 
    express_delivery_days, economy_delivery_days
) VALUES 
('YTO', '圆通速递', 'https://cdn.example.com/logo/yto.png', '400-111-1111', 1, 100, 3, 2, 5),
('STO', '申通快递', 'https://cdn.example.com/logo/sto.png', '400-222-2222', 1, 110, 3, 2, 5),
('ZTO', '中通快递', 'https://cdn.example.com/logo/zto.png', '400-333-3333', 1, 90, 3, 2, 5),
('YD', '韵达速递', 'https://cdn.example.com/logo/yunda.png', '400-444-4444', 1, 120, 3, 2, 5),
('SF', '顺丰速运', 'https://cdn.example.com/logo/sf.png', '400-555-5555', 1, 10, 2, 1, 4),
('JD', '京东物流', 'https://cdn.example.com/logo/jd.png', '400-666-6666', 1, 20, 2, 1, 4),
('EMS', '中国邮政', 'https://cdn.example.com/logo/ems.png', '400-777-7777', 1, 200, 5, 3, 7);
```

### 8.3 初始化配置数据

```sql
-- 支付配置
INSERT INTO payment_config (config_key, config_value, config_type, config_group, description) VALUES
('payment.default_expire_minutes', '15', 'int', 'payment', '默认支付过期时间（分钟）'),
('payment.max_amount', '99999.99', 'float', 'payment', '单笔支付最大金额'),
('payment.min_amount', '0.01', 'float', 'payment', '单笔支付最小金额'),
('payment.simulation_enabled', 'true', 'bool', 'payment', '是否启用支付模拟'),
('payment.auto_success_rate', '0.8', 'float', 'payment', '自动支付成功率'),
('payment.notify_max_times', '5', 'int', 'payment', '支付通知最大次数'),
('refund.max_days', '30', 'int', 'refund', '退款申请最大天数'),
('refund.auto_approve', 'false', 'bool', 'refund', '是否自动批准退款');

-- 物流配置
INSERT INTO logistics_config (config_key, config_value, config_type, config_group, description) VALUES
('logistics.auto_ship_hours', '2', 'int', 'logistics', '自动发货延迟时间（小时）'),
('logistics.auto_complete_days', '7', 'int', 'logistics', '自动确认收货天数'),
('logistics.track_update_interval', '30', 'int', 'logistics', '轨迹更新间隔（分钟）'),
('logistics.default_company', '5', 'int', 'logistics', '默认物流公司ID'),
('logistics.simulation_enabled', 'true', 'bool', 'logistics', '是否启用物流模拟'),
('logistics.max_weight', '30', 'float', 'logistics', '最大重量限制（kg）'),
('logistics.max_volume', '100000', 'float', 'logistics', '最大体积限制（cm³）'),
('logistics.insurance_rate', '0.005', 'float', 'logistics', '保价费率');
```

## 9. 索引优化建议

### 9.1 分区表设计

```sql
-- 按时间分区的物流轨迹表
CREATE TABLE logistics_tracks_partitioned (
    -- 表结构同logistics_tracks
) PARTITION BY RANGE (YEAR(created_at)) (
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION p2025 VALUES LESS THAN (2026),
    PARTITION p2026 VALUES LESS THAN (2027),
    PARTITION pmax VALUES LESS THAN MAXVALUE
);

-- 按用户ID分区的支付订单表
CREATE TABLE payment_orders_partitioned (
    -- 表结构同payment_orders
) PARTITION BY HASH (user_id) PARTITIONS 32;
```

### 9.2 复合索引建议

```sql
-- 支付订单复合索引
ALTER TABLE payment_orders ADD INDEX idx_user_status_created (user_id, payment_status, created_at);
ALTER TABLE payment_orders ADD INDEX idx_status_expired (payment_status, expired_at);
ALTER TABLE payment_orders ADD INDEX idx_method_status_amount (payment_method, payment_status, amount);

-- 物流订单复合索引
ALTER TABLE logistics_orders ADD INDEX idx_user_status_created (user_id, logistics_status, created_at);
ALTER TABLE logistics_orders ADD INDEX idx_company_city_status (logistics_company, receiver_city, logistics_status);
ALTER TABLE logistics_orders ADD INDEX idx_courier_status (courier_code, logistics_status);

-- 物流轨迹复合索引
ALTER TABLE logistics_tracks ADD INDEX idx_tracking_time (tracking_number, track_time);
ALTER TABLE logistics_tracks ADD INDEX idx_city_status_time (city, status_code, track_time);
```

## 10. 数据归档策略

### 10.1 归档配置

```sql
-- 创建归档表
CREATE TABLE payment_orders_archive LIKE payment_orders;
CREATE TABLE logistics_tracks_archive LIKE logistics_tracks;
CREATE TABLE payment_logs_archive LIKE payment_logs;

-- 归档策略：保留3年的数据，超过3年的数据归档
-- 可以通过定时任务执行归档操作
```

### 10.2 数据清理策略

```sql
-- 清理策略：
-- 1. 支付日志保留1年
-- 2. 物流轨迹保留2年
-- 3. 订单状态日志保留3年
-- 4. 定时任务记录保留3个月
```

这个数据库设计提供了完整的支付和物流服务数据存储方案，包括了表结构、索引、分区、归档等多个方面的设计，能够支撑大规模电商系统的运营需求。