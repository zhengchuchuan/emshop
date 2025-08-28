#!/bin/bash

# EMShop 增强API服务监控系统部署脚本
# 自动配置Prometheus、Grafana和AlertManager
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
    echo -e "${BLUE}EMShop 监控系统部署脚本${NC}"
    echo
    echo "用法: $0 [选项]"
    echo
    echo "选项:"
    echo "  -h, --help         显示此帮助信息"
    echo "  --setup-all        部署完整监控系统"
    echo "  --prometheus       仅部署Prometheus"
    echo "  --grafana         仅部署Grafana"
    echo "  --alertmanager    仅部署AlertManager"
    echo "  --cleanup         清理监控系统"
    echo "  --status          查看监控系统状态"
    echo
    echo "示例:"
    echo "  $0 --setup-all                        # 部署完整监控系统"
    echo "  $0 --prometheus                       # 仅部署Prometheus"
    echo "  $0 --status                          # 查看系统状态"
}

# 检查Docker和Docker Compose
check_prerequisites() {
    log_info "检查前置条件..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        return 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        return 1
    fi
    
    # 检查网络
    if ! docker network ls | grep -q emshop-network; then
        log_info "创建Docker网络..."
        docker network create emshop-network
        log_success "✅ Docker网络创建成功"
    fi
    
    log_success "✅ 前置条件检查通过"
    return 0
}

# 创建监控配置目录
setup_directories() {
    log_info "创建监控配置目录..."
    
    local dirs=(
        "$PROJECT_ROOT/monitoring/prometheus/data"
        "$PROJECT_ROOT/monitoring/prometheus/config"
        "$PROJECT_ROOT/monitoring/grafana/data"
        "$PROJECT_ROOT/monitoring/grafana/dashboards"
        "$PROJECT_ROOT/monitoring/grafana/provisioning/datasources"
        "$PROJECT_ROOT/monitoring/grafana/provisioning/dashboards"
        "$PROJECT_ROOT/monitoring/alertmanager/data"
        "$PROJECT_ROOT/monitoring/alertmanager/config"
        "$PROJECT_ROOT/logs/monitoring"
    )
    
    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
        # 设置适当的权限（Grafana需要472用户组）
        sudo chown -R 472:472 "$PROJECT_ROOT/monitoring/grafana" 2>/dev/null || true
        sudo chown -R nobody:nogroup "$PROJECT_ROOT/monitoring/prometheus" 2>/dev/null || true
    done
    
    log_success "✅ 监控目录创建完成"
}

# 复制配置文件
copy_configs() {
    log_info "复制监控配置文件..."
    
    # 复制Prometheus配置
    if [ -f "$PROJECT_ROOT/configs/prometheus/emshop-api-monitoring.yml" ]; then
        cp "$PROJECT_ROOT/configs/prometheus/emshop-api-monitoring.yml" \
           "$PROJECT_ROOT/monitoring/prometheus/config/prometheus.yml"
        log_success "✅ Prometheus配置文件复制成功"
    else
        log_error "❌ Prometheus配置文件不存在"
        return 1
    fi
    
    # 复制告警规则
    if [ -f "$PROJECT_ROOT/configs/prometheus/rules/emshop-api-alerts.yml" ]; then
        mkdir -p "$PROJECT_ROOT/monitoring/prometheus/config/rules"
        cp "$PROJECT_ROOT/configs/prometheus/rules/emshop-api-alerts.yml" \
           "$PROJECT_ROOT/monitoring/prometheus/config/rules/"
        log_success "✅ 告警规则文件复制成功"
    fi
    
    # 复制Grafana仪表板
    if [ -f "$PROJECT_ROOT/configs/grafana/emshop-api-dashboard.json" ]; then
        cp "$PROJECT_ROOT/configs/grafana/emshop-api-dashboard.json" \
           "$PROJECT_ROOT/monitoring/grafana/dashboards/"
        log_success "✅ Grafana仪表板复制成功"
    fi
    
    return 0
}

# 创建Grafana数据源配置
create_grafana_datasource() {
    log_info "创建Grafana数据源配置..."
    
    cat > "$PROJECT_ROOT/monitoring/grafana/provisioning/datasources/prometheus.yml" << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
    jsonData:
      timeInterval: 15s
      queryTimeout: 60s
      httpMethod: POST
EOF

    log_success "✅ Grafana数据源配置创建完成"
}

