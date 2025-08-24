#!/bin/bash

# 专门测试goods服务和Canal消费者的脚本
echo "🧪 测试goods服务Canal消费者功能"

# 颜色定义  
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "📋 检查基础服务状态"
echo "================================="

# 检查Docker容器状态
echo "Docker容器状态:"
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(mysql|elasticsearch|rocketmq|canal)" | sed 's/^/  /'

# 检查服务端口
echo ""
echo "端口监听状态:"
for port in 3306 9200 9876 11111 18089; do
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        echo -e "  端口 $port: ${GREEN}✅ 监听中${NC}"
    else
        echo -e "  端口 $port: ${RED}❌ 未监听${NC}"
    fi
done

echo ""
echo "🔧 编译和配置检查"
echo "================================="

# 检查配置文件
if [ -f "configs/goods-canal.yaml" ]; then
    echo -e "✅ 配置文件存在: ${GREEN}configs/goods-canal.yaml${NC}"
    echo "配置内容预览:"
    head -n 15 configs/goods-canal.yaml | sed 's/^/  /'
else
    echo -e "❌ 配置文件不存在: ${RED}configs/goods-canal.yaml${NC}"
    exit 1
fi

# 编译goods服务
echo ""
if [ ! -f "./bin/goods" ]; then
    echo -n "🔨 编译goods服务... "
    if go build -o ./bin/goods ./cmd/goods/ 2>/tmp/build-error.log; then
        echo -e "${GREEN}✅ 编译成功${NC}"
    else
        echo -e "${RED}❌ 编译失败${NC}"
        echo "编译错误:"
        cat /tmp/build-error.log | sed 's/^/  /'
        exit 1
    fi
else
    echo -e "✅ goods服务已编译"
fi

echo ""
echo "🚀 启动goods服务测试"
echo "================================="

# 停止可能存在的goods进程
EXISTING_PID=$(pgrep -f "./bin/goods")
if [ -n "$EXISTING_PID" ]; then
    echo "🛑 停止现有goods服务 (PID: $EXISTING_PID)"
    kill $EXISTING_PID
    sleep 2
fi

# 启动goods服务
echo "启动goods服务..."
./bin/goods --config=configs/goods-canal.yaml > /tmp/goods-service-test.log 2>&1 &
GOODS_PID=$!

echo "Goods服务PID: $GOODS_PID"
echo "日志文件: /tmp/goods-service-test.log"

# 等待服务启动
echo -n "等待服务启动"
for i in {1..10}; do
    sleep 1
    echo -n "."
    if ! kill -0 $GOODS_PID 2>/dev/null; then
        echo -e "\n${RED}❌ 服务启动失败${NC}"
        echo "错误日志:"
        cat /tmp/goods-service-test.log | tail -n 20 | sed 's/^/  /'
        exit 1
    fi
done
echo -e "\n${GREEN}✅ 服务启动成功${NC}"

# 监控服务运行5分钟
echo ""
echo "📊 监控服务运行状态 (30秒)"
echo "================================="

for i in {1..30}; do
    if ! kill -0 $GOODS_PID 2>/dev/null; then
        echo -e "${RED}❌ 服务意外停止${NC}"
        break
    fi
    
    # 显示最新日志
    if [ -f "/tmp/goods-service-test.log" ]; then
        NEW_LOGS=$(tail -n 1 /tmp/goods-service-test.log 2>/dev/null)
        if [ -n "$NEW_LOGS" ] && [ "$NEW_LOGS" != "$LAST_LOG" ]; then
            echo -e "${BLUE}📄 $(date '+%H:%M:%S')${NC} $NEW_LOGS"
            LAST_LOG="$NEW_LOGS"
        fi
    fi
    
    sleep 1
done

echo ""
echo "📋 服务运行总结"
echo "================================="

if kill -0 $GOODS_PID 2>/dev/null; then
    echo -e "✅ goods服务状态: ${GREEN}正在运行${NC}"
else
    echo -e "❌ goods服务状态: ${RED}已停止${NC}"
fi

# 分析日志
if [ -f "/tmp/goods-service-test.log" ]; then
    echo ""
    echo "📊 日志分析:"
    
    TOTAL_LINES=$(wc -l < /tmp/goods-service-test.log)
    ERROR_COUNT=$(grep -c -i "error\|fatal\|panic" /tmp/goods-service-test.log)
    CANAL_COUNT=$(grep -c -i "canal" /tmp/goods-service-test.log)
    ROCKETMQ_COUNT=$(grep -c -i "rocketmq\|consumer" /tmp/goods-service-test.log)
    
    echo "  总日志行数: $TOTAL_LINES"
    echo "  错误日志数: $ERROR_COUNT"
    echo "  Canal相关: $CANAL_COUNT"
    echo "  RocketMQ相关: $ROCKETMQ_COUNT"
    
    if [ $ERROR_COUNT -eq 0 ]; then
        echo -e "  状态评估: ${GREEN}✅ 无错误${NC}"
    else
        echo -e "  状态评估: ${YELLOW}⚠️ 发现错误${NC}"
    fi
    
    echo ""
    echo "🔍 关键日志内容:"
    echo "================================="
    
    # 显示启动日志
    echo -e "${BLUE}启动日志:${NC}"
    head -n 10 /tmp/goods-service-test.log | sed 's/^/  /'
    
    # 显示Canal相关日志
    if [ $CANAL_COUNT -gt 0 ]; then
        echo -e "\n${BLUE}Canal相关日志:${NC}"
        grep -i "canal" /tmp/goods-service-test.log | tail -n 5 | sed 's/^/  /'
    fi
    
    # 显示错误日志
    if [ $ERROR_COUNT -gt 0 ]; then
        echo -e "\n${BLUE}错误日志:${NC}"
        grep -i "error\|fatal\|panic" /tmp/goods-service-test.log | tail -n 5 | sed 's/^/  /'
    fi
    
    # 显示最新日志
    echo -e "\n${BLUE}最新日志:${NC}"
    tail -n 5 /tmp/goods-service-test.log | sed 's/^/  /'
fi

echo ""
echo "🎯 测试结论"
echo "================================="

if kill -0 $GOODS_PID 2>/dev/null; then
    if [ $ERROR_COUNT -eq 0 ]; then
        echo -e "${GREEN}🎉 goods服务运行正常，Canal消费者已集成${NC}"
        echo ""
        echo "下一步建议:"
        echo "• 保持服务运行"
        echo "• 手动插入MySQL数据测试同步"
        echo "• 监控Canal Admin界面: http://localhost:18089"
        
    else
        echo -e "${YELLOW}⚠️ goods服务运行但有错误${NC}"
        echo "建议检查错误日志并修复"
    fi
    
    echo ""
    echo -e "停止服务请运行: ${BLUE}kill $GOODS_PID${NC}"
else
    echo -e "${RED}❌ goods服务启动失败${NC}"
    echo "请检查配置文件和依赖服务"
fi

echo ""
echo "📄 完整日志文件: /tmp/goods-service-test.log"