#!/bin/bash

# Canal+RocketMQ环境检查脚本
echo "🔍 开始检查Canal+RocketMQ环境..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查结果统计
PASS_COUNT=0
FAIL_COUNT=0

check_service() {
    local service_name=$1
    local host=$2
    local port=$3
    
    echo -n "检查 $service_name ($host:$port)..."
    
    if timeout 5 bash -c "echo >/dev/tcp/$host/$port" 2>/dev/null; then
        echo -e " ${GREEN}✅ 正常${NC}"
        ((PASS_COUNT++))
        return 0
    else
        echo -e " ${RED}❌ 失败${NC}"
        ((FAIL_COUNT++))
        return 1
    fi
}

check_command() {
    local cmd=$1
    local desc=$2
    
    echo -n "检查 $desc..."
    
    if command -v $cmd >/dev/null 2>&1; then
        echo -e " ${GREEN}✅ 已安装${NC}"
        ((PASS_COUNT++))
        return 0
    else
        echo -e " ${RED}❌ 未安装${NC}"
        ((FAIL_COUNT++))
        return 1
    fi
}

echo "📋 基础工具检查"
echo "================================="
check_command "mysql" "MySQL客户端"
check_command "curl" "HTTP客户端"
check_command "jq" "JSON处理工具"
check_command "telnet" "网络测试工具"

echo ""
echo "🌐 网络服务检查"
echo "================================="
check_service "MySQL" "localhost" "3306"
check_service "Elasticsearch" "localhost" "9200"  
check_service "RocketMQ NameServer" "localhost" "9876"
check_service "Canal Server" "localhost" "11111"
check_service "Canal Admin" "localhost" "18089"

echo ""
echo "🗄️ 数据库检查"
echo "================================="

# 检查MySQL连接和binlog
echo -n "检查MySQL binlog配置..."
BINLOG_STATUS=$(mysql -h localhost -P 3306 -u root -proot -e "SHOW VARIABLES LIKE 'log_bin';" 2>/dev/null | grep log_bin | awk '{print $2}')

if [ "$BINLOG_STATUS" = "ON" ]; then
    echo -e " ${GREEN}✅ binlog已开启${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${RED}❌ binlog未开启${NC}"
    echo "   请在MySQL配置文件中添加: log-bin=mysql-bin"
    ((FAIL_COUNT++))
fi

# 检查binlog格式
echo -n "检查MySQL binlog格式..."
BINLOG_FORMAT=$(mysql -h localhost -P 3306 -u root -proot -e "SHOW VARIABLES LIKE 'binlog_format';" 2>/dev/null | grep binlog_format | awk '{print $2}')

if [ "$BINLOG_FORMAT" = "ROW" ]; then
    echo -e " ${GREEN}✅ 格式正确 (ROW)${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${YELLOW}⚠️ 格式为 $BINLOG_FORMAT，建议使用ROW格式${NC}"
    ((PASS_COUNT++))
fi

# 检查目标数据库是否存在
echo -n "检查emshop数据库..."
DB_EXISTS=$(mysql -h localhost -P 3306 -u root -proot -e "SHOW DATABASES LIKE 'emshop';" 2>/dev/null | grep emshop)

if [ -n "$DB_EXISTS" ]; then
    echo -e " ${GREEN}✅ 数据库存在${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${RED}❌ emshop数据库不存在${NC}"
    echo "   创建数据库: CREATE DATABASE emshop;"
    ((FAIL_COUNT++))
fi

echo ""
echo "🔍 Elasticsearch检查"
echo "================================="

# 检查ES连接
echo -n "检查Elasticsearch连接..."
ES_STATUS=$(curl -s http://localhost:9200/_cluster/health 2>/dev/null | jq -r '.status' 2>/dev/null)

if [ "$ES_STATUS" = "green" ] || [ "$ES_STATUS" = "yellow" ]; then
    echo -e " ${GREEN}✅ 连接正常 (状态: $ES_STATUS)${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${RED}❌ 连接失败${NC}"
    ((FAIL_COUNT++))
fi

# 检查goods索引
echo -n "检查goods索引..."
GOODS_INDEX=$(curl -s http://localhost:9200/goods 2>/dev/null | jq -r '.error.type' 2>/dev/null)

if [ "$GOODS_INDEX" != "index_not_found_exception" ] && [ -n "$(curl -s http://localhost:9200/goods 2>/dev/null)" ]; then
    echo -e " ${GREEN}✅ 索引存在${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${YELLOW}⚠️ 索引不存在，将自动创建${NC}"
    ((PASS_COUNT++))
fi

echo ""
echo "🚀 RocketMQ检查"
echo "================================="

# 检查RocketMQ控制台 (如果有的话)
echo -n "检查RocketMQ控制台..."
if timeout 5 bash -c "echo >/dev/tcp/localhost/8080" 2>/dev/null; then
    echo -e " ${GREEN}✅ 控制台可访问 (localhost:8080)${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${YELLOW}⚠️ 控制台不可访问 (可选)${NC}"
    ((PASS_COUNT++))
fi

echo ""
echo "📁 配置文件检查"
echo "================================="

# 检查Canal配置文件
echo -n "检查Canal实例配置..."
if [ -f "components/mysql-canal/canal-server/conf/example/instance.properties" ]; then
    echo -e " ${GREEN}✅ 配置文件存在${NC}"
    ((PASS_COUNT++))
    
    # 检查关键配置项
    if grep -q "emshop\\\\.goods" components/mysql-canal/canal-server/conf/example/instance.properties; then
        echo -e "   ${GREEN}✅ goods表监听已配置${NC}"
    else
        echo -e "   ${RED}❌ goods表监听未配置${NC}"
        ((FAIL_COUNT++))
    fi
else
    echo -e " ${RED}❌ Canal配置文件不存在${NC}"
    ((FAIL_COUNT++))
fi

# 检查goods服务配置
echo -n "检查goods服务配置..."
if [ -f "configs/goods-canal.yaml" ]; then
    echo -e " ${GREEN}✅ 配置文件存在${NC}"
    ((PASS_COUNT++))
else
    echo -e " ${RED}❌ goods服务配置不存在${NC}"
    ((FAIL_COUNT++))
fi

echo ""
echo "📊 环境检查总结"
echo "================================="
echo -e "通过检查项: ${GREEN}$PASS_COUNT${NC}"
echo -e "失败检查项: ${RED}$FAIL_COUNT${NC}"

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}🎉 环境检查全部通过，可以开始测试！${NC}"
    exit 0
else
    echo -e "${RED}❌ 发现 $FAIL_COUNT 个问题，请修复后再进行测试${NC}"
    echo ""
    echo "💡 修复建议:"
    echo "1. 确保所有服务正常运行"
    echo "2. 检查MySQL binlog配置"
    echo "3. 创建必要的数据库和表"
    echo "4. 验证网络连通性"
    exit 1
fi