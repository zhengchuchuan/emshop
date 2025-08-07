package restserver

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// initTrans 初始化翻译器
// 功能：为验证器设置多语言支持，将验证错误消息翻译为指定语言
// 参数：locale - 语言环境标识符（"zh"=中文, "en"=英文）
// 返回：error - 初始化过程中的错误信息
//
// 工作原理：
// 1. 获取gin框架的validator验证引擎
// 2. 注册自定义字段名函数，使用JSON标签作为字段名
// 3. 创建go-i18n/v2的Bundle和Localizer实例
// 4. 从配置文件或内置消息加载翻译内容
// 5. 根据locale参数选择对应的语言环境
func (s *Server) initTrans(locale string) (err error) {
	// 修改gin框架中的validator引擎属性, 实现定制
	// 通过类型断言获取validator.Validate实例，以便进行自定义配置
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		
		// 注册一个获取json标签的自定义方法
		// 作用：当验证失败时，错误消息中显示的字段名使用JSON标签而不是结构体字段名
		// 例如：User{Name string `json:"username"`} 验证失败时显示"username"而不是"Name"
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			// 获取结构体字段的json标签，如果有多个选项用逗号分隔，只取第一个
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			// 如果json标签为"-"，表示该字段在JSON序列化时被忽略，返回空字符串
			if name == "-" {
				return ""
			}
			return name
		})

		// 创建go-i18n/v2的Bundle实例，默认语言为英文
		// Bundle用于管理所有支持的语言和消息
		bundle := i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
		
		// 加载翻译消息：优先从文件加载，如果文件不存在则使用内置消息
		if err := s.loadTranslationMessages(bundle); err != nil {
			return fmt.Errorf("failed to load translation messages: %w", err)
		}
		
		// 根据传入的locale参数创建对应的Localizer实例
		// Localizer负责根据语言环境获取翻译后的消息
		var acceptLanguage string
		switch locale {
		case "zh":
			acceptLanguage = "zh-CN,zh,en"
		case "en":
			acceptLanguage = "en,zh-CN,zh"
		default:
			acceptLanguage = "en,zh-CN,zh"
		}
		
		s.localizer = i18n.NewLocalizer(bundle, acceptLanguage)
		s.locale = locale

		// 注册自定义的验证错误翻译函数
		// 这个函数会在验证失败时被调用，用于翻译错误消息
		if err := s.registerValidationTranslations(v); err != nil {
			return fmt.Errorf("failed to register validation translations: %w", err)
		}
		return
	}
	return
}

// loadTranslationMessages 加载翻译消息
// 优先从指定的翻译文件目录加载，如果不存在则使用内置翻译消息
func (s *Server) loadTranslationMessages(bundle *i18n.Bundle) error {
	if s.localesDir != "" {
		// 从文件加载翻译
		return s.loadTranslationFromFiles(bundle)
	}
	
	// 使用内置翻译消息
	return s.addBuiltinValidationMessages(bundle)
}

// loadTranslationFromFiles 从文件系统加载翻译文件
func (s *Server) loadTranslationFromFiles(bundle *i18n.Bundle) error {
	// 支持的语言文件
	languages := []string{"en.json", "zh-CN.json"}
	
	loadedFiles := 0
	for _, langFile := range languages {
		filePath := filepath.Join(s.localesDir, langFile)
		
		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue // 跳过不存在的文件
		}
		
		// 加载消息文件
		_, err := bundle.LoadMessageFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load translation file %s: %w", filePath, err)
		}
		
		loadedFiles++
	}
	
	// 如果没有加载任何文件，使用内置翻译作为备用
	if loadedFiles == 0 {
		return s.addBuiltinValidationMessages(bundle)
	}
	
	return nil
}

