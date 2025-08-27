# emshop API设计文档 - 新服务接口

## 项目概述

为emshop商城系统设计C端API接口，集成coupon（优惠券）、logistics（物流）、payment（支付）三个微服务，为前端商城页面提供完整的业务支持。

## 架构设计

### 整体架构
采用现有emshop的四层架构模式：
- **Controller层**：处理HTTP请求，参数验证，响应格式化
- **Service层**：业务逻辑处理，调用RPC客户端
- **Data层**：RPC客户端封装，与微服务通信
- **Domain层**：DTO对象定义，请求响应结构

### 技术栈
- **Web框架**：Gin
- **服务发现**：Consul
- **RPC通信**：gRPC
- **认证方式**：JWT
- **响应格式**：统一使用core.WriteResponse

## API接口设计

### 1. 优惠券服务API

#### 路由组：`/v1/coupons`

##### 1.1 获取可领取优惠券列表
```
GET /v1/coupons/available
```
**描述**：获取当前可领取的优惠券模板列表  
**认证**：不需要  
**参数**：
- `page` (query, int): 页码，默认1
- `pageSize` (query, int): 页大小，默认10

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "items": [
      {
        "id": 1,
        "name": "新用户专享券",
        "discount_type": 1,
        "discount_value": 10.0,
        "min_order_amount": 100.0,
        "valid_days": 30,
        "total_count": 1000,
        "received_count": 345,
        "per_user_limit": 1,
        "description": "新用户专享优惠券",
        "status": 1
      }
    ]
  }
}
```

##### 1.2 领取优惠券
```
POST /v1/coupons/receive
```
**描述**：用户领取指定优惠券  
**认证**：需要JWT  
**请求体**：
```json
{
  "coupon_template_id": 1
}
```

**响应示例**：
```json
{
  "code": 0,
  "message": "领取成功",
  "data": {
    "id": 123,
    "coupon_code": "COUP20250827001",
    "template_name": "新用户专享券",
    "discount_value": 10.0,
    "expired_at": "2025-09-27T00:00:00Z",
    "status": 1
  }
}
```

##### 1.3 我的优惠券列表
```
GET /v1/coupons/my
```
**描述**：获取用户的优惠券列表  
**认证**：需要JWT  
**参数**：
- `status` (query, int): 状态筛选，1-未使用，2-已使用，3-已过期
- `page` (query, int): 页码，默认1
- `pageSize` (query, int): 页大小，默认10

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "items": [
      {
        "id": 123,
        "coupon_code": "COUP20250827001",
        "template": {
          "name": "新用户专享券",
          "discount_type": 1,
          "discount_value": 10.0,
          "min_order_amount": 100.0
        },
        "status": 1,
        "received_at": "2025-08-27T10:30:00Z",
        "expired_at": "2025-09-27T00:00:00Z"
      }
    ]
  }
}
```

##### 1.4 获取可用优惠券（下单使用）
```
GET /v1/coupons/available-for-order
```
**描述**：根据订单金额获取用户可用的优惠券  
**认证**：需要JWT  
**参数**：
- `order_amount` (query, float): 订单金额

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "available_coupons": [
      {
        "id": 123,
        "coupon_code": "COUP20250827001",
        "template": {
          "name": "满100减10",
          "discount_value": 10.0,
          "min_order_amount": 100.0
        },
        "can_use": true,
        "discount_amount": 10.0
      }
    ]
  }
}
```

##### 1.5 计算优惠折扣
```
POST /v1/coupons/calculate-discount
```
**描述**：计算指定优惠券的折扣金额  
**认证**：需要JWT  
**请求体**：
```json
{
  "coupon_ids": [123, 124],
  "order_amount": 150.0,
  "order_items": [
    {
      "goods_id": 1,
      "quantity": 2,
      "price": 75.0
    }
  ]
}
```

### 2. 支付服务API

#### 路由组：`/v1/payment`

##### 2.1 创建支付订单
```
POST /v1/payment/create
```
**描述**：创建支付订单  
**认证**：需要JWT  
**请求体**：
```json
{
  "order_sn": "ORD20250827001",
  "amount": 140.0,
  "payment_method": 1,
  "expired_minutes": 15
}
```

##### 2.2 查询支付状态
```
GET /v1/payment/:payment_sn/status
```
**描述**：查询支付订单状态  
**认证**：需要JWT  

##### 2.3 模拟支付成功（测试用）
```
POST /v1/payment/:payment_sn/simulate-success
```
**描述**：模拟支付成功，用于测试  
**认证**：需要JWT  

##### 2.4 模拟支付失败（测试用）
```
POST /v1/payment/:payment_sn/simulate-failure
```
**描述**：模拟支付失败，用于测试  
**认证**：需要JWT  

### 3. 物流服务API

#### 路由组：`/v1/logistics`

##### 3.1 查询物流信息
```
GET /v1/logistics/info
```
**描述**：根据订单号或物流单号查询物流信息  
**认证**：需要JWT  
**参数**：
- `order_sn` (query, string): 订单号
- `logistics_sn` (query, string): 物流单号
- `tracking_number` (query, string): 快递单号

##### 3.2 查看物流轨迹
```
GET /v1/logistics/tracks
```
**描述**：查看详细的物流轨迹  
**认证**：需要JWT  
**参数**：
- `order_sn` (query, string): 订单号
- `logistics_sn` (query, string): 物流单号

##### 3.3 计算运费
```
POST /v1/logistics/calculate-fee
```
**描述**：计算订单的运费  
**认证**：不需要  
**请求体**：
```json
{
  "receiver_address": "北京市朝阳区...",
  "items": [
    {
      "goods_id": 1,
      "quantity": 2,
      "weight": 1.5,
      "volume": 0.01
    }
  ],
  "logistics_company": 1,
  "shipping_method": 1
}
```

##### 3.4 获取物流公司列表
```
GET /v1/logistics/companies
```
**描述**：获取支持的物流公司列表  
**认证**：不需要  

## 错误处理

所有API遵循统一的错误响应格式：

```json
{
  "code": 40001,
  "message": "参数错误：缺少必需的参数order_amount",
  "data": null
}
```

### 错误码说明
- `40001-40999`: 客户端错误
- `50001-50999`: 服务器错误  
- `60001-60999`: 业务逻辑错误

## 安全考虑

1. **认证授权**：敏感操作需要JWT认证
2. **参数验证**：所有输入参数进行严格验证
3. **限流防护**：对频繁操作（如领券）进行限流
4. **数据脱敏**：响应数据中敏感信息脱敏处理

## 实施计划

### 阶段一：基础设施（第1-2天）
1. 扩展data层接口定义
2. 创建RPC客户端封装
3. 创建DTO对象定义

### 阶段二：核心功能（第3-4天）  
1. 实现服务层业务逻辑
2. 实现控制器层API端点
3. 更新路由配置

### 阶段三：集成测试（第5天）
1. 集成测试和调试
2. API文档完善
3. 性能优化

## 依赖关系

- 依赖于coupon、logistics、payment三个RPC服务
- 需要JWT认证中间件
- 需要Consul服务发现
- 需要Redis缓存（用于限流等功能）

## 注意事项

1. **向后兼容**：新API不影响现有功能
2. **性能考虑**：合理使用缓存，避免频繁RPC调用
3. **监控告警**：添加关键业务指标监控
4. **文档维护**：及时更新API文档和用户手册