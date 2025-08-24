#!/bin/bash

# 启动Goods服务并集成Canal + RocketMQ消费者
# 使用方法: ./scripts/start-goods-service-with-canal.sh

set -e

echo "🚀 启动Goods服务 (集成Canal + RocketMQ)"

# 设置工作目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# 检查配置文件
CONFIG_FILE="configs/goods-service-with-canal.yaml"
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "❌ 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 检查依赖服务
echo "📋 检查依赖服务状态..."

# 检查MySQL
if ! docker ps | grep -q emshop-mysql; then
    echo "❌ MySQL服务未运行，请先启动MySQL"
    exit 1
fi

# 检查RocketMQ NameServer
if ! docker ps | grep -q rmqnamesrv; then
    echo "❌ RocketMQ NameServer未运行，请先启动RocketMQ"
    exit 1
fi

# 检查RocketMQ Broker
if ! docker ps | grep -q rmqbroker; then
    echo "❌ RocketMQ Broker未运行，请先启动RocketMQ"
    exit 1
fi

# 检查Canal服务
if ! docker ps | grep -q emshop-canal-server; then
    echo "❌ Canal服务未运行，请先启动Canal"
    exit 1
fi

echo "✅ 所有依赖服务正常运行"

# 构建服务
echo "🔨 构建Goods服务..."
go mod tidy
go build -o bin/goods-srv ./cmd/goods-srv/

# 启动服务
echo "🎯 启动Goods服务，配置文件: $CONFIG_FILE"
echo "📥 Canal消费者将自动启动并监听RocketMQ消息"
echo "🔄 MySQL变更将自动同步到Elasticsearch"
echo "🛑 按Ctrl+C停止服务"
echo ""

./bin/goods-srv --config="$CONFIG_FILE"