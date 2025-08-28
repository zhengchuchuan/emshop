# EmShop API 使用文档

## 目录
- [项目概览](#项目概览)
- [认证机制](#认证机制)
- [用户服务 API](#用户服务-api)
- [基础服务 API](#基础服务-api)
- [商品服务 API](#商品服务-api)
- [订单服务 API](#订单服务-api)
- [用户操作 API](#用户操作-api)
- [优惠券服务 API](#优惠券服务-api)
- [支付服务 API](#支付服务-api)
- [物流服务 API](#物流服务-api)
- [错误处理](#错误处理)

---

## 项目概览

EmShop 是一个基于微服务架构的电商平台 API，使用 Go 语言开发，采用 Gin 框架构建 RESTful API。

### 技术栈
- **框架**: Gin Web Framework
- **认证**: JWT Token
- **架构**: 微服务架构
- **数据格式**: JSON

### API 基础信息
- **API 版本**: v1
- **基础路径**: `/v1`
- **请求格式**: JSON
- **响应格式**: JSON

---

## 认证机制

### JWT Token 认证
大部分 API 需要在请求头中携带 JWT Token：

```
Authorization: Bearer <your-jwt-token>
```

### 获取 Token
通过用户登录接口获取 Token：
- 登录成功后会返回 `token` 和 `expiredAt`
- Token 过期需要重新登录获取

---

## 用户服务 API

### 用户登录
**POST** `/v1/user/pwd_login`

**请求体:**
```json
{
  "mobile": "13800138000",
  "password": "password123"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "nickName": "用户昵称",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiredAt": 1640995200
  }
}
```

### 用户注册
**POST** `/v1/user/register`

**请求体:**
```json
{
  "mobile": "13800138000",
  "password": "password123",
  "name": "用户昵称",
  "captcha": "1234",
  "captchaId": "captcha_id_123"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "nickName": "用户昵称",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiredAt": 1640995200
  }
}
```

### 获取用户详情
**GET** `/v1/user/detail`

**认证**: 需要 JWT Token

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "mobile": "13800138000",
    "name": "用户昵称",
    "birthday": "1990-01-01",
    "gender": "male",
    "role": 1
  }
}
```

### 更新用户信息
**PATCH** `/v1/user/update`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "name": "新昵称",
  "gender": "female",
  "birthday": "1995-05-15"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "更新成功"
  }
}
```

---

## 基础服务 API

### 发送短信验证码
**POST** `/v1/base/send_sms`

**请求体:**
```json
{
  "mobile": "13800138000",
  "type": 1
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "发送成功"
  }
}
```

### 获取图形验证码
**GET** `/v1/base/captcha`

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "captchaId": "captcha_id_123",
    "picPath": "base64://iVBORw0KGgoAAAANSUhEUgAA..."
  }
}
```

---

## 商品服务 API

### 商品列表
**GET** `/v1/goods`

**查询参数:**
- `pages` (可选): 页码，默认 1
- `pagePerNums` (可选): 每页数量，默认 10
- `priceMin` (可选): 最低价格
- `priceMax` (可选): 最高价格
- `isHot` (可选): 是否热销商品
- `isNew` (可选): 是否新品
- `topCategory` (可选): 分类ID
- `brand` (可选): 品牌ID
- `keyWords` (可选): 关键词搜索

**示例请求:**
```
GET /v1/goods?pages=1&pagePerNums=10&isHot=true
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "data": [
      {
        "id": 1,
        "name": "商品名称",
        "goodsBrief": "商品简介",
        "desc": "商品描述",
        "shopPrice": 99.99,
        "frontImage": "https://example.com/image.jpg",
        "images": ["https://example.com/1.jpg"],
        "category": {
          "id": 1,
          "name": "分类名称"
        },
        "brand": {
          "id": 1,
          "name": "品牌名称",
          "logo": "https://example.com/logo.jpg"
        },
        "isHot": true,
        "isNew": false,
        "onSale": true
      }
    ]
  }
}
```

### 商品详情
**GET** `/v1/goods/:id`

**路径参数:**
- `id`: 商品ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "商品名称",
    "goodsBrief": "商品简介",
    "desc": "商品描述",
    "shopPrice": 99.99,
    "frontImage": "https://example.com/image.jpg",
    "images": ["https://example.com/1.jpg", "https://example.com/2.jpg"],
    "descImages": ["https://example.com/desc1.jpg"],
    "category": {
      "id": 1,
      "name": "分类名称"
    },
    "brand": {
      "id": 1,
      "name": "品牌名称",
      "logo": "https://example.com/logo.jpg"
    },
    "shipFree": true,
    "isHot": true,
    "isNew": false,
    "onSale": true
  }
}
```

### 商品库存
**GET** `/v1/goods/:id/stocks`

**路径参数:**
- `id`: 商品ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "stocks": 100,
    "goodsId": 1
  }
}
```

### 商品分类列表
**GET** `/v1/categorys`

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "分类名称",
      "level": 1,
      "parentCategory": 0,
      "isTab": true,
      "subCategories": [
        {
          "id": 2,
          "name": "子分类",
          "level": 2,
          "parentCategory": 1,
          "isTab": false
        }
      ]
    }
  ]
}
```

### 分类详情
**GET** `/v1/categorys/:id`

**路径参数:**
- `id`: 分类ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "分类名称",
    "level": 1,
    "parentCategory": 0,
    "isTab": true,
    "subCategories": [
      {
        "id": 2,
        "name": "子分类",
        "level": 2,
        "parentCategory": 1,
        "isTab": false
      }
    ]
  }
}
```

### 品牌列表
**GET** `/v1/brands`

**查询参数:**
- `pages` (可选): 页码，默认 1
- `pagePerNums` (可选): 每页数量，默认 10

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "data": [
      {
        "id": 1,
        "name": "品牌名称",
        "logo": "https://example.com/logo.jpg"
      }
    ]
  }
}
```

### 轮播图列表
**GET** `/v1/banners`

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "data": [
      {
        "id": 1,
        "index": 1,
        "image": "https://example.com/banner.jpg",
        "url": "https://example.com/link"
      }
    ]
  }
}
```

---

## 订单服务 API

### 订单列表
**GET** `/v1/orders`

**认证**: 需要 JWT Token

**查询参数:**
- `pages` (可选): 页码，默认 1
- `pagePerNums` (可选): 每页数量，默认 10

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 10,
    "data": [
      {
        "id": 1,
        "orderSn": "ORDER20231201001",
        "status": "TRADE_SUCCESS",
        "payType": "ALIPAY",
        "total": 299.99,
        "address": "北京市朝阳区xxx",
        "name": "张三",
        "mobile": "13800138000",
        "addTime": "2023-12-01T10:00:00Z"
      }
    ]
  }
}
```

