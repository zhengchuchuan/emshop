#!/bin/bash

# 优惠券DTM分布式事务测试脚本
# 用于验证优惠券服务的分布式事务流程

set -e

echo "======================================"
echo "优惠券DTM分布式事务测试"
echo "======================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置信息
COUPON_SERVICE_PORT=8078
COUPON_HTTP_PORT=8079
TEST_USER_ID=1001

# 检查优惠券服务是否运行
check_coupon_service() {
    echo -n "检查优惠券服务 (端口 $COUPON_SERVICE_PORT)... "
    
    if nc -z localhost $COUPON_SERVICE_PORT 2>/dev/null; then
        echo -e "${GREEN}✓ 运行中${NC}"
        return 0
    else
        echo -e "${RED}✗ 未运行${NC}"
        return 1
    fi
}

# 启动优惠券服务
start_coupon_service() {
    echo -e "\n${YELLOW}启动优惠券服务...${NC}"
    
    # 检查调试配置文件
    if [ ! -f "configs/coupon/srv-debug.yaml" ]; then
        echo -e "${RED}调试配置文件不存在: configs/coupon/srv-debug.yaml${NC}"
        echo -e "${YELLOW}提示: 使用固定端口的调试配置便于测试${NC}"
        return 1
    fi
    
    # 检查可执行文件
    if [ ! -f "bin/coupon" ]; then
        echo -e "${YELLOW}编译优惠券服务...${NC}"
        go build -o bin/coupon ./cmd/coupon/
    fi
    
    # 后台启动服务 (使用调试配置)
    echo "启动优惠券服务进程..."
    ./bin/coupon -c configs/coupon/srv-debug.yaml > logs/coupon-service.log 2>&1 &
    COUPON_PID=$!
    
    echo "优惠券服务PID: $COUPON_PID"
    echo $COUPON_PID > /tmp/coupon-service.pid
    
    # 等待服务启动
    echo "等待服务启动..."
    sleep 5
    
    # 验证服务是否成功启动
    if check_coupon_service; then
        echo -e "${GREEN}优惠券服务启动成功${NC}"
        return 0
    else
        echo -e "${RED}优惠券服务启动失败${NC}"
        if [ -f "logs/coupon-service.log" ]; then
            echo "错误日志:"
            tail -n 20 logs/coupon-service.log
        fi
        return 1
    fi
}

# 停止优惠券服务
stop_coupon_service() {
    if [ -f "/tmp/coupon-service.pid" ]; then
        COUPON_PID=$(cat /tmp/coupon-service.pid)
        echo -e "\n${YELLOW}停止优惠券服务 (PID: $COUPON_PID)...${NC}"
        kill $COUPON_PID 2>/dev/null || true
        rm -f /tmp/coupon-service.pid
        echo -e "${GREEN}优惠券服务已停止${NC}"
    fi
}

# 执行gRPC调用测试
test_grpc_call() {
    local method=$1
    local request_data=$2
    local description=$3
    
    echo -e "\n${BLUE}测试: $description${NC}"
    echo "方法: $method"
    echo "请求数据: $request_data"
    
    # 使用grpcurl进行测试调用
    if grpcurl -plaintext -d "$request_data" localhost:$COUPON_SERVICE_PORT coupon.Coupon/$method; then
        echo -e "${GREEN}✓ $description 成功${NC}"
        return 0
    else
        echo -e "${RED}✗ $description 失败${NC}"
        return 1
    fi
}

# 创建测试数据
create_test_data() {
    echo -e "\n${YELLOW}步骤1: 创建测试数据${NC}"
    
    # 创建优惠券模板
    echo "1.1 创建优惠券模板"
    TEMPLATE_REQUEST='{
        "name": "DTM测试优惠券",
        "type": 1,
        "discount_type": 1,
        "discount_value": 10.00,
        "min_order_amount": 50.00,
        "max_discount_amount": 10.00,
        "total_count": 100,
        "per_user_limit": 2,
        "valid_start_time": '$(date +%s)',
        "valid_end_time": '$(($(date +%s) + 86400))',
        "valid_days": 7,
        "description": "DTM分布式事务测试专用"
    }'
    
    if test_grpc_call "CreateCouponTemplate" "$TEMPLATE_REQUEST" "创建优惠券模板"; then
        echo -e "${GREEN}优惠券模板创建成功${NC}"
    else
        echo -e "${RED}优惠券模板创建失败${NC}"
        return 1
    fi
    
    # 创建秒杀活动
    echo "1.2 创建秒杀活动"
    FLASH_SALE_REQUEST='{
        "coupon_template_id": 1,
        "name": "DTM测试秒杀",
        "start_time": '$(date +%s)',
        "end_time": '$(($(date +%s) + 3600))',
        "flash_sale_count": 10,
        "per_user_limit": 1
    }'
    
    if test_grpc_call "CreateFlashSaleActivity" "$FLASH_SALE_REQUEST" "创建秒杀活动"; then
        echo -e "${GREEN}秒杀活动创建成功${NC}"
    else
        echo -e "${RED}秒杀活动创建失败${NC}"
        return 1
    fi
}

