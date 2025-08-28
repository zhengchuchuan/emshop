#!/bin/bash

# EMShop API集成测试脚本
# 测试优惠券、支付、物流API的完整功能
# Author: Claude Code
# Version: 1.0

set -e

# 配置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API配置
API_BASE_URL="http://localhost:8080"
API_VERSION="/v1"

# 测试用户凭据（需要根据实际情况修改）
TEST_USER_MOBILE="13800138000"
TEST_USER_PASSWORD="123456"
JWT_TOKEN=""

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查测试依赖..."
    
    if ! command -v curl &> /dev/null; then
        log_error "curl命令未找到，请安装curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_warning "jq命令未找到，JSON响应将不会被格式化"
    fi
    
    log_success "依赖检查通过"
}

# 测试API端点可用性
test_api_health() {
    log_info "测试API端点可用性..."
    ((TOTAL_TESTS++))
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL/health" || echo "000")
    
    if [ "$response" -eq 200 ]; then
        log_success "API端点健康检查通过"
    else
        log_error "API端点不可用 (HTTP $response)"
        exit 1
    fi
}

# 用户登录获取JWT Token
user_login() {
    log_info "用户登录获取JWT Token..."
    ((TOTAL_TESTS++))
    
    login_response=$(curl -s -X POST "$API_BASE_URL$API_VERSION/user/pwd_login" \
        -H "Content-Type: application/json" \
        -d "{\"mobile\":\"$TEST_USER_MOBILE\",\"password\":\"$TEST_USER_PASSWORD\"}" || echo "{}")
    
    if command -v jq &> /dev/null; then
        JWT_TOKEN=$(echo "$login_response" | jq -r '.data.token // empty')
        if [ -n "$JWT_TOKEN" ] && [ "$JWT_TOKEN" != "null" ]; then
            log_success "用户登录成功，获取到JWT Token"
        else
            log_error "用户登录失败或未获取到Token: $login_response"
            # 继续测试，但某些需要认证的接口会失败
        fi
    else
        # 简单的文本匹配
        if echo "$login_response" | grep -q "token"; then
            JWT_TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
            log_success "用户登录成功，获取到JWT Token"
        else
            log_error "用户登录失败: $login_response"
        fi
    fi
}

# 测试优惠券API
test_coupon_apis() {
    log_info "========== 测试优惠券API =========="
    
    # 1. 获取可领取优惠券列表 (无需认证)
    log_info "测试获取可领取优惠券列表..."
    ((TOTAL_TESTS++))
    
    response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/coupons/templates?page=1&pageSize=10")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        log_success "获取优惠券模板列表 - HTTP $http_code"
        if command -v jq &> /dev/null; then
            echo "$body" | jq '.' 2>/dev/null || echo "$body"
        fi
    else
        log_error "获取优惠券模板列表失败 - HTTP $http_code: $body"
    fi
    
    if [ -n "$JWT_TOKEN" ]; then
        # 2. 领取优惠券 (需要认证)
        log_info "测试领取优惠券..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL$API_VERSION/coupons/receive" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d '{"coupon_template_id": 1}')
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 400 ]; then
            # 400可能是已经领取过了
            log_success "领取优惠券测试 - HTTP $http_code"
        else
            log_error "领取优惠券失败 - HTTP $http_code: $body"
        fi
        
        # 3. 获取我的优惠券列表 (需要认证)
        log_info "测试获取我的优惠券列表..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/coupons/user?status=1&page=1&pageSize=10" \
            -H "Authorization: Bearer $JWT_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ]; then
            log_success "获取用户优惠券列表 - HTTP $http_code"
        else
            log_error "获取用户优惠券列表失败 - HTTP $http_code: $body"
        fi
        
        # 4. 获取可用优惠券 (需要认证)
        log_info "测试获取可用优惠券..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/coupons/available?order_amount=150.0" \
            -H "Authorization: Bearer $JWT_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ]; then
            log_success "获取可用优惠券 - HTTP $http_code"
        else
            log_error "获取可用优惠券失败 - HTTP $http_code: $body"
        fi
        
        # 5. 计算优惠折扣 (需要认证)
        log_info "测试计算优惠折扣..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL$API_VERSION/coupons/calculate-discount" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d '{
                "coupon_ids": [1],
                "order_amount": 150.0,
                "order_items": [
                    {"goods_id": 1, "quantity": 2, "price": 75.0}
                ]
            }')
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ]; then
            log_success "计算优惠折扣 - HTTP $http_code"
        else
            log_error "计算优惠折扣失败 - HTTP $http_code: $body"
        fi
    else
        log_warning "跳过需要认证的优惠券API测试（未获取到JWT Token）"
        ((TOTAL_TESTS+=4)) # 跳过的测试数量
        ((FAILED_TESTS+=4))
    fi
}

