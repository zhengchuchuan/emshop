# 支付服务模拟设计文档

## 1. 概述

本文档设计了一个适用于学习项目的支付服务模拟系统，旨在模拟真实电商系统的支付流程，但无需对接真实的第三方支付接口。

## 2. 设计目标

- 模拟完整的支付流程体验
- 支持多种支付方式选择
- 提供支付状态管理
- 集成到现有的分布式事务架构
- 便于学习和演示

## 3. 支付状态定义

```go
type PaymentStatus int32

const (
    PaymentPending   PaymentStatus = 1 // 待支付
    PaymentPaid      PaymentStatus = 2 // 支付成功
    PaymentFailed    PaymentStatus = 3 // 支付失败
    PaymentCancelled PaymentStatus = 4 // 已取消
    PaymentRefunding PaymentStatus = 5 // 退款中
    PaymentRefunded  PaymentStatus = 6 // 已退款
)
```

## 4. 支付方式定义

```go
type PaymentMethod int32

const (
    PaymentMethodWechat    PaymentMethod = 1 // 微信支付
    PaymentMethodAlipay    PaymentMethod = 2 // 支付宝
    PaymentMethodUnionPay  PaymentMethod = 3 // 银联支付
    PaymentMethodBank      PaymentMethod = 4 // 网银支付
    PaymentMethodBalance   PaymentMethod = 5 // 余额支付
)
```

## 5. 数据模型设计

### 5.1 支付订单表 (payment_orders)

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

### 5.2 支付记录表 (payment_logs)

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

## 6. 服务接口设计

### 6.1 gRPC 服务定义

```protobuf
service Payment {
    // 创建支付订单
    rpc CreatePayment(CreatePaymentRequest) returns (CreatePaymentResponse);
    
    // 查询支付状态
    rpc GetPaymentStatus(GetPaymentStatusRequest) returns (GetPaymentStatusResponse);
    
    // 模拟支付成功（仅用于测试和演示）
    rpc SimulatePaymentSuccess(SimulatePaymentRequest) returns (google.protobuf.Empty);
    
    // 模拟支付失败（仅用于测试和演示）
    rpc SimulatePaymentFailure(SimulatePaymentRequest) returns (google.protobuf.Empty);
    
    // 取消支付
    rpc CancelPayment(CancelPaymentRequest) returns (google.protobuf.Empty);
    
    // 申请退款
    rpc RequestRefund(RefundRequest) returns (RefundResponse);
    
    // 查询退款状态
    rpc GetRefundStatus(GetRefundStatusRequest) returns (GetRefundStatusResponse);
}
```

### 6.2 消息定义

```protobuf
message CreatePaymentRequest {
    string order_sn = 1;
    int32 user_id = 2;
    double amount = 3;
    int32 payment_method = 4;
    int32 expired_minutes = 5; // 支付过期时间（分钟）
}

message CreatePaymentResponse {
    string payment_sn = 1;
    string payment_url = 2; // 模拟支付链接
    int64 expired_at = 3;
}

message GetPaymentStatusRequest {
    string payment_sn = 1;
}

message GetPaymentStatusResponse {
    string payment_sn = 1;
    int32 payment_status = 2;
    double amount = 3;
    int32 payment_method = 4;
    int64 paid_at = 5;
}

message SimulatePaymentRequest {
    string payment_sn = 1;
    optional string third_party_sn = 2;
}

message CancelPaymentRequest {
    string payment_sn = 1;
}

message RefundRequest {
    string payment_sn = 1;
    double refund_amount = 2;
    string reason = 3;
}

message RefundResponse {
    string refund_sn = 1;
    int32 refund_status = 2;
}

message GetRefundStatusRequest {
    string refund_sn = 1;
}

message GetRefundStatusResponse {
    string refund_sn = 1;
    int32 refund_status = 2;
    double refund_amount = 3;
    int64 refunded_at = 4;
}
```

## 7. 核心业务逻辑

### 7.1 支付订单创建流程

1. 验证订单是否存在且状态为待支付
2. 生成唯一的支付单号
3. 设置支付过期时间（默认15分钟）
4. 创建支付订单记录
5. 返回支付信息（包含模拟的支付链接）

### 7.2 支付状态查询

