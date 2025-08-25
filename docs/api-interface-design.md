# API接口设计文档

## 1. 概述

本文档定义了电商系统中支付服务和物流服务的完整API接口设计，包括gRPC服务定义和HTTP RESTful API设计。

## 2. gRPC服务定义

### 2.1 支付服务 (Payment Service)

#### 2.1.1 服务定义文件 `api/payment/v1/payment.proto`

```protobuf
syntax = "proto3";
package payment.v1;
import "google/protobuf/empty.proto";
option go_package = ".;proto";

service Payment {
    // 创建支付订单
    rpc CreatePayment(CreatePaymentRequest) returns (CreatePaymentResponse);
    
    // 查询支付状态
    rpc GetPaymentStatus(GetPaymentStatusRequest) returns (GetPaymentStatusResponse);
    
    // 查询支付详情
    rpc GetPaymentDetail(GetPaymentDetailRequest) returns (GetPaymentDetailResponse);
    
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
    
    // 支付成功回调（内部调用）
    rpc PaymentSuccessCallback(PaymentCallbackRequest) returns (google.protobuf.Empty);
    
    // 支付失败回调（内部调用）
    rpc PaymentFailureCallback(PaymentCallbackRequest) returns (google.protobuf.Empty);
}

// 创建支付请求
message CreatePaymentRequest {
    string order_sn = 1;           // 订单号
    int32 user_id = 2;             // 用户ID
    double amount = 3;             // 支付金额
    int32 payment_method = 4;      // 支付方式
    int32 expired_minutes = 5;     // 支付过期时间（分钟）
    string return_url = 6;         // 支付完成跳转URL
    string notify_url = 7;         // 支付结果通知URL
    string remark = 8;             // 备注
}

// 创建支付响应
message CreatePaymentResponse {
    string payment_sn = 1;         // 支付单号
    string payment_url = 2;        // 支付链接
    string qr_code = 3;            // 支付二维码（模拟）
    int64 expired_at = 4;          // 过期时间戳
    double amount = 5;             // 支付金额
}

// 查询支付状态请求
message GetPaymentStatusRequest {
    oneof query {
        string payment_sn = 1;     // 支付单号
        string order_sn = 2;       // 订单号
    }
}

// 查询支付状态响应
message GetPaymentStatusResponse {
    string payment_sn = 1;         // 支付单号
    string order_sn = 2;           // 订单号
    int32 payment_status = 3;      // 支付状态
    double amount = 4;             // 支付金额
    int32 payment_method = 5;      // 支付方式
    string third_party_sn = 6;     // 第三方支付单号
    int64 paid_at = 7;             // 支付时间
    int64 expired_at = 8;          // 过期时间
}

// 查询支付详情请求
message GetPaymentDetailRequest {
    string payment_sn = 1;         // 支付单号
}

// 查询支付详情响应
message GetPaymentDetailResponse {
    string payment_sn = 1;         // 支付单号
    string order_sn = 2;           // 订单号
    int32 user_id = 3;             // 用户ID
    double amount = 4;             // 支付金额
    int32 payment_method = 5;      // 支付方式
    int32 payment_status = 6;      // 支付状态
    string third_party_sn = 7;     // 第三方支付单号
    int64 created_at = 8;          // 创建时间
    int64 paid_at = 9;             // 支付时间
    int64 expired_at = 10;         // 过期时间
    string remark = 11;            // 备注
}

// 模拟支付请求
message SimulatePaymentRequest {
    string payment_sn = 1;         // 支付单号
    optional string third_party_sn = 2; // 第三方支付单号
    optional string remark = 3;    // 备注
}

// 取消支付请求
message CancelPaymentRequest {
    string payment_sn = 1;         // 支付单号
    string reason = 2;             // 取消原因
}

// 退款请求
message RefundRequest {
    string payment_sn = 1;         // 支付单号
    double refund_amount = 2;      // 退款金额
    string reason = 3;             // 退款原因
    string operator_id = 4;        // 操作员ID
}

// 退款响应
message RefundResponse {
    string refund_sn = 1;          // 退款单号
    int32 refund_status = 2;       // 退款状态
    double refund_amount = 3;      // 退款金额
    int64 expected_at = 4;         // 预计到账时间
}

// 查询退款状态请求
message GetRefundStatusRequest {
    oneof query {
        string refund_sn = 1;      // 退款单号
        string payment_sn = 2;     // 支付单号
    }
}

// 查询退款状态响应
message GetRefundStatusResponse {
    string refund_sn = 1;          // 退款单号
    string payment_sn = 2;         // 支付单号
    int32 refund_status = 3;       // 退款状态
    double refund_amount = 4;      // 退款金额
    string reason = 5;             // 退款原因
    int64 created_at = 6;          // 申请时间
    int64 refunded_at = 7;         // 退款完成时间
}

// 支付回调请求
message PaymentCallbackRequest {
    string payment_sn = 1;         // 支付单号
    string third_party_sn = 2;     // 第三方支付单号
    double amount = 3;             // 支付金额
    string callback_data = 4;      // 回调数据
}
```

