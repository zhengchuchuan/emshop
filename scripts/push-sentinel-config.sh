#!/bin/bash

# Sentinel配置推送到Nacos脚本
# 使用方法: ./push-sentinel-config.sh [nacos-server] [namespace]

set -e

NACOS_SERVER=${1:-"127.0.0.1:8848"}
NAMESPACE=${2:-""}
GROUP="sentinel-go"
CONFIG_DIR="./configs/sentinel"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Nacos服务器是否可访问
check_nacos() {
    log_info "检查Nacos服务器连通性: $NACOS_SERVER"
    if ! curl -s "http://$NACOS_SERVER/nacos/v1/ns/operator/metrics" > /dev/null; then
        log_error "无法连接到Nacos服务器: $NACOS_SERVER"
        log_info "请确保Nacos服务器正在运行并且地址正确"
        exit 1
    fi
    log_info "Nacos服务器连接正常"
}

# 推送配置文件到Nacos
push_config() {
    local data_id=$1
    local config_file=$2
    
    if [[ ! -f "$config_file" ]]; then
        log_warn "配置文件不存在: $config_file"
        return 1
    fi
    
    log_info "推送配置: $data_id"
    
    local url="http://$NACOS_SERVER/nacos/v1/cs/configs"
    local content=$(cat "$config_file")
    
    # 构建请求参数
    local params="dataId=${data_id}&group=${GROUP}&content=${content}"
    
    if [[ -n "$NAMESPACE" ]]; then
        params="${params}&tenant=${NAMESPACE}"
    fi
    
    # 发送HTTP请求
    local response=$(curl -s -X POST "$url" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "$params")
    
    if [[ "$response" == "true" ]]; then
        log_info "✅ 成功推送: $data_id"
    else
        log_error "❌ 推送失败: $data_id, 响应: $response"
        return 1
    fi
}

# 推送所有优惠券服务的Sentinel配置
push_coupon_configs() {
    log_info "推送优惠券服务Sentinel配置..."
    
    push_config "coupon-flow-rules" "$CONFIG_DIR/coupon-flow-rules.json"
    push_config "coupon-circuit-breaker-rules" "$CONFIG_DIR/coupon-circuit-breaker-rules.json"
    push_config "coupon-hotspot-rules" "$CONFIG_DIR/coupon-hotspot-rules.json"
    push_config "coupon-system-rules" "$CONFIG_DIR/coupon-system-rules.json"
}

# 推送用户服务的Sentinel配置
push_user_configs() {
    log_info "推送用户服务Sentinel配置..."
    
    # 生成用户服务的流控规则
    cat > "$CONFIG_DIR/user-flow-rules.json" << 'EOF'
[
  {
    "resource": "user-srv:CreateUser",
    "tokenCalculateStrategy": 0,
    "controlBehavior": 0,
    "threshold": 100.0,
    "relationStrategy": 0,
    "refResource": "",
    "maxQueueingTimeMs": 500,
    "warmUpPeriodSec": 10,
    "warmUpColdFactor": 3,
    "statIntervalInMs": 1000
  },
  {
    "resource": "user-srv:GetUserByMobile",
    "tokenCalculateStrategy": 0,
    "controlBehavior": 0,
    "threshold": 300.0,
    "relationStrategy": 0,
    "refResource": "",
    "maxQueueingTimeMs": 200,
    "warmUpPeriodSec": 5,
    "warmUpColdFactor": 3,
    "statIntervalInMs": 1000
  },
  {
    "resource": "user-srv:GetUserById",
    "tokenCalculateStrategy": 0,
    "controlBehavior": 0,
    "threshold": 1000.0,
    "relationStrategy": 0,
    "refResource": "",
    "maxQueueingTimeMs": 100,
    "warmUpPeriodSec": 5,
    "warmUpColdFactor": 3,
    "statIntervalInMs": 1000
  }
]
EOF

    # 生成用户服务的熔断规则
    cat > "$CONFIG_DIR/user-circuit-breaker-rules.json" << 'EOF'
[
  {
    "resource": "user-srv:CreateUser",
    "strategy": 0,
    "retryTimeoutMs": 10000,
    "minRequestAmount": 10,
    "statIntervalMs": 1000,
    "statSlidingWindowBucketCount": 10,
    "maxAllowedRtMs": 1000,
    "threshold": 0.4
  },
  {
    "resource": "user-srv:GetUserByMobile",
    "strategy": 1,
    "retryTimeoutMs": 5000,
    "minRequestAmount": 15,
    "statIntervalMs": 1000,
    "statSlidingWindowBucketCount": 10,
    "maxAllowedRtMs": 500,
    "threshold": 0.6
  }
]
EOF

    # 推送配置
    push_config "user-flow-rules" "$CONFIG_DIR/user-flow-rules.json"
    push_config "user-circuit-breaker-rules" "$CONFIG_DIR/user-circuit-breaker-rules.json"
}

# 推送库存服务的Sentinel配置
push_inventory_configs() {
    log_info "推送库存服务Sentinel配置..."
    
    # 生成库存服务的流控规则
    cat > "$CONFIG_DIR/inventory-flow-rules.json" << 'EOF'
[
  {
    "resource": "inventory-srv:Sell",
    "tokenCalculateStrategy": 0,
    "controlBehavior": 0,
    "threshold": 300.0,
    "relationStrategy": 0,
    "refResource": "",
    "maxQueueingTimeMs": 1000,
    "warmUpPeriodSec": 15,
    "warmUpColdFactor": 5,
    "statIntervalInMs": 1000
  },
  {
    "resource": "inventory-srv:InvDetail",
    "tokenCalculateStrategy": 0,
    "controlBehavior": 0,
    "threshold": 2000.0,
    "relationStrategy": 0,
    "refResource": "",
    "maxQueueingTimeMs": 100,
    "warmUpPeriodSec": 5,
    "warmUpColdFactor": 3,
    "statIntervalInMs": 1000
  },
  {
    "resource": "inventory-srv:Reback",
    "tokenCalculateStrategy": 0,
    "controlBehavior": 0,
    "threshold": 200.0,
    "relationStrategy": 0,
    "refResource": "",
    "maxQueueingTimeMs": 500,
    "warmUpPeriodSec": 10,
    "warmUpColdFactor": 3,
    "statIntervalInMs": 1000
  }
]
EOF

    # 推送配置
    push_config "inventory-flow-rules" "$CONFIG_DIR/inventory-flow-rules.json"
}

# 主函数
main() {
    log_info "=== EMShop Sentinel配置推送工具 ==="
    log_info "Nacos服务器: $NACOS_SERVER"
    log_info "命名空间: ${NAMESPACE:-default}"
    log_info "配置组: $GROUP"
    echo ""
    
    # 创建配置目录
    mkdir -p "$CONFIG_DIR"
    
    # 检查Nacos连通性
    check_nacos
    echo ""
    
    # 推送各服务配置
    push_coupon_configs
    echo ""
    
    push_user_configs
    echo ""
    
    push_inventory_configs
    echo ""
    
    log_info "🎉 所有Sentinel配置推送完成！"
    log_info ""
    log_info "📋 配置检查命令:"
    log_info "  curl 'http://$NACOS_SERVER/nacos/v1/cs/configs?dataId=coupon-flow-rules&group=$GROUP'"
    log_info ""
    log_info "🔧 Nacos控制台:"
    log_info "  http://$NACOS_SERVER/nacos"
}

# 执行主函数
main "$@"