# 物流服务模拟设计文档

## 1. 概述

本文档设计了一个适用于学习项目的物流服务模拟系统，旨在模拟真实电商系统的物流配送流程，但无需对接真实的物流公司接口。

## 2. 设计目标

- 模拟完整的物流配送流程
- 提供物流轨迹跟踪功能
- 支持多种配送方式
- 集成到现有的订单管理系统
- 便于学习和演示

## 3. 物流状态定义

```go
type LogisticsStatus int32

const (
    LogisticsPending     LogisticsStatus = 1 // 待发货
    LogisticsShipped     LogisticsStatus = 2 // 已发货
    LogisticsInTransit   LogisticsStatus = 3 // 运输中
    LogisticsDelivering  LogisticsStatus = 4 // 配送中
    LogisticsDelivered   LogisticsStatus = 5 // 已签收
    LogisticsRejected    LogisticsStatus = 6 // 拒收
    LogisticsReturning   LogisticsStatus = 7 // 退货中
    LogisticsReturned    LogisticsStatus = 8 // 已退货
)
```

## 4. 配送方式定义

```go
type ShippingMethod int32

const (
    ShippingStandard   ShippingMethod = 1 // 标准配送
    ShippingExpress    ShippingMethod = 2 // 急速配送
    ShippingEconomy    ShippingMethod = 3 // 经济配送
    ShippingSelfPickup ShippingMethod = 4 // 自提
)
```

## 5. 物流公司定义

```go
type LogisticsCompany int32

const (
    CompanyYTO      LogisticsCompany = 1 // 圆通速递
    CompanySTO      LogisticsCompany = 2 // 申通快递
    CompanyZTO      LogisticsCompany = 3 // 中通快递
    CompanyYunda    LogisticsCompany = 4 // 韵达速递
    CompanySF       LogisticsCompany = 5 // 顺丰速运
    CompanyJD       LogisticsCompany = 6 // 京东物流
    CompanyEMS      LogisticsCompany = 7 // 中国邮政
)
```

## 6. 数据模型设计

### 6.1 物流订单表 (logistics_orders)

```sql
CREATE TABLE logistics_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    logistics_sn VARCHAR(64) NOT NULL UNIQUE COMMENT '物流单号',
    order_sn VARCHAR(64) NOT NULL COMMENT '订单号',
    user_id INT NOT NULL COMMENT '用户ID',
    logistics_company TINYINT NOT NULL COMMENT '物流公司',
    shipping_method TINYINT NOT NULL COMMENT '配送方式',
    tracking_number VARCHAR(64) NOT NULL COMMENT '快递单号',
    logistics_status TINYINT NOT NULL DEFAULT 1 COMMENT '物流状态',
    
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
    
    remark TEXT COMMENT '备注信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_logistics_sn (logistics_sn),
    INDEX idx_order_sn (order_sn),
    INDEX idx_tracking_number (tracking_number),
    INDEX idx_user_id (user_id),
    INDEX idx_status (logistics_status)
);
```

### 6.2 物流轨迹表 (logistics_tracks)

```sql
CREATE TABLE logistics_tracks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    logistics_sn VARCHAR(64) NOT NULL COMMENT '物流单号',
    tracking_number VARCHAR(64) NOT NULL COMMENT '快递单号',
    location VARCHAR(128) NOT NULL COMMENT '当前位置',
    description TEXT NOT NULL COMMENT '轨迹描述',
    track_time TIMESTAMP NOT NULL COMMENT '轨迹时间',
    operator_name VARCHAR(64) COMMENT '操作员',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_logistics_sn (logistics_sn),
    INDEX idx_tracking_number (tracking_number),
    INDEX idx_track_time (track_time)
);
```

### 6.3 物流配送员表 (logistics_couriers)

```sql
CREATE TABLE logistics_couriers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    courier_code VARCHAR(32) NOT NULL UNIQUE COMMENT '配送员编号',
    courier_name VARCHAR(64) NOT NULL COMMENT '配送员姓名',
    phone VARCHAR(32) NOT NULL COMMENT '联系电话',
    logistics_company TINYINT NOT NULL COMMENT '所属物流公司',
    service_area VARCHAR(128) COMMENT '服务区域',
    status TINYINT DEFAULT 1 COMMENT '状态：1-在职，0-离职',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_courier_code (courier_code),
    INDEX idx_company (logistics_company)
);
```

## 7. 服务接口设计

### 7.1 gRPC 服务定义