### 创建订单
**POST** `/v1/orders`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "address": "北京市朝阳区xxx街道xxx号",
  "name": "张三",
  "mobile": "13800138000",
  "post": "100000"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "订单创建成功"
  }
}
```

### 订单详情
**GET** `/v1/orders/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 订单ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "orderSn": "ORDER20231201001",
    "status": "TRADE_SUCCESS",
    "payType": "ALIPAY",
    "total": 299.99,
    "address": "北京市朝阳区xxx",
    "name": "张三",
    "mobile": "13800138000",
    "addTime": "2023-12-01T10:00:00Z",
    "goods": [
      {
        "id": 1,
        "goodsId": 101,
        "goodsName": "商品名称",
        "goodsImage": "https://example.com/goods.jpg",
        "goodsPrice": 99.99,
        "nums": 2
      }
    ]
  }
}
```

### 购物车列表
**GET** `/v1/shopcarts`

**认证**: 需要 JWT Token

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 3,
    "data": [
      {
        "id": 1,
        "goodsId": 101,
        "nums": 2,
        "checked": true
      }
    ]
  }
}
```

### 添加到购物车
**POST** `/v1/shopcarts`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "goods": 101,
  "nums": 2
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "goodsId": 101,
    "nums": 2,
    "checked": true
  }
}
```

### 更新购物车商品
**PATCH** `/v1/shopcarts/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 购物车条目ID

**请求体:**
```json
{
  "nums": 3,
  "checked": false
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "更新成功"
  }
}
```

### 删除购物车商品
**DELETE** `/v1/shopcarts/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 购物车条目ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "删除成功"
  }
}
```

---

## 用户操作 API

### 用户收藏列表
**GET** `/v1/userfavs`

**认证**: 需要 JWT Token

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "data": [
      {
        "userId": 1,
        "goodsId": 101
      }
    ]
  }
}
```

### 添加收藏
**POST** `/v1/userfavs`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "goods": 101
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "userId": 1,
    "goodsId": 101
  }
}
```

