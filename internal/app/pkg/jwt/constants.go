package jwt

// Context keys for storing user information in gin.Context
const (
	KeyUserID   = "userID"
	KeyUserRole = "userRole"
	KeyUsername = "username"
	KeyUserIP   = "userIP"
)

// Role constants for emshop business
const (
	RoleUser  = 1 // 普通用户
	RoleAdmin = 2 // 管理员
	RoleSuper = 3 // 超级管理员
)

// Permission constants for emshop business (扩展用)
const (
	PermissionUserRead    = "user:read"
	PermissionUserWrite   = "user:write"
	PermissionOrderRead   = "order:read"
	PermissionOrderWrite  = "order:write"
	PermissionGoodsRead   = "goods:read"
	PermissionGoodsWrite  = "goods:write"
)

// JWT issuer for emshop
const (
	IssuerEmshopAPI   = "emshop-api"
	IssuerEmshopAdmin = "emshop-admin"
)