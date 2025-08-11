package admin

import (
	"strings"
	
	"github.com/gin-gonic/gin"
	"emshop/internal/app/pkg/options"
	"emshop/gin-micro/server/rest-server/middlewares"
	"emshop/pkg/common/core"
	"emshop/pkg/errors"
	"emshop/gin-micro/code"
)

func newJWTAuth(opts *options.JwtOptions) middlewares.AuthStrategy {
	return &customJWTAuth{opts: opts}
}

type customJWTAuth struct {
	opts *options.JwtOptions
}

func (c *customJWTAuth) AuthFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := extractToken(ctx)
		if token == "" {
			core.WriteResponse(ctx, errors.WithCode(code.ErrSignatureInvalid, "Authorization token is missing"), nil)
			ctx.Abort()
			return
		}

		j := middlewares.NewJWT(c.opts.Key)
		claims, err := j.ParseToken(token)
		if err != nil {
			core.WriteResponse(ctx, errors.WithCode(code.ErrSignatureInvalid, "Invalid token: %s", err.Error()), nil)
			ctx.Abort()
			return
		}

		ctx.Set(middlewares.KeyUserID, float64(claims.ID))
		ctx.Next()
	}
}

func extractToken(c *gin.Context) string {
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
