# 订单状态流转设计文档

## 1. 概述

本文档设计了完整的电商订单状态流转机制，整合支付服务和物流服务，确保订单从创建到完成的全生命周期管理。

## 2. 设计目标

- 定义完整的订单状态流转规则
- 集成支付和物流服务
- 支持异常情况处理
- 提供状态回滚机制
- 确保数据一致性

## 3. 订单状态定义

### 3.1 订单主状态

```go
type OrderStatus int32

const (
    OrderStatusPending    OrderStatus = 1  // 待支付
    OrderStatusPaid       OrderStatus = 2  // 已支付
    OrderStatusShipped    OrderStatus = 3  // 已发货
    OrderStatusDelivered  OrderStatus = 4  // 已送达
    OrderStatusCompleted  OrderStatus = 5  // 已完成
    OrderStatusCancelled  OrderStatus = 6  // 已取消
    OrderStatusRefunding  OrderStatus = 7  // 退款中
    OrderStatusRefunded   OrderStatus = 8  // 已退款
    OrderStatusReturning  OrderStatus = 9  // 退货中
    OrderStatusReturned   OrderStatus = 10 // 已退货
)
```

### 3.2 订单子状态（支付相关）

```go
type PaymentSubStatus int32

const (
    PaymentSubStatusNone      PaymentSubStatus = 0  // 无
    PaymentSubStatusCreated   PaymentSubStatus = 1  // 支付订单已创建
    PaymentSubStatusPaying    PaymentSubStatus = 2  // 支付中
    PaymentSubStatusPaid      PaymentSubStatus = 3  // 支付成功
    PaymentSubStatusFailed    PaymentSubStatus = 4  // 支付失败
    PaymentSubStatusExpired   PaymentSubStatus = 5  // 支付过期
)
```

### 3.3 订单子状态（物流相关）

```go
type LogisticsSubStatus int32

const (
    LogisticsSubStatusNone       LogisticsSubStatus = 0  // 无
    LogisticsSubStatusPreparing  LogisticsSubStatus = 1  // 备货中
    LogisticsSubStatusShipped    LogisticsSubStatus = 2  // 已发货
    LogisticsSubStatusInTransit  LogisticsSubStatus = 3  // 运输中
    LogisticsSubStatusDelivering LogisticsSubStatus = 4  // 配送中
    LogisticsSubStatusDelivered  LogisticsSubStatus = 5  // 已送达
    LogisticsSubStatusRejected   LogisticsSubStatus = 6  // 拒收
)
```

## 4. 状态流转图

```
订单创建 → 待支付 → 已支付 → 已发货 → 已送达 → 已完成
    ↓        ↓        ↓        ↓        ↓
   取消    支付超时   申请退款  申请退货   申请退货
             ↓        ↓        ↓        ↓
           已取消    退款中    退货中    退货中
                      ↓        ↓        ↓
                    已退款    已退货    已退货
```

## 5. 数据模型扩展

### 5.1 订单表扩展

```sql
ALTER TABLE order_info ADD COLUMN payment_status TINYINT DEFAULT 0 COMMENT '支付子状态';
ALTER TABLE order_info ADD COLUMN logistics_status TINYINT DEFAULT 0 COMMENT '物流子状态';
ALTER TABLE order_info ADD COLUMN payment_sn VARCHAR(64) COMMENT '支付单号';
ALTER TABLE order_info ADD COLUMN logistics_sn VARCHAR(64) COMMENT '物流单号';
ALTER TABLE order_info ADD COLUMN tracking_number VARCHAR(64) COMMENT '快递单号';
ALTER TABLE order_info ADD COLUMN paid_at TIMESTAMP NULL COMMENT '支付时间';
ALTER TABLE order_info ADD COLUMN shipped_at TIMESTAMP NULL COMMENT '发货时间';
ALTER TABLE order_info ADD COLUMN delivered_at TIMESTAMP NULL COMMENT '送达时间';
ALTER TABLE order_info ADD COLUMN completed_at TIMESTAMP NULL COMMENT '完成时间';
ALTER TABLE order_info ADD COLUMN cancelled_at TIMESTAMP NULL COMMENT '取消时间';
ALTER TABLE order_info ADD COLUMN cancel_reason TEXT COMMENT '取消原因';
ALTER TABLE order_info ADD COLUMN auto_complete_at TIMESTAMP NULL COMMENT '自动完成时间';

-- 添加索引
ALTER TABLE order_info ADD INDEX idx_payment_sn (payment_sn);
ALTER TABLE order_info ADD INDEX idx_logistics_sn (logistics_sn);
ALTER TABLE order_info ADD INDEX idx_tracking_number (tracking_number);
ALTER TABLE order_info ADD INDEX idx_payment_status (payment_status);
ALTER TABLE order_info ADD INDEX idx_logistics_status (logistics_status);
```

