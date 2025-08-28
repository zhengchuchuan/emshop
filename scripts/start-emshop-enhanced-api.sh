#!/bin/bash

# EMShop 增强API服务启动脚本
# 提供开发和生产环境的启动选项
# Author: Claude Code
# Version: 1.0

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 默认配置
DEFAULT_ENV="development"
DEFAULT_CONFIG="$PROJECT_ROOT/configs/emshop/api.yaml"
DEFAULT_PORT="8052"

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

show_help() {
    echo -e "${BLUE}EMShop 增强API服务启动脚本${NC}"
    echo
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -h, --help              显示此帮助信息"
    echo "  -e, --env ENV           指定环境 (development|production) [默认: development]"
    echo "  -c, --config FILE       指定配置文件路径"
    echo "  -p, --port PORT         指定HTTP端口 [默认: 8052]"
    echo "  -d, --daemon            后台运行"
    echo "  --docker                使用Docker运行"
    echo "  --build                 重新构建（仅Docker模式）"
    echo "  --logs                  查看日志（仅Docker模式）"
    echo "  --stop                  停止服务（仅Docker模式）"
    echo "  --status                查看服务状态"
    echo
    echo "示例:"
    echo "  $0                                    # 开发环境运行"
    echo "  $0 -e production                     # 生产环境运行"
    echo "  $0 --docker                          # Docker方式运行"
    echo "  $0 --docker --build                  # 重新构建并运行"
    echo "  $0 --logs                            # 查看Docker日志"
    echo "  $0 --stop                            # 停止Docker服务"
}

# 检查前置条件
check_prerequisites() {
    log_info "检查前置条件..."
    
    # 检查Go版本
    if command -v go >/dev/null 2>&1; then
        go_version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
        log_success "✅ Go版本: $go_version"
    else
        log_error "❌ 未安装Go语言"
        return 1
    fi
    
    # 检查项目依赖
    cd "$PROJECT_ROOT"
    if [ -f "go.mod" ]; then
        log_success "✅ 项目依赖配置正常"
    else
        log_error "❌ 未找到go.mod文件"
        return 1
    fi
    
    return 0
}

# 检查服务状态
check_service_status() {
    local port=${1:-$DEFAULT_PORT}
    
    if curl -s "http://localhost:$port/healthz" >/dev/null 2>&1; then
        log_success "✅ API服务运行正常 (端口: $port)"
        return 0
    else
        log_warning "⚠️ API服务未运行或不健康 (端口: $port)"
        return 1
    fi
}

# 原生方式启动
start_native() {
    local env="$1"
    local config="$2"
    local port="$3"
    local daemon="$4"
    
    log_info "以原生方式启动EMShop增强API服务..."
    log_info "环境: $env"
    log_info "配置文件: $config"
    log_info "HTTP端口: $port"
    
    if [ ! -f "$config" ]; then
        log_error "❌ 配置文件不存在: $config"
        return 1
    fi
    
    cd "$PROJECT_ROOT"
    
    # 确保日志目录存在
    mkdir -p logs
    
    # 设置环境变量
    export ENV="$env"
    export CONFIG_PATH="$config"
    
    # 构建命令
    build_cmd="go run cmd/api/emshop/main.go -c $config"
    
    if [ "$daemon" = true ]; then
        log_info "后台启动服务..."
        nohup $build_cmd > logs/emshop-enhanced-api.out 2>&1 &
        local pid=$!
        echo $pid > logs/emshop-enhanced-api.pid
        log_success "✅ 服务已在后台启动 (PID: $pid)"
        
        # 等待服务启动
        sleep 3
        if check_service_status "$port"; then
            log_success "🎉 EMShop增强API服务启动成功!"
            echo -e "${BLUE}服务信息:${NC}"
            echo "  - HTTP API: http://localhost:$port"
            echo "  - 健康检查: http://localhost:$port/healthz"
            echo "  - 监控指标: http://localhost:$port/metrics"
            echo "  - PID文件: $PROJECT_ROOT/logs/emshop-enhanced-api.pid"
            echo "  - 日志文件: $PROJECT_ROOT/logs/emshop-enhanced-api.out"
        else
            log_error "❌ 服务启动失败，请检查日志"
            return 1
        fi
    else
        log_info "前台启动服务..."
        log_success "🎉 启动EMShop增强API服务..."
        echo -e "${BLUE}服务将在以下地址提供服务:${NC}"
        echo "  - HTTP API: http://localhost:$port"
        echo "  - 健康检查: http://localhost:$port/healthz"
        echo "  - 监控指标: http://localhost:$port/metrics"
        echo
        exec $build_cmd
    fi
}

