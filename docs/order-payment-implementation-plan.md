# 基于DTM的订单支付分布式事务完善方案

## 1. 总体架构设计

### 1.1 事务拆分策略
将现有的订单创建流程重构为两个独立的分布式事务：

**事务1：订单提交事务（Order Submit Saga）**
```
Step1: 创建订单(待支付状态) - Order/CreateOrder ↔ Order/CreateOrderCom
Step2: 创建支付订单 - Payment/CreatePayment ↔ Payment/CancelPayment  
Step3: 预留库存 - Inventory/ReserveStock ↔ Inventory/ReleaseReserved
```

**事务2：支付成功事务（Payment Success Saga）**
```
Step1: 确认支付成功 - Payment/ConfirmPayment ↔ Payment/RefundPayment
Step2: 更新订单状态为已支付 - Order/UpdatePaidStatus ↔ Order/RevertPaidStatus
Step3: 确认扣减库存 - Inventory/ConfirmSell ↔ Inventory/Reback
Step4: 创建物流订单(可选) - Logistics/CreateOrder ↔ Logistics/CancelOrder
```

### 1.2 服务间协作流程
```
用户下单 
    ↓
订单提交事务 (Saga1)
    ↓ 成功
等待用户支付 (15分钟超时)
    ↓ 支付成功
支付成功事务 (Saga2) 
    ↓ 成功
订单完成，进入物流环节
```

## 2. 实施步骤

### Phase 1: 保存计划文档并创建支付服务框架 (第1天)
1. **保存实施计划**到 `/home/zcc/project/golang/emshop/emshop/docs/order-payment-implementation-plan.md`
2. **创建支付服务目录结构**
   ```
   internal/app/payment/
   ├── srv/
   │   ├── app.go
   │   ├── rpc.go  
   │   ├── controller/payment/v1/
   │   ├── service/v1/
   │   ├── data/v1/mysql/
   │   └── domain/
   cmd/payment/
   └── payment.go
   api/payment/v1/
   └── payment.proto
   ```
3. **定义支付服务gRPC接口**
4. **创建支付相关数据表**

### Phase 2: 实现支付服务核心功能 (第2天)
1. **实现支付服务基础功能**
   - CreatePayment() - 创建支付订单
   - CancelPayment() - 取消支付订单（补偿）
   - GetPaymentStatus() - 查询支付状态
   - ConfirmPayment() - 确认支付成功
   - RefundPayment() - 退款（补偿）
2. **实现支付状态管理**
3. **添加支付日志记录**

### Phase 3: 库存服务改造 (第3天)  
1. **新增库存预留接口**
   - ReserveStock() - 预留库存（基于现有TrySell改造）
   - ReleaseReserved() - 释放预留库存（补偿）
2. **完善TCC模式的ConfirmSell接口**
3. **优化库存状态管理**

### Phase 4: 订单服务改造 (第4天)
1. **扩展订单数据模型**
   ```sql
   ALTER TABLE order_info 
   ADD COLUMN payment_status TINYINT DEFAULT 0,
   ADD COLUMN payment_sn VARCHAR(64),
   ADD COLUMN paid_at TIMESTAMP NULL;
   ```
2. **新增订单支付状态管理接口**
   - UpdatePaidStatus() - 更新订单为已支付
   - RevertPaidStatus() - 回滚支付状态（补偿）
3. **重构订单提交流程**

### Phase 5: 分布式事务集成 (第5天)
1. **实现订单提交Saga事务**
   ```go
   // 新的Submit流程：创建订单 → 创建支付 → 预留库存
   saga := dtmgrpc.NewSagaGrpc(dtmServer, orderSn).
       Add(orderSrv+"/CreateOrder", orderSrv+"/CreateOrderCom", orderReq).
       Add(paymentSrv+"/CreatePayment", paymentSrv+"/CancelPayment", paymentReq).
       Add(inventorySrv+"/ReserveStock", inventorySrv+"/ReleaseReserved", stockReq)
   ```
2. **实现支付成功Saga事务**
   ```go
   // 支付成功流程：确认支付 → 更新订单 → 确认扣库存
   saga := dtmgrpc.NewSagaGrpc(dtmServer, paymentSn).
       Add(paymentSrv+"/ConfirmPayment", paymentSrv+"/RefundPayment", confirmReq).
       Add(orderSrv+"/UpdatePaidStatus", orderSrv+"/RevertPaidStatus", updateReq).
       Add(inventorySrv+"/ConfirmSell", inventorySrv+"/Reback", sellReq)
   ```

### Phase 6: 自动化和完善功能 (第6天)
1. **实现支付超时自动取消**
2. **实现库存预留超时释放**
3. **完善补偿机制和异常处理**
4. **添加完整的监控日志**

### Phase 7: 测试和优化 (第7天)
1. **单元测试编写**
2. **集成测试验证**
3. **分布式事务场景测试**
4. **性能测试和优化**

