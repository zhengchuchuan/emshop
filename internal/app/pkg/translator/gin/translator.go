package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"emshop/gin-micro/server/rest-server"
	"net/http"
	"strings"
)

func removeTopStruct(fields map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fields {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}

func HandleValidatorError(c *gin.Context, err error, trans restserver.I18nTranslator) {
	errs, ok := err.(validator.ValidationErrors)
    if !ok {
        // Non-validation binding/parsing error: treat as a bad request.
        c.JSON(http.StatusBadRequest, gin.H{
            "msg": err.Error(),
        })
        return
    }
	
	// 使用新的翻译系统
	errorMessages := make(map[string]string)
	for _, fieldError := range errs {
		field := fieldError.Field()
		tag := fieldError.Tag()
		param := fieldError.Param()
		
		// 尝试翻译错误消息
		translatedMsg := trans.T(tag, map[string]interface{}{
			"Field": field,
			"Param": param,
		})
		
		errorMessages[field] = translatedMsg
	}
	
	c.JSON(http.StatusBadRequest, gin.H{
		"error": errorMessages,
	})
	return
}
