package middleware

import (
	"strings"
	
	"github.com/gin-gonic/gin"
	"emshop/internal/app/pkg/options"
	"emshop/gin-micro/server/rest-server/middlewares"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	"emshop/gin-micro/code"
	"emshop/pkg/log"
)

// 角色常量定义
const (
	RoleUser  = 1 // 普通用户
	RoleAdmin = 2 // 管理员
	RoleSuper = 3 // 超级管理员（预留）
)

// 权限常量定义（扩展用）
const (
	PermissionUserRead    = "user:read"
	PermissionUserWrite   = "user:write"
	PermissionOrderRead   = "order:read"
	PermissionOrderWrite  = "order:write"
	PermissionGoodsRead   = "goods:read"
	PermissionGoodsWrite  = "goods:write"
)

// 中间件键名常量
const (
	KeyUserID   = "userID"
	KeyUserRole = "userRole"
)

// JWTAuth 创建基础JWT认证中间件
func JWTAuth(opts *options.JwtOptions) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if token == "" {
			core.WriteResponse(ctx, errors.WithCode(code.ErrSignatureInvalid, "Authorization token is missing"), nil)
			ctx.Abort()
			return
		}

		j := middlewares.NewJWT(opts.Key)
		claims, err := j.ParseToken(token)
		if err != nil {
			core.WriteResponse(ctx, errors.WithCode(code.ErrSignatureInvalid, "Invalid token: %s", err.Error()), nil)
			ctx.Abort()
			return
		}

		// 设置基础用户信息到上下文
		ctx.Set(KeyUserID, int(claims.ID))
		ctx.Set(KeyUserRole, int(claims.AuthorityId))
		ctx.Next()
	})
}

// AdminAuth 创建管理员权限验证中间件
func AdminAuth(opts *options.JwtOptions) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if token == "" {
			log.Warn("Admin access denied: missing token")
			core.WriteResponse(ctx, errors.WithCode(code.ErrSignatureInvalid, "管理员认证失败：缺少认证令牌"), nil)
			ctx.Abort()
			return
		}

		j := middlewares.NewJWT(opts.Key)
		claims, err := j.ParseToken(token)
		if err != nil {
			log.Warnf("Admin access denied: invalid token - %v", err)
			core.WriteResponse(ctx, errors.WithCode(code.ErrSignatureInvalid, "管理员认证失败：无效的认证令牌"), nil)
			ctx.Abort()
			return
		}

		// 验证管理员权限：角色必须 >= 管理员
		userRole := int(claims.AuthorityId)
		if userRole < RoleAdmin {
			log.Warnf("Admin access denied: insufficient privileges - userID: %d, role: %d", claims.ID, userRole)
			core.WriteResponse(ctx, errors.WithCode(code.ErrPermissionDenied, "权限不足：需要管理员权限"), nil)
			ctx.Abort()
			return
		}

		// 记录管理员访问日志
		log.Infof("Admin access granted - userID: %d, role: %d, path: %s", claims.ID, userRole, ctx.Request.URL.Path)

		// 设置用户信息到上下文
		ctx.Set(KeyUserID, int(claims.ID))
		ctx.Set(KeyUserRole, userRole)
		ctx.Next()
	})
}

// ExtractToken 从请求中提取JWT token
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

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(ctx *gin.Context) (int, bool) {
	userID, exists := ctx.Get(KeyUserID)
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

// GetUserRoleFromContext 从上下文中获取用户角色
func GetUserRoleFromContext(ctx *gin.Context) (int, bool) {
	userRole, exists := ctx.Get(KeyUserRole)
	if !exists {
		return 0, false
	}
	return userRole.(int), true
}

// CheckAdminPermission 检查管理员权限
func CheckAdminPermission(ctx *gin.Context) bool {
	userRole, exists := GetUserRoleFromContext(ctx)
	if !exists {
		return false
	}
	return userRole >= RoleAdmin
}

// CheckSuperAdminPermission 检查超级管理员权限
func CheckSuperAdminPermission(ctx *gin.Context) bool {
	userRole, exists := GetUserRoleFromContext(ctx)
	if !exists {
		return false
	}
	return userRole >= RoleSuper
}