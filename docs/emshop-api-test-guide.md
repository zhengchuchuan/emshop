# EMShop API 测试指南

**版本**: v2.1-stable  
**生成日期**: 2025-08-26  
**API基础地址**: `http://localhost:8051`

## 🚀 快速开始

### 环境要求
- EMShop Shop API服务运行在端口8051
- 所有微服务正常运行
- JWT认证已启用

### 测试流程概览
1. 用户注册/登录获取JWT Token
2. 浏览商品并添加到购物车
3. 创建订单并验证分布式事务
4. 查看订单状态和详情

---

## 📋 API 端点总览

### 🔐 认证相关
- `POST /v1/user/register` - 用户注册
- `POST /v1/user/pwd_login` - 用户登录
- `GET /v1/user/detail` - 获取用户详情 (需认证)
- `PATCH /v1/user/update` - 更新用户信息 (需认证)

### 🛒 购物车管理
- `GET /v1/shopcarts` - 获取购物车列表 (需认证)
- `POST /v1/shopcarts` - 添加商品到购物车 (需认证)  
- `PATCH /v1/shopcarts/:id` - 更新购物车商品 (需认证)
- `DELETE /v1/shopcarts/:id` - 删除购物车商品 (需认证)

### 📦 订单管理
- `GET /v1/orders` - 获取订单列表 (需认证)
- `POST /v1/orders` - 创建订单 (需认证)
- `GET /v1/orders/:id` - 获取订单详情 (需认证)

### 🛍️ 商品浏览
- `GET /v1/goods` - 商品列表
- `GET /v1/goods/:id` - 商品详情
- `GET /v1/goods/:id/stocks` - 商品库存
- `GET /v1/categorys` - 商品分类列表
- `GET /v1/brands` - 品牌列表

---

## 🔧 详细测试用例

### 1. 用户认证流程

#### 1.1 用户注册
```bash
curl -X POST http://localhost:8051/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "13800138001",
    "password": "123456",
    "code": "123456"
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "注册成功",
  "data": {
    "id": 1001,
    "mobile": "13800138001",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 1.2 用户登录
```bash
curl -X POST http://localhost:8051/v1/user/pwd_login \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "13800138001",
    "password": "123456"
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "登录成功", 
  "data": {
    "id": 1001,
    "mobile": "13800138001",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expire": "2025-08-27T10:30:00Z"
  }
}
```

#### 1.3 获取用户详情
```bash
curl -X GET http://localhost:8051/v1/user/detail \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1001,
    "mobile": "13800138001",
    "name": "测试用户",
    "gender": "male",
    "birthday": "1990-01-01"
  }
}
```

### 2. 商品浏览

#### 2.1 获取商品列表
```bash
# 基础商品列表
curl -X GET "http://localhost:8051/v1/goods"

# 带筛选条件的商品列表  
curl -X GET "http://localhost:8051/v1/goods?pages=1&pagePerNums=10&isHot=true&priceMin=10&priceMax=1000"

# 关键词搜索
curl -X GET "http://localhost:8051/v1/goods?keyWords=手机&pages=1&pagePerNums=5"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 50,
    "data": [
      {
        "id": 1,
        "name": "iPhone 15 Pro",
        "goodsSn": "IP15P001", 
        "marketPrice": 8999.0,
        "shopPrice": 7999.0,
        "goodsFrontImage": "https://example.com/iphone15.jpg",
        "isNew": true,
        "isHot": true,
        "categoryId": 1,
        "brandId": 1
      }
    ]
  }
}
```

#### 2.2 获取商品详情
```bash
curl -X GET "http://localhost:8051/v1/goods/1"
```

#### 2.3 获取商品库存
```bash
curl -X GET "http://localhost:8051/v1/goods/1/stocks"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success", 
  "data": {
    "goodsId": 1,
    "stocks": 100,
    "sold": 50
  }
}
```

#### 2.4 获取商品分类
```bash
curl -X GET "http://localhost:8051/v1/categorys"
```

### 3. 购物车操作流程

#### 3.1 添加商品到购物车
```bash
curl -X POST http://localhost:8051/v1/shopcarts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "goods": 1,
    "nums": 2
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "添加成功",
  "data": {
    "id": 1,
    "goodsId": 1,
    "nums": 2,
    "checked": true
  }
}
```

#### 3.2 获取购物车列表
```bash
curl -X GET http://localhost:8051/v1/shopcarts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 1,
    "data": [
      {
        "id": 1,
        "goodsId": 1,
        "nums": 2,
        "checked": true
      }
    ]
  }
}
```

#### 3.3 更新购物车商品数量
```bash
curl -X PATCH http://localhost:8051/v1/shopcarts/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nums": 3,
    "checked": true
  }'