### 2.2 物流服务 (Logistics Service)

#### 2.2.1 服务定义文件 `api/logistics/v1/logistics.proto`

```protobuf
syntax = "proto3";
package logistics.v1;
import "google/protobuf/empty.proto";
option go_package = ".;proto";

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
    
    // 获取物流公司列表
    rpc GetLogisticsCompanies(google.protobuf.Empty) returns (LogisticsCompaniesResponse);
    
    // 批量查询物流状态
    rpc BatchGetLogisticsStatus(BatchGetLogisticsStatusRequest) returns (BatchGetLogisticsStatusResponse);
}

// 创建物流订单请求
message CreateLogisticsOrderRequest {
    string order_sn = 1;           // 订单号
    int32 user_id = 2;             // 用户ID
    int32 logistics_company = 3;   // 物流公司
    int32 shipping_method = 4;     // 配送方式
    
    // 发货信息
    string sender_name = 5;        // 发货人姓名
    string sender_phone = 6;       // 发货人电话
    string sender_address = 7;     // 发货地址
    
    // 收货信息
    string receiver_name = 8;      // 收货人姓名
    string receiver_phone = 9;     // 收货人电话
    string receiver_address = 10;  // 收货地址
    
    // 商品信息
    repeated OrderItem items = 11; // 商品列表
    
    string remark = 12;            // 备注
    bool need_insurance = 13;      // 是否需要保价
    double goods_value = 14;       // 商品价值（保价用）
}

// 商品信息
message OrderItem {
    int32 goods_id = 1;            // 商品ID
    string goods_name = 2;         // 商品名称
    int32 quantity = 3;            // 数量
    double weight = 4;             // 重量(kg)
    double volume = 5;             // 体积(cm³)
    double price = 6;              // 单价
}

// 创建物流订单响应
message CreateLogisticsOrderResponse {
    string logistics_sn = 1;       // 物流单号
    string tracking_number = 2;    // 快递单号
    double shipping_fee = 3;       // 运费
    double insurance_fee = 4;      // 保价费
    double total_fee = 5;          // 总费用
    int64 estimated_delivery_at = 6; // 预计送达时间
    string courier_name = 7;       // 配送员姓名
    string courier_phone = 8;      // 配送员电话
}

// 查询物流信息请求
message GetLogisticsInfoRequest {
    oneof query {
        string logistics_sn = 1;   // 物流单号
        string order_sn = 2;       // 订单号
        string tracking_number = 3; // 快递单号
    }
}

// 查询物流信息响应
message GetLogisticsInfoResponse {
    string logistics_sn = 1;       // 物流单号
    string order_sn = 2;           // 订单号
    string tracking_number = 3;    // 快递单号
    int32 logistics_company = 4;   // 物流公司
    string company_name = 5;       // 物流公司名称
    int32 shipping_method = 6;     // 配送方式
    int32 logistics_status = 7;    // 物流状态
    
    string sender_name = 8;        // 发货人姓名
    string sender_phone = 9;       // 发货人电话
    string sender_address = 10;    // 发货地址
    
    string receiver_name = 11;     // 收货人姓名
    string receiver_phone = 12;    // 收货人电话
    string receiver_address = 13;  // 收货地址
    
    double shipping_fee = 14;      // 运费
    double insurance_fee = 15;     // 保价费
    int64 created_at = 16;         // 创建时间
    int64 shipped_at = 17;         // 发货时间
    int64 delivered_at = 18;       // 签收时间
    int64 estimated_delivery_at = 19; // 预计送达时间
    
    string courier_name = 20;      // 配送员姓名
    string courier_phone = 21;     // 配送员电话
    string remark = 22;            // 备注信息
}

// 查询物流轨迹请求
message GetLogisticsTracksRequest {
    oneof query {
        string logistics_sn = 1;   // 物流单号
        string tracking_number = 2; // 快递单号
    }
}

// 物流轨迹信息
message LogisticsTrack {
    string location = 1;           // 当前位置
    string description = 2;        // 轨迹描述
    int64 track_time = 3;          // 轨迹时间
    string operator_name = 4;      // 操作员
    int32 status = 5;              // 状态编码
}

// 查询物流轨迹响应
message GetLogisticsTracksResponse {
    string logistics_sn = 1;       // 物流单号
    string tracking_number = 2;    // 快递单号
    int32 current_status = 3;      // 当前状态
    repeated LogisticsTrack tracks = 4; // 轨迹列表
}

// 更新物流状态请求
message UpdateLogisticsStatusRequest {
    string logistics_sn = 1;       // 物流单号
    int32 new_status = 2;          // 新状态
    string location = 3;           // 当前位置
    string description = 4;        // 描述
    string operator_name = 5;      // 操作员
    string remark = 6;             // 备注
}

// 模拟发货请求
message SimulateShipmentRequest {
    string logistics_sn = 1;       // 物流单号
    string courier_name = 2;       // 配送员姓名
    string courier_phone = 3;      // 配送员电话
    string departure_time = 4;     // 发车时间
}

// 模拟签收请求
message SimulateDeliveryRequest {
    string logistics_sn = 1;       // 物流单号
    string receiver_name = 2;      // 实际收货人
    string delivery_remark = 3;    // 签收备注
    int64 delivery_time = 4;       // 签收时间
}

// 计算运费请求
message CalculateShippingFeeRequest {
    string sender_address = 1;     // 发货地址
    string receiver_address = 2;   // 收货地址
    int32 shipping_method = 3;     // 配送方式
    double total_weight = 4;       // 总重量
    double total_volume = 5;       // 总体积
    double goods_value = 6;        // 商品价值
    bool need_insurance = 7;       // 是否需要保价
    int32 logistics_company = 8;   // 物流公司
}

// 计算运费响应
message CalculateShippingFeeResponse {
    double shipping_fee = 1;       // 运费
    double insurance_fee = 2;      // 保价费
    double total_fee = 3;          // 总费用
    int32 estimated_days = 4;      // 预计天数
    string fee_detail = 5;         // 费用明细
}

// 物流公司信息
message LogisticsCompany {
    int32 company_id = 1;          // 公司ID
    string company_code = 2;       // 公司编码
    string company_name = 3;       // 公司名称
    string logo_url = 4;           // Logo地址
    bool is_available = 5;         // 是否可用
    repeated int32 support_methods = 6; // 支持的配送方式
}

// 物流公司列表响应
message LogisticsCompaniesResponse {
    repeated LogisticsCompany companies = 1;
}

// 批量查询物流状态请求
message BatchGetLogisticsStatusRequest {
    repeated string logistics_sns = 1; // 物流单号列表
}

// 物流状态信息
message LogisticsStatusInfo {
    string logistics_sn = 1;       // 物流单号
    string tracking_number = 2;    // 快递单号
    int32 logistics_status = 3;    // 物流状态
    string status_description = 4; // 状态描述
    string current_location = 5;   // 当前位置
    int64 updated_at = 6;          // 更新时间
}

// 批量查询物流状态响应
message BatchGetLogisticsStatusResponse {
    repeated LogisticsStatusInfo statuses = 1;
}
```

