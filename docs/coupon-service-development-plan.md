# 优惠券服务开发计划

<div align="center">

![Status](https://img.shields.io/badge/Status-80%25%20Complete-green.svg)
![Version](https://img.shields.io/badge/Version-v1.0-green.svg)
![Architecture](https://img.shields.io/badge/Architecture-Microservice-orange.svg)

**基于EMShop微服务架构的高性能优惠券服务开发计划**

</div>

## 📋 项目概述

本项目旨在为EMShop电商系统构建一个高性能、高可用的优惠券服务，重点解决**优惠券秒杀**和**支付抵用**两大核心场景。

### 🎯 设计理念

**分场景设计策略**：
- **优惠券秒杀**：采用轻量级Redis方案，追求极致性能
- **支付抵用**：集成DTM分布式事务，确保数据强一致性

### 🚀 核心目标

- **高性能**：优惠券秒杀支持50,000+ QPS
- **零超卖**：基于Redis原子操作的精确库存控制
- **强一致性**：支付场景下的分布式事务保障
- **易维护**：清晰的架构设计和完善的监控体系

## 🏗️ 系统架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    优惠券服务架构                        │
├─────────────────────────────────────────────────────────┤
│  接入层    │ Kong API网关 + 限流 + 负载均衡             │
├─────────────────────────────────────────────────────────┤
│  应用层    │ Coupon Service (gRPC) + 本地缓存           │
├─────────────────────────────────────────────────────────┤
│  缓存层    │ Redis集群 (库存控制 + 用户状态)            │
├─────────────────────────────────────────────────────────┤
│  消息层    │ RocketMQ (异步处理 + 事务消息)             │
├─────────────────────────────────────────────────────────┤
│  存储层    │ MySQL (持久化) + Elasticsearch (搜索)      │
├─────────────────────────────────────────────────────────┤
│  事务层    │ DTM (分布式事务协调器)                      │
└─────────────────────────────────────────────────────────┘
```

### 技术栈选择

| 组件类型 | 技术选型 | 用途说明 |
|---------|---------|----------|
| **微服务框架** | Go + gRPC | 与现有服务保持一致 |
| **服务发现** | Consul | 已有基础设施 |
| **缓存** | Redis 集群 | 高性能库存控制 |
| **数据库** | MySQL 8.0 | 持久化存储 |
| **消息队列** | RocketMQ | 异步处理和事务消息 |
| **分布式事务** | DTM Saga | 支付场景强一致性 |
| **监控** | Prometheus + Grafana | 性能监控和告警 |

## 🎯 核心功能设计

### 场景1：高性能优惠券秒杀

#### 技术方案
```
用户请求 → Kong限流 → Ristretto缓存(L1) → Redis缓存(L2) → MySQL(L3)
                                     ↓
                        Canal数据变更 → 缓存失效 → 多级缓存更新
```

#### 核心特性
- **Redis Lua脚本**：确保库存扣减的原子性
- **用户防重**：Redis记录用户抢购状态，防止重复抢购
- **异步处理**：成功抢购后，通过RocketMQ异步处理数据库写入
- **Ristretto本地缓存**：采用TinyLFU算法，95%+命中率，热门优惠券信息缓存到应用内存
- **Canal缓存一致性**：利用现有Canal+RocketMQ机制，实时同步缓存更新

#### 关键实现

**Redis Lua脚本**：
```lua
-- 优惠券秒杀原子操作脚本
local couponKey = KEYS[1]        -- 优惠券库存key
local userKey = KEYS[2]          -- 用户抢购记录key  
local userId = ARGV[1]           -- 用户ID
local decreNum = tonumber(ARGV[2]) -- 扣减数量

-- 检查用户是否已抢购
if redis.call('EXISTS', userKey) == 1 then
    return -2  -- 用户已抢购过
end

-- 检查并扣减库存
local stock = redis.call('GET', couponKey)
if not stock or tonumber(stock) < decreNum then
    return -1  -- 库存不足
end

-- 原子操作：扣库存 + 记录用户
redis.call('DECRBY', couponKey, decreNum)
redis.call('SETEX', userKey, 1800, userId)  -- 30分钟过期

return tonumber(stock) - decreNum  -- 返回剩余库存
```

#### 性能目标
- **QPS**: 60,000+（单机12,000+ QPS × 5个实例，Ristretto性能提升）
- **响应时间**: < 30ms（P99）
- **成功率**: 99.9%+
- **零超卖**: 100%准确的库存控制

### 场景2：支付优惠券使用

#### DTM分布式事务方案

**Saga事务流程**：
```
1. [Coupon] 锁定优惠券 ← → 释放锁定
2. [Order] 计算优惠金额 ← → 恢复原价
3. [Payment] 创建支付订单 ← → 取消支付
4. [Inventory] 扣减商品库存 ← → 恢复库存
```

#### 关键实现

**优惠券锁定接口**：
```protobuf
service Coupon {
    // Saga正向操作
    rpc LockUserCoupon(LockCouponRequest) returns (google.protobuf.Empty);
    rpc UseLockedCoupon(UseCouponRequest) returns (UseCouponResponse);
    
    // Saga补偿操作  
    rpc UnlockUserCoupon(UnlockCouponRequest) returns (google.protobuf.Empty);
    rpc RevertUsedCoupon(RevertCouponRequest) returns (google.protobuf.Empty);
}
// 注意: 可选字段添加optional 
```

**事务协调逻辑**：
```go
func (s *CouponService) ProcessOrderPaymentWithCoupon(ctx context.Context, req *OrderPaymentRequest) error {
    // 创建DTM Saga事务
    saga := dtmcli.NewSaga(s.dtmServer, dtmcli.MustGenGid(s.dtmServer))
    
    // 构建事务请求
    lockReq := &LockCouponRequest{
        UserId: req.UserId,
        CouponId: req.CouponId,
        OrderSn: req.OrderSn,
    }
    
    // 添加事务步骤
    saga.Add(s.couponSrv+"/LockUserCoupon", s.couponSrv+"/UnlockUserCoupon", lockReq)
    saga.Add(s.orderSrv+"/CalculateDiscount", s.orderSrv+"/RevertDiscount", req)
    saga.Add(s.paymentSrv+"/CreatePaymentOrder", s.paymentSrv+"/CancelPaymentOrder", req)
    saga.Add(s.inventorySrv+"/ReserveStock", s.inventorySrv+"/ReleaseStock", req)
    
    // 提交事务
    return saga.Submit()
}
```

## 📊 数据模型设计

### 核心数据表

#### 1. 优惠券模板表 (coupon_templates)
```sql
CREATE TABLE coupon_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '优惠券名称',
    type TINYINT NOT NULL COMMENT '类型：1-满减，2-折扣，3-免邮',
    discount_type TINYINT NOT NULL COMMENT '优惠类型：1-金额，2-比例',
    discount_value DECIMAL(10,2) NOT NULL COMMENT '优惠值',
    min_amount DECIMAL(10,2) DEFAULT 0 COMMENT '最低消费金额',
    total_count INT NOT NULL COMMENT '发放总数',
    used_count INT DEFAULT 0 COMMENT '已使用数量',
    per_user_limit INT DEFAULT 1 COMMENT '单用户限制',
    valid_start_time TIMESTAMP NOT NULL COMMENT '有效期开始',
    valid_end_time TIMESTAMP NOT NULL COMMENT '有效期结束',
    applicable_goods TEXT COMMENT '适用商品ID列表(JSON)',
    applicable_categories TEXT COMMENT '适用分类ID列表(JSON)',
    status TINYINT DEFAULT 1 COMMENT '状态：0-草稿，1-发布，2-停用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_status_time (status, valid_start_time, valid_end_time),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表';
```

#### 2. 用户优惠券表 (user_coupons)
```sql
CREATE TABLE user_coupons (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    coupon_id BIGINT NOT NULL COMMENT '优惠券模板ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    coupon_sn VARCHAR(32) UNIQUE NOT NULL COMMENT '优惠券编号',
    status TINYINT DEFAULT 1 COMMENT '状态：1-未使用，2-已锁定，3-已使用，4-已过期',
    obtain_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '获取时间',
    used_time TIMESTAMP NULL COMMENT '使用时间',
    order_sn VARCHAR(64) NULL COMMENT '使用订单号',
    valid_start_time TIMESTAMP NOT NULL COMMENT '有效期开始',
    valid_end_time TIMESTAMP NOT NULL COMMENT '有效期结束',
    
    UNIQUE KEY uk_coupon_sn (coupon_sn),
    INDEX idx_user_status (user_id, status),
    INDEX idx_coupon_id (coupon_id),
    INDEX idx_valid_time (valid_start_time, valid_end_time),
    FOREIGN KEY (coupon_id) REFERENCES coupon_templates(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';
```

#### 3. 优惠券使用记录表 (coupon_usage_logs)
```sql
CREATE TABLE coupon_usage_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_coupon_id BIGINT NOT NULL COMMENT '用户优惠券ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    original_amount DECIMAL(10,2) NOT NULL COMMENT '原始金额',
    discount_amount DECIMAL(10,2) NOT NULL COMMENT '优惠金额',
    final_amount DECIMAL(10,2) NOT NULL COMMENT '最终金额',
    used_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '使用时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_order_sn (order_sn),
    INDEX idx_used_time (used_time),
    FOREIGN KEY (user_coupon_id) REFERENCES user_coupons(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券使用记录表';
```

#### 4. 秒杀活动表 (flash_sale_activities)  
```sql
CREATE TABLE flash_sale_activities (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    coupon_id BIGINT NOT NULL COMMENT '优惠券模板ID',
    name VARCHAR(100) NOT NULL COMMENT '活动名称',
    total_count INT NOT NULL COMMENT '总投放数量',
    success_count INT DEFAULT 0 COMMENT '成功抢购数量',
    start_time TIMESTAMP NOT NULL COMMENT '开始时间',
    end_time TIMESTAMP NOT NULL COMMENT '结束时间',
    per_user_limit INT DEFAULT 1 COMMENT '单用户抢购限制',
    status TINYINT DEFAULT 1 COMMENT '状态：1-待开始，2-进行中，3-已结束',
    
    INDEX idx_coupon_id (coupon_id),
    INDEX idx_time (start_time, end_time),
    INDEX idx_status (status),
    FOREIGN KEY (coupon_id) REFERENCES coupon_templates(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';
```

### Redis数据结构设计

#### 库存控制
```
# 优惠券库存
coupon:stock:{coupon_id} = {available_count}

# 用户抢购记录  
coupon:user:{activity_id}:{user_id} = {timestamp}

# 优惠券模板缓存
coupon:template:{coupon_id} = {template_json}
```

#### 分布式锁
```
# 用户优惠券锁定（支付场景）
coupon:lock:{user_coupon_id} = {order_sn}

# 秒杀活动锁定
coupon:activity:lock:{activity_id} = {server_instance}
```

## 🔧 gRPC接口设计

### 优惠券管理接口
```protobuf
syntax = "proto3";
import "google/protobuf/empty.proto";
option go_package = ".;proto";

service Coupon {
    // 优惠券模板管理
    rpc CreateCouponTemplate(CreateTemplateRequest) returns (TemplateResponse);
    rpc UpdateCouponTemplate(UpdateTemplateRequest) returns (google.protobuf.Empty);
    rpc GetCouponTemplate(GetTemplateRequest) returns (TemplateResponse);
    rpc ListCouponTemplates(ListTemplateRequest) returns (ListTemplateResponse);
    
    // 秒杀相关
    rpc StartFlashSale(StartFlashSaleRequest) returns (google.protobuf.Empty);
    rpc FlashSaleCoupon(FlashSaleRequest) returns (FlashSaleResponse);
    rpc GetFlashSaleStatus(FlashSaleStatusRequest) returns (FlashSaleStatusResponse);
    
    // 用户优惠券  
    rpc GetUserCoupons(GetUserCouponsRequest) returns (UserCouponsResponse);
    rpc GetAvailableCouponsForOrder(OrderCouponsRequest) returns (AvailableCouponsResponse);
    
    // 优惠券使用（DTM分布式事务接口）
    rpc LockUserCoupon(LockCouponRequest) returns (google.protobuf.Empty);
    rpc UnlockUserCoupon(UnlockCouponRequest) returns (google.protobuf.Empty);
    rpc UseLockedCoupon(UseCouponRequest) returns (UseCouponResponse);
    rpc RevertUsedCoupon(RevertCouponRequest) returns (google.protobuf.Empty);
    
    // 优惠计算
    rpc CalculateDiscount(CalculateDiscountRequest) returns (DiscountResponse);
}
```

### 关键消息定义

#### 秒杀请求/响应
<!-- 注意:可选字段使用optional关键字修饰 -->
```protobuf
message FlashSaleRequest {
    int64 activity_id = 1;     // 秒杀活动ID
    int64 user_id = 2;         // 用户ID
    string client_ip = 3;      // 客户端IP（防刷）
}

message FlashSaleResponse {
    bool success = 1;          // 是否成功
    string message = 2;        // 响应消息
    string coupon_sn = 3;      // 优惠券编号（成功时）
    int64 remaining_count = 4; // 剩余库存
}
```

#### 优惠计算请求/响应
```protobuf
message CalculateDiscountRequest {
    int64 user_id = 1;                        // 用户ID
    repeated int64 coupon_ids = 2;            // 要使用的优惠券ID列表
    repeated OrderItem order_items = 3;       // 订单商品列表
    double shipping_fee = 4;                  // 运费
}

message DiscountResponse {
    double original_amount = 1;               // 原始金额
    double discount_amount = 2;               // 优惠金额
    double final_amount = 3;                  // 最终金额
    repeated CouponDiscount coupon_discounts = 4; // 每个优惠券的优惠详情
}

message OrderItem {
    int64 goods_id = 1;       // 商品ID
    int32 quantity = 2;       // 数量
    double price = 3;         // 单价
    int64 category_id = 4;    // 商品分类ID
}

message CouponDiscount {
    int64 coupon_id = 1;      // 优惠券ID
    double discount_amount = 2; // 本券优惠金额
    string discount_reason = 3; // 优惠说明
}
```

## 📊 实施状态更新 (2025-08-27)

### ✅ 已完成的核心功能 (80%+)

#### 🏗️ 基础架构完备
- [x] **项目结构**: 完整的服务目录结构和配置文件
- [x] **数据层实现**: GORM模型、Repository接口、MySQL实现
- [x] **gRPC服务**: Protobuf接口定义、代码生成、服务器实现
- [x] **服务注册**: Consul集成和健康检查

#### 🔥 秒杀引擎完备  
- [x] **Redis Lua脚本**: 原子操作脚本，确保零超卖
- [x] **库存管理器**: 高性能StockManager，支持预热、回滚
- [x] **秒杀核心**: FlashSaleSrvCore完整业务逻辑实现
- [x] **用户防重**: Redis记录用户抢购状态

#### 📦 缓存系统完备
- [x] **三层缓存**: Ristretto + Redis + MySQL架构
- [x] **Canal集成**: 缓存一致性保障机制  
- [x] **缓存管理器**: CacheManager接口和实现
- [x] **预热机制**: 支持热门数据预加载

#### 💳 分布式事务完备
- [x] **DTM集成**: Saga事务协调器封装
- [x] **优惠券锁定**: 锁定/解锁逻辑实现
- [x] **优惠计算**: 多策略优惠计算引擎
- [x] **事务处理器**: DTMManager和相关接口

#### 📨 消息系统基础
- [x] **消费者结构**: FlashSaleConsumer完整实现
- [x] **事件定义**: 秒杀成功/失败事件结构  
- [x] **幂等处理**: 避免重复消费机制
- [x] **Canal消费者**: 缓存一致性消息处理

### 🔧 需要完善的关键环节

## 🚀 开发计划 (剩余核心完善工作)

### 阶段1: RocketMQ事务消息增强

#### 1.1 事务消息实现  
- [x] RocketMQ Producer基础实现已完成
- [ ] 实现事务消息支持（确保消息和业务的一致性）
- [ ] 添加事务消息回查机制
- [ ] 实现消息重试策略和失败处理
- [ ] 优化消息序列化和传输可靠性

**交付物**：
- RocketMQ事务消息完整实现
- 业务操作与消息发送的强一致性保障
- 可靠的消息重试和失败恢复机制

### 阶段2: 功能测试和验证

#### 2.1 集成测试优化
- [ ] 优化现有test-coupon-*.sh脚本
- [ ] 验证秒杀功能的并发安全性和库存准确性
- [ ] 测试分布式事务的完整性和一致性
- [ ] 验证RocketMQ消息的可靠投递

#### 2.2 核心功能验证
- [ ] 优惠券模板管理功能测试
- [ ] 秒杀活动完整流程测试
- [ ] 支付抵用分布式事务测试
- [ ] 缓存一致性和性能测试

**交付物**：
- 完整的功能回归测试
- 验证所有核心业务流程
- 确保服务稳定性和数据一致性

## 🎯 交付目标

通过简化的开发计划，将实现：

### 📈 功能完整性
- **优惠券管理**: 模板创建、更新、查询等完整功能
- **秒杀系统**: 高并发秒杀，零超卖保障
- **分布式事务**: 支付场景的数据强一致性
- **消息可靠性**: RocketMQ事务消息确保业务一致性

### 🛡️ 系统稳定性
- **功能验证**: 所有核心业务流程测试通过
- **并发安全**: 高并发场景下的数据准确性
- **服务可用**: 服务可以稳定启动并处理请求
- **集成完整**: 各个组件协作无异常

### 💼 业务价值
- **营销支持**: 支撑秒杀活动和优惠券营销
- **用户体验**: 快速响应的抢购体验
- **数据一致**: 严格的库存控制和事务保障
- **易于维护**: 清晰的代码结构和接口设计

## 📁 服务目录结构

```
internal/app/coupon/
├── srv/
│   ├── app.go                          # 应用入口
│   ├── config/
│   │   └── config.go                   # 配置结构定义
│   ├── controller/
│   │   └── v1/
│   │       ├── coupon.go               # gRPC接口实现
│   │       ├── flashsale.go            # 秒杀接口实现  
│   │       └── dtm.go                  # DTM事务接口
│   ├── service/
│   │   └── v1/
│   │       ├── service.go              # 服务接口定义
│   │       ├── coupon.go               # 优惠券业务逻辑
│   │       ├── flashsale.go            # 秒杀业务逻辑
│   │       ├── discount.go             # 优惠计算引擎
│   │       └── dtm_manager.go          # DTM事务管理
│   ├── data/
│   │   └── v1/
│   │       ├── factory_manager.go      # 数据工厂管理
│   │       ├── interfaces/             # Repository接口
│   │       │   ├── coupon.go           
│   │       │   ├── user_coupon.go      
│   │       │   └── flashsale.go        
│   │       ├── mysql/                  # MySQL实现
│   │       │   ├── factory.go          
│   │       │   ├── coupon.go           
│   │       │   ├── user_coupon.go      
│   │       │   └── flashsale.go        
│   │       └── redis/                  # Redis实现
│   │           ├── coupon_cache.go     
│   │           ├── stock_manager.go    
│   │           └── lua_scripts.go      
│   ├── domain/
│   │   ├── do/                         # 数据对象
│   │   │   ├── coupon.go               
│   │   │   ├── user_coupon.go          
│   │   │   └── flashsale.go            
│   │   └── dto/                        # 传输对象
│   │       ├── coupon.go               
│   │       └── flashsale.go            
│   ├── pkg/
│   │   ├── calculator/                 # 优惠计算器
│   │   │   ├── calculator.go           
│   │   │   ├── fullcut.go              # 满减计算
│   │   │   ├── discount.go             # 折扣计算
│   │   │   └── freeshipping.go         # 免邮计算
│   │   ├── validator/                  # 业务验证器
│   │   │   ├── coupon_validator.go     
│   │   │   └── order_validator.go      
│   │   └── constants/                  # 常量定义
│   │       ├── coupon_status.go        
│   │       └── error_codes.go          
│   ├── consumer/                       # 消息消费者
│   │   ├── coupon_consumer.go          # 优惠券消息处理
│   │   └── dtm_consumer.go             # DTM事务消息
│   └── rpc.go                          # gRPC服务启动
```

## ⚙️ 配置文件模板

### configs/coupon/srv.yaml
```yaml
# 服务器配置
server:
  name: "coupon"
  host: "0.0.0.0"
  port: 0 # 随机分配
  http-port: 8056
  healthz: true
  enable-metrics: true
  profiling: true

# 日志配置  
log:
  name: emshop-coupon-srv
  development: true
  level: debug
  format: json
  enable-color: false
  disable-caller: false
  disable-stacktrace: false
  output-paths: logs/emshop-coupon-srv.log,stdout
  error-output-paths: logs/emshop-coupon-srv.error.log

# 服务注册与发现
registry:
  address: localhost:8500
  scheme: http

# 链路追踪配置
telemetry:
  name: coupon
  endpoint: http://localhost:14268/api/traces
  sampler: 1.0
  batcher: jaeger

# MySQL配置
mysql:
  host: "localhost"
  port: 3306
  password: "root" 
  username: "root"
  database: "emshop_coupon_srv"
  max_idle_connections: 10
  max_open_connections: 100
  max_connection_life_time: "1h"
  log_level: 4

# Redis配置
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 50
  min_idle_conns: 10
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"

# RocketMQ配置
rocketmq:
  nameservers: ["localhost:9876"]
  producer_group: "coupon-producer-group"
  consumer_group: "coupon-consumer-group"
  topic: "coupon-events"
  max_reconsume: 3

# DTM配置
dtm:
  grpc_server: "localhost:36790"
  http_server: "localhost:36789"
  timeout: "30s"

# Ristretto本地缓存配置
ristretto:
  num_counters: 1000000           # 1M个key的统计信息
  max_cost: 104857600             # 100MB最大内存
  buffer_items: 64                # 缓冲区大小
  metrics: true                   # 开启监控指标
  
# Canal缓存一致性配置  
canal:
  consumer_group: "coupon-cache-sync-consumer"  # Canal消费者组
  topic: "coupon-binlog-topic"                  # Canal消息主题
  watch_tables:                                 # 监听的表
    - "coupon_templates"
    - "user_coupons" 
    - "flash_sale_activities"
  batch_size: 32                               # 批量处理大小
  
# 业务配置
business:
  # 秒杀配置
  flashsale:
    max_qps_per_user: 5           # 单用户最大QPS
    stock_cache_ttl: "300s"       # 库存缓存TTL
    user_limit_ttl: "1800s"       # 用户限制TTL
    batch_size: 100               # 批量处理大小
    
  # 优惠券配置
  coupon:
    max_stack_count: 5            # 最大叠加数量
    lock_ttl: "900s"              # 锁定TTL（15分钟）
    calc_timeout: "5s"            # 计算超时时间
    
  # 缓存配置
  cache:
    l1_ttl: "10m"                 # L1缓存TTL
    l2_ttl: "30m"                 # L2缓存TTL  
    warmup_count: 100             # 预热优惠券数量
    enable_warmup: true           # 是否开启预热
```

## 🔍 关键实现细节

### 1. Redis Lua脚本实现

```lua
-- coupon_flash_sale.lua
-- 优惠券秒杀原子操作脚本

local couponKey = KEYS[1]        -- 库存key: coupon:stock:{coupon_id}
local userKey = KEYS[2]          -- 用户key: coupon:user:{activity_id}:{user_id}
local logKey = KEYS[3]           -- 日志key: coupon:log:{activity_id}

local userId = ARGV[1]           -- 用户ID
local activityId = ARGV[2]       -- 活动ID
local decreNum = tonumber(ARGV[3]) -- 扣减数量
local ttl = tonumber(ARGV[4])    -- TTL秒数

-- 1. 检查用户是否已参与
if redis.call('EXISTS', userKey) == 1 then
    return {-2, 0, "用户已参与"}
end

-- 2. 检查库存
local stock = redis.call('GET', couponKey)
if not stock then
    return {-1, 0, "活动不存在"}
end

stock = tonumber(stock)
if stock < decreNum then
    return {-1, stock, "库存不足"}
end

-- 3. 原子操作：扣库存 + 记录用户 + 写日志
local remainStock = stock - decreNum
redis.call('SET', couponKey, remainStock)
redis.call('SETEX', userKey, ttl, userId)

-- 4. 记录抢购日志（可选）
local logData = string.format("%s:%s:%d", userId, activityId, redis.call('TIME')[1])
redis.call('LPUSH', logKey, logData)
redis.call('EXPIRE', logKey, ttl)

return {1, remainStock, "秒杀成功"}
```

### 2. Ristretto三层缓存管理器

```go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/dgraph-io/ristretto"
    "github.com/go-redis/redis/v8"
    "emshop/pkg/log"
)

// CouponCacheManager 三层缓存管理器
type CouponCacheManager struct {
    // L1: Ristretto本地缓存 (1ms响应，95%命中率)
    localCache  *ristretto.Cache
    // L2: Redis集群缓存 (5ms响应，90%命中率)  
    redis       *redis.Client
    // L3: MySQL数据库 (20ms响应，100%命中率)
    repository  CouponRepository
}

func NewCouponCacheManager(redis *redis.Client, repo CouponRepository) *CouponCacheManager {
    // 初始化Ristretto缓存
    cache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: 1000000,   // 1M个key的统计信息
        MaxCost:     100 << 20, // 100MB最大内存
        BufferItems: 64,        // 缓冲区大小
        Metrics:     true,      // 开启监控指标
    })
    if err != nil {
        log.Fatalf("初始化Ristretto缓存失败: %v", err)
    }
    
    return &CouponCacheManager{
        localCache: cache,
        redis:      redis,
        repository: repo,
    }
}

// GetCouponTemplate 获取优惠券模板（三层缓存查询）
func (ccm *CouponCacheManager) GetCouponTemplate(ctx context.Context, couponID int64) (*CouponTemplate, error) {
    key := fmt.Sprintf("coupon:template:%d", couponID)
    
    // L1: Ristretto本地缓存查询
    if value, found := ccm.localCache.Get(key); found {
        template := value.(*CouponTemplate)
        log.Debugf("命中L1缓存, couponID: %d", couponID)
        return template, nil
    }
    
    // L2: Redis缓存查询
    if data := ccm.redis.Get(ctx, key).Val(); data != "" {
        var template CouponTemplate
        if err := json.Unmarshal([]byte(data), &template); err == nil {
            // 回写L1缓存 (成本为1，TTL 10分钟)
            ccm.localCache.SetWithTTL(key, &template, 1, 10*time.Minute)
            log.Debugf("命中L2缓存, couponID: %d", couponID)
            return &template, nil
        }
    }
    
    // L3: 数据库查询
    template, err := ccm.repository.GetCouponTemplate(ctx, couponID)
    if err != nil {
        return nil, err
    }
    
    // 回写L2缓存 (TTL 30分钟)
    data, _ := json.Marshal(template)
    ccm.redis.SetEX(ctx, key, data, 30*time.Minute)
    
    // 回写L1缓存 (成本为1，TTL 10分钟)  
    ccm.localCache.SetWithTTL(key, template, 1, 10*time.Minute)
    
    log.Debugf("命中L3数据库, couponID: %d", couponID)
    return template, nil
}

// InvalidateCache 缓存失效 (Canal调用)
func (ccm *CouponCacheManager) InvalidateCache(keys ...string) {
    for _, key := range keys {
        // 删除L1缓存
        ccm.localCache.Del(key)
        // 删除L2缓存  
        ccm.redis.Del(context.Background(), key)
        log.Infof("缓存失效: %s", key)
    }
}

// WarmupCache 缓存预热
func (ccm *CouponCacheManager) WarmupCache(ctx context.Context) error {
    log.Info("开始缓存预热...")
    
    // 查询热门优惠券模板
    hotCoupons, err := ccm.repository.GetHotCouponTemplates(ctx, 100)
    if err != nil {
        return fmt.Errorf("获取热门优惠券失败: %v", err)
    }
    
    // 批量预热到L1和L2缓存
    for _, coupon := range hotCoupons {
        key := fmt.Sprintf("coupon:template:%d", coupon.ID)
        
        // 写入L2 Redis缓存
        data, _ := json.Marshal(coupon)
        ccm.redis.SetEX(ctx, key, data, 30*time.Minute)
        
        // 写入L1 Ristretto缓存 (高成本保证不被淘汰)
        ccm.localCache.SetWithTTL(key, coupon, 10, 10*time.Minute)
    }
    
    log.Infof("缓存预热完成，预热%d个优惠券模板", len(hotCoupons))
    return nil
}

// GetCacheStats 获取缓存统计信息
func (ccm *CouponCacheManager) GetCacheStats() map[string]interface{} {
    metrics := ccm.localCache.Metrics
    
    return map[string]interface{}{
        "ristretto_hits":        metrics.Hits(),
        "ristretto_misses":      metrics.Misses(),
        "ristretto_hit_ratio":   metrics.Ratio(),
        "ristretto_keys_added":  metrics.KeysAdded(),
        "ristretto_keys_evicted": metrics.KeysEvicted(),
        "ristretto_cost_added":  metrics.CostAdded(),
        "ristretto_cost_evicted": metrics.CostEvicted(),
    }
}
```

### 3. Canal缓存一致性集成

```go
package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"
    
    "github.com/apache/rocketmq-client-go/v2/consumer"
    "github.com/apache/rocketmq-client-go/v2/primitive"
    "emshop/internal/app/coupon/srv/pkg/cache"
    "emshop/pkg/log"
)

// CouponCanalConsumer 优惠券Canal消费者
type CouponCanalConsumer struct {
    cacheManager *cache.CouponCacheManager
    consumer     rocketmq.PushConsumer
}

// 监听的数据库表
var WatchTables = map[string]bool{
    "coupon_templates":      true, // 优惠券模板
    "user_coupons":         true, // 用户优惠券
    "flash_sale_activities": true, // 秒杀活动
}

func NewCouponCanalConsumer(cacheManager *cache.CouponCacheManager) *CouponCanalConsumer {
    return &CouponCanalConsumer{
        cacheManager: cacheManager,
    }
}

// ConsumeCanalMessage 消费Canal消息，实现缓存一致性
func (ccc *CouponCanalConsumer) ConsumeCanalMessage(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
    for _, msg := range msgs {
        var canalMsg CanalMessage
        if err := json.Unmarshal(msg.Body, &canalMsg); err != nil {
            log.Errorf("Canal消息解析失败: %v", err)
            continue
        }
        
        // 只处理优惠券相关表
        if !WatchTables[canalMsg.Table] {
            continue
        }
        
        log.Infof("收到Canal消息: database=%s, table=%s, type=%s", 
            canalMsg.Database, canalMsg.Table, canalMsg.Type)
        
        // 根据表名和操作类型处理缓存更新
        switch canalMsg.Table {
        case "coupon_templates":
            ccc.handleCouponTemplateChange(&canalMsg)
        case "user_coupons":
            ccc.handleUserCouponChange(&canalMsg)
        case "flash_sale_activities":
            ccc.handleFlashSaleChange(&canalMsg)
        }
    }
    
    return consumer.ConsumeSuccess, nil
}

// handleCouponTemplateChange 处理优惠券模板变更
func (ccc *CouponCanalConsumer) handleCouponTemplateChange(msg *CanalMessage) {
    for _, data := range msg.Data {
        couponIDStr, ok := data["id"].(string)
        if !ok {
            continue
        }
        
        couponID, err := strconv.ParseInt(couponIDStr, 10, 64)
        if err != nil {
            continue
        }
        
        // 构建需要失效的缓存key
        keys := []string{
            fmt.Sprintf("coupon:template:%d", couponID),
        }
        
        // 如果是删除操作，还需要清理相关缓存
        if msg.Type == "DELETE" {
            keys = append(keys, 
                fmt.Sprintf("coupon:list:user:*"), // 用户可用优惠券列表
                fmt.Sprintf("coupon:valid:%d:*", couponID), // 优惠券有效性缓存
            )
        }
        
        // 执行缓存失效
        ccc.cacheManager.InvalidateCache(keys...)
        
        log.Infof("优惠券模板缓存失效: couponID=%d, type=%s", couponID, msg.Type)
    }
}

// handleUserCouponChange 处理用户优惠券变更  
func (ccc *CouponCanalConsumer) handleUserCouponChange(msg *CanalMessage) {
    for _, data := range msg.Data {
        userIDStr, ok := data["user_id"].(string)
        if !ok {
            continue
        }
        
        userID, err := strconv.ParseInt(userIDStr, 10, 64)
        if err != nil {
            continue
        }
        
        // 失效用户相关缓存
        keys := []string{
            fmt.Sprintf("coupon:user:list:%d", userID),      // 用户优惠券列表
            fmt.Sprintf("coupon:user:available:%d", userID), // 用户可用优惠券
            fmt.Sprintf("coupon:user:count:%d", userID),     // 用户优惠券数量
        }
        
        ccc.cacheManager.InvalidateCache(keys...)
        log.Infof("用户优惠券缓存失效: userID=%d, type=%s", userID, msg.Type)
    }
}

// handleFlashSaleChange 处理秒杀活动变更
func (ccc *CouponCanalConsumer) handleFlashSaleChange(msg *CanalMessage) {
    for _, data := range msg.Data {
        activityIDStr, ok := data["id"].(string)
        if !ok {
            continue
        }
        
        activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
        if err != nil {
            continue
        }
        
        // 失效秒杀活动相关缓存
        keys := []string{
            fmt.Sprintf("flashsale:activity:%d", activityID),  // 秒杀活动信息
            fmt.Sprintf("flashsale:status:%d", activityID),    // 秒杀状态
        }
        
        ccc.cacheManager.InvalidateCache(keys...)
        log.Infof("秒杀活动缓存失效: activityID=%d, type=%s", activityID, msg.Type)
    }
}
```

### 4. 高性能库存管理器

```go
package redis

import (
    "context"
    "emshop/pkg/log"
    "github.com/go-redis/redis/v8"
    "time"
)

type StockManager struct {
    redis  *redis.Client
    script *redis.Script
}

func NewStockManager(rdb *redis.Client) *StockManager {
    // 加载Lua脚本
    script := redis.NewScript(flashSaleLuaScript)
    
    return &StockManager{
        redis:  rdb,
        script: script,
    }
}

// FlashSale 执行秒杀
func (sm *StockManager) FlashSale(ctx context.Context, req *FlashSaleRequest) (*FlashSaleResult, error) {
    keys := []string{
        fmt.Sprintf("coupon:stock:%d", req.CouponID),
        fmt.Sprintf("coupon:user:%d:%d", req.ActivityID, req.UserID),
        fmt.Sprintf("coupon:log:%d", req.ActivityID),
    }
    
    args := []interface{}{
        req.UserID,
        req.ActivityID, 
        1,              // 固定扣减1个
        1800,           // 30分钟TTL
    }
    
    // 执行Lua脚本
    result, err := sm.script.Run(ctx, sm.redis, keys, args...).Result()
    if err != nil {
        log.Errorf("秒杀执行失败: %v", err)
        return nil, err
    }
    
    // 解析结果
    resultSlice := result.([]interface{})
    code := resultSlice[0].(int64)
    stock := resultSlice[1].(int64)
    message := resultSlice[2].(string)
    
    return &FlashSaleResult{
        Code:      int(code),
        Stock:     int(stock),
        Message:   message,
        Success:   code == 1,
        Timestamp: time.Now().Unix(),
    }, nil
}

// PrewarmStock 预热库存到Redis
func (sm *StockManager) PrewarmStock(ctx context.Context, couponID int64, totalStock int) error {
    key := fmt.Sprintf("coupon:stock:%d", couponID)
    
    // 设置库存，TTL为1小时
    err := sm.redis.Set(ctx, key, totalStock, time.Hour).Err()
    if err != nil {
        log.Errorf("库存预热失败, couponID: %d, err: %v", couponID, err)
        return err
    }
    
    log.Infof("库存预热成功, couponID: %d, stock: %d", couponID, totalStock)
    return nil
}
```

### 3. 异步消息处理器

```go 
package consumer

import (
    "context"
    "encoding/json"
    "emshop/internal/app/coupon/srv/service/v1"
    "emshop/pkg/log"
    "github.com/apache/rocketmq-client-go/v2/consumer"
    "github.com/apache/rocketmq-client-go/v2/primitive"
)

type CouponConsumer struct {
    couponSrv v1.CouponServiceInterface
    consumer  rocketmq.PushConsumer
}

func NewCouponConsumer(couponSrv v1.CouponServiceInterface) *CouponConsumer {
    return &CouponConsumer{
        couponSrv: couponSrv,
    }
}

// 处理秒杀成功消息
func (cc *CouponConsumer) HandleFlashSaleSuccess(ctx context.Context, 
    msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
    
    for _, msg := range msgs {
        var event FlashSaleSuccessEvent
        if err := json.Unmarshal(msg.Body, &event); err != nil {
            log.Errorf("消息解析失败: %v", err)
            continue
        }
        
        // 异步创建用户优惠券记录
        err := cc.couponSrv.CreateUserCouponAsync(ctx, &CreateUserCouponRequest{
            CouponID:   event.CouponID,
            UserID:     event.UserID,
            ActivityID: event.ActivityID,
            CouponSn:   event.CouponSn,
            Source:     "flash_sale",
        })
        
        if err != nil {
            log.Errorf("创建用户优惠券失败: %v", err)
            return consumer.ConsumeRetryLater, err
        }
        
        log.Infof("异步创建用户优惠券成功, userID: %d, couponSn: %s", 
            event.UserID, event.CouponSn)
    }
    
    return consumer.ConsumeSuccess, nil
}

type FlashSaleSuccessEvent struct {
    CouponID   int64  `json:"coupon_id"`
    UserID     int64  `json:"user_id"`
    ActivityID int64  `json:"activity_id"`
    CouponSn   string `json:"coupon_sn"`
    Timestamp  int64  `json:"timestamp"`
}
```

### 4. DTM Saga事务管理器

```go
package v1

import (
    "context"
    "emshop/internal/app/pkg/options"
    "emshop/pkg/log"
    "github.com/dtm-labs/client/dtmcli"
    "github.com/dtm-labs/client/dtmcli/dtmimp"
)

type DTMManager struct {
    dtmServer  string
    couponSrv  string  
    orderSrv   string
    paymentSrv string
}

func NewDTMManager(dtmOpts *options.DtmOptions) *DTMManager {
    return &DTMManager{
        dtmServer:  dtmOpts.GrpcServer,
        couponSrv:  "discovery:///emshop-coupon-srv",
        orderSrv:   "discovery:///emshop-order-srv",
        paymentSrv: "discovery:///emshop-payment-srv",
    }
}

// ProcessCouponPayment 处理优惠券支付分布式事务
func (dm *DTMManager) ProcessCouponPayment(ctx context.Context, req *CouponPaymentRequest) error {
    // 生成全局事务ID
    gid := dtmcli.MustGenGid(dm.dtmServer)
    
    // 创建Saga事务
    saga := dtmcli.NewSaga(dm.dtmServer, gid).
        Add(dm.couponSrv+"/LockUserCoupon", dm.couponSrv+"/UnlockUserCoupon", &LockCouponRequest{
            UserCouponId: req.UserCouponId,
            OrderSn:      req.OrderSn,
            LockTimeout:  900, // 15分钟锁定
        }).
        Add(dm.orderSrv+"/CalculateOrderDiscount", dm.orderSrv+"/RevertOrderDiscount", &CalculateDiscountRequest{
            OrderSn:      req.OrderSn,
            CouponIds:    req.CouponIds,
            OriginalAmount: req.OriginalAmount,
        }).
        Add(dm.paymentSrv+"/CreatePaymentWithDiscount", dm.paymentSrv+"/CancelPaymentOrder", &CreatePaymentRequest{
            OrderSn:        req.OrderSn,
            UserId:         req.UserId,
            OriginalAmount: req.OriginalAmount,
            DiscountAmount: req.DiscountAmount,
            FinalAmount:    req.FinalAmount,
        })
    
    // 提交事务
    err := saga.Submit()
    if err != nil {
        log.Errorf("优惠券支付事务提交失败, gid: %s, err: %v", gid, err)
        return err
    }
    
    log.Infof("优惠券支付事务提交成功, gid: %s, orderSn: %s", gid, req.OrderSn)
    return nil
}
```


## 📝 总结

本开发计划专注于优惠券服务的核心功能完善和可用性验证。通过**分场景设计**的策略，在保证高性能的同时，确保系统的功能完整性和数据一致性。

### 🎯 核心亮点

1. **高性能秒杀**：基于Redis Lua脚本的原子操作，确保零超卖
2. **分布式事务**：DTM Saga模式确保支付场景的数据强一致性  
3. **架构清晰**：简单场景简单处理，复杂场景用分布式事务
4. **消息可靠**：RocketMQ事务消息保证业务一致性

### 🚀 交付成果

- **功能完整**: 优惠券管理、秒杀活动、分布式事务支付全流程可用
- **数据一致**: 严格的库存控制和事务保障
- **消息可靠**: RocketMQ事务消息确保数据一致性
- **测试验证**: 完整的功能回归测试通过

通过简化的开发计划，我们将交付一个功能完整、稳定可靠的优惠券服务，为EMShop电商系统提供核心营销能力支持。

## 📝 技术实现亮点

基于本项目的核心实现，技术亮点包括：

**"基于Ristretto+Redis+MySQL三层缓存架构的高并发优惠券秒杀系统，采用TinyLFU算法实现高命中率本地缓存，结合Canal数据同步机制保证缓存一致性，使用Redis Lua脚本实现零超卖的精确库存控制"**

### 关键技术特性
- **Ristretto本地缓存**: TinyLFU算法，高效的内存缓存管理
- **Canal缓存一致性**: 实时数据同步，保证缓存数据准确性  
- **三层缓存协调**: L1+L2+L3完整缓存体系，优化系统性能
- **DTM分布式事务**: 支付场景的强一致性保障
- **Redis Lua脚本**: 原子操作确保库存零超卖
- **RocketMQ事务消息**: 业务操作与消息发送的一致性保障

---

<div align="center">

**EMShop优惠券服务 - 高性能营销解决方案**

Created with ❤️ by EMShop Team

</div>