## 3. 核心技术要点

### 3.1 关键改进点
1. **业务流程优化**：将库存扣减延后到支付成功后执行
2. **状态管理完善**：订单、支付、库存状态协同管理
3. **异常处理增强**：完善的补偿和回滚机制
4. **超时处理**：自动处理支付超时和库存预留超时

### 3.2 技术难点解决
1. **幂等性保证**：使用业务唯一ID确保操作幂等
2. **状态一致性**：通过DTM Saga确保最终一致性  
3. **并发控制**：使用分布式锁避免库存超卖
4. **补偿机制**：每个操作都有对应的补偿操作

### 3.3 数据模型设计

#### 支付订单表 (payment_orders)
```sql
CREATE TABLE payment_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_sn VARCHAR(64) NOT NULL UNIQUE COMMENT '支付单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    amount DECIMAL(10,2) NOT NULL COMMENT '支付金额',
    payment_method TINYINT NOT NULL COMMENT '支付方式',
    payment_status TINYINT NOT NULL DEFAULT 1 COMMENT '支付状态',
    third_party_sn VARCHAR(128) COMMENT '第三方支付单号(模拟)',
    paid_at TIMESTAMP NULL COMMENT '支付完成时间',
    expired_at TIMESTAMP NOT NULL COMMENT '支付过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_payment_sn (payment_sn),
    INDEX idx_order_sn (order_sn),
    INDEX idx_user_id (user_id),
    INDEX idx_status (payment_status)
);
```

#### 支付记录表 (payment_logs)
```sql
CREATE TABLE payment_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_sn VARCHAR(64) NOT NULL COMMENT '支付单号',
    action VARCHAR(32) NOT NULL COMMENT '操作类型',
    status_from TINYINT COMMENT '状态变更前',
    status_to TINYINT COMMENT '状态变更后',
    remark TEXT COMMENT '备注信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_payment_sn (payment_sn)
);
```

### 3.4 核心接口定义

#### 支付服务接口
```go
type PaymentSrv interface {
    // Saga正向操作
    CreatePayment(ctx context.Context, req *CreatePaymentRequest) error
    ConfirmPayment(ctx context.Context, req *ConfirmPaymentRequest) error
    
    // Saga补偿操作
    CancelPayment(ctx context.Context, req *CancelPaymentRequest) error
    RefundPayment(ctx context.Context, req *RefundPaymentRequest) error
    
    // 查询操作
    GetPaymentStatus(ctx context.Context, req *GetPaymentStatusRequest) (*PaymentStatusResponse, error)
}
```

#### 库存服务扩展接口
```go
type InventorySrv interface {
    // 现有接口
    Sell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
    Reback(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
    
    // 新增预留接口
    ReserveStock(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
    ReleaseReserved(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
    
    // TCC接口（已实现，用于支付成功后确认扣减）
    TrySell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
    ConfirmSell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
    CancelSell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error
}
```

### 3.5 状态管理

#### 支付状态定义
```go
const (
    PaymentPending   = 1 // 待支付
    PaymentPaid      = 2 // 支付成功
    PaymentFailed    = 3 // 支付失败
    PaymentCancelled = 4 // 已取消
    PaymentRefunding = 5 // 退款中
    PaymentRefunded  = 6 // 已退款
)
```

#### 订单状态扩展
```go
const (
    OrderStatusPending    = 1  // 待支付
    OrderStatusPaid       = 2  // 已支付
    OrderStatusShipped    = 3  // 已发货
    OrderStatusDelivered  = 4  // 已送达
    OrderStatusCompleted  = 5  // 已完成
    OrderStatusCancelled  = 6  // 已取消
)
```

## 4. 实施注意事项

### 4.1 兼容性考虑
1. **保持现有接口兼容**：新功能作为扩展，不影响现有功能
2. **数据迁移策略**：新增字段使用默认值，确保现有数据正常
3. **渐进式部署**：先部署新服务，再切换流程

### 4.2 测试策略
1. **单元测试**：每个服务接口都要有完整的单元测试
2. **集成测试**：验证分布式事务的正向流程和补偿流程
3. **异常测试**：模拟各种异常情况，验证补偿机制
4. **性能测试**：验证分布式事务对性能的影响

### 4.3 监控和日志
1. **业务日志**：记录每个Saga步骤的执行情况
2. **状态跟踪**：记录订单、支付、库存状态变更
3. **异常监控**：监控事务失败和补偿执行情况
4. **性能监控**：监控分布式事务执行时间

## 5. 预期结果

完成后将实现：
1. **标准电商流程**：下单 → 支付 → 发货的完整闭环
2. **数据一致性保证**：跨服务的强一致性
3. **高可用性**：完善的异常处理和恢复机制
4. **良好扩展性**：模块化设计，便于后续扩展

这个方案基于现有架构，采用渐进式改造，风险可控，预计7天完成核心功能。