# Go-i18n/v2 翻译系统使用说明

## 概述

项目已从 `universal-translator` 迁移到 `github.com/nicksnyder/go-i18n/v2/i18n`，支持从外部配置文件加载翻译内容。

## 特性

1. **配置文件驱动**：翻译内容存储在 JSON 配置文件中，便于管理和维护
2. **多语言支持**：支持中文 (zh-CN) 和英文 (en)
3. **灵活配置**：支持自定义翻译文件路径，也可使用内置翻译
4. **向下兼容**：保持与现有控制器代码的兼容性
5. **业务消息翻译**：除了验证错误，还支持业务错误消息的翻译

## 使用方法

### 1. 基本配置

```go
// 使用内置翻译
server := restserver.NewServer(
    restserver.WithTransNames("zh"),
)

// 使用外部翻译文件
server := restserver.NewServer(
    restserver.WithTransNames("zh"),
    restserver.WithLocalesDir("./locales"), // 翻译文件目录
)
```

### 2. 翻译文件格式

创建 `locales/en.json` 和 `locales/zh-CN.json` 文件：

```json
{
  "validation": {
    "required": {
      "other": "{{.Field}} is a required field"
    },
    "min": {
      "other": "{{.Field}} must be at least {{.Param}} characters long"
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
}
```

### 3. 在控制器中使用

#### 验证错误翻译（自动）
```go
func (us *userServer) Login(ctx *gin.Context) {
    if err := ctx.ShouldBind(&form); err != nil {
        gin2.HandleValidatorError(ctx, err, us.trans) // 自动翻译验证错误
        return
    }
}
```

#### 业务错误翻译
```go
func (us *userServer) Login(ctx *gin.Context) {
    if loginFailed {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "msg": us.trans.T("business.login_failed"), // 翻译业务错误
        })
        return
    }
}
```

## 支持的翻译类型

### 验证错误消息
- `validation.required` - 必填字段验证
- `validation.min` - 最小长度验证
- `validation.max` - 最大长度验证
- `validation.email` - 邮箱格式验证
- `validation.mobile` - 手机号格式验证
- 等...

### 业务错误消息
- `business.login_failed` - 登录失败
- `business.captcha_error` - 验证码错误
- `business.sms_send_failed` - 短信发送失败
- `business.user_not_found` - 用户不存在
- 等...

## 模板变量

翻译消息支持模板变量：
- `{{.Field}}` - 字段名称
- `{{.Param}}` - 验证参数
- 其他自定义变量

## 错误处理

1. 如果翻译文件不存在，系统会回退到内置翻译
2. 如果翻译消息不存在，返回原始消息键名
3. 翻译文件格式错误会在启动时报告

## 最佳实践

1. **统一消息管理**：所有用户可见的消息都应通过翻译系统
2. **命名规范**：使用点号分隔的层级结构 (`category.message_key`)
3. **回退策略**：为所有翻译消息提供英文版本作为回退
4. **版本控制**：将翻译文件纳入版本控制

## 示例项目结构

```
project/
├── locales/
│   ├── en.json          # 英文翻译
│   └── zh-CN.json       # 中文翻译
├── gin-micro/
│   └── server/rest-server/
│       ├── translator.go # 翻译器实现
│       └── server.go     # 服务器配置
└── main.go
```