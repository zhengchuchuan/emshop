#!/bin/bash

# Canal + RocketMQ + Elasticsearch 完整集成测试脚本
# 测试 MySQL → Canal → RocketMQ → Goods Service → Elasticsearch 数据流

set -e

echo "🧪 Canal集成端到端测试"
echo "========================================"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. 检查服务状态
log_info "检查依赖服务状态..."

check_service() {
    local service_name=$1
    local container_pattern=$2
    
    if docker ps | grep -q "$container_pattern"; then
        log_success "$service_name 服务运行正常"
        return 0
    else
        log_error "$service_name 服务未运行"
        return 1
    fi
}

# 检查所有必需服务
services_ok=true
check_service "MySQL" "emshop-mysql" || services_ok=false
check_service "RocketMQ NameServer" "rmqnamesrv" || services_ok=false
check_service "RocketMQ Broker" "rmqbroker" || services_ok=false
check_service "Canal Server" "emshop-canal-server" || services_ok=false

if [ "$services_ok" = false ]; then
    log_error "请先启动所有必需的服务"
    exit 1
fi

# 2. 启动RocketMQ消费者（后台运行）
log_info "启动Canal消息监控器..."
timeout 60 go run test-canal-consumer.go > /tmp/canal_integration_test.log 2>&1 &
CONSUMER_PID=$!
log_success "消费者已启动 (PID: $CONSUMER_PID)"

# 等待消费者完全启动
sleep 3

# 3. 执行数据库操作测试
log_info "开始数据库操作测试..."

# 测试1: 插入新商品
log_info "测试1: 插入新商品..."
INSERT_SQL="INSERT INTO goods (add_time, update_time, category_id, brand_id, on_sale, goods_sn, name, click_num, sold_num, fav_num, market_price, shop_price, goods_brief, ship_free, images, desc_images, goods_front_image, is_new, is_hot) VALUES (NOW(), NOW(), 136595, 656, 1, 'CANAL_INTEG_TEST_001', 'Canal集成测试商品-INSERT', 0, 0, 0, 999.99, 799.99, '集成测试商品，用于验证Canal+RocketMQ+ES同步', 1, '[]', '[]', 'integration-test.jpg', 1, 1);"

docker exec emshop-mysql mysql -u root -proot -e "USE emshop_goods_srv; $INSERT_SQL" 2>/dev/null
if [ $? -eq 0 ]; then
    log_success "插入测试商品成功"
else
    log_error "插入测试商品失败"
fi

sleep 2

# 测试2: 更新商品信息
log_info "测试2: 更新商品信息..."
UPDATE_SQL="UPDATE goods SET name='Canal集成测试商品-UPDATED', shop_price=699.99, goods_brief='更新后的商品描述，验证Canal更新同步', update_time=NOW() WHERE goods_sn='CANAL_INTEG_TEST_001';"

docker exec emshop-mysql mysql -u root -proot -e "USE emshop_goods_srv; $UPDATE_SQL" 2>/dev/null
if [ $? -eq 0 ]; then
    log_success "更新测试商品成功"
else
    log_error "更新测试商品失败"
fi

sleep 2

# 测试3: 批量操作
log_info "测试3: 批量插入测试..."
BATCH_SQL="INSERT INTO goods (add_time, update_time, category_id, brand_id, on_sale, goods_sn, name, click_num, sold_num, fav_num, market_price, shop_price, goods_brief, ship_free, images, desc_images, goods_front_image, is_new, is_hot) VALUES 
(NOW(), NOW(), 136595, 656, 1, 'CANAL_BATCH_001', 'Canal批量测试商品1', 10, 5, 2, 199.99, 159.99, '批量测试商品1', 1, '[]', '[]', 'batch1.jpg', 0, 1),
(NOW(), NOW(), 136595, 656, 1, 'CANAL_BATCH_002', 'Canal批量测试商品2', 20, 8, 3, 299.99, 239.99, '批量测试商品2', 1, '[]', '[]', 'batch2.jpg', 1, 0),
(NOW(), NOW(), 136595, 656, 1, 'CANAL_BATCH_003', 'Canal批量测试商品3', 15, 3, 1, 399.99, 319.99, '批量测试商品3', 0, '[]', '[]', 'batch3.jpg', 1, 1);"

docker exec emshop-mysql mysql -u root -proot -e "USE emshop_goods_srv; $BATCH_SQL" 2>/dev/null
if [ $? -eq 0 ]; then
    log_success "批量插入测试商品成功"
else
    log_error "批量插入测试商品失败"
fi

sleep 3

# 4. 等待消息处理
log_info "等待Canal消息处理（10秒）..."
sleep 10

# 5. 检查消费者日志
log_info "检查Canal消息消费情况..."
if [ -f "/tmp/canal_integration_test.log" ]; then
    message_count=$(grep -c "Received Canal Message" /tmp/canal_integration_test.log 2>/dev/null || echo "0")
    if [ "$message_count" -gt 0 ]; then
        log_success "检测到 $message_count 条Canal消息"
        echo ""
        echo "📄 消息内容预览:"
        echo "----------------------------------------"
        grep -A 10 "Received Canal Message" /tmp/canal_integration_test.log | head -20
        echo "----------------------------------------"
    else
        log_warning "未检测到Canal消息，可能存在配置问题"
        echo ""
        echo "📄 消费者日志:"
        echo "----------------------------------------"
        tail -20 /tmp/canal_integration_test.log
        echo "----------------------------------------"
    fi
else
    log_error "消费者日志文件不存在"
fi

# 6. 停止消费者
log_info "停止消费者进程..."
if kill $CONSUMER_PID 2>/dev/null; then
    log_success "消费者进程已停止"
else
    log_warning "消费者进程可能已经结束"
fi

# 7. 清理测试数据（可选）
read -p "🗑️  是否清理测试数据? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "清理测试数据..."
    CLEANUP_SQL="DELETE FROM goods WHERE goods_sn LIKE 'CANAL_%';"
    docker exec emshop-mysql mysql -u root -proot -e "USE emshop_goods_srv; $CLEANUP_SQL" 2>/dev/null
    log_success "测试数据已清理"
fi

echo ""
echo "========================================"
log_info "集成测试完成！"
echo ""
echo "📊 测试结果总结:"
echo "   - 数据库操作: 执行了INSERT/UPDATE/BATCH操作"
echo "   - Canal消息: 检测到 ${message_count:-0} 条消息"
echo "   - 完整日志: /tmp/canal_integration_test.log"
echo ""
echo "💡 下一步:"
echo "   1. 检查Elasticsearch中是否有同步数据"
echo "   2. 启动完整的Goods服务进行业务测试"
echo "   3. 配置生产环境的监控和告警"
echo ""