# 创建Grafana仪表板配置
create_grafana_dashboard_config() {
    log_info "创建Grafana仪表板配置..."
    
    cat > "$PROJECT_ROOT/monitoring/grafana/provisioning/dashboards/dashboards.yml" << 'EOF'
apiVersion: 1

providers:
  - name: 'EMShop API Dashboards'
    orgId: 1
    folder: 'EMShop'
    type: file
    disableDeletion: false
    editable: true
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

    log_success "✅ Grafana仪表板配置创建完成"
}

# 创建AlertManager配置
create_alertmanager_config() {
    log_info "创建AlertManager配置..."
    
    cat > "$PROJECT_ROOT/monitoring/alertmanager/config/alertmanager.yml" << 'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@emshop.com'

route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
  - match:
      severity: critical
    receiver: 'critical-alerts'
  - match:
      severity: warning
    receiver: 'warning-alerts'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://localhost:5001/webhook'
    
- name: 'critical-alerts'
  email_configs:
  - to: 'admin@emshop.com'
    subject: '[CRITICAL] EMShop API 告警'
    body: |
      {{ range .Alerts }}
      告警: {{ .Annotations.summary }}
      描述: {{ .Annotations.description }}
      时间: {{ .StartsAt }}
      {{ end }}
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK_URL'
    channel: '#alerts-critical'
    title: 'EMShop API 紧急告警'
    text: |
      {{ range .Alerts }}
      {{ .Annotations.summary }}
      {{ .Annotations.description }}
      {{ end }}

- name: 'warning-alerts'
  email_configs:
  - to: 'ops@emshop.com'
    subject: '[WARNING] EMShop API 警告'
    body: |
      {{ range .Alerts }}
      告警: {{ .Annotations.summary }}
      描述: {{ .Annotations.description }}
      时间: {{ .StartsAt }}
      {{ end }}
EOF

    log_success "✅ AlertManager配置创建完成"
}

# 创建Docker Compose监控配置
create_monitoring_compose() {
    log_info "创建监控Docker Compose配置..."
    
    cat > "$PROJECT_ROOT/docker-compose.monitoring.yml" << 'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v2.40.0
    container_name: emshop-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/config:/etc/prometheus:ro
      - ./monitoring/prometheus/data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
    networks:
      - emshop-network
    labels:
      - "service=prometheus"
      - "monitoring=emshop"

  grafana:
    image: grafana/grafana:9.2.0
    container_name: emshop-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - ./monitoring/grafana/data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=emshop123
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_SERVER_DOMAIN=localhost
      - GF_SMTP_ENABLED=true
      - GF_SMTP_HOST=smtp.gmail.com:587
      - GF_SMTP_USER=your-email@gmail.com
      - GF_SMTP_PASSWORD=your-password
      - GF_SMTP_FROM_ADDRESS=your-email@gmail.com
      - GF_SMTP_FROM_NAME=EMShop Monitoring
    networks:
      - emshop-network
    depends_on:
      - prometheus
    labels:
      - "service=grafana"
      - "monitoring=emshop"

  alertmanager:
    image: prom/alertmanager:v0.25.0
    container_name: emshop-alertmanager
    restart: unless-stopped
    ports:
      - "9093:9093"
    volumes:
      - ./monitoring/alertmanager/config:/etc/alertmanager:ro
      - ./monitoring/alertmanager/data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    networks:
      - emshop-network
    labels:
      - "service=alertmanager"
      - "monitoring=emshop"

  node-exporter:
    image: prom/node-exporter:v1.4.0
    container_name: emshop-node-exporter
    restart: unless-stopped
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.ignored-mount-points=^/(sys|proc|dev|host|etc)($$|/)'
    networks:
      - emshop-network
    labels:
      - "service=node-exporter"
      - "monitoring=emshop"

networks:
  emshop-network:
    external: true
EOF

    log_success "✅ 监控Docker Compose配置创建完成"
}