```protobuf
service Logistics {
    // 创建物流订单
    rpc CreateLogisticsOrder(CreateLogisticsOrderRequest) returns (CreateLogisticsOrderResponse);
    
    // 查询物流信息
    rpc GetLogisticsInfo(GetLogisticsInfoRequest) returns (GetLogisticsInfoResponse);
    
    // 查询物流轨迹
    rpc GetLogisticsTracks(GetLogisticsTracksRequest) returns (GetLogisticsTracksResponse);
    
    // 更新物流状态（内部调用）
    rpc UpdateLogisticsStatus(UpdateLogisticsStatusRequest) returns (google.protobuf.Empty);
    
    // 模拟发货
    rpc SimulateShipment(SimulateShipmentRequest) returns (google.protobuf.Empty);
    
    // 模拟签收
    rpc SimulateDelivery(SimulateDeliveryRequest) returns (google.protobuf.Empty);
    
    // 计算运费
    rpc CalculateShippingFee(CalculateShippingFeeRequest) returns (CalculateShippingFeeResponse);
}
```

### 7.2 消息定义

```protobuf
message CreateLogisticsOrderRequest {
    string order_sn = 1;
    int32 user_id = 2;
    int32 logistics_company = 3;
    int32 shipping_method = 4;
    
    // 发货信息
    string sender_name = 5;
    string sender_phone = 6;
    string sender_address = 7;
    
    // 收货信息
    string receiver_name = 8;
    string receiver_phone = 9;
    string receiver_address = 10;
    
    // 商品信息
    repeated OrderItem items = 11;
    
    string remark = 12;
}

message OrderItem {
    int32 goods_id = 1;
    string goods_name = 2;
    int32 quantity = 3;
    double weight = 4; // 重量(kg)
    double volume = 5; // 体积(cm³)
}

message CreateLogisticsOrderResponse {
    string logistics_sn = 1;
    string tracking_number = 2;
    double shipping_fee = 3;
    int64 estimated_delivery_at = 4;
}

message GetLogisticsInfoRequest {
    oneof query {
        string logistics_sn = 1;
        string order_sn = 2;
        string tracking_number = 3;
    }
}

message GetLogisticsInfoResponse {
    string logistics_sn = 1;
    string order_sn = 2;
    string tracking_number = 3;
    int32 logistics_company = 4;
    int32 shipping_method = 5;
    int32 logistics_status = 6;
    
    string sender_name = 7;
    string sender_phone = 8;
    string sender_address = 9;
    
    string receiver_name = 10;
    string receiver_phone = 11;
    string receiver_address = 12;
    
    double shipping_fee = 13;
    int64 shipped_at = 14;
    int64 delivered_at = 15;
    int64 estimated_delivery_at = 16;
    
    string remark = 17;
}

message GetLogisticsTracksRequest {
    oneof query {
        string logistics_sn = 1;
        string tracking_number = 2;
    }
}

message LogisticsTrack {
    string location = 1;
    string description = 2;
    int64 track_time = 3;
    string operator_name = 4;
}

message GetLogisticsTracksResponse {
    string logistics_sn = 1;
    string tracking_number = 2;
    repeated LogisticsTrack tracks = 3;
}

message UpdateLogisticsStatusRequest {
    string logistics_sn = 1;
    int32 new_status = 2;
    string remark = 3;
}

message SimulateShipmentRequest {
    string logistics_sn = 1;
    string courier_name = 2;
    string courier_phone = 3;
}

message SimulateDeliveryRequest {
    string logistics_sn = 1;
    string receiver_name = 2;
    string delivery_remark = 3;
}

message CalculateShippingFeeRequest {
    string sender_address = 1;
    string receiver_address = 2;
    int32 shipping_method = 3;
    double total_weight = 4;
    double total_volume = 5;
    double goods_value = 6;
    bool need_insurance = 7;
}

message CalculateShippingFeeResponse {
    double shipping_fee = 1;
    double insurance_fee = 2;
    double total_fee = 3;
    int32 estimated_days = 4;
}
```

## 8. 核心业务逻辑

### 8.1 物流订单创建流程

```go
func CreateLogisticsOrder(req *CreateLogisticsOrderRequest) (*CreateLogisticsOrderResponse, error) {
    // 1. 生成物流单号和快递单号
    logisticsSn := generateLogisticsSn()
    trackingNumber := generateTrackingNumber(req.LogisticsCompany)
    
    // 2. 计算运费
    shippingFee := calculateShippingFee(req)
    
    // 3. 计算预计送达时间
    estimatedDelivery := calculateEstimatedDelivery(req.ShippingMethod, req.ReceiverAddress)
    
    // 4. 创建物流订单
    logisticsOrder := &LogisticsOrder{
        LogisticsSn:     logisticsSn,
        OrderSn:         req.OrderSn,
        UserId:          req.UserId,
        TrackingNumber:  trackingNumber,
        // ... 其他字段
        EstimatedDeliveryAt: estimatedDelivery,
    }
    
    // 5. 保存到数据库
    err := db.Create(logisticsOrder).Error
    if err != nil {
        return nil, err
    }
    
    // 6. 创建初始轨迹记录
    initialTrack := &LogisticsTrack{
        LogisticsSn:    logisticsSn,
        TrackingNumber: trackingNumber,
        Location:       "商家仓库",
        Description:    "商家正在准备发货",
        TrackTime:      time.Now(),
    }
    db.Create(initialTrack)
    
    return &CreateLogisticsOrderResponse{
        LogisticsSn:         logisticsSn,
        TrackingNumber:      trackingNumber,
        ShippingFee:         shippingFee,
        EstimatedDeliveryAt: estimatedDelivery.Unix(),
    }, nil
}
```