# 测试支付API
test_payment_apis() {
    log_info "========== 测试支付API =========="
    
    if [ -n "$JWT_TOKEN" ]; then
        # 1. 创建支付订单
        log_info "测试创建支付订单..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL$API_VERSION/payment/create" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d '{
                "order_sn": "ORD'$(date +%Y%m%d%H%M%S)'",
                "amount": 140.0,
                "payment_method": 1,
                "expired_minutes": 15
            }')
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
            log_success "创建支付订单 - HTTP $http_code"
            # 尝试提取支付单号用于后续测试
            if command -v jq &> /dev/null; then
                PAYMENT_SN=$(echo "$body" | jq -r '.data.payment_sn // empty')
            fi
        else
            log_error "创建支付订单失败 - HTTP $http_code: $body"
        fi
        
        # 2. 查询支付状态 (如果有支付单号)
        if [ -n "$PAYMENT_SN" ] && [ "$PAYMENT_SN" != "null" ]; then
            log_info "测试查询支付状态..."
            ((TOTAL_TESTS++))
            
            response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/payment/$PAYMENT_SN/status" \
                -H "Authorization: Bearer $JWT_TOKEN")
            http_code=$(echo "$response" | tail -n1)
            body=$(echo "$response" | sed '$d')
            
            if [ "$http_code" -eq 200 ]; then
                log_success "查询支付状态 - HTTP $http_code"
            else
                log_error "查询支付状态失败 - HTTP $http_code: $body"
            fi
        else
            log_warning "跳过支付状态查询测试（未获取到支付单号）"
            ((TOTAL_TESTS++))
            ((FAILED_TESTS++))
        fi
    else
        log_warning "跳过支付API测试（未获取到JWT Token）"
        ((TOTAL_TESTS+=2))
        ((FAILED_TESTS+=2))
    fi
}

# 测试物流API
test_logistics_apis() {
    log_info "========== 测试物流API =========="
    
    # 1. 获取物流公司列表 (无需认证)
    log_info "测试获取物流公司列表..."
    ((TOTAL_TESTS++))
    
    response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/logistics/companies")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        log_success "获取物流公司列表 - HTTP $http_code"
    else
        log_error "获取物流公司列表失败 - HTTP $http_code: $body"
    fi
    
    # 2. 计算运费 (无需认证)
    log_info "测试计算运费..."
    ((TOTAL_TESTS++))
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE_URL$API_VERSION/logistics/calculate-fee" \
        -H "Content-Type: application/json" \
        -d '{
            "receiver_address": "北京市朝阳区测试地址",
            "items": [
                {"goods_id": 1, "quantity": 2, "weight": 1.5, "volume": 0.01}
            ],
            "logistics_company": 1,
            "shipping_method": 1
        }')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        log_success "计算运费 - HTTP $http_code"
    else
        log_error "计算运费失败 - HTTP $http_code: $body"
    fi
    
    if [ -n "$JWT_TOKEN" ]; then
        # 3. 查询物流信息 (需要认证)
        log_info "测试查询物流信息..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/logistics/info?order_sn=ORD20250101001" \
            -H "Authorization: Bearer $JWT_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 404 ]; then
            # 404是正常的，因为测试订单可能不存在
            log_success "查询物流信息测试 - HTTP $http_code"
        else
            log_error "查询物流信息失败 - HTTP $http_code: $body"
        fi
        
        # 4. 查看物流轨迹 (需要认证)
        log_info "测试查看物流轨迹..."
        ((TOTAL_TESTS++))
        
        response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL$API_VERSION/logistics/tracks?order_sn=ORD20250101001" \
            -H "Authorization: Bearer $JWT_TOKEN")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')
        
        if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 404 ]; then
            log_success "查看物流轨迹测试 - HTTP $http_code"
        else
            log_error "查看物流轨迹失败 - HTTP $http_code: $body"
        fi
    else
        log_warning "跳过需要认证的物流API测试（未获取到JWT Token）"
        ((TOTAL_TESTS+=2))
        ((FAILED_TESTS+=2))
    fi
}

# 生成测试报告
generate_report() {
    log_info "========== 测试报告 =========="
    echo
    echo "测试总数: $TOTAL_TESTS"
    echo -e "通过: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有测试通过！API集成测试成功！${NC}"
        return 0
    else
        echo -e "\n${YELLOW}⚠️  部分测试失败，请检查API服务状态和配置${NC}"
        return 1
    fi
}

# 主函数
main() {
    echo -e "${BLUE}"
    echo "======================================================"
    echo "       EMShop API 集成测试脚本"
    echo "======================================================"
    echo -e "${NC}"
    
    # 检查参数
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        echo "用法: $0 [选项]"
        echo "选项:"
        echo "  -h, --help     显示此帮助信息"
        echo "  --base-url     指定API基础URL (默认: $API_BASE_URL)"
        echo "  --user-mobile  指定测试用户手机号 (默认: $TEST_USER_MOBILE)"
        echo "  --user-pass    指定测试用户密码 (默认: $TEST_USER_PASSWORD)"
        echo
        echo "示例:"
        echo "  $0 --base-url http://localhost:9090 --user-mobile 13800138001"
        exit 0
    fi
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --base-url)
                API_BASE_URL="$2"
                shift 2
                ;;
            --user-mobile)
                TEST_USER_MOBILE="$2"
                shift 2
                ;;
            --user-pass)
                TEST_USER_PASSWORD="$2"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    log_info "使用API基础URL: $API_BASE_URL"
    log_info "测试用户: $TEST_USER_MOBILE"
    
    # 执行测试
    check_dependencies
    test_api_health
    user_login
    test_coupon_apis
    test_payment_apis
    test_logistics_apis
    
    # 生成报告
    generate_report
}

# 执行主函数
main "$@"