### 5.2 订单状态变更日志表

```sql
CREATE TABLE order_status_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    status_from TINYINT COMMENT '变更前状态',
    status_to TINYINT NOT NULL COMMENT '变更后状态',
    sub_status_from TINYINT COMMENT '变更前子状态',
    sub_status_to TINYINT COMMENT '变更后子状态',
    change_type ENUM('payment', 'logistics', 'manual', 'system') NOT NULL COMMENT '变更类型',
    operator_id INT COMMENT '操作员ID',
    operator_type ENUM('user', 'admin', 'system') NOT NULL COMMENT '操作员类型',
    remark TEXT COMMENT '变更说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_order_sn (order_sn),
    INDEX idx_created_at (created_at)
);
```

## 6. 状态流转规则

### 6.1 状态流转矩阵

```go
type StatusTransition struct {
    From      OrderStatus
    To        OrderStatus
    Condition func(*Order) bool
    Handler   func(*Order) error
}

var allowedTransitions = []StatusTransition{
    // 正常流程
    {OrderStatusPending, OrderStatusPaid, checkPaymentSuccess, handlePaymentSuccess},
    {OrderStatusPaid, OrderStatusShipped, checkCanShip, handleShipment},
    {OrderStatusShipped, OrderStatusDelivered, checkDelivered, handleDelivery},
    {OrderStatusDelivered, OrderStatusCompleted, checkCanComplete, handleCompletion},
    
    // 取消流程
    {OrderStatusPending, OrderStatusCancelled, checkCanCancel, handleCancellation},
    {OrderStatusPaid, OrderStatusCancelled, checkCanCancelPaid, handlePaidCancellation},
    
    // 退款流程
    {OrderStatusPaid, OrderStatusRefunding, checkCanRefund, handleRefundStart},
    {OrderStatusShipped, OrderStatusRefunding, checkCanRefund, handleRefundStart},
    {OrderStatusRefunding, OrderStatusRefunded, checkRefundSuccess, handleRefundComplete},
    {OrderStatusRefunding, OrderStatusPaid, checkRefundFailed, handleRefundFailed},
    
    // 退货流程
    {OrderStatusDelivered, OrderStatusReturning, checkCanReturn, handleReturnStart},
    {OrderStatusReturning, OrderStatusReturned, checkReturnSuccess, handleReturnComplete},
    {OrderStatusReturning, OrderStatusDelivered, checkReturnFailed, handleReturnFailed},
}
```

### 6.2 状态变更核心逻辑

```go
type OrderStateMachine struct {
    orderRepo   OrderRepository
    paymentSrv  PaymentService
    logisticsSrv LogisticsService
    eventBus    EventBus
}

func (osm *OrderStateMachine) TransitionTo(orderSn string, targetStatus OrderStatus, operator *Operator) error {
    // 1. 获取当前订单信息
    order, err := osm.orderRepo.GetByOrderSn(orderSn)
    if err != nil {
        return err
    }
    
    // 2. 检查状态转换是否合法
    transition := osm.findTransition(order.Status, targetStatus)
    if transition == nil {
        return errors.New("invalid status transition")
    }
    
    // 3. 检查转换条件
    if transition.Condition != nil && !transition.Condition(order) {
        return errors.New("transition condition not met")
    }
    
    // 4. 执行状态转换处理逻辑
    if transition.Handler != nil {
        if err := transition.Handler(order); err != nil {
            return err
        }
    }
    
    // 5. 更新订单状态
    oldStatus := order.Status
    order.Status = targetStatus
    order.UpdatedAt = time.Now()
    
    // 6. 保存订单和状态变更日志
    tx := osm.orderRepo.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    if err := tx.UpdateOrder(order); err != nil {
        tx.Rollback()
        return err
    }
    
    statusLog := &OrderStatusLog{
        OrderSn:      orderSn,
        StatusFrom:   int(oldStatus),
        StatusTo:     int(targetStatus),
        OperatorID:   operator.ID,
        OperatorType: operator.Type,
        Remark:       transition.Description,
        CreatedAt:    time.Now(),
    }
    
    if err := tx.CreateStatusLog(statusLog); err != nil {
        tx.Rollback()
        return err
    }
    
    tx.Commit()
    
    // 7. 发布状态变更事件
    osm.eventBus.Publish(&OrderStatusChangedEvent{
        OrderSn:    orderSn,
        OldStatus:  oldStatus,
        NewStatus:  targetStatus,
        Operator:   operator,
        ChangedAt:  time.Now(),
    })
    
    return nil
}
```