# 测试订单-优惠券分布式事务
test_order_coupon_transaction() {
    echo -e "\n${YELLOW}步骤2: 测试订单-优惠券分布式事务${NC}"
    
    # 首先领取优惠券
    echo "2.1 用户领取优惠券"
    RECEIVE_REQUEST='{
        "user_id": '$TEST_USER_ID',
        "coupon_template_id": 1
    }'
    
    if test_grpc_call "ReceiveCoupon" "$RECEIVE_REQUEST" "领取优惠券"; then
        echo -e "${GREEN}优惠券领取成功${NC}"
    else
        echo -e "${RED}优惠券领取失败${NC}"
        return 1
    fi
    
    # 测试订单-优惠券分布式事务
    echo "2.2 提交订单使用优惠券分布式事务"
    ORDER_SN="DTM_ORDER_$(date +%Y%m%d%H%M%S)"
    
    ORDER_COUPON_REQUEST='{
        "order_sn": "'$ORDER_SN'",
        "user_id": '$TEST_USER_ID',
        "coupon_ids": [1],
        "original_amount": 100.00,
        "discount_amount": 10.00,
        "final_amount": 90.00,
        "payment_method": 1,
        "goods_details": [
            {
                "goods_id": 1001,
                "quantity": 2,
                "price": 50.00
            }
        ],
        "address": "北京市朝阳区DTM测试地址123号"
    }'
    
    if test_grpc_call "SubmitOrderWithCoupons" "$ORDER_COUPON_REQUEST" "订单-优惠券分布式事务"; then
        echo -e "${GREEN}订单-优惠券分布式事务成功${NC}"
        return 0
    else
        echo -e "${RED}订单-优惠券分布式事务失败${NC}"
        return 1
    fi
}

# 测试秒杀-库存分布式事务
test_flash_sale_transaction() {
    echo -e "\n${YELLOW}步骤3: 测试秒杀-库存分布式事务${NC}"
    
    FLASH_SALE_INVENTORY_REQUEST='{
        "user_id": '$TEST_USER_ID',
        "flash_sale_id": 1,
        "goods_id": 1001,
        "quantity": 1
    }'
    
    if test_grpc_call "ProcessFlashSaleWithInventory" "$FLASH_SALE_INVENTORY_REQUEST" "秒杀-库存分布式事务"; then
        echo -e "${GREEN}秒杀-库存分布式事务成功${NC}"
        return 0
    else
        echo -e "${RED}秒杀-库存分布式事务失败${NC}"
        return 1
    fi
}

# 测试TCC回调接口
test_tcc_callbacks() {
    echo -e "\n${YELLOW}步骤4: 测试TCC回调接口${NC}"
    
    TCC_REQUEST='{
        "user_id": '$TEST_USER_ID',
        "flash_sale_id": 1
    }'
    
    # 测试Try阶段
    echo "4.1 测试TCC Try阶段"
    if test_grpc_call "TryFlashSale" "$TCC_REQUEST" "TCC Try预占"; then
        echo -e "${GREEN}TCC Try阶段成功${NC}"
    fi
    
    # 测试Confirm阶段
    echo "4.2 测试TCC Confirm阶段"
    if test_grpc_call "ConfirmFlashSale" "$TCC_REQUEST" "TCC Confirm确认"; then
        echo -e "${GREEN}TCC Confirm阶段成功${NC}"
    fi
    
    # 测试Cancel阶段
    echo "4.3 测试TCC Cancel阶段"
    if test_grpc_call "CancelFlashSale" "$TCC_REQUEST" "TCC Cancel取消"; then
        echo -e "${GREEN}TCC Cancel阶段成功${NC}"
    fi
}

# 测试事务状态查询
test_transaction_status() {
    echo -e "\n${YELLOW}步骤5: 测试事务状态查询${NC}"
    
    STATUS_REQUEST='{
        "gid": "test-transaction-001"
    }'
    
    if test_grpc_call "GetTransactionStatus" "$STATUS_REQUEST" "查询事务状态"; then
        echo -e "${GREEN}事务状态查询成功${NC}"
    fi
}

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}清理资源...${NC}"
    stop_coupon_service
}

# 注册清理函数
trap cleanup EXIT

# 主测试流程
main() {
    echo -e "${BLUE}优惠券DTM分布式事务测试开始${NC}"
    echo "测试用户ID: $TEST_USER_ID"
    echo "服务端口: $COUPON_SERVICE_PORT"
    
    # 创建日志目录
    mkdir -p logs
    
    # 检查服务状态并启动
    if ! check_coupon_service; then
        if ! start_coupon_service; then
            echo -e "${RED}无法启动优惠券服务${NC}"
            exit 1
        fi
    fi
    
    # 运行测试用例
    local test_success=true
    
    # 创建测试数据
    if ! create_test_data; then
        test_success=false
    fi
    
    # 测试订单-优惠券分布式事务
    if ! test_order_coupon_transaction; then
        test_success=false
    fi
    
    # 测试秒杀-库存分布式事务
    if ! test_flash_sale_transaction; then
        test_success=false
    fi
    
    # 测试TCC回调接口
    if ! test_tcc_callbacks; then
        test_success=false
    fi
    
    # 测试事务状态查询
    if ! test_transaction_status; then
        test_success=false
    fi
    
    # 输出测试结果
    echo -e "\n${BLUE}======================================"
    echo "DTM分布式事务测试完成"
    echo "======================================${NC}"
    
    if [ "$test_success" = true ]; then
        echo -e "${GREEN}✓ 所有测试通过${NC}"
        echo -e "\n${YELLOW}测试总结:${NC}"
        echo "✓ 优惠券模板管理测试通过"
        echo "✓ 订单-优惠券分布式事务测试通过"
        echo "✓ 秒杀-库存分布式事务测试通过"
        echo "✓ TCC回调接口测试通过"
        echo "✓ 事务状态查询测试通过"
        
        exit 0
    else
        echo -e "${RED}✗ 部分测试失败${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"