```

#### 3.4 删除购物车商品
```bash
curl -X DELETE http://localhost:8051/v1/shopcarts/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 4. 订单管理流程

#### 4.1 创建订单 (分布式事务)
```bash
curl -X POST http://localhost:8051/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "北京市朝阳区测试街道123号",
    "name": "张三",
    "mobile": "13800138001",
    "post": "100000"
  }'
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "订单创建成功",
  "data": {
    "orderSn": "ORDER20250826001",
    "status": "WAIT_BUYER_PAY",
    "total": 15998.0
  }
}
```

#### 4.2 获取订单列表
```bash
# 基础订单列表
curl -X GET http://localhost:8051/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 分页查询
curl -X GET "http://localhost:8051/v1/orders?pages=1&pagePerNums=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 5,
    "data": [
      {
        "id": 1,
        "orderSn": "ORDER20250826001",
        "status": "WAIT_BUYER_PAY",
        "payType": 0,
        "total": 15998.0,
        "address": "北京市朝阳区测试街道123号",
        "name": "张三",
        "mobile": "13800138001",
        "addTime": "2025-08-26T10:30:00Z"
      }
    ]
  }
}
```

#### 4.3 获取订单详情
```bash
curl -X GET http://localhost:8051/v1/orders/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "orderSn": "ORDER20250826001", 
    "status": "WAIT_BUYER_PAY",
    "payType": 0,
    "total": 15998.0,
    "address": "北京市朝阳区测试街道123号",
    "name": "张三",
    "mobile": "13800138001",
    "addTime": "2025-08-26T10:30:00Z",
    "goods": [
      {
        "id": 1,
        "goodsId": 1,
        "goodsName": "iPhone 15 Pro",
        "goodsImage": "https://example.com/iphone15.jpg",
        "goodsPrice": 7999.0,
        "nums": 2
      }
    ]
  }
}
```

### 5. 用户地址管理

#### 5.1 获取地址列表
```bash
curl -X GET http://localhost:8051/v1/address \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 5.2 创建收货地址
```bash
curl -X POST http://localhost:8051/v1/address \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "province": "北京市",
    "city": "朝阳区", 
    "district": "建外街道",
    "address": "测试大厦1号楼101室",
    "signerName": "张三",
    "signerMobile": "13800138001",
    "postCode": "100000"
  }'
```

#### 5.3 更新地址
```bash
curl -X PUT http://localhost:8051/v1/address/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "province": "北京市",
    "city": "海淀区",
    "district": "中关村街道",
    "address": "科技大厦2号楼201室",
    "signerName": "张三",
    "signerMobile": "13800138001",
    "postCode": "100080"
  }'
```

#### 5.4 删除地址
```bash
curl -X DELETE http://localhost:8051/v1/address/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 6. 用户收藏管理

#### 6.1 获取收藏列表
```bash
curl -X GET http://localhost:8051/v1/userfavs \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 6.2 添加商品收藏
```bash
curl -X POST http://localhost:8051/v1/userfavs \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "goods": 1
  }'
```

#### 6.3 取消商品收藏
```bash
curl -X DELETE http://localhost:8051/v1/userfavs/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 6.4 查看商品是否已收藏
```bash
curl -X GET http://localhost:8051/v1/userfavs/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🧪 完整测试流程示例

### 测试场景: 完整的下单流程

```bash
#!/bin/bash