## 7. 具体状态处理逻辑

### 7.1 支付成功处理

```go
func handlePaymentSuccess(order *Order) error {
    // 更新支付相关信息
    order.PaymentStatus = int(PaymentSubStatusPaid)
    order.PaidAt = &time.Now()
    
    // 自动进入备货流程
    go func() {
        time.Sleep(1 * time.Hour) // 1小时后自动发货
        osm.TransitionTo(order.OrderSn, OrderStatusShipped, &Operator{
            ID:   0,
            Type: "system",
            Name: "系统自动发货",
        })
    }()
    
    return nil
}
```

### 7.2 发货处理

```go
func handleShipment(order *Order) error {
    // 1. 创建物流订单
    logisticsReq := &CreateLogisticsOrderRequest{
        OrderSn:         order.OrderSn,
        UserId:          order.User,
        LogisticsCompany: selectOptimalCompany(order.Address),
        ShippingMethod:   ShippingStandard,
        SenderName:      "商家仓库",
        SenderPhone:     "400-123-4567",
        SenderAddress:   "北京市朝阳区商家仓库",
        ReceiverName:    order.SignerName,
        ReceiverPhone:   order.SingerMobile,
        ReceiverAddress: order.Address,
        // Items: convertOrderItems(order.OrderGoods),
    }
    
    logisticsResp, err := osm.logisticsSrv.CreateLogisticsOrder(context.Background(), logisticsReq)
    if err != nil {
        return err
    }
    
    // 2. 更新订单物流信息
    order.LogisticsSn = logisticsResp.LogisticsSn
    order.TrackingNumber = logisticsResp.TrackingNumber
    order.LogisticsStatus = int(LogisticsSubStatusShipped)
    order.ShippedAt = &time.Now()
    
    // 3. 扣减库存（如果还没扣减的话）
    // 这里假设在支付成功时已经扣减了库存
    
    return nil
}
```

### 7.3 送达处理

```go
func handleDelivery(order *Order) error {
    order.LogisticsStatus = int(LogisticsSubStatusDelivered)
    order.DeliveredAt = &time.Now()
    
    // 设置自动完成时间（7天后）
    autoCompleteTime := time.Now().AddDate(0, 0, 7)
    order.AutoCompleteAt = &autoCompleteTime
    
    // 启动自动完成定时任务
    scheduleAutoCompletion(order.OrderSn, autoCompleteTime)
    
    return nil
}
```

### 7.4 订单完成处理

```go
func handleCompletion(order *Order) error {
    order.CompletedAt = &time.Now()
    
    // 发放积分、优惠券等奖励
    go func() {
        osm.rewardSrv.GrantOrderRewards(order.User, order.OrderSn, order.Total)
    }()
    
    return nil
}
```

### 7.5 退款处理

```go
func handleRefundStart(order *Order) error {
    // 1. 创建退款订单
    refundReq := &RefundRequest{
        PaymentSn:    order.PaymentSn,
        RefundAmount: order.Total,
        Reason:       "用户申请退款",
    }
    
    refundResp, err := osm.paymentSrv.RequestRefund(context.Background(), refundReq)
    if err != nil {
        return err
    }
    
    order.PaymentStatus = int(PaymentSubStatusRefunding)
    // 保存退款单号等信息
    
    return nil
}
```

## 8. 自动化流程

### 8.1 支付超时自动取消

```go
func AutoCancelExpiredOrders() {
    expiredOrders := osm.orderRepo.FindExpiredPendingOrders(15 * time.Minute)
    
    for _, order := range expiredOrders {
        osm.TransitionTo(order.OrderSn, OrderStatusCancelled, &Operator{
            ID:   0,
            Type: "system",
            Name: "支付超时自动取消",
        })
    }
}
```