1. 根据支付单号查询支付订单
2. 返回当前支付状态和相关信息
3. 如果支付已过期，自动更新状态为已取消

### 7.3 模拟支付成功

1. 验证支付单状态为待支付
2. 生成模拟的第三方支付单号
3. 更新支付状态为支付成功
4. 记录支付完成时间
5. 触发支付成功回调（更新订单状态）
6. 记录操作日志

### 7.4 模拟支付失败

1. 验证支付单状态为待支付
2. 更新支付状态为支付失败
3. 记录失败原因
4. 记录操作日志

## 8. 定时任务设计

### 8.1 支付过期处理

```go
// 每分钟检查一次过期的支付订单
func ProcessExpiredPayments() {
    // 查询已过期但状态仍为待支付的订单
    expiredPayments := findExpiredPendingPayments()
    
    for _, payment := range expiredPayments {
        // 更新状态为已取消
        updatePaymentStatus(payment.PaymentSn, PaymentCancelled)
        
        // 记录日志
        logPaymentAction(payment.PaymentSn, "auto_cancel", "支付过期自动取消")
    }
}
```

### 8.2 自动支付模拟（可选）

```go
// 模拟随机支付成功/失败，用于演示
func AutoSimulatePayments() {
    pendingPayments := findPendingPayments()
    
    for _, payment := range pendingPayments {
        // 根据一定概率模拟支付结果
        if rand.Float32() < 0.8 { // 80%成功率
            simulatePaymentSuccess(payment.PaymentSn)
        } else {
            simulatePaymentFailure(payment.PaymentSn)
        }
    }
}
```

## 9. 与订单服务的集成

### 9.1 支付成功回调

当支付成功时，调用订单服务更新订单状态：

```go
func OnPaymentSuccess(paymentSn string, orderSn string) error {
    // 调用订单服务的支付成功接口
    _, err := orderClient.PaymentSuccess(context.Background(), &order.PaymentSuccessRequest{
        OrderSn:   orderSn,
        PaymentSn: paymentSn,
        PaidAt:    time.Now().Unix(),
    })
    return err
}
```

### 9.2 支付失败处理

当支付失败时，通知订单服务：

```go
func OnPaymentFailure(paymentSn string, orderSn string) error {
    // 调用订单服务的支付失败接口
    _, err := orderClient.PaymentFailure(context.Background(), &order.PaymentFailureRequest{
        OrderSn:   orderSn,
        PaymentSn: paymentSn,
        Reason:    "支付失败",
    })
    return err
}
```

## 10. 前端集成

### 10.1 支付页面设计

1. 显示订单信息和金额
2. 提供支付方式选择
3. 生成模拟的支付二维码或链接
4. 轮询查询支付状态
5. 显示支付结果

### 10.2 管理后台功能

1. 支付订单列表查询
2. 支付状态统计
3. 手动模拟支付成功/失败（用于测试）
4. 支付日志查看

## 11. 监控和日志

### 11.1 关键指标监控

- 支付成功率
- 支付响应时间
- 支付过期率
- 退款成功率

### 11.2 日志记录

- 支付订单创建日志
- 支付状态变更日志
- 支付回调日志
- 异常错误日志

## 12. 安全考虑

虽然是模拟支付，但仍需注意：

1. 支付单号生成要保证唯一性
2. 金额字段要进行精度控制
3. 状态变更要进行合理性验证
4. 接口要进行身份认证
5. 敏感信息要进行日志脱敏

## 13. 扩展性设计

为未来对接真实支付接口预留扩展空间：

1. 使用适配器模式封装支付逻辑
2. 配置化支付方式启用/禁用
3. 支持插件式支付方式扩展
4. 预留第三方回调接口

## 14. 测试策略

### 14.1 单元测试

- 支付订单创建测试
- 状态流转测试
- 过期处理测试
- 回调处理测试

### 14.2 集成测试

- 与订单服务集成测试
- 分布式事务测试
- 并发支付测试

### 14.3 演示场景

- 正常支付流程演示
- 支付失败处理演示
- 支付过期处理演示
- 退款流程演示

这个设计方案提供了一个完整的支付服务模拟框架，既保持了真实支付系统的核心逻辑，又避免了复杂的第三方接口对接，非常适合用于学习和演示电商系统的支付功能。