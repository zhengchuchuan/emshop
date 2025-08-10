#!/bin/bash

# EMShop 一键启动脚本
echo "=== EMShop 微服务一键启动脚本 ==="

# 检查是否在正确的目录
if [ ! -d "cmd" ] || [ ! -d "configs" ]; then
    echo "错误: 请在项目根目录运行此脚本"
    exit 1
fi

# 设置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
if ! command -v go &> /dev/null; then
    log_error "Go 未安装或未在PATH中"
    exit 1
fi

log_info "Go 版本: $(go version)"

# 定义服务配置
declare -A services
services[user]="cmd/user/user.go configs/user/srv.yaml"
services[goods]="cmd/goods/goods.go configs/goods/srv.yaml"
services[inventory]="cmd/inventory/inventory.go configs/inventory/srv.yaml"
services[order]="cmd/order/order.go configs/order/srv.yaml"
services[admin]="cmd/admin/admin.go configs/admin/admin.yaml"
services[shop]="cmd/shop/api.go configs/shop/api.yaml"

# 存储进程PID
declare -A pids

# 清理函数
cleanup() {
    log_warn "正在停止所有服务..."
    for service in "${!pids[@]}"; do
        if [ -n "${pids[$service]}" ]; then
            log_info "停止服务: $service (PID: ${pids[$service]})"
            kill -TERM "${pids[$service]}" 2>/dev/null
        fi
    done
    
    # 等待进程结束
    sleep 2
    
    # 强制杀死未结束的进程
    for service in "${!pids[@]}"; do
        if [ -n "${pids[$service]}" ]; then
            if kill -0 "${pids[$service]}" 2>/dev/null; then
                log_warn "强制停止服务: $service"
                kill -KILL "${pids[$service]}" 2>/dev/null
            fi
        fi
    done
    
    log_info "所有服务已停止"
    exit 0
}

# 注册信号处理
trap cleanup SIGINT SIGTERM

# 启动单个服务的函数
start_service() {
    local service=$1
    local cmd_file=$2
    local config_file=$3
    
    # 检查文件是否存在
    if [ ! -f "$cmd_file" ]; then
        log_error "服务文件不存在: $cmd_file"
        return 1
    fi
    
    if [ ! -f "$config_file" ]; then
        log_error "配置文件不存在: $config_file"
        return 1
    fi
    
    log_info "启动服务: $service"
    log_info "命令文件: $cmd_file"
    log_info "配置文件: $config_file"
    
    # 启动服务
    go run "$cmd_file" -c "$config_file" > "logs/${service}.log" 2>&1 &
    local pid=$!
    pids[$service]=$pid
    
    # 等待一下检查服务是否启动成功
    sleep 1
    if kill -0 $pid 2>/dev/null; then
        log_info "✓ 服务 $service 启动成功 (PID: $pid)"
        return 0
    else
        log_error "✗ 服务 $service 启动失败"
        return 1
    fi
}

# 创建日志目录
if [ ! -d "logs" ]; then
    mkdir -p logs
    log_info "创建日志目录: logs/"
fi

# 启动所有服务
log_info "开始启动所有服务..."
echo

failed_services=()

for service in "${!services[@]}"; do
    IFS=' ' read -r cmd_file config_file <<< "${services[$service]}"
    if ! start_service "$service" "$cmd_file" "$config_file"; then
        failed_services+=("$service")
    fi
    echo
done

# 检查启动结果
if [ ${#failed_services[@]} -eq 0 ]; then
    log_info "🎉 所有服务启动成功！"
else
    log_warn "以下服务启动失败:"
    for service in "${failed_services[@]}"; do
        log_error "  - $service"
    done
fi

# 显示运行状态
echo
log_info "=== 服务运行状态 ==="
for service in "${!pids[@]}"; do
    local pid=${pids[$service]}
    if [ -n "$pid" ] && kill -0 $pid 2>/dev/null; then
        log_info "✓ $service (PID: $pid) - 运行中"
    else
        log_error "✗ $service - 已停止"
    fi
done

echo
log_info "日志文件位置: logs/"
log_info "按 Ctrl+C 停止所有服务"
echo

# 保持脚本运行，监控服务状态
while true; do
    sleep 10
    
    # 检查服务状态
    for service in "${!pids[@]}"; do
        local pid=${pids[$service]}
        if [ -n "$pid" ] && ! kill -0 $pid 2>/dev/null; then
            log_warn "服务 $service 意外停止，正在重启..."
            IFS=' ' read -r cmd_file config_file <<< "${services[$service]}"
            start_service "$service" "$cmd_file" "$config_file"
        fi
    done
done