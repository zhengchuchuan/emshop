# 错误码管理最佳实践

## 概述

EMShop项目采用统一的错误码管理机制，支持语义化的HTTP状态码和自定义业务错误码。

## HTTP状态码规范

### 支持的状态码范围

```go
// 成功响应
200 - OK                    // 请求成功
201 - Created              // 资源创建成功
202 - Accepted             // 请求已接受，异步处理
204 - No Content           // 成功但无返回内容

// 客户端错误 (4xx)
400 - Bad Request          // 请求参数错误
401 - Unauthorized         // 未认证
403 - Forbidden           // 无权限访问
404 - Not Found           // 资源不存在
405 - Method Not Allowed  // HTTP方法不允许
409 - Conflict            // 冲突（并发、重复操作）
422 - Unprocessable Entity // 业务逻辑验证失败
429 - Too Many Requests   // 请求频率限制

// 服务器错误 (5xx)
500 - Internal Server Error // 服务器内部错误
502 - Bad Gateway          // 网关错误
503 - Service Unavailable  // 服务不可用
504 - Gateway Timeout      // 网关超时
```

## 状态码选择指南

### 4xx 客户端错误的区分

- **400 Bad Request**: 请求格式错误、参数类型错误
- **409 Conflict**: 资源状态冲突，如优惠券已使用、重复参与秒杀
- **422 Unprocessable Entity**: 业务逻辑验证失败，如优惠券已过期、活动未开始
- **429 Too Many Requests**: 频率限制，如用户超过领取限额

### 业务场景映射

```go
// 优惠券服务错误码示例
register(101010, 422, "Coupon has expired")              // 业务逻辑：过期
register(101011, 409, "Coupon has been used")            // 冲突：已使用
register(101013, 429, "Coupon usage limit exceeded")     // 限流：超限额
register(101016, 409, "Flash sale stock is empty")       // 冲突：库存竞争
```

## 错误码分配规则

### 服务级错误码范围

```
100001-100999: 通用错误码
101001-101999: 优惠券服务 (coupon)
102001-102999: 订单服务 (order)
103001-103999: 支付服务 (payment)
104001-104999: 商品服务 (goods)
105001-105999: 用户服务 (user)
106001-106999: 库存服务 (inventory)
107001-107999: 物流服务 (logistics)
```

### 功能子模块分配

```
// 优惠券服务内部分配
101001-101099: 基础错误
101100-101199: 优惠券模板管理
101200-101299: 用户优惠券操作
101300-101399: 秒杀活动管理
101400-101499: 分布式事务相关
```

## 错误处理最佳实践

### 1. 错误码定义

```go
package code

const (
    // 使用语义化命名和清晰注释
    ErrCouponExpired int = iota + 101010  // 优惠券已过期
    ErrCouponUsed                         // 优惠券已使用
    ErrFlashSaleStockEmpty               // 秒杀库存为空
)
```

### 2. 错误码注册

```go
// 选择合适的HTTP状态码
register(101010, 422, "Coupon has expired")           // 业务逻辑错误用422
register(101011, 409, "Coupon has been used")         // 状态冲突用409
register(101013, 429, "Usage limit exceeded")         // 限流用429
```

### 3. 业务层使用

```go
func (s *Service) UseCoupon(ctx context.Context, req *dto.UseCouponDTO) error {
    coupon, err := s.data.GetUserCoupon(req.CouponID)
    if err != nil {
        return errors.WithCode(err, ErrCouponNotFound)
    }
    
    if coupon.Status == CouponStatusUsed {
        return errors.WithCode(nil, ErrCouponUsed)  // 409冲突
    }
    
    if time.Now().After(coupon.ExpireTime) {
        return errors.WithCode(nil, ErrCouponExpired)  // 422业务逻辑错误
    }
    
    return nil
}
```

### 4. 客户端处理

```javascript
// 前端根据状态码进行不同处理
switch (response.status) {
    case 409: // Conflict
        showMessage('优惠券状态冲突，请刷新后重试');
        break;
    case 422: // Unprocessable Entity  
        showMessage('优惠券不满足使用条件');
        break;
    case 429: // Too Many Requests
        showMessage('操作过于频繁，请稍后再试');
        break;
    default:
        showMessage('系统错误，请联系客服');
}
```

## 监控和告警

### 错误码统计

```go
// 按状态码统计错误分布
metrics.Counter("http_requests_total", 
    prometheus.Labels{
        "method": "POST",
        "endpoint": "/coupon/use", 
        "status": "409",
        "error_code": "101011",
    }).Inc()
```

### 告警规则

```yaml
# 业务错误率告警
- alert: HighCouponConflictRate
  expr: rate(http_requests_total{status="409",service="coupon"}[5m]) > 0.1
  for: 2m
  annotations:
    description: "优惠券冲突率过高，可能存在并发问题"

- alert: HighRateLimitRate  
  expr: rate(http_requests_total{status="429",service="coupon"}[5m]) > 0.05
  for: 1m
  annotations:
    description: "优惠券限流触发频繁，需要调整限流策略"
```

## 迁移指南

### 从旧系统迁移

1. **识别现有错误场景**
   - 统计现有错误类型和频率
   - 分析业务语义和处理逻辑

2. **重新映射状态码**
   - 将通用400错误细分为409/422/429
   - 区分客户端错误和服务器错误

3. **渐进式更新**
   - 优先更新高频错误场景
   - 保持向后兼容性

## 总结

通过扩展HTTP状态码支持和语义化错误分类，我们实现了：

1. **更准确的语义表达** - 409冲突、422业务逻辑错误、429限流
2. **更好的客户端体验** - 前端可以根据状态码做差异化处理
3. **更精准的监控告警** - 可以针对不同错误类型设置专门的告警规则
4. **更清晰的问题定位** - 运维人员可以快速识别问题类型

这种改进使错误处理更加规范化和工程化，提升了整个微服务架构的可维护性。