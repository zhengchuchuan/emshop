package middlewares

import "github.com/gin-gonic/gin"

const (
	UsernameKey = "username"
	KeyUserID   = "userid"
	UserIP      = "ip"
)

// 为每个请求添加上下文, django
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		//TODO 扩展
		c.Next()
	}
}