# 部署监控系统
deploy_monitoring() {
    local component="$1"
    
    log_info "部署监控系统..."
    
    cd "$PROJECT_ROOT"
    
    case "$component" in
        "all"|"")
            log_info "部署完整监控系统..."
            docker-compose -f docker-compose.monitoring.yml up -d
            ;;
        "prometheus")
            log_info "部署Prometheus..."
            docker-compose -f docker-compose.monitoring.yml up -d prometheus
            ;;
        "grafana")
            log_info "部署Grafana..."
            docker-compose -f docker-compose.monitoring.yml up -d grafana
            ;;
        "alertmanager")
            log_info "部署AlertManager..."
            docker-compose -f docker-compose.monitoring.yml up -d alertmanager
            ;;
        *)
            log_error "未知组件: $component"
            return 1
            ;;
    esac
    
    log_success "✅ 监控系统部署成功"
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10
    
    # 显示访问信息
    show_access_info
}

# 显示访问信息
show_access_info() {
    log_info "监控系统访问信息:"
    echo
    echo -e "${BLUE}服务访问地址:${NC}"
    echo "  - Prometheus: http://localhost:9090"
    echo "  - Grafana:    http://localhost:3000 (admin/emshop123)"
    echo "  - AlertManager: http://localhost:9093"
    echo "  - Node Exporter: http://localhost:9100"
    echo
    echo -e "${BLUE}EMShop API监控:${NC}"
    echo "  - API指标: http://localhost:8052/metrics"
    echo "  - 健康检查: http://localhost:8052/healthz"
    echo
    echo -e "${BLUE}有用的命令:${NC}"
    echo "  - 查看日志: docker-compose -f docker-compose.monitoring.yml logs -f"
    echo "  - 重启服务: docker-compose -f docker-compose.monitoring.yml restart"
    echo "  - 停止监控: $0 --cleanup"
}

# 查看监控系统状态
check_monitoring_status() {
    log_info "检查监控系统状态..."
    
    cd "$PROJECT_ROOT"
    
    if [ -f "docker-compose.monitoring.yml" ]; then
        docker-compose -f docker-compose.monitoring.yml ps
        echo
        
        # 检查服务健康状态
        services=("prometheus:9090" "grafana:3000" "alertmanager:9093")
        for service in "${services[@]}"; do
            local name=$(echo "$service" | cut -d':' -f1)
            local port=$(echo "$service" | cut -d':' -f2)
            
            if curl -s "http://localhost:$port" >/dev/null 2>&1; then
                log_success "✅ $name 服务运行正常"
            else
                log_warning "⚠️ $name 服务可能存在问题"
            fi
        done
    else
        log_warning "⚠️ 监控系统未部署"
    fi
}

# 清理监控系统
cleanup_monitoring() {
    log_info "清理监控系统..."
    
    cd "$PROJECT_ROOT"
    
    if [ -f "docker-compose.monitoring.yml" ]; then
        docker-compose -f docker-compose.monitoring.yml down -v
        log_success "✅ 监控容器已停止和删除"
        
        # 可选：删除数据目录
        read -p "是否删除监控数据目录? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            sudo rm -rf "$PROJECT_ROOT/monitoring"
            rm -f "$PROJECT_ROOT/docker-compose.monitoring.yml"
            log_success "✅ 监控数据已清理"
        fi
    else
        log_warning "⚠️ 未找到监控系统配置"
    fi
}

# 主函数
main() {
    local action=""
    local component=""
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --setup-all)
                action="setup"
                component="all"
                shift
                ;;
            --prometheus)
                action="setup"
                component="prometheus"
                shift
                ;;
            --grafana)
                action="setup"
                component="grafana"
                shift
                ;;
            --alertmanager)
                action="setup"
                component="alertmanager"
                shift
                ;;
            --cleanup)
                action="cleanup"
                shift
                ;;
            --status)
                action="status"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    if [ -z "$action" ]; then
        show_help
        exit 1
    fi
    
    echo -e "${BLUE}"
    echo "======================================================"
    echo "       EMShop 监控系统部署脚本"
    echo "======================================================"
    echo -e "${NC}"
    
    # 执行相应操作
    case "$action" in
        "setup")
            if ! check_prerequisites; then
                exit 1
            fi
            setup_directories
            copy_configs
            create_grafana_datasource
            create_grafana_dashboard_config
            create_alertmanager_config
            create_monitoring_compose
            deploy_monitoring "$component"
            ;;
        "status")
            check_monitoring_status
            ;;
        "cleanup")
            cleanup_monitoring
            ;;
    esac
    
    log_success "🎉 监控系统操作完成！"
}

# 捕获中断信号
trap 'log_info "收到中断信号，正在退出..."; exit 0' INT TERM

# 执行主函数
main "$@"