package middleware

import (
	"github.com/gin-gonic/gin"
	"emshop/internal/app/pkg/options"
	"emshop/gin-micro/server/rest-server/middlewares"
	"emshop/gin-micro/code"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	jwtpkg "emshop/internal/app/pkg/jwt"
	"emshop/pkg/log"
)

// 业务层JWT认证策略 - 实现框架层的 AuthStrategy 接口
type businessJWTStrategy struct {
	jwtTool *jwtpkg.EmshopJWT
}

var _ middlewares.AuthStrategy = &businessJWTStrategy{}

// newBusinessJWTStrategy 创建业务JWT策略
func newBusinessJWTStrategy(signingKey string) *businessJWTStrategy {
	return &businessJWTStrategy{
		jwtTool: jwtpkg.NewEmshopJWT(signingKey),
	}
}

// AuthFunc 实现基础业务JWT认证中间件
func (b *businessJWTStrategy) AuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用框架层的通用token提取工具
		token := middlewares.ExtractToken(c)
		if token == "" {
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "Authorization token is missing"), nil)
			c.Abort()
			return
		}

		// 使用业务JWT工具解析
		claims, err := b.jwtTool.ParseToken(token)
		if err != nil {
			core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, "Invalid token: %s", err.Error()), nil)
			c.Abort()
			return
		}

		// 设置业务上下文
		c.Set(jwtpkg.KeyUserID, int(claims.ID))
		c.Set(jwtpkg.KeyUserRole, int(claims.AuthorityId))
		c.Next()
	}
}

// 业务层管理员认证策略
type adminAuthStrategy struct {
	jwtTool *jwtpkg.EmshopJWT
}

var _ middlewares.AuthStrategy = &adminAuthStrategy{}

// newAdminAuthStrategy 创建管理员认证策略
func newAdminAuthStrategy(signingKey string) *adminAuthStrategy {
	return &adminAuthStrategy{
		jwtTool: jwtpkg.NewEmshopJWT(signingKey),
	}
}

// AuthFunc 实现管理员权限验证中间件
func (a *adminAuthStrategy) AuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用框架层的通用token提取工具
		token := middlewares.ExtractToken(c)
		if token == "" {
			log.Warn("Admin access denied: missing token")
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "管理员认证失败：缺少认证令牌"), nil)
			c.Abort()
			return
		}

		// 使用业务JWT工具解析
		claims, err := a.jwtTool.ParseToken(token)
		if err != nil {
			log.Warnf("Admin access denied: invalid token - %v", err)
			core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, "管理员认证失败：无效的认证令牌"), nil)
			c.Abort()
			return
		}

		// 验证管理员权限：角色必须 >= 管理员
		userRole := int(claims.AuthorityId)
		if userRole < jwtpkg.RoleAdmin {
			log.Warnf("Admin access denied: insufficient privileges - userID: %d, role: %d", claims.ID, userRole)
			core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "权限不足：需要管理员权限"), nil)
			c.Abort()
			return
		}

		// 记录管理员访问日志
		log.Infof("Admin access granted - userID: %d, role: %d, path: %s", claims.ID, userRole, c.Request.URL.Path)

		// 设置业务上下文
		c.Set(jwtpkg.KeyUserID, int(claims.ID))
		c.Set(jwtpkg.KeyUserRole, userRole)
		c.Next()
	}
}

// 业务层超级管理员认证策略
type superAdminAuthStrategy struct {
	jwtTool *jwtpkg.EmshopJWT
}

var _ middlewares.AuthStrategy = &superAdminAuthStrategy{}

// newSuperAdminAuthStrategy 创建超级管理员认证策略
func newSuperAdminAuthStrategy(signingKey string) *superAdminAuthStrategy {
	return &superAdminAuthStrategy{
		jwtTool: jwtpkg.NewEmshopJWT(signingKey),
	}
}

// AuthFunc 实现超级管理员权限验证中间件
func (s *superAdminAuthStrategy) AuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用框架层的通用token提取工具
		token := middlewares.ExtractToken(c)
		if token == "" {
			log.Warn("Super admin access denied: missing token")
			core.WriteResponse(c, errors.WithCode(code.ErrInvalidAuthHeader, "超级管理员认证失败：缺少认证令牌"), nil)
			c.Abort()
			return
		}

		// 使用业务JWT工具解析
		claims, err := s.jwtTool.ParseToken(token)
		if err != nil {
			log.Warnf("Super admin access denied: invalid token - %v", err)
			core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, "超级管理员认证失败：无效的认证令牌"), nil)
			c.Abort()
			return
		}

		// 验证超级管理员权限：角色必须 >= 超级管理员
		userRole := int(claims.AuthorityId)
		if userRole < jwtpkg.RoleSuper {
			log.Warnf("Super admin access denied: insufficient privileges - userID: %d, role: %d", claims.ID, userRole)
			core.WriteResponse(c, errors.WithCode(code.ErrPermissionDenied, "权限不足：需要超级管理员权限"), nil)
			c.Abort()
			return
		}

		// 记录超级管理员访问日志
		log.Infof("Super admin access granted - userID: %d, role: %d, path: %s", claims.ID, userRole, c.Request.URL.Path)

		// 设置业务上下文
		c.Set(jwtpkg.KeyUserID, int(claims.ID))
		c.Set(jwtpkg.KeyUserRole, userRole)
		c.Next()
	}
}

// 公开的中间件构造函数 - 基于框架层策略模式，业务层实现

// JWTAuth 创建基础JWT认证中间件 - 基于框架层策略模式
func JWTAuth(opts *options.JwtOptions) gin.HandlerFunc {
	// 业务层实现策略，使用框架层的策略模式
	strategy := newBusinessJWTStrategy(opts.Key)
	operator := &middlewares.AuthOperator{}
	operator.SetStrategy(strategy)
	return operator.AuthFunc()
}

// AdminAuth 创建管理员权限验证中间件 - 基于框架层策略模式
func AdminAuth(opts *options.JwtOptions) gin.HandlerFunc {
	// 业务层实现策略，使用框架层的策略模式
	strategy := newAdminAuthStrategy(opts.Key)
	operator := &middlewares.AuthOperator{}
	operator.SetStrategy(strategy)
	return operator.AuthFunc()
}

// SuperAdminAuth 创建超级管理员权限验证中间件 - 基于框架层策略模式
func SuperAdminAuth(opts *options.JwtOptions) gin.HandlerFunc {
	// 业务层实现策略，使用框架层的策略模式
	strategy := newSuperAdminAuthStrategy(opts.Key)
	operator := &middlewares.AuthOperator{}
	operator.SetStrategy(strategy)
	return operator.AuthFunc()
}

// 业务层辅助函数 - 供其他业务逻辑使用

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(ctx *gin.Context) (int, bool) {
	userID, exists := ctx.Get(jwtpkg.KeyUserID)
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

// GetUserRoleFromContext 从上下文中获取用户角色
func GetUserRoleFromContext(ctx *gin.Context) (int, bool) {
	userRole, exists := ctx.Get(jwtpkg.KeyUserRole)
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
	return userRole >= jwtpkg.RoleAdmin
}

// CheckSuperAdminPermission 检查超级管理员权限
func CheckSuperAdminPermission(ctx *gin.Context) bool {
	userRole, exists := GetUserRoleFromContext(ctx)
	if !exists {
		return false
	}
	return userRole >= jwtpkg.RoleSuper
}

// ExtractToken 代理到框架层的通用Token提取工具，保持向后兼容
func ExtractToken(c *gin.Context) string {
	return middlewares.ExtractToken(c)
}