# 设置API基础URL
BASE_URL="http://localhost:8051"
MOBILE="13800138888"
PASSWORD="123456"

echo "=== EMShop 完整下单流程测试 ==="

# 1. 用户注册
echo "1. 用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST ${BASE_URL}/v1/user/register \
  -H "Content-Type: application/json" \
  -d "{
    \"mobile\": \"${MOBILE}\",
    \"password\": \"${PASSWORD}\",
    \"code\": \"123456\"
  }")

echo "注册响应: $REGISTER_RESPONSE"

# 2. 用户登录获取Token
echo "2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST ${BASE_URL}/v1/user/pwd_login \
  -H "Content-Type: application/json" \
  -d "{
    \"mobile\": \"${MOBILE}\",
    \"password\": \"${PASSWORD}\"
  }")

# 提取Token (需要根据实际响应格式调整)
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
echo "Token: $TOKEN"

# 3. 浏览商品
echo "3. 获取商品列表..."
curl -s -X GET "${BASE_URL}/v1/goods?pages=1&pagePerNums=5" | jq '.'

# 4. 添加商品到购物车
echo "4. 添加商品到购物车..."
curl -s -X POST ${BASE_URL}/v1/shopcarts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "goods": 1,
    "nums": 2
  }' | jq '.'

# 5. 查看购物车
echo "5. 查看购物车..."
curl -s -X GET ${BASE_URL}/v1/shopcarts \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 6. 创建订单
echo "6. 创建订单..."
ORDER_RESPONSE=$(curl -s -X POST ${BASE_URL}/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "北京市朝阳区测试街道123号",
    "name": "测试用户",
    "mobile": "'${MOBILE}'",
    "post": "100000"
  }')

echo "订单创建响应: $ORDER_RESPONSE"

# 7. 查看订单列表
echo "7. 查看订单列表..."
curl -s -X GET ${BASE_URL}/v1/orders \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo "=== 测试完成 ==="
```

---

## 🔍 错误处理

### 常见错误码
- `400` - 请求参数错误
- `401` - 未授权 (Token无效或过期)
- `404` - 资源不存在  
- `409` - 资源冲突 (如用户已存在)
- `500` - 服务器内部错误

### 错误响应示例
```json
{
  "code": 401,
  "msg": "Token无效或已过期",
  "data": null
}
```

### 调试建议
1. **Token过期**: 重新登录获取新Token
2. **参数错误**: 检查请求体格式和必填字段
3. **权限问题**: 确认接口是否需要认证
4. **服务不可用**: 检查微服务是否正常运行

---

## 🎯 分布式事务测试

### DTM Saga事务验证

EMShop使用DTM Saga模式处理分布式事务，订单创建涉及以下步骤:

1. **库存扣减** (`Inventory/Sell`) ↔ **库存归还** (`Inventory/Reback`) 
2. **订单创建** (`Order/CreateOrder`) ↔ **订单删除** (`Order/CreateOrderCom`)

### 测试补偿机制

```bash
# 创建订单后立即检查库存变化
curl -X GET "http://localhost:8051/v1/goods/1/stocks"

# 如果订单创建失败，验证库存是否已回滚
# 检查DTM事务状态 (需要DTM管理界面或日志)
```

### 监控事务状态

```bash
# 检查DTM服务状态
curl -X GET "http://localhost:36789/health"

# 查看服务日志确认事务执行情况
docker logs dtm -f
```

---

## 📊 性能测试

### 并发测试示例
```bash
# 使用ab工具进行并发测试
ab -n 100 -c 10 -H "Authorization: Bearer YOUR_TOKEN" \
   -T "application/json" \
   -p order_data.json \
   http://localhost:8051/v1/orders
```

其中 `order_data.json`:
```json
{
  "address": "性能测试地址",
  "name": "测试用户",
  "mobile": "13800138000",
  "post": "100000"
}
```

---

**文档更新**: 请根据实际API响应格式调整示例中的JSON结构。  
**联系支持**: 如遇问题请查看微服务日志或联系开发团队。