### 8.2 轨迹号生成策略

```go
func generateTrackingNumber(company LogisticsCompany) string {
    switch company {
    case CompanySF:
        return "SF" + generateRandomNumber(12)
    case CompanyJD:
        return "JD" + generateRandomNumber(13)
    case CompanyYTO:
        return "YT" + generateRandomNumber(12)
    default:
        return "EX" + generateRandomNumber(12)
    }
}

func generateLogisticsSn() string {
    return "LG" + time.Now().Format("20060102150405") + generateRandomNumber(6)
}
```

### 8.3 运费计算逻辑

```go
func calculateShippingFee(req *CalculateShippingFeeRequest) (float64, float64, float64) {
    // 基础运费计算
    baseDistance := calculateDistance(req.SenderAddress, req.ReceiverAddress)
    baseFee := calculateBaseFee(baseDistance, req.TotalWeight)
    
    // 配送方式加成
    methodMultiplier := getMethodMultiplier(req.ShippingMethod)
    shippingFee := baseFee * methodMultiplier
    
    // 保价费计算
    var insuranceFee float64
    if req.NeedInsurance {
        insuranceFee = req.GoodsValue * 0.005 // 0.5%保价费率
    }
    
    return shippingFee, insuranceFee, shippingFee + insuranceFee
}
```

## 9. 物流轨迹模拟

### 9.1 轨迹模板定义

```go
type TrackTemplate struct {
    LocationPattern string
    DescriptionPattern string
    DelayMinutes int
}

var trackTemplates = map[ShippingMethod][]TrackTemplate{
    ShippingStandard: {
        {"商家仓库", "商家正在准备发货", 0},
        {"商家仓库", "商品已打包完成", 60},
        {"{{SenderCity}}集散中心", "快件已发出", 120},
        {"{{SenderCity}}集散中心", "快件已到达{{SenderCity}}集散中心", 180},
        {"{{ReceiverCity}}集散中心", "快件已到达{{ReceiverCity}}集散中心", 1440}, // 1天后
        {"{{ReceiverCity}}配送站", "快件已到达{{ReceiverCity}}配送站", 1500},
        {"{{ReceiverCity}}配送站", "配送员{{CourierName}}正在配送中", 1560},
        {"{{ReceiverAddress}}", "快件已签收", 1620},
    },
    ShippingExpress: {
        {"商家仓库", "商家正在准备发货", 0},
        {"商家仓库", "商品已打包完成", 30},
        {"{{SenderCity}}集散中心", "快件已发出", 60},
        {"{{ReceiverCity}}集散中心", "快件已到达{{ReceiverCity}}集散中心", 720}, // 12小时后
        {"{{ReceiverCity}}配送站", "配送员{{CourierName}}正在配送中", 780},
        {"{{ReceiverAddress}}", "快件已签收", 840},
    },
}
```

### 9.2 自动轨迹生成

```go
func GenerateLogisticsTracks(logisticsSn string) error {
    // 获取物流订单信息
    order, err := getLogisticsOrder(logisticsSn)
    if err != nil {
        return err
    }
    
    // 选择轨迹模板
    templates := trackTemplates[order.ShippingMethod]
    
    // 生成配送员信息
    courier := assignRandomCourier(order.LogisticsCompany)
    
    // 解析地址信息
    senderCity := parseCity(order.SenderAddress)
    receiverCity := parseCity(order.ReceiverAddress)
    
    // 生成轨迹记录
    startTime := order.ShippedAt
    for i, template := range templates {
        trackTime := startTime.Add(time.Duration(template.DelayMinutes) * time.Minute)
        
        // 替换模板变量
        location := strings.ReplaceAll(template.LocationPattern, "{{SenderCity}}", senderCity)
        location = strings.ReplaceAll(location, "{{ReceiverCity}}", receiverCity)
        location = strings.ReplaceAll(location, "{{ReceiverAddress}}", order.ReceiverAddress)
        
        description := strings.ReplaceAll(template.DescriptionPattern, "{{CourierName}}", courier.Name)
        description = strings.ReplaceAll(description, "{{SenderCity}}", senderCity)
        description = strings.ReplaceAll(description, "{{ReceiverCity}}", receiverCity)
        
        track := &LogisticsTrack{
            LogisticsSn:    logisticsSn,
            TrackingNumber: order.TrackingNumber,
            Location:       location,
            Description:    description,
            TrackTime:      trackTime,
            OperatorName:   courier.Name,
        }
        
        // 延迟创建轨迹记录
        scheduleTrackCreation(track, trackTime)
        
        // 最后一个轨迹是签收，更新物流状态
        if i == len(templates)-1 {
            scheduleStatusUpdate(logisticsSn, LogisticsDelivered, trackTime)
        }
    }
    
    return nil
}
```