### 8.2 自动发货

```go
func AutoShipOrders() {
    paidOrders := osm.orderRepo.FindPaidOrdersReadyForShipment()
    
    for _, order := range paidOrders {
        // 检查库存和商品状态
        if osm.canShipOrder(order) {
            osm.TransitionTo(order.OrderSn, OrderStatusShipped, &Operator{
                ID:   0,
                Type: "system", 
                Name: "系统自动发货",
            })
        }
    }
}
```

### 8.3 自动确认收货

```go
func AutoCompleteDeliveredOrders() {
    deliveredOrders := osm.orderRepo.FindDeliveredOrdersReadyForCompletion()
    
    for _, order := range deliveredOrders {
        if time.Now().After(*order.AutoCompleteAt) {
            osm.TransitionTo(order.OrderSn, OrderStatusCompleted, &Operator{
                ID:   0,
                Type: "system",
                Name: "系统自动确认收货",
            })
        }
    }
}
```

## 9. 事件驱动架构

### 9.1 订单状态变更事件

```go
type OrderStatusChangedEvent struct {
    OrderSn    string      `json:"order_sn"`
    OldStatus  OrderStatus `json:"old_status"`
    NewStatus  OrderStatus `json:"new_status"`
    Operator   *Operator   `json:"operator"`
    ChangedAt  time.Time   `json:"changed_at"`
}

func (osm *OrderStateMachine) handleOrderStatusChanged(event *OrderStatusChangedEvent) {
    switch event.NewStatus {
    case OrderStatusPaid:
        osm.onOrderPaid(event)
    case OrderStatusShipped:
        osm.onOrderShipped(event)
    case OrderStatusDelivered:
        osm.onOrderDelivered(event)
    case OrderStatusCompleted:
        osm.onOrderCompleted(event)
    case OrderStatusCancelled:
        osm.onOrderCancelled(event)
    }
}
```

### 9.2 事件处理器

```go
func (osm *OrderStateMachine) onOrderPaid(event *OrderStatusChangedEvent) {
    // 发送支付成功通知
    osm.notificationSrv.SendOrderPaidNotification(event.OrderSn)
    
    // 更新用户积分
    osm.userSrv.UpdateUserPoints(event.OrderSn, "order_paid", 10)
    
    // 记录支付成功日志
    log.Infof("订单 %s 支付成功", event.OrderSn)
}

func (osm *OrderStateMachine) onOrderShipped(event *OrderStatusChangedEvent) {
    // 发送发货通知
    osm.notificationSrv.SendOrderShippedNotification(event.OrderSn)
    
    // 记录发货日志
    log.Infof("订单 %s 已发货", event.OrderSn)
}

func (osm *OrderStateMachine) onOrderCompleted(event *OrderStatusChangedEvent) {
    // 发送完成通知
    osm.notificationSrv.SendOrderCompletedNotification(event.OrderSn)
    
    // 发放订单完成奖励
    osm.rewardSrv.GrantCompletionReward(event.OrderSn)
    
    // 记录完成日志
    log.Infof("订单 %s 已完成", event.OrderSn)
}
```

## 10. 异常处理机制

### 10.1 状态回滚

```go
func (osm *OrderStateMachine) RollbackStatus(orderSn string, targetStatus OrderStatus, reason string) error {
    order, err := osm.orderRepo.GetByOrderSn(orderSn)
    if err != nil {
        return err
    }
    
    // 记录回滚日志
    rollbackLog := &OrderStatusLog{
        OrderSn:      orderSn,
        StatusFrom:   int(order.Status),
        StatusTo:     int(targetStatus),
        ChangeType:   "rollback",
        OperatorType: "system",
        Remark:       "状态回滚: " + reason,
        CreatedAt:    time.Now(),
    }
    
    // 执行状态回滚
    order.Status = targetStatus
    
    tx := osm.orderRepo.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    if err := tx.UpdateOrder(order); err != nil {
        tx.Rollback()
        return err
    }
    
    if err := tx.CreateStatusLog(rollbackLog); err != nil {
        tx.Rollback()
        return err
    }
    
    tx.Commit()
    return nil
}
```

### 10.2 状态一致性检查