## 3. HTTP RESTful API 设计

### 3.1 支付服务 REST API

#### 3.1.1 创建支付订单
```
POST /api/v1/payments
Content-Type: application/json

Request Body:
{
    "order_sn": "ORD202501241234567890",
    "user_id": 123,
    "amount": 99.99,
    "payment_method": 1,
    "expired_minutes": 15,
    "return_url": "https://example.com/payment/success",
    "notify_url": "https://example.com/payment/notify",
    "remark": "商品购买"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "payment_sn": "PAY202501241234567890",
        "payment_url": "https://mock-pay.example.com/pay/PAY202501241234567890",
        "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
        "expired_at": 1642982400,
        "amount": 99.99
    }
}
```

#### 3.1.2 查询支付状态
```
GET /api/v1/payments/{payment_sn}/status
GET /api/v1/payments/status?order_sn={order_sn}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "payment_sn": "PAY202501241234567890",
        "order_sn": "ORD202501241234567890",
        "payment_status": 2,
        "amount": 99.99,
        "payment_method": 1,
        "third_party_sn": "WX20250124123456",
        "paid_at": 1642982100,
        "expired_at": 1642982400
    }
}
```

#### 3.1.3 模拟支付成功
```
POST /api/v1/payments/{payment_sn}/simulate-success
Content-Type: application/json

Request Body:
{
    "third_party_sn": "WX20250124123456",
    "remark": "模拟支付成功"
}

Response:
{
    "code": 0,
    "message": "支付模拟成功"
}
```

