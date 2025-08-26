package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// Validator 验证器接口
type Validator interface {
	Validate(interface{}) error
	ValidateStruct(interface{}) []string
}

// CustomValidator 自定义验证器
type CustomValidator struct {
	validator  *validator.Validate
	translator ut.Translator
}

// NewValidator 创建新的验证器实例
func NewValidator(locale string) *CustomValidator {
	v := validator.New()
	
	// 创建翻译器
	en := en.New()
	zh := zh.New()
	uni := ut.New(zh, en, zh)
	
	var trans ut.Translator
	var ok bool
	
	switch locale {
	case "en", "en_US":
		trans, ok = uni.GetTranslator("en")
		if ok {
			en_translations.RegisterDefaultTranslations(v, trans)
		}
	default:
		trans, ok = uni.GetTranslator("zh")
		if ok {
			zh_translations.RegisterDefaultTranslations(v, trans)
		}
	}
	
	// 注册自定义字段名
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &CustomValidator{
		validator:  v,
		translator: trans,
	}
}

// Validate 验证结构体
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return cv.translateError(err)
	}
	return nil
}

// ValidateStruct 验证结构体并返回错误列表
func (cv *CustomValidator) ValidateStruct(s interface{}) []string {
	var errors []string
	
	err := cv.validator.Struct(s)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors {
			errors = append(errors, e.Translate(cv.translator))
		}
	}
	
	return errors
}

// translateError 翻译错误信息
func (cv *CustomValidator) translateError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}
	
	var messages []string
	for _, e := range validationErrors {
		messages = append(messages, e.Translate(cv.translator))
	}
	
	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

// ValidateVar 验证变量
func (cv *CustomValidator) ValidateVar(field interface{}, tag string) error {
	if err := cv.validator.Var(field, tag); err != nil {
		return cv.translateError(err)
	}
	return nil
}

// RegisterValidation 注册自定义验证规则
func (cv *CustomValidator) RegisterValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// RegisterTranslation 注册翻译
func (cv *CustomValidator) RegisterTranslation(tag string, text string, translation string) error {
	cv.validator.RegisterTranslation(tag, cv.translator, func(ut ut.Translator) error {
		return ut.Add(tag, text, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(tag, fe.Field(), translation)
		return t
	})
	return nil
}

// 全局验证器实例
var defaultValidator = NewValidator("zh")

// Validate 全局验证函数
func Validate(i interface{}) error {
	return defaultValidator.Validate(i)
}

// ValidateStruct 全局验证结构体函数
func ValidateStruct(s interface{}) []string {
	return defaultValidator.ValidateStruct(s)
}

// ValidateVar 全局验证变量函数
func ValidateVar(field interface{}, tag string) error {
	return defaultValidator.ValidateVar(field, tag)
}