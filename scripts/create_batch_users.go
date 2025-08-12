package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 用户创建请求结构体
type CreateUserRequest struct {
	NickName string `json:"nickName"`
	PassWord string `json:"passWord"`
	Mobile   string `json:"mobile"`
}

// 用户信息响应结构体
type UserInfoResponse struct {
	ID       int32  `json:"id"`
	Mobile   string `json:"mobile"`
	NickName string `json:"nickName"`
	Gender   string `json:"gender"`
	Role     int32  `json:"role"`
}

func main() {
	baseURL := "http://localhost:8022"
	password := "admin123"
	
	// 生成用户数据
	users := []CreateUserRequest{
		{NickName: "张三", PassWord: password, Mobile: "13800000001"},
		{NickName: "李四", PassWord: password, Mobile: "13800000002"},
		{NickName: "王五", PassWord: password, Mobile: "13800000003"},
		{NickName: "赵六", PassWord: password, Mobile: "13800000004"},
		{NickName: "钱七", PassWord: password, Mobile: "13800000005"},
		{NickName: "孙八", PassWord: password, Mobile: "13800000006"},
		{NickName: "周九", PassWord: password, Mobile: "13800000007"},
		{NickName: "吴十", PassWord: password, Mobile: "13800000008"},
		{NickName: "郑十一", PassWord: password, Mobile: "13800000009"},
		{NickName: "王十二", PassWord: password, Mobile: "13800000010"},
		{NickName: "管理员", PassWord: password, Mobile: "13900000001"},
		{NickName: "测试用户1", PassWord: password, Mobile: "13900000002"},
		{NickName: "测试用户2", PassWord: password, Mobile: "13900000003"},
		{NickName: "测试用户3", PassWord: password, Mobile: "13900000004"},
		{NickName: "客服", PassWord: password, Mobile: "13900000005"},
	}

	fmt.Printf("开始批量创建 %d 个用户...\n", len(users))
	
	successCount := 0
	for i, user := range users {
		fmt.Printf("正在创建用户 %d/%d: %s (%s)\n", i+1, len(users), user.NickName, user.Mobile)
		
		if err := createUser(baseURL, user); err != nil {
			fmt.Printf("❌ 创建用户失败: %s - %v\n", user.NickName, err)
		} else {
			fmt.Printf("✅ 创建用户成功: %s\n", user.NickName)
			successCount++
		}
		
		// 避免请求过快，稍微延迟
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("\n✅ 批量创建完成！成功创建 %d/%d 个用户\n", successCount, len(users))
	fmt.Printf("密码统一为: %s\n", password)
}

func createUser(baseURL string, user CreateUserRequest) error {
	// 序列化请求数据
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 创建HTTP请求
	url := baseURL + "/v1/user/create"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var userResp UserInfoResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return fmt.Errorf("解析响应失败: %v, 响应: %s", err, string(body))
	}

	return nil
}