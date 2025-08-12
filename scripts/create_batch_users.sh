#!/bin/bash

# 批量创建用户脚本
# 使用方法: ./create_batch_users.sh

BASE_URL="http://localhost:8022"
PASSWORD="admin123"

echo "开始批量创建用户..."
echo "服务地址: $BASE_URL"
echo "统一密码: $PASSWORD"
echo "=========================="

# 用户数据数组
declare -a users=(
    "张三:13800000001"
    "李四:13800000002" 
    "王五:13800000003"
    "赵六:13800000004"
    "钱七:13800000005"
    "孙八:13800000006"
    "周九:13800000007"
    "吴十:13800000008"
    "郑十一:13800000009"
    "王十二:13800000010"
    "管理员:13900000001"
    "测试用户1:13900000002"
    "测试用户2:13900000003"
    "测试用户3:13900000004"
    "客服:13900000005"
)

success_count=0
total_count=${#users[@]}

# 遍历创建用户
for i in "${!users[@]}"; do
    IFS=':' read -r nickname mobile <<< "${users[$i]}"
    
    echo "正在创建用户 $((i+1))/$total_count: $nickname ($mobile)"
    
    # 构造JSON请求体
    json_data=$(cat <<EOF
{
    "nickName": "$nickname",
    "passWord": "$PASSWORD", 
    "mobile": "$mobile"
}
EOF
)
    
    # 发送POST请求
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/user/create" \
        -H "Content-Type: application/json" \
        -d "$json_data")
    
    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        echo "✅ 创建成功: $nickname"
        ((success_count++))
    else
        echo "❌ 创建失败: $nickname (HTTP: $http_code)"
        echo "   错误信息: $response_body"
    fi
    
    # 避免请求过快
    sleep 0.1
done

echo "=========================="
echo "✅ 批量创建完成！"
echo "成功创建: $success_count/$total_count 个用户"
echo "统一密码: $PASSWORD"