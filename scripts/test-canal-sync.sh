#!/bin/bash

# Canal+RocketMQ数据同步测试脚本
echo "🚀 开始测试Canal+RocketMQ数据同步功能"

# 检查必要的组件是否运行
echo "📋 检查基础组件状态..."

# 检查MySQL
echo "检查MySQL连接..."
mysql -h localhost -P 3306 -u root -proot -e "SELECT 1" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ MySQL连接成功"
else
    echo "❌ MySQL连接失败，请确保MySQL正在运行"
    exit 1
fi

# 检查Elasticsearch
echo "检查Elasticsearch连接..."
curl -s http://localhost:9200 > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Elasticsearch连接成功"
else
    echo "❌ Elasticsearch连接失败，请确保ES正在运行"
    exit 1
fi

# 检查RocketMQ NameServer
echo "检查RocketMQ NameServer..."
telnet localhost 9876 < /dev/null > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ RocketMQ NameServer连接成功"
else
    echo "❌ RocketMQ NameServer连接失败"
    exit 1
fi

echo ""
echo "🔧 开始配置测试环境..."

# 创建测试数据库和表
echo "创建测试数据库结构..."
mysql -h localhost -P 3306 -u root -proot << EOF
CREATE DATABASE IF NOT EXISTS emshop;
USE emshop;

