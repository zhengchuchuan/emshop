#!/bin/bash

# Canal+RocketMQ同步功能测试脚本
echo "🧪 开始Canal+RocketMQ同步功能测试..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试结果统计
TEST_COUNT=0
PASS_COUNT=0
FAIL_COUNT=0

# 工具函数
log_test() {
    echo -e "${BLUE}📋 测试 $((++TEST_COUNT)): $1${NC}"
}

log_pass() {
    echo -e "   ${GREEN}✅ $1${NC}"
    ((PASS_COUNT++))
}

log_fail() {
    echo -e "   ${RED}❌ $1${NC}"
    ((FAIL_COUNT++))
}

log_info() {
    echo -e "   ${YELLOW}ℹ️ $1${NC}"
}

wait_for_sync() {
    echo -n "   ⏳ 等待同步"
    for i in {1..5}; do
        sleep 1
        echo -n "."
    done
    echo ""
}

# 测试开始
echo "🚀 开始功能测试..."
echo "================================="

# 测试1: 准备测试数据库结构
log_test "准备测试数据库结构"

mysql -h localhost -P 3306 -u root -proot << 'EOF' 2>/dev/null
CREATE DATABASE IF NOT EXISTS emshop;
USE emshop;

-- 创建测试表结构
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

-- 创建基础测试数据
INSERT IGNORE INTO goods (id, name, category_id, brand_id, market_price, shop_price, goods_brief) VALUES
(999, 'Canal测试清理商品', 1, 1, 99.99, 89.99, '用于清理的测试商品');

-- 清理可能存在的测试数据
DELETE FROM goods WHERE name LIKE 'Canal测试%' AND id != 999;
EOF

if [ $? -eq 0 ]; then
    log_pass "数据库结构准备完成"
else
    log_fail "数据库结构准备失败"
fi

# 测试2: 创建Elasticsearch索引
log_test "创建Elasticsearch索引"

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
}' >/dev/null 2>&1

if [ $? -eq 0 ]; then
    log_pass "Elasticsearch索引创建成功"
else
    log_info "Elasticsearch索引可能已存在"
fi

# 测试3: 验证Canal服务状态
log_test "验证Canal服务状态"