### 取消收藏
**DELETE** `/v1/userfavs/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 商品ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "取消收藏成功"
  }
}
```

### 查看收藏状态
**GET** `/v1/userfavs/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 商品ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "userId": 1,
    "goodsId": 101
  }
}
```

### 用户地址列表
**GET** `/v1/address`

**认证**: 需要 JWT Token

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 3,
    "data": [
      {
        "id": 1,
        "province": "北京市",
        "city": "北京市",
        "district": "朝阳区",
        "address": "xxx街道xxx号",
        "signerName": "张三",
        "signerMobile": "13800138000"
      }
    ]
  }
}
```

### 创建地址
**POST** `/v1/address`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "province": "北京市",
  "city": "北京市",
  "district": "朝阳区",
  "address": "xxx街道xxx号",
  "signerName": "张三",
  "signerMobile": "13800138000"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "province": "北京市",
    "city": "北京市",
    "district": "朝阳区",
    "address": "xxx街道xxx号",
    "signerName": "张三",
    "signerMobile": "13800138000"
  }
}
```

### 更新地址
**PUT** `/v1/address/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 地址ID

**请求体:**
```json
{
  "province": "上海市",
  "city": "上海市",
  "district": "浦东新区",
  "address": "xxx路xxx号",
  "signerName": "李四",
  "signerMobile": "13900139000"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "更新成功"
  }
}
```

### 删除地址
**DELETE** `/v1/address/:id`

**认证**: 需要 JWT Token

**路径参数:**
- `id`: 地址ID

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "msg": "删除成功"
  }
}
```

### 用户留言列表
**GET** `/v1/message`

**认证**: 需要 JWT Token

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "data": [
      {
        "id": 1,
        "messageType": 1,
        "subject": "留言主题",
        "message": "留言内容",
        "file": "https://example.com/attachment.pdf"
      }
    ]
  }
}
```

### 创建留言
**POST** `/v1/message`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "type": 1,
  "subject": "留言主题",
  "message": "留言内容",
  "file": "https://example.com/attachment.pdf"
}
```

**字段说明:**
- `type`: 留言类型，1-5之间的数字

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "messageType": 1,
    "subject": "留言主题",
    "message": "留言内容",
    "file": "https://example.com/attachment.pdf"
  }
}
```

---

## 优惠券服务 API

### 获取优惠券模板列表
**GET** `/v1/coupons/templates`

**查询参数:**
- `page`: 页码，必填，最小值 1
- `pageSize`: 每页数量，必填，1-50之间

**示例请求:**
```
GET /v1/coupons/templates?page=1&pageSize=10
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "templates": [
      {
        "id": 1,
        "name": "满100减10优惠券",
        "type": 1,
        "value": 10.00,
        "minOrderAmount": 100.00,
        "description": "满100元可用",
        "startTime": "2023-12-01T00:00:00Z",
        "endTime": "2023-12-31T23:59:59Z",
        "totalCount": 1000,
        "usedCount": 100,
        "status": 1
      }
    ],
    "total": 50,
    "page": 1,
    "pageSize": 10
  }
}
```

### 用户领取优惠券
**POST** `/v1/coupons/receive`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "template_id": 1
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "couponId": 123,
    "templateId": 1,
    "userId": 1,
    "status": 1,
    "receivedAt": "2023-12-01T10:00:00Z",
    "expireAt": "2023-12-31T23:59:59Z"
  }
}
```

### 获取用户优惠券列表
**GET** `/v1/coupons/user`

**认证**: 需要 JWT Token

**查询参数:**
- `status` (可选): 优惠券状态筛选
- `page`: 页码，必填，最小值 1
- `pageSize`: 每页数量，必填，1-50之间

**示例请求:**
```
GET /v1/coupons/user?page=1&pageSize=10&status=1
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "coupons": [
      {
        "id": 123,
        "templateId": 1,
        "name": "满100减10优惠券",
        "type": 1,
        "value": 10.00,
        "minOrderAmount": 100.00,
        "status": 1,
        "receivedAt": "2023-12-01T10:00:00Z",
        "expireAt": "2023-12-31T23:59:59Z"
      }
    ],
    "total": 5,
    "page": 1,
    "pageSize": 10
  }
}
```

### 获取用户可用优惠券
**GET** `/v1/coupons/available`

**认证**: 需要 JWT Token

