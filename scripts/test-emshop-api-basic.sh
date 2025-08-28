#!/bin/bash

# EMShop API基础功能测试脚本
# 测试API服务的基础编译和路由配置
# Author: Claude Code
# Version: 1.0

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试编译状态
test_compilation() {
    log_info "测试EMShop API编译状态..."
    
    cd /home/zcc/project/golang/emshop/emshop
    
    if go build ./internal/app/api/emshop/... > /dev/null 2>&1; then
        log_success "✅ API服务编译成功"
    else
        log_error "❌ API服务编译失败"
        go build ./internal/app/api/emshop/...
        return 1
    fi
}

# 测试路由配置
test_router_config() {
    log_info "验证路由配置完整性..."
    
    # 检查路由文件是否存在新的API路由
    router_file="/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/router.go"
    
    if [ ! -f "$router_file" ]; then
        log_error "路由配置文件不存在: $router_file"
        return 1
    fi
    
    # 检查是否包含新增的API路由
    missing_routes=()
    
    if ! grep -q "coupons" "$router_file"; then
        missing_routes+=("优惠券API路由")
    fi
    
    if ! grep -q "payment" "$router_file"; then
        missing_routes+=("支付API路由")
    fi
    
    if ! grep -q "logistics" "$router_file"; then
        missing_routes+=("物流API路由")
    fi
    
    if [ ${#missing_routes[@]} -eq 0 ]; then
        log_success "✅ 路由配置完整"
    else
        log_error "❌ 缺少路由配置: ${missing_routes[*]}"
        return 1
    fi
}

# 测试Controller文件
test_controllers() {
    log_info "验证Controller文件完整性..."
    
    controllers=(
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/controller/coupon/v1/coupon.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/controller/payment/v1/payment.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/controller/logistics/v1/logistics.go"
    )
    
    missing_controllers=()
    
    for controller in "${controllers[@]}"; do
        if [ ! -f "$controller" ]; then
            missing_controllers+=("$(basename "$controller")")
        fi
    done
    
    if [ ${#missing_controllers[@]} -eq 0 ]; then
        log_success "✅ Controller文件完整"
    else
        log_error "❌ 缺少Controller文件: ${missing_controllers[*]}"
        return 1
    fi
}

# 测试Service文件
test_services() {
    log_info "验证Service文件完整性..."
    
    services=(
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/service/coupon/v1/coupon.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/service/payment/v1/payment.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/service/logistics/v1/logistics.go"
    )
    
    missing_services=()
    
    for service in "${services[@]}"; do
        if [ ! -f "$service" ]; then
            missing_services+=("$(basename "$service")")
        fi
    done
    
    if [ ${#missing_services[@]} -eq 0 ]; then
        log_success "✅ Service文件完整"
    else
        log_error "❌ 缺少Service文件: ${missing_services[*]}"
        return 1
    fi
}

# 测试Data层文件
test_data_layer() {
    log_info "验证Data层文件完整性..."
    
    data_files=(
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/data/rpc/coupon.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/data/rpc/payment.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/data/rpc/logistics.go"
    )
    
    missing_data_files=()
    
    for data_file in "${data_files[@]}"; do
        if [ ! -f "$data_file" ]; then
            missing_data_files+=("$(basename "$data_file")")
        fi
    done
    
    if [ ${#missing_data_files[@]} -eq 0 ]; then
        log_success "✅ Data层文件完整"
    else
        log_error "❌ 缺少Data层文件: ${missing_data_files[*]}"
        return 1
    fi
}

# 测试DTO文件
test_dto_files() {
    log_info "验证DTO文件完整性..."
    
    request_files=(
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/domain/dto/request/coupon.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/domain/dto/request/payment.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/domain/dto/request/logistics.go"
    )
    
    response_files=(
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/domain/dto/response/coupon.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/domain/dto/response/payment.go"
        "/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop/domain/dto/response/logistics.go"
    )
    
    missing_files=()
    
    for file in "${request_files[@]}" "${response_files[@]}"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$(basename "$file")")
        fi
    done
    
    if [ ${#missing_files[@]} -eq 0 ]; then
        log_success "✅ DTO文件完整"
    else
        log_error "❌ 缺少DTO文件: ${missing_files[*]}"
        return 1
    fi
}

# 统计代码行数
count_code_lines() {
    log_info "统计新增代码行数..."
    
    total_lines=0
    api_dir="/home/zcc/project/golang/emshop/emshop/internal/app/api/emshop"
    
    # 统计新增的三个服务的代码行数
    for service in coupon payment logistics; do
        if [ -d "$api_dir/controller/$service" ]; then
            lines=$(find "$api_dir/controller/$service" -name "*.go" -exec wc -l {} + | tail -1 | awk '{print $1}' || echo 0)
            total_lines=$((total_lines + lines))
        fi
        
        if [ -d "$api_dir/service/$service" ]; then
            lines=$(find "$api_dir/service/$service" -name "*.go" -exec wc -l {} + | tail -1 | awk '{print $1}' || echo 0)
            total_lines=$((total_lines + lines))
        fi
        
        if [ -f "$api_dir/data/rpc/$service.go" ]; then
            lines=$(wc -l < "$api_dir/data/rpc/$service.go" || echo 0)
            total_lines=$((total_lines + lines))
        fi
        
        if [ -f "$api_dir/domain/dto/request/$service.go" ]; then
            lines=$(wc -l < "$api_dir/domain/dto/request/$service.go" || echo 0)
            total_lines=$((total_lines + lines))
        fi
        
        if [ -f "$api_dir/domain/dto/response/$service.go" ]; then
            lines=$(wc -l < "$api_dir/domain/dto/response/$service.go" || echo 0)
            total_lines=$((total_lines + lines))
        fi
    done
    
    log_success "📊 新增代码总行数: $total_lines 行"
}

# 生成测试报告
generate_report() {
    log_info "========== 基础功能测试报告 =========="
    echo
    echo -e "${BLUE}测试项目:${NC}"
    echo "✓ 编译状态检查"
    echo "✓ 路由配置验证"
    echo "✓ Controller文件完整性"
    echo "✓ Service文件完整性"  
    echo "✓ Data层文件完整性"
    echo "✓ DTO文件完整性"
    echo "✓ 代码行数统计"
    echo
    count_code_lines
    echo
    log_success "🎉 基础功能测试完成！EMShop API新增服务架构完整！"
}

# 主函数
main() {
    echo -e "${BLUE}"
    echo "======================================================"
    echo "       EMShop API 基础功能测试"
    echo "======================================================"
    echo -e "${NC}"
    
    # 执行所有测试
    test_compilation || exit 1
    test_router_config || exit 1
    test_controllers || exit 1
    test_services || exit 1
    test_data_layer || exit 1
    test_dto_files || exit 1
    
    # 生成报告
    generate_report
    
    echo -e "\n${GREEN}✅ 所有基础测试通过！可以进行下一步的集成测试。${NC}"
    echo -e "${YELLOW}💡 运行完整集成测试: ./scripts/test-emshop-api-integration.sh${NC}"
}

# 执行主函数
main "$@"