CANAL_STATUS=$(curl -s http://localhost:18089 2>/dev/null | grep -o "title" | wc -l)
if [ "$CANAL_STATUS" -gt 0 ]; then
    log_pass "Canal Admin服务正常"
else
    log_fail "Canal Admin服务异常"
fi

# 测试4: 检查goods服务是否编译成功
log_test "检查goods服务编译状态"

if [ -f "./bin/goods" ]; then
    log_pass "goods服务已编译"
else
    log_info "正在编译goods服务..."
    go build -o ./bin/goods ./cmd/goods/
    if [ $? -eq 0 ]; then
        log_pass "goods服务编译成功"
    else
        log_fail "goods服务编译失败"
    fi
fi

# 测试5: 启动goods服务(后台运行)
log_test "启动goods服务"

# 检查是否已有goods服务在运行
GOODS_PID=$(pgrep -f "./bin/goods")
if [ -n "$GOODS_PID" ]; then
    log_info "goods服务已在运行 (PID: $GOODS_PID)"
else
    log_info "启动goods服务(后台运行)..."
    nohup ./bin/goods --config=configs/goods-canal.yaml > /tmp/goods-test.log 2>&1 &
    sleep 3
    
    GOODS_PID=$(pgrep -f "./bin/goods")
    if [ -n "$GOODS_PID" ]; then
        log_pass "goods服务启动成功 (PID: $GOODS_PID)"
        echo "   📄 日志文件: /tmp/goods-test.log"
    else
        log_fail "goods服务启动失败"
        echo "   📄 检查日志: /tmp/goods-test.log"
    fi
fi

# 等待服务完全启动
sleep 5

# 测试6: 数据插入同步测试
log_test "数据插入同步测试"

log_info "插入测试商品数据..."
INSERT_RESULT=$(mysql -h localhost -P 3306 -u root -proot -se "
USE emshop;
INSERT INTO goods (name, category_id, brand_id, market_price, shop_price, goods_brief) 
VALUES ('Canal测试商品INSERT', 1, 1, 299.99, 199.99, 'Canal插入同步测试商品');
SELECT LAST_INSERT_ID();
" 2>/dev/null)

if [ -n "$INSERT_RESULT" ]; then
    GOODS_ID=$INSERT_RESULT
    log_pass "商品插入成功 (ID: $GOODS_ID)"
    
    wait_for_sync
    
    # 检查ES中的数据
    log_info "检查Elasticsearch同步结果..."
    ES_COUNT=$(curl -s "localhost:9200/goods/_search?q=id:$GOODS_ID" 2>/dev/null | jq -r '.hits.total.value' 2>/dev/null)
    
    if [ "$ES_COUNT" = "1" ]; then
        log_pass "数据成功同步到Elasticsearch"
        
        # 验证数据内容
        ES_NAME=$(curl -s "localhost:9200/goods/_doc/$GOODS_ID" 2>/dev/null | jq -r '._source.name' 2>/dev/null)
        if [[ "$ES_NAME" == *"INSERT"* ]]; then
            log_pass "同步数据内容正确"
        else
            log_fail "同步数据内容不正确: $ES_NAME"
        fi
    else
        log_fail "数据未同步到Elasticsearch (ES count: $ES_COUNT)"
    fi
else
    log_fail "商品插入失败"
    GOODS_ID=""
fi

# 测试7: 数据更新同步测试
if [ -n "$GOODS_ID" ]; then
    log_test "数据更新同步测试"
    
    log_info "更新测试商品数据..."
    mysql -h localhost -P 3306 -u root -proot -e "
    USE emshop;
    UPDATE goods SET 
        name='Canal测试商品UPDATE', 
        shop_price=159.99,
        is_hot=TRUE
    WHERE id=$GOODS_ID;
    " 2>/dev/null
    
    if [ $? -eq 0 ]; then
        log_pass "商品更新成功"
        
        wait_for_sync
        
        # 检查ES中的更新
        log_info "检查Elasticsearch更新结果..."
        ES_NAME=$(curl -s "localhost:9200/goods/_doc/$GOODS_ID" 2>/dev/null | jq -r '._source.name' 2>/dev/null)
        ES_PRICE=$(curl -s "localhost:9200/goods/_doc/$GOODS_ID" 2>/dev/null | jq -r '._source.shop_price' 2>/dev/null)
        
        if [[ "$ES_NAME" == *"UPDATE"* ]] && [[ "$ES_PRICE" == "159.99" ]]; then
            log_pass "更新数据成功同步到Elasticsearch"
        else
            log_fail "更新数据未正确同步 (名称: $ES_NAME, 价格: $ES_PRICE)"
        fi
    else
        log_fail "商品更新失败"
    fi
    
    # 测试8: 数据删除同步测试
    log_test "数据删除同步测试"
    
    log_info "删除测试商品数据..."
    mysql -h localhost -P 3306 -u root -proot -e "
    USE emshop;
    DELETE FROM goods WHERE id=$GOODS_ID;
    " 2>/dev/null
    
    if [ $? -eq 0 ]; then
        log_pass "商品删除成功"
        
        wait_for_sync
        
        # 检查ES中的删除
        log_info "检查Elasticsearch删除结果..."
        ES_EXISTS=$(curl -s "localhost:9200/goods/_doc/$GOODS_ID" 2>/dev/null | jq -r '.found' 2>/dev/null)
        
        if [ "$ES_EXISTS" = "false" ]; then
            log_pass "删除操作成功同步到Elasticsearch"
        else
            log_fail "删除操作未同步到Elasticsearch (found: $ES_EXISTS)"
        fi
    else
        log_fail "商品删除失败"
    fi
fi

# 测试9: 性能测试 - 批量插入
log_test "性能测试 - 批量插入"

log_info "批量插入10条测试数据..."
START_TIME=$(date +%s)

for i in {1..10}; do
    mysql -h localhost -P 3306 -u root -proot -e "
    USE emshop;
    INSERT INTO goods (name, category_id, brand_id, market_price, shop_price, goods_brief) 
    VALUES ('Canal批量测试商品$i', 1, 1, $(($i * 10 + 99)).99, $(($i * 10 + 50)).99, '批量性能测试商品$i');
    " 2>/dev/null
done

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log_pass "批量插入完成，耗时: ${DURATION}秒"

wait_for_sync

# 检查批量同步结果
log_info "检查批量同步结果..."
BATCH_COUNT=$(curl -s "localhost:9200/goods/_search?q=name:Canal批量测试*" 2>/dev/null | jq -r '.hits.total.value' 2>/dev/null)

if [ "$BATCH_COUNT" -ge 8 ]; then
    log_pass "批量数据同步成功 ($BATCH_COUNT/10条)"
else
    log_fail "批量数据同步不完整 ($BATCH_COUNT/10条)"
fi

# 清理批量测试数据
mysql -h localhost -P 3306 -u root -proot -e "
USE emshop;
DELETE FROM goods WHERE name LIKE 'Canal批量测试%';
" 2>/dev/null

# 测试10: 检查goods服务日志
log_test "检查goods服务运行日志"

if [ -f "/tmp/goods-test.log" ]; then
    ERROR_COUNT=$(grep -i "error\|fatal\|panic" /tmp/goods-test.log | wc -l)
    CANAL_LOG_COUNT=$(grep -i "canal" /tmp/goods-test.log | wc -l)
    
    if [ "$ERROR_COUNT" -eq 0 ]; then
        log_pass "服务运行无错误"
    else
        log_fail "发现 $ERROR_COUNT 个错误"
    fi
    
    if [ "$CANAL_LOG_COUNT" -gt 0 ]; then
        log_pass "Canal消费者日志正常 ($CANAL_LOG_COUNT 条消息)"
    else
        log_fail "未发现Canal消费者日志"
    fi
    
    echo "   📄 最新日志:"
    tail -n 5 /tmp/goods-test.log | sed 's/^/   /'
fi

# 停止测试服务
log_info "停止测试服务..."
if [ -n "$GOODS_PID" ]; then
    kill $GOODS_PID 2>/dev/null
    sleep 2
    log_info "goods服务已停止"
fi

echo ""
echo "📊 测试结果总结"
echo "================================="
echo -e "总测试项目: ${BLUE}$TEST_COUNT${NC}"
echo -e "通过测试: ${GREEN}$PASS_COUNT${NC}"
echo -e "失败测试: ${RED}$FAIL_COUNT${NC}"

SUCCESS_RATE=$(( PASS_COUNT * 100 / TEST_COUNT ))
echo -e "成功率: ${BLUE}$SUCCESS_RATE%${NC}"

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！Canal+RocketMQ同步方案运行正常${NC}"
    echo ""
    echo "✅ 验证结果:"
    echo "• 数据插入同步: 正常"
    echo "• 数据更新同步: 正常" 
    echo "• 数据删除同步: 正常"
    echo "• 批量操作同步: 正常"
    echo "• 服务稳定性: 正常"
    exit 0
elif [ $SUCCESS_RATE -ge 80 ]; then
    echo -e "${YELLOW}⚠️ 大部分测试通过，但需要关注失败项${NC}"
    exit 1
else
    echo -e "${RED}❌ 测试失败较多，需要排查问题${NC}"
    echo ""
    echo "🔧 建议检查:"
    echo "• Canal服务配置"
    echo "• RocketMQ连接状态"
    echo "• goods服务日志: /tmp/goods-test.log"
    echo "• MySQL binlog配置"
    exit 2
fi