// addBuiltinValidationMessages 添加内置的验证错误消息到i18n bundle
// 当没有配置翻译文件时使用内置翻译
func (s *Server) addBuiltinValidationMessages(bundle *i18n.Bundle) error {
	// 添加英文验证消息
	englishMessages := []*i18n.Message{
		{ID: "required", Other: "{0} is a required field"},
		{ID: "min", Other: "{0} must be at least {1}"},
		{ID: "max", Other: "{0} must be at most {1}"},
		{ID: "len", Other: "{0} must be {1} characters long"},
		{ID: "email", Other: "{0} must be a valid email address"},
		{ID: "oneof", Other: "{0} must be one of [{1}]"},
		{ID: "numeric", Other: "{0} must be a valid numeric value"},
		{ID: "alphanum", Other: "{0} can only contain alphanumeric characters"},
		{ID: "mobile", Other: "{0} is not a valid mobile number"},
	}
	
	for _, msg := range englishMessages {
		if err := bundle.AddMessages(language.English, msg); err != nil {
			return fmt.Errorf("failed to add English message %s: %w", msg.ID, err)
		}
	}
	
	// 添加中文验证消息
	chineseMessages := []*i18n.Message{
		{ID: "required", Other: "{0}为必填字段"},
		{ID: "min", Other: "{0}长度必须至少为{1}"},
		{ID: "max", Other: "{0}长度不能超过{1}"},
		{ID: "len", Other: "{0}长度必须为{1}"},
		{ID: "email", Other: "{0}必须是一个有效的邮箱地址"},
		{ID: "oneof", Other: "{0}必须是[{1}]中的一个"},
		{ID: "numeric", Other: "{0}必须是一个有效的数值"},
		{ID: "alphanum", Other: "{0}只能包含字母和数字"},
		{ID: "mobile", Other: "{0}不是一个有效的手机号码"},
	}
	
	// 获取中文语言标识
	chineseLang, err := language.Parse("zh-CN")
	if err != nil {
		return fmt.Errorf("failed to parse Chinese language: %w", err)
	}
	
	for _, msg := range chineseMessages {
		if err := bundle.AddMessages(chineseLang, msg); err != nil {
			return fmt.Errorf("failed to add Chinese message %s: %w", msg.ID, err)
		}
	}
	
	return nil
}

// registerValidationTranslations 注册验证翻译函数
func (s *Server) registerValidationTranslations(_ *validator.Validate) error {
	// 由于go-i18n/v2的设计不同，我们不在这里预先注册翻译规则
	// 而是在Translate方法中动态翻译消息
	return nil
}

// Translate 翻译验证错误消息
func (s *Server) Translate(errs validator.ValidationErrors) []string {
	var messages []string
	for _, err := range errs {
		// 构造翻译消息ID：validation.{tag}
		messageID := fmt.Sprintf("validation.%s", err.Tag())
		
		config := &i18n.LocalizeConfig{
			MessageID: messageID,
			TemplateData: map[string]interface{}{
				"Field": err.Field(),
				"Param": err.Param(),
			},
			DefaultMessage: &i18n.Message{
				ID:    messageID,
				Other: fmt.Sprintf("{{.Field}} validation failed on tag '%s'", err.Tag()),
			},
		}
		
		translated, translationErr := s.localizer.Localize(config)
		if translationErr != nil {
			translated = fmt.Sprintf("%s validation failed on tag '%s'", err.Field(), err.Tag())
		}
		
		messages = append(messages, translated)
	}
	
	return messages
}

// TranslateBusiness 翻译业务错误消息
func (s *Server) TranslateBusiness(key string, templateData ...map[string]interface{}) string {
	// 构造翻译消息ID：business.{key}
	messageID := fmt.Sprintf("business.%s", key)
	
	config := &i18n.LocalizeConfig{
		MessageID: messageID,
		DefaultMessage: &i18n.Message{
			ID:    messageID,
			Other: key, // 翻译失败时返回原始key
		},
	}
	
	// 如果提供了模板数据，使用第一个
	if len(templateData) > 0 {
		config.TemplateData = templateData[0]
	}
	
	result, err := s.localizer.Localize(config)
	if err != nil {
		return key // 翻译失败时返回原始key
	}
	
	return result
}