#### 3.1.4 申请退款
```
POST /api/v1/payments/{payment_sn}/refund
Content-Type: application/json

Request Body:
{
    "refund_amount": 99.99,
    "reason": "用户申请退款",
    "operator_id": "admin123"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "refund_sn": "REF202501241234567890",
        "refund_status": 1,
        "refund_amount": 99.99,
        "expected_at": 1643068800
    }
}
```

### 3.2 物流服务 REST API

#### 3.2.1 创建物流订单
```
POST /api/v1/logistics/orders
Content-Type: application/json

Request Body:
{
    "order_sn": "ORD202501241234567890",
    "user_id": 123,
    "logistics_company": 1,
    "shipping_method": 1,
    "sender_name": "商家仓库",
    "sender_phone": "400-123-4567",
    "sender_address": "北京市朝阳区商家仓库",
    "receiver_name": "张三",
    "receiver_phone": "13812345678",
    "receiver_address": "上海市浦东新区某某街道123号",
    "items": [
        {
            "goods_id": 1001,
            "goods_name": "iPhone 14",
            "quantity": 1,
            "weight": 0.5,
            "volume": 1000,
            "price": 5999.00
        }
    ],
    "remark": "易碎品，请轻拿轻放",
    "need_insurance": true,
    "goods_value": 5999.00
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "logistics_sn": "LOG202501241234567890",
        "tracking_number": "SF1234567890123",
        "shipping_fee": 15.00,
        "insurance_fee": 29.99,
        "total_fee": 44.99,
        "estimated_delivery_at": 1643068800,
        "courier_name": "李师傅",
        "courier_phone": "13987654321"
    }
}
```