**查询参数:**
- `order_amount`: 订单金额，必填，最小值 0.01

**示例请求:**
```
GET /v1/coupons/available?order_amount=150.00
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "availableCoupons": [
      {
        "id": 123,
        "templateId": 1,
        "name": "满100减10优惠券",
        "type": 1,
        "value": 10.00,
        "minOrderAmount": 100.00,
        "canUse": true,
        "reason": ""
      }
    ]
  }
}
```

### 计算优惠券折扣
**POST** `/v1/coupons/calculate-discount`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "coupon_ids": [123, 124],
  "order_amount": 200.00,
  "order_items": [
    {
      "goods_id": 101,
      "quantity": 2,
      "price": 50.00
    },
    {
      "goods_id": 102,
      "quantity": 1,
      "price": 100.00
    }
  ]
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "originalAmount": 200.00,
    "discountAmount": 15.00,
    "finalAmount": 185.00,
    "appliedCoupons": [
      {
        "couponId": 123,
        "name": "满100减10优惠券",
        "discountAmount": 10.00
      },
      {
        "couponId": 124,
        "name": "满200减5优惠券",
        "discountAmount": 5.00
      }
    ],
    "unusedCoupons": []
  }
}
```

---

## 支付服务 API

### 创建支付订单
**POST** `/v1/payments`

**认证**: 需要 JWT Token

**请求体:**
```json
{
  "orderSn": "ORDER20231201001",
  "payType": "ALIPAY",
  "amount": 299.99,
  "subject": "商品支付",
  "description": "订单商品支付"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "paymentSn": "PAY20231201001",
    "orderId": 1,
    "amount": 299.99,
    "payType": "ALIPAY",
    "status": "WAIT_BUYER_PAY",
    "payUrl": "https://openapi.alipaydev.com/gateway.do?...",
    "createdAt": "2023-12-01T10:00:00Z",
    "expireAt": "2023-12-01T10:30:00Z"
  }
}
```

### 获取支付状态
**GET** `/v1/payments/:paymentSN/status`

**路径参数:**
- `paymentSN`: 支付单号

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "paymentSn": "PAY20231201001",
    "status": "TRADE_SUCCESS",
    "amount": 299.99,
    "payType": "ALIPAY",
    "paidAt": "2023-12-01T10:05:00Z",
    "tradeNo": "2023120122001..."
  }
}
```

### 模拟支付
**POST** `/v1/payments/:paymentSN/simulate`

**路径参数:**
- `paymentSN`: 支付单号

**请求体:**
```json
{
  "success": true,
  "tradeNo": "2023120122001..."
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "paymentSn": "PAY20231201001",
    "status": "TRADE_SUCCESS",
    "simulatedAt": "2023-12-01T10:05:00Z",
    "tradeNo": "2023120122001..."
  }
}
```

---

## 物流服务 API

### 获取物流信息
**GET** `/v1/logistics/info`

**查询参数:**
- `orderSn`: 订单号，必填
- `logisticsCompany`: 物流公司代码，可选

**示例请求:**
```
GET /v1/logistics/info?orderSn=ORDER20231201001&logisticsCompany=SF
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "orderSn": "ORDER20231201001",
    "logisticsNumber": "SF1234567890",
    "logisticsCompany": "SF",
    "logisticsCompanyName": "顺丰速运",
    "status": "DELIVERED",
    "statusName": "已签收",
    "senderInfo": {
      "name": "发货人",
      "phone": "400-111-1111",
      "address": "广东省深圳市..."
    },
    "receiverInfo": {
      "name": "收货人",
      "phone": "13800138000",
      "address": "北京市朝阳区..."
    }
  }
}
```

### 获取物流轨迹
**GET** `/v1/logistics/tracks`

**查询参数:**
- `logisticsNumber`: 物流单号，必填
- `logisticsCompany`: 物流公司代码，可选

**示例请求:**
```
GET /v1/logistics/tracks?logisticsNumber=SF1234567890&logisticsCompany=SF
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "logisticsNumber": "SF1234567890",
    "logisticsCompany": "SF",
    "status": "DELIVERED",
    "tracks": [
      {
        "time": "2023-12-01T14:30:00Z",
        "location": "北京市朝阳区",
        "status": "DELIVERED",
        "description": "快件已签收，签收人：本人签收"
      },
      {
        "time": "2023-12-01T10:00:00Z",
        "location": "北京市朝阳区",
        "status": "OUT_FOR_DELIVERY",
        "description": "快件正在派送中"
      },
      {
        "time": "2023-11-30T18:00:00Z",
        "location": "北京转运中心",
        "status": "IN_TRANSIT",
        "description": "快件已到达北京转运中心"
      }
    ]
  }
}
```

