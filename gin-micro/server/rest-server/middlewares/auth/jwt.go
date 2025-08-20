package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
	"emshop/gin-micro/server/rest-server/middlewares"
	"emshop/gin-micro/code"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
)

// AuthzAudience defines the value of jwt audience field.
const AuthzAudience = "emshop.com"

// JWTStrategy defines jwt bearer authentication strategy using golang-jwt/v5.
type JWTStrategy struct {
	jwtTool *middlewares.JWT[jwt.Claims] // Use generic JWT tool
	key     string                       // signing key
}

var _ middlewares.AuthStrategy = &JWTStrategy{}

// NewJWTStrategy create jwt bearer strategy with signing key.
func NewJWTStrategy(signingKey string) JWTStrategy {
	return JWTStrategy{
		jwtTool: middlewares.NewJWT[jwt.Claims](signingKey),
		key:     signingKey,
	}
}

// AuthFunc defines jwt bearer strategy as the gin authentication middleware.
func (j JWTStrategy) AuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		token := middlewares.ExtractToken(c)
		if token == "" {
			core.WriteResponse(
				c,
				errors.WithCode(code.ErrInvalidAuthHeader, "Authorization header is missing or invalid"),
				nil,
			)
			c.Abort()
			return
		}

		// Parse and validate token
		claims := jwt.MapClaims{}
		parsedClaims, err := j.jwtTool.ParseToken(token, claims)
		if err != nil {
			core.WriteResponse(
				c,
				errors.WithCode(code.ErrSignatureInvalid, "Invalid token: %s", err.Error()),
				nil,
			)
			c.Abort()
			return
		}

		// Set claims in context for downstream handlers
		c.Set("jwt_claims", parsedClaims)
		c.Next()
	}
}