#### 3.2.2 查询物流信息
```
GET /api/v1/logistics/orders/{logistics_sn}
GET /api/v1/logistics/orders?order_sn={order_sn}
GET /api/v1/logistics/orders?tracking_number={tracking_number}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "logistics_sn": "LOG202501241234567890",
        "order_sn": "ORD202501241234567890",
        "tracking_number": "SF1234567890123",
        "logistics_company": 1,
        "company_name": "顺丰速运",
        "shipping_method": 1,
        "logistics_status": 3,
        "sender_name": "商家仓库",
        "sender_phone": "400-123-4567",
        "sender_address": "北京市朝阳区商家仓库",
        "receiver_name": "张三",
        "receiver_phone": "13812345678",
        "receiver_address": "上海市浦东新区某某街道123号",
        "shipping_fee": 15.00,
        "insurance_fee": 29.99,
        "created_at": 1642982100,
        "shipped_at": 1642985700,
        "delivered_at": null,
        "estimated_delivery_at": 1643068800,
        "courier_name": "李师傅",
        "courier_phone": "13987654321",
        "remark": "易碎品，请轻拿轻放"
    }
}
```

#### 3.2.3 查询物流轨迹
```
GET /api/v1/logistics/tracks/{logistics_sn}
GET /api/v1/logistics/tracks?tracking_number={tracking_number}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "logistics_sn": "LOG202501241234567890",
        "tracking_number": "SF1234567890123",
        "current_status": 3,
        "tracks": [
            {
                "location": "北京朝阳集散中心",
                "description": "快件已发出",
                "track_time": 1642985700,
                "operator_name": "张师傅",
                "status": 2
            },
            {
                "location": "上海浦东集散中心",
                "description": "快件已到达上海浦东集散中心",
                "track_time": 1643025700,
                "operator_name": "王师傅",
                "status": 3
            },
            {
                "location": "上海浦东配送站",
                "description": "配送员李师傅正在配送中",
                "track_time": 1643061700,
                "operator_name": "李师傅",
                "status": 4
            }
        ]
    }
}
```

#### 3.2.4 计算运费
```
POST /api/v1/logistics/calculate-fee
Content-Type: application/json

Request Body:
{
    "sender_address": "北京市朝阳区",
    "receiver_address": "上海市浦东新区",
    "shipping_method": 1,
    "total_weight": 0.5,
    "total_volume": 1000,
    "goods_value": 5999.00,
    "need_insurance": true,
    "logistics_company": 1
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "shipping_fee": 15.00,
        "insurance_fee": 29.99,
        "total_fee": 44.99,
        "estimated_days": 2,
        "fee_detail": "首重1kg内：12元，续重0.5kg：3元，保价费率：0.5%"
    }
}
```

## 4. 错误码定义

### 4.1 支付服务错误码

```go
const (
    // 支付相关错误码 (20000-29999)
    ErrPaymentNotFound           = 20001 // 支付订单不存在
    ErrPaymentExpired            = 20002 // 支付已过期
    ErrPaymentAlreadyPaid        = 20003 // 支付订单已支付
    ErrPaymentAlreadyCancelled   = 20004 // 支付订单已取消
    ErrPaymentAmountMismatch     = 20005 // 支付金额不匹配
    ErrPaymentMethodNotSupported = 20006 // 不支持的支付方式
    ErrPaymentCreateFailed       = 20007 // 创建支付订单失败
    ErrPaymentStatusInvalid      = 20008 // 支付状态无效
    
    // 退款相关错误码 (20100-20199)
    ErrRefundNotFound         = 20101 // 退款订单不存在
    ErrRefundAmountExceeded   = 20102 // 退款金额超出支付金额
    ErrRefundAlreadyProcessed = 20103 // 退款已处理
    ErrRefundNotAllowed       = 20104 // 不允许退款
    ErrRefundCreateFailed     = 20105 // 创建退款订单失败
)
```

### 4.2 物流服务错误码

```go
const (
    // 物流相关错误码 (30000-39999)
    ErrLogisticsNotFound         = 30001 // 物流订单不存在
    ErrLogisticsAlreadyShipped   = 30002 // 物流订单已发货
    ErrLogisticsCompanyInvalid   = 30003 // 无效的物流公司
    ErrLogisticsCreateFailed     = 30004 // 创建物流订单失败
    ErrLogisticsAddressInvalid   = 30005 // 地址信息无效
    ErrLogisticsWeightExceeded   = 30006 // 重量超出限制
    ErrLogisticsVolumeExceeded   = 30007 // 体积超出限制
    ErrLogisticsStatusInvalid    = 30008 // 物流状态无效
    ErrLogisticsTrackNotFound    = 30009 // 物流轨迹不存在
    ErrLogisticsDelivered        = 30010 // 物流订单已送达
    ErrLogisticsCancelled        = 30011 // 物流订单已取消
)
```

