package main

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "emshop/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 连接gRPC服务
	conn, err := grpc.Dial("localhost:8021", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接gRPC服务失败: %v", err)
	}
	defer conn.Close()

	// 创建用户客户端
	client := v1.NewUserClient(conn)

	password := "admin123"
	
	// 用户数据
	users := []struct {
		NickName string
		Mobile   string
	}{
		{"张三", "13800000001"},
		{"李四", "13800000002"},
		{"王五", "13800000003"},
		{"赵六", "13800000004"},
		{"钱七", "13800000005"},
		{"孙八", "13800000006"},
		{"周九", "13800000007"},
		{"吴十", "13800000008"},
		{"郑十一", "13800000009"},
		{"王十二", "13800000010"},
		{"管理员", "13900000001"},
		{"测试用户1", "13900000002"},
		{"测试用户2", "13900000003"},
		{"测试用户3", "13900000004"},
		{"客服", "13900000005"},
	}

	fmt.Printf("开始批量创建 %d 个用户...\n", len(users))
	fmt.Printf("gRPC服务地址: localhost:8021\n")
	fmt.Printf("统一密码: %s\n", password)
	fmt.Println("==========================")

	successCount := 0
	for i, user := range users {
		fmt.Printf("正在创建用户 %d/%d: %s (%s)\n", i+1, len(users), user.NickName, user.Mobile)

		// 创建用户请求
		req := &v1.CreateUserInfo{
			NickName: user.NickName,
			PassWord: password,
			Mobile:   user.Mobile,
		}

		// 调用gRPC接口
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := client.CreateUser(ctx, req)
		cancel()

		if err != nil {
			fmt.Printf("❌ 创建用户失败: %s - %v\n", user.NickName, err)
		} else {
			fmt.Printf("✅ 创建用户成功: %s (ID: %d)\n", user.NickName, resp.Id)
			successCount++
		}

		// 避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("==========================")
	fmt.Printf("✅ 批量创建完成！\n")
	fmt.Printf("成功创建: %d/%d 个用户\n", successCount, len(users))
	fmt.Printf("统一密码: %s\n", password)
}