package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Cors 跨域资源共享（CORS）中间件
// 用于设置 HTTP 响应头，允许前端跨域请求
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// 设置允许所有域名跨域
		c.Header("Access-Control-Allow-Origin", "*")
		// 设置允许的请求头
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")
		// 设置允许的请求方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		// 设置暴露的响应头
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		// 允许携带 Cookie
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求（OPTIONS），直接返回 204
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}