# Docker方式启动
start_docker() {
    local build="$1"
    
    log_info "以Docker方式启动EMShop增强API服务..."
    
    cd "$PROJECT_ROOT"
    
    if [ "$build" = true ]; then
        log_info "重新构建Docker镜像..."
        docker-compose -f docker-compose.emshop-api.yml build --no-cache
    fi
    
    # 确保网络存在
    docker network create emshop-network 2>/dev/null || true
    
    # 启动服务
    docker-compose -f docker-compose.emshop-api.yml up -d
    
    log_success "✅ Docker服务启动成功"
    
    # 等待服务就绪
    log_info "等待服务就绪..."
    local retries=0
    local max_retries=30
    
    while [ $retries -lt $max_retries ]; do
        if docker-compose -f docker-compose.emshop-api.yml ps | grep -q "Up"; then
            if check_service_status "8052"; then
                log_success "🎉 EMShop增强API服务启动成功!"
                docker-compose -f docker-compose.emshop-api.yml ps
                echo
                echo -e "${BLUE}服务信息:${NC}"
                echo "  - HTTP API: http://localhost:8052"
                echo "  - 健康检查: http://localhost:8052/healthz"
                echo "  - 监控指标: http://localhost:8052/metrics"
                return 0
            fi
        fi
        
        sleep 2
        ((retries++))
    done
    
    log_error "❌ 服务启动超时，请检查Docker日志"
    docker-compose -f docker-compose.emshop-api.yml logs --tail=50
    return 1
}

# 查看Docker日志
show_docker_logs() {
    cd "$PROJECT_ROOT"
    docker-compose -f docker-compose.emshop-api.yml logs -f
}

# 停止Docker服务
stop_docker() {
    cd "$PROJECT_ROOT"
    log_info "停止EMShop增强API服务..."
    docker-compose -f docker-compose.emshop-api.yml down
    log_success "✅ 服务已停止"
}

# 停止原生服务
stop_native() {
    local pid_file="$PROJECT_ROOT/logs/emshop-enhanced-api.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            log_info "停止服务 (PID: $pid)..."
            kill "$pid"
            rm -f "$pid_file"
            log_success "✅ 服务已停止"
        else
            log_warning "⚠️ 服务进程不存在"
            rm -f "$pid_file"
        fi
    else
        log_warning "⚠️ 未找到PID文件，服务可能未在后台运行"
    fi
}

# 主函数
main() {
    local env="$DEFAULT_ENV"
    local config="$DEFAULT_CONFIG"
    local port="$DEFAULT_PORT"
    local daemon=false
    local docker=false
    local build=false
    local logs=false
    local stop=false
    local status=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -e|--env)
                env="$2"
                shift 2
                ;;
            -c|--config)
                config="$2"
                shift 2
                ;;
            -p|--port)
                port="$2"
                shift 2
                ;;
            -d|--daemon)
                daemon=true
                shift
                ;;
            --docker)
                docker=true
                shift
                ;;
            --build)
                build=true
                shift
                ;;
            --logs)
                logs=true
                shift
                ;;
            --stop)
                stop=true
                shift
                ;;
            --status)
                status=true
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 根据环境调整配置文件
    if [ "$env" = "production" ] && [ "$config" = "$DEFAULT_CONFIG" ]; then
        config="$PROJECT_ROOT/configs/emshop/api-production.yaml"
    fi
    
    echo -e "${BLUE}"
    echo "======================================================"
    echo "       EMShop 增强API服务启动脚本"
    echo "======================================================"
    echo -e "${NC}"
    
    # 执行相应操作
    if [ "$status" = true ]; then
        check_service_status "$port"
    elif [ "$stop" = true ]; then
        if [ "$docker" = true ]; then
            stop_docker
        else
            stop_native
        fi
    elif [ "$logs" = true ]; then
        if [ "$docker" = true ]; then
            show_docker_logs
        else
            log_error "原生模式请直接查看日志文件: $PROJECT_ROOT/logs/"
        fi
    else
        # 检查前置条件
        if ! check_prerequisites; then
            exit 1
        fi
        
        # 启动服务
        if [ "$docker" = true ]; then
            start_docker "$build"
        else
            start_native "$env" "$config" "$port" "$daemon"
        fi
    fi
}

# 捕获中断信号
trap 'log_info "收到中断信号，正在退出..."; exit 0' INT TERM

# 执行主函数
main "$@"