## 5. API认证和鉴权

### 5.1 JWT Token 认证

```go
// HTTP Headers
Authorization: Bearer {jwt_token}
Content-Type: application/json
```

### 5.2 gRPC 认证

```go
// gRPC Metadata
authorization: bearer {jwt_token}
user-id: {user_id}
```

## 6. API版本管理

### 6.1 URL版本控制
```
https://api.example.com/v1/payments
https://api.example.com/v2/payments
```

### 6.2 gRPC版本控制
```
package payment.v1;
package payment.v2;
```

## 7. 限流和熔断

### 7.1 接口限流配置

```yaml
rate_limit:
  payments:
    create: 100/min    # 创建支付限流
    query: 1000/min    # 查询支付限流
  logistics:
    create: 50/min     # 创建物流限流
    track: 500/min     # 查询轨迹限流
```

### 7.2 熔断配置

```yaml
circuit_breaker:
  failure_rate: 50%      # 失败率阈值
  slow_call_rate: 50%    # 慢调用率阈值
  slow_call_duration: 2s # 慢调用时长阈值
  min_requests: 10       # 最小请求数
  wait_duration: 30s     # 等待时长
```

## 8. 接口文档生成

### 8.1 OpenAPI/Swagger文档

使用grpc-gateway生成HTTP接口的OpenAPI文档：

```yaml
swagger: "2.0"
info:
  title: "电商系统API"
  version: "1.0.0"
host: "api.example.com"
schemes:
  - "https"
  - "http"
basePath: "/api/v1"
```

### 8.2 文档自动化

通过CI/CD流程自动生成和更新API文档：

```bash
# 生成protobuf文件
protoc --go_out=. --go-grpc_out=. api/payment/v1/payment.proto
protoc --go_out=. --go-grpc_out=. api/logistics/v1/logistics.proto

# 生成grpc-gateway代码
protoc --grpc-gateway_out=. api/payment/v1/payment.proto
protoc --grpc-gateway_out=. api/logistics/v1/logistics.proto

# 生成OpenAPI文档
protoc --openapiv2_out=. api/payment/v1/payment.proto
protoc --openapiv2_out=. api/logistics/v1/logistics.proto
```

## 9. 测试支持

### 9.1 模拟接口

为支持开发和测试，提供模拟接口：

```go
// 支付服务模拟器
type PaymentSimulator struct {
    successRate float64 // 支付成功率
    delayMs     int     // 模拟延迟
}

// 物流服务模拟器
type LogisticsSimulator struct {
    trackUpdateInterval time.Duration // 轨迹更新间隔
    deliveryDays        int           // 配送天数
}
```

### 9.2 接口测试

提供完整的接口测试套件：

```go
func TestPaymentAPI(t *testing.T) {
    // 测试创建支付
    payment := createTestPayment(t)
    assert.NotEmpty(t, payment.PaymentSn)
    
    // 测试查询支付状态
    status := getPaymentStatus(t, payment.PaymentSn)
    assert.Equal(t, PaymentStatusPending, status.PaymentStatus)
    
    // 测试模拟支付成功
    simulatePaymentSuccess(t, payment.PaymentSn)
    
    // 验证状态变更
    status = getPaymentStatus(t, payment.PaymentSn)
    assert.Equal(t, PaymentStatusPaid, status.PaymentStatus)
}
```

这个API接口设计文档提供了完整的gRPC和RESTful API定义，支持支付和物流服务的所有核心功能，同时考虑了认证、鉴权、限流、熔断等生产环境的关键需求。