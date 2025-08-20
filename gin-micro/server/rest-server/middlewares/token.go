package middlewares

import (
	"strings"
	"github.com/gin-gonic/gin"
)

// ExtractToken 通用Token提取工具 - 框架层工具函数
// 支持从多个位置提取token：Authorization header、query参数、cookie
func ExtractToken(c *gin.Context) string {
	// 1. Try to get token from Authorization header
	auth := c.Request.Header.Get("Authorization")
	if auth != "" {
		if token, found := strings.CutPrefix(auth, "Bearer "); found {
			return token
		}
		// If no Bearer prefix, try the token directly
		return auth
	}

	// 2. Try to get token from query parameter
	if token := c.Query("token"); token != "" {
		return token
	}

	// 3. Try to get token from cookie
	if token, err := c.Cookie("jwt"); err == nil && token != "" {
		return token
	}

	return ""
}