```go
func (osm *OrderStateMachine) ValidateOrderConsistency(orderSn string) error {
    order, err := osm.orderRepo.GetByOrderSn(orderSn)
    if err != nil {
        return err
    }
    
    // 检查支付状态一致性
    if order.Status == OrderStatusPaid {
        paymentStatus, err := osm.paymentSrv.GetPaymentStatus(context.Background(), &GetPaymentStatusRequest{
            PaymentSn: order.PaymentSn,
        })
        if err != nil {
            return err
        }
        
        if paymentStatus.PaymentStatus != int32(PaymentPaid) {
            return errors.New("payment status inconsistent")
        }
    }
    
    // 检查物流状态一致性
    if order.Status == OrderStatusShipped {
        logisticsInfo, err := osm.logisticsSrv.GetLogisticsInfo(context.Background(), &GetLogisticsInfoRequest{
            LogisticsSn: order.LogisticsSn,
        })
        if err != nil {
            return err
        }
        
        if logisticsInfo.LogisticsStatus < int32(LogisticsShipped) {
            return errors.New("logistics status inconsistent")
        }
    }
    
    return nil
}
```

## 11. 监控和报警

### 11.1 状态流转监控

```go
type OrderStatusMetrics struct {
    statusTransitionCounter *prometheus.CounterVec
    statusDurationHistogram *prometheus.HistogramVec
    abnormalStatusGauge     *prometheus.GaugeVec
}

func (osm *OrderStateMachine) recordStatusTransition(orderSn string, from, to OrderStatus, duration time.Duration) {
    osm.metrics.statusTransitionCounter.WithLabelValues(
        from.String(),
        to.String(),
    ).Inc()
    
    osm.metrics.statusDurationHistogram.WithLabelValues(
        from.String(),
        to.String(),
    ).Observe(duration.Seconds())
}
```

### 11.2 异常状态报警

```go
func (osm *OrderStateMachine) checkAbnormalOrders() {
    // 检查长时间待支付的订单
    longPendingOrders := osm.orderRepo.FindLongPendingOrders(2 * time.Hour)
    if len(longPendingOrders) > 100 {
        osm.alertManager.SendAlert("too_many_pending_orders", len(longPendingOrders))
    }
    
    // 检查长时间未发货的订单
    longPaidOrders := osm.orderRepo.FindLongPaidOrders(24 * time.Hour)
    if len(longPaidOrders) > 50 {
        osm.alertManager.SendAlert("delayed_shipment", len(longPaidOrders))
    }
    
    // 检查长时间运输中的订单
    longShippedOrders := osm.orderRepo.FindLongShippedOrders(7 * 24 * time.Hour)
    if len(longShippedOrders) > 20 {
        osm.alertManager.SendAlert("delayed_delivery", len(longShippedOrders))
    }
}
```

## 12. API 接口设计

### 12.1 订单状态查询

```go
func (oc *OrderController) GetOrderStatus(c *gin.Context) {
    orderSn := c.Param("order_sn")
    
    order, err := oc.orderSrv.GetOrderWithStatus(orderSn)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
        return
    }
    
    statusInfo := &OrderStatusInfo{
        OrderSn:         order.OrderSn,
        Status:          order.Status,
        PaymentStatus:   order.PaymentStatus,
        LogisticsStatus: order.LogisticsStatus,
        PaymentSn:       order.PaymentSn,
        LogisticsSn:     order.LogisticsSn,
        TrackingNumber:  order.TrackingNumber,
        CreatedAt:       order.CreatedAt,
        PaidAt:          order.PaidAt,
        ShippedAt:       order.ShippedAt,
        DeliveredAt:     order.DeliveredAt,
        CompletedAt:     order.CompletedAt,
    }
    
    c.JSON(http.StatusOK, gin.H{"data": statusInfo})
}
```

### 12.2 手动状态变更

```go
func (oc *OrderController) UpdateOrderStatus(c *gin.Context) {
    var req struct {
        OrderSn   string `json:"order_sn" binding:"required"`
        NewStatus int    `json:"new_status" binding:"required"`
        Remark    string `json:"remark"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 获取操作员信息
    operator := getOperatorFromContext(c)
    
    err := oc.stateMachine.TransitionTo(req.OrderSn, OrderStatus(req.NewStatus), operator)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}
```

这个订单状态流转设计提供了完整的订单生命周期管理，集成了支付和物流服务，确保了状态变更的合法性和数据的一致性，同时支持异常处理和监控报警。