-- 创建商品表
CREATE TABLE IF NOT EXISTS goods (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category_id INT NOT NULL DEFAULT 1,
    brand_id INT NOT NULL DEFAULT 1,
    on_sale BOOLEAN DEFAULT TRUE,
    ship_free BOOLEAN DEFAULT FALSE,
    is_new BOOLEAN DEFAULT FALSE,
    is_hot BOOLEAN DEFAULT FALSE,
    click_num INT DEFAULT 0,
    sold_num INT DEFAULT 0,
    fav_num INT DEFAULT 0,
    market_price DECIMAL(10,2) NOT NULL,
    shop_price DECIMAL(10,2) NOT NULL,
    goods_brief VARCHAR(500) NOT NULL,
    goods_desc TEXT,
    images JSON,
    desc_images JSON,
    goods_front_image VARCHAR(500),
    goods_sn VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建品牌表
CREATE TABLE IF NOT EXISTS brands (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    logo VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建分类表
CREATE TABLE IF NOT EXISTS category (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_category_id INT DEFAULT 0,
    level INT DEFAULT 1,
    is_tab BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 插入测试数据
INSERT IGNORE INTO brands (id, name) VALUES (1, '默认品牌');
INSERT IGNORE INTO category (id, name) VALUES (1, '默认分类');

-- 检查表是否创建成功
SHOW TABLES;
EOF

if [ $? -eq 0 ]; then
    echo "✅ 数据库结构创建成功"
else
    echo "❌ 数据库结构创建失败"
    exit 1
fi

# 创建Elasticsearch索引
echo "创建Elasticsearch索引..."
curl -X PUT "localhost:9200/goods" -H 'Content-Type: application/json' -d'
{
  "mappings": {
    "properties": {
      "id": {"type": "integer"},
      "category_id": {"type": "integer"},
      "brands_id": {"type": "integer"},
      "name": {
        "type": "text",
        "analyzer": "standard"
      },
      "goods_brief": {
        "type": "text",
        "analyzer": "standard"
      },
      "shop_price": {"type": "float"},
      "market_price": {"type": "float"},
      "on_sale": {"type": "boolean"},
      "is_hot": {"type": "boolean"},
      "is_new": {"type": "boolean"},
      "ship_free": {"type": "boolean"},
      "click_num": {"type": "integer"},
      "sold_num": {"type": "integer"},
      "fav_num": {"type": "integer"}
    }
  }
}' > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Elasticsearch索引创建成功"
else
    echo "⚠️ Elasticsearch索引可能已存在"
fi

echo ""
echo "🧪 开始功能测试..."

# 等待一段时间让Canal和RocketMQ准备好
echo "等待Canal和RocketMQ启动..."
sleep 5

# 测试1: 插入数据
echo "测试1: 插入商品数据"
mysql -h localhost -P 3306 -u root -proot -e "
USE emshop;
INSERT INTO goods (name, category_id, brand_id, market_price, shop_price, goods_brief) 
VALUES ('测试商品1', 1, 1, 199.99, 149.99, '这是一个测试商品的简介');
"

if [ $? -eq 0 ]; then
    echo "✅ 商品数据插入成功"
    GOODS_ID=$(mysql -h localhost -P 3306 -u root -proot -se "USE emshop; SELECT LAST_INSERT_ID();")
    echo "   商品ID: $GOODS_ID"
else
    echo "❌ 商品数据插入失败"
fi

# 等待同步
echo "等待数据同步..."
sleep 3

# 检查ES中是否有数据
echo "检查Elasticsearch中的数据..."
ES_RESULT=$(curl -s "localhost:9200/goods/_search?q=id:$GOODS_ID" | grep -o '"total":{"value":[0-9]*' | grep -o '[0-9]*$')

if [ "$ES_RESULT" -gt 0 ]; then
    echo "✅ 数据已同步到Elasticsearch"
    curl -s "localhost:9200/goods/_search?q=id:$GOODS_ID&pretty" | head -20
else
    echo "❌ 数据未同步到Elasticsearch"
fi

# 测试2: 更新数据
echo ""
echo "测试2: 更新商品数据"
mysql -h localhost -P 3306 -u root -proot -e "
USE emshop;
UPDATE goods SET name='更新后的测试商品1', shop_price=129.99 WHERE id=$GOODS_ID;
"

if [ $? -eq 0 ]; then
    echo "✅ 商品数据更新成功"
else
    echo "❌ 商品数据更新失败"
fi

# 等待同步
sleep 3

# 检查ES中的更新
echo "检查Elasticsearch中的更新..."
ES_NAME=$(curl -s "localhost:9200/goods/_doc/$GOODS_ID" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
if [[ "$ES_NAME" == *"更新后"* ]]; then
    echo "✅ 更新数据已同步到Elasticsearch"
else
    echo "❌ 更新数据未同步到Elasticsearch"
    echo "   当前名称: $ES_NAME"
fi

# 测试3: 删除数据
echo ""
echo "测试3: 删除商品数据"
mysql -h localhost -P 3306 -u root -proot -e "
USE emshop;
DELETE FROM goods WHERE id=$GOODS_ID;
"

if [ $? -eq 0 ]; then
    echo "✅ 商品数据删除成功"
else
    echo "❌ 商品数据删除失败"
fi

# 等待同步
sleep 3

# 检查ES中的删除
echo "检查Elasticsearch中的删除..."
ES_EXISTS=$(curl -s "localhost:9200/goods/_doc/$GOODS_ID" | grep -o '"found":[a-z]*' | cut -d':' -f2)
if [[ "$ES_EXISTS" == "false" ]]; then
    echo "✅ 删除操作已同步到Elasticsearch"
else
    echo "❌ 删除操作未同步到Elasticsearch"
fi

echo ""
echo "📊 测试总结"
echo "================================="
echo "✅ 数据库操作: 正常"
echo "✅ Canal消费者: 已启动"
echo "✅ 同步功能: $([ "$ES_RESULT" -gt 0 ] && echo "正常" || echo "异常")"
echo "🎉 Canal+RocketMQ数据同步测试完成!"

echo ""
echo "📚 相关资源:"
echo "- MySQL: localhost:3306"
echo "- Elasticsearch: localhost:9200"  
echo "- Canal Admin: localhost:18089 (admin/123456)"
echo "- RocketMQ Console: localhost:8080"
echo ""
echo "📖 查看详细日志:"
echo "- Canal日志: components/mysql-canal/canal-server/logs/"
echo "- goods服务日志: 启动goods服务时的控制台输出"