## 10. 定时任务设计

### 10.1 物流状态自动更新

```go
func AutoUpdateLogisticsStatus() {
    // 查询待处理的物流订单
    pendingOrders := findPendingLogisticsOrders()
    
    for _, order := range pendingOrders {
        switch order.LogisticsStatus {
        case LogisticsPending:
            // 模拟发货：创建后2小时自动发货
            if time.Since(order.CreatedAt) > 2*time.Hour {
                simulateShipment(order.LogisticsSn)
            }
            
        case LogisticsShipped:
            // 模拟配送流程
            if time.Since(order.ShippedAt) > getExpectedDeliveryDuration(order.ShippingMethod) {
                simulateDelivery(order.LogisticsSn)
            }
        }
    }
}
```

### 10.2 轨迹记录自动生成

```go
func ProcessScheduledTracks() {
    // 处理预定的轨迹记录
    scheduledTracks := findScheduledTracks()
    now := time.Now()
    
    for _, track := range scheduledTracks {
        if track.ScheduledTime.Before(now) {
            // 创建轨迹记录
            createLogisticsTrack(track)
            
            // 删除预定记录
            deleteScheduledTrack(track.ID)
        }
    }
}
```

## 11. 与订单服务的集成

### 11.1 发货通知

```go
func OnShipmentCreated(logisticsSn string, orderSn string) error {
    // 获取物流信息
    logistics, err := getLogisticsInfo(logisticsSn)
    if err != nil {
        return err
    }
    
    // 通知订单服务
    _, err = orderClient.UpdateOrderStatus(context.Background(), &order.UpdateOrderStatusRequest{
        OrderSn: orderSn,
        Status:  "shipped",
        LogisticsInfo: &order.LogisticsInfo{
            LogisticsSn:    logisticsSn,
            TrackingNumber: logistics.TrackingNumber,
            Company:        logistics.LogisticsCompany,
        },
    })
    
    return err
}
```

### 11.2 签收通知

```go
func OnDeliveryCompleted(logisticsSn string, orderSn string) error {
    // 通知订单服务订单已完成
    _, err := orderClient.CompleteOrder(context.Background(), &order.CompleteOrderRequest{
        OrderSn:      orderSn,
        DeliveredAt:  time.Now().Unix(),
        LogisticsSn:  logisticsSn,
    })
    
    return err
}
```

## 12. 前端集成功能

### 12.1 物流跟踪页面

1. 输入快递单号查询
2. 显示物流轨迹时间线
3. 显示当前物流状态
4. 提供配送员联系方式

### 12.2 管理后台功能

1. 物流订单管理
2. 手动发货操作
3. 物流轨迹管理
4. 配送员管理
5. 运费规则配置

## 13. 监控和统计

### 13.1 关键指标

- 平均配送时长
- 签收成功率
- 客户满意度
- 物流成本分析

### 13.2 异常监控

- 超时未签收订单
- 物流轨迹异常
- 配送失败统计

## 14. 扩展功能

### 14.1 智能配送

```go
// 根据收货地址智能选择最优物流公司
func SelectOptimalLogisticsCompany(receiverAddress string) LogisticsCompany {
    region := parseRegion(receiverAddress)
    
    // 根据地区选择最优物流公司
    switch region {
    case "一线城市":
        return CompanyJD // 京东物流覆盖好
    case "偏远地区":
        return CompanyEMS // 邮政覆盖全
    default:
        return CompanySF // 顺丰服务好
    }
}
```

### 14.2 配送时效预测

```go
func PredictDeliveryTime(senderAddr, receiverAddr string, method ShippingMethod) time.Time {
    distance := calculateDistance(senderAddr, receiverAddr)
    baseHours := getBaseDeliveryHours(method)
    
    // 根据距离调整预计时间
    if distance > 1000 { // 跨省
        baseHours += 24
    } else if distance > 100 { // 跨市
        baseHours += 12
    }
    
    return time.Now().Add(time.Duration(baseHours) * time.Hour)
}
```

这个物流服务设计提供了完整的物流管理功能模拟，包括订单创建、轨迹跟踪、状态管理等核心功能，既保持了真实物流系统的完整性，又避免了复杂的第三方接口对接问题。