### 计算运费
**POST** `/v1/logistics/shipping-fee`

**请求体:**
```json
{
  "fromProvince": "广东省",
  "fromCity": "深圳市",
  "toProvince": "北京市",
  "toCity": "北京市",
  "weight": 1.5,
  "volume": 0.01,
  "logisticsCompany": "SF"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "shippingFee": 15.00,
    "logisticsCompany": "SF",
    "logisticsCompanyName": "顺丰速运",
    "estimatedDeliveryTime": "1-2个工作日",
    "feeDetails": [
      {
        "type": "BASIC_FEE",
        "name": "基础费用",
        "fee": 12.00
      },
      {
        "type": "WEIGHT_FEE",
        "name": "重量费用",
        "fee": 3.00
      }
    ]
  }
}
```

### 获取物流公司列表
**GET** `/v1/logistics/companies`

**响应:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "companies": [
      {
        "code": "SF",
        "name": "顺丰速运",
        "logo": "https://example.com/sf-logo.png",
        "phone": "400-111-1111",
        "website": "https://www.sf-express.com",
        "supportCod": true,
        "supportInsurance": true
      },
      {
        "code": "YTO",
        "name": "圆通速递",
        "logo": "https://example.com/yto-logo.png",
        "phone": "400-222-2222",
        "website": "https://www.yto.net.cn",
        "supportCod": false,
        "supportInsurance": true
      }
    ]
  }
}
```

---

## 错误处理

### 错误响应格式
所有错误响应都遵循统一格式：

```json
{
  "code": 400,
  "message": "请求参数错误",
  "data": null
}
```

### 常见错误代码

| 错误代码 | 说明 | 解决方案 |
|---------|------|----------|
| 400 | 请求参数错误 | 检查请求参数格式和必填项 |
| 401 | 未授权/Token无效 | 重新登录获取Token |
| 403 | 权限不足 | 检查用户权限或联系管理员 |
| 404 | 资源不存在 | 确认请求的资源ID是否正确 |
| 409 | 资源冲突 | 检查是否存在重复数据 |
| 429 | 请求频率过高 | 降低请求频率或稍后重试 |
| 500 | 服务器内部错误 | 联系技术支持 |

### 参数验证错误
当请求参数验证失败时，返回详细的错误信息：

```json
{
  "code": 400,
  "message": "参数验证失败",
  "data": {
    "errors": [
      {
        "field": "mobile",
        "message": "手机号格式不正确"
      },
      {
        "field": "password",
        "message": "密码长度不能少于6位"
      }
    ]
  }
}
```

### 业务逻辑错误
业务相关错误会返回具体的错误信息：

```json
{
  "code": 4001,
  "message": "商品库存不足",
  "data": {
    "goodsId": 101,
    "requestedNum": 5,
    "availableNum": 2
  }
}
```

---

## 附录

### 状态枚举值

#### 订单状态
- `TRADE_SUCCESS`: 交易成功
- `WAIT_BUYER_PAY`: 等待买家付款
- `TRADE_CLOSED`: 交易关闭
- `TRADE_FINISHED`: 交易完结

#### 支付状态
- `WAIT_BUYER_PAY`: 等待买家付款
- `TRADE_SUCCESS`: 交易成功
- `TRADE_CLOSED`: 交易关闭
- `TRADE_FINISHED`: 交易完结

#### 物流状态
- `PENDING`: 待发货
- `SHIPPED`: 已发货
- `IN_TRANSIT`: 运输中
- `OUT_FOR_DELIVERY`: 派送中
- `DELIVERED`: 已签收
- `RETURNED`: 已退回

#### 优惠券状态
- `1`: 可用
- `2`: 已使用
- `3`: 已过期
- `4`: 已作废

### 时间格式
所有时间字段均使用 ISO 8601 格式：`2023-12-01T10:00:00Z`

### 金额格式
所有金额字段均为浮点数，单位为元，保留两位小数。

---

**最后更新**: 2023-12-01  
**API版本**: v1.0.0  
**文档版本**: 1.0.0