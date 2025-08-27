// cd internal/app/pkg/code && ../../../../tools/codegen/codegen -type=int 
// 可选 -doc -output ../../../../docs/error_code_generated.md


package code

import (
	"fmt"
	"emshop/pkg/errors"
	"net/http"

	"github.com/novalagung/gubrak"
)

type ErrCode struct {
	//错误码
	C int

	//http的状态码
	HTTP int

	//扩展字段
	Ext string

	//引用文档
	Ref string
}

func (e ErrCode) HTTPStatus() int {
	return e.HTTP
}

func (e ErrCode) String() string {
	return e.Ext
}

func (e ErrCode) Reference() string {
	return e.Ref
}

func (e ErrCode) Code() int {
	if e.C == 0 {
		return http.StatusInternalServerError
	}
	return e.C
}

func register(code int, httpStatus int, message string, refs ...string) {
	// 支持常用的HTTP状态码，覆盖微服务架构中的典型场景
	allowedCodes := []int{
		// 成功响应
		200, // OK
		201, // Created
		202, // Accepted
		204, // No Content
		
		// 客户端错误
		400, // Bad Request - 请求参数错误
		401, // Unauthorized - 未认证
		403, // Forbidden - 无权限
		404, // Not Found - 资源不存在
		405, // Method Not Allowed - 方法不允许
		409, // Conflict - 冲突（并发、重复操作）
		422, // Unprocessable Entity - 业务逻辑验证失败
		429, // Too Many Requests - 限流
		
		// 服务器错误
		500, // Internal Server Error - 服务器内部错误
		502, // Bad Gateway - 网关错误
		503, // Service Unavailable - 服务不可用
		504, // Gateway Timeout - 网关超时
	}
	
	found, _ := gubrak.Includes(allowedCodes, httpStatus)
	if !found {
		panic(fmt.Sprintf("HTTP状态码 %d 不在允许的范围内，支持的状态码: %v", httpStatus, allowedCodes))
	}
	var ref string
	if len(refs) > 0 {
		ref = refs[0]
	}
	coder := ErrCode{
		C:    code,
		HTTP: httpStatus,
		Ext:  message,
		Ref:  ref,
	}

	errors.MustRegister(coder)
}

var _ errors.Coder = (*ErrCode)(nil)
