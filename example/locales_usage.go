package main

import (
	"fmt"
	"log"

	"emshop/gin-micro/server/rest-server"
)

// 演示如何使用新的翻译系统
func main() {
	// 示例1：使用内置翻译（不配置翻译文件路径）
	fmt.Println("=== 使用内置翻译 ===")
	server1 := restserver.NewServer(
		restserver.WithTransNames("zh"),
	)
	// 这里只是演示，实际使用时会在Start()方法中初始化
	fmt.Printf("服务器语言设置: %s\n", server1.GetLocale())

	// 示例2：使用外部翻译文件
	fmt.Println("\n=== 使用外部翻译文件 ===")
	server2 := restserver.NewServer(
		restserver.WithTransNames("en"),
		restserver.WithLocalesDir("./locales"), // 指向包含 en.json, zh-CN.json 的目录
	)
	fmt.Printf("服务器语言设置: %s\n", server2.GetLocale())

	fmt.Println("\n=== 翻译文件格式示例 ===")
	fmt.Println("locales/en.json 和 locales/zh-CN.json 应包含以下格式的内容：")
	fmt.Println(`{
  "validation": {
    "required": {
      "other": "{{.Field}} is a required field"
    },
    "mobile": {
      "other": "{{.Field}} is not a valid mobile number"
    }
  },
  "business": {
    "login_failed": {
      "other": "Login failed"
    },
    "captcha_error": {
      "other": "Captcha code is incorrect"
    }
  }
}`)

	log.Println("翻译系统配置完成!")
}