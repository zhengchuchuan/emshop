package core

import (
	"fmt"
	"net/http"

	"emshop/pkg/errors"

	"github.com/gin-gonic/gin"
)

 // ErrResponse 定义了发生错误时的返回消息结构。
 // 如果 Reference 不存在，则不会返回该字段。
 // swagger:model
type ErrResponse struct {
	 // Code 业务错误码。
		Code int `json:"code"`

	 // Message 包含该消息的详细信息。
	 // 此消息适合对外暴露。
		Message string `json:"msg"`

	Detail string `json:"detail"`

	 // Reference 返回可能有助于解决该错误的参考文档。
		Reference string `json:"reference,omitempty"`
}

 // WriteResponse 将错误或响应数据写入 HTTP 响应体。
 // 它使用 errors.ParseCoder 将任意 error 解析为 errors.Coder。
 // errors.Coder 包含错误码、可安全暴露的错误信息和 HTTP 状态码。
func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err != nil {
		errStr := fmt.Sprintf("%#+v", err)
		coder := errors.ParseCoder(err)
		c.JSON(coder.HTTPStatus(), ErrResponse{
			Code:      coder.Code(),
			Message:   coder.String(),
			Detail:    errStr,
			Reference: coder.Reference(),
		})

		return
	}

	c.JSON(http.StatusOK, data)
}
