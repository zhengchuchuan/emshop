package admin

import (
	"github.com/gin-gonic/gin"
	"emshop/internal/app/pkg/options"
	"emshop/gin-micro/server/rest-server/middlewares"
	"emshop/gin-micro/server/rest-server/middlewares/auth"

	ginjwt "github.com/appleboy/gin-jwt/v2"
)

func newJWTAuth(opts *options.JwtOptions) middlewares.AuthStrategy {
	gjwt, _ := ginjwt.New(&ginjwt.GinJWTMiddleware{
		Realm:            opts.Realm,
		SigningAlgorithm: "HS256",
		Key:              []byte(opts.Key),
		Timeout:          opts.Timeout,
		MaxRefresh:       opts.MaxRefresh,
		LogoutResponse: func(c *gin.Context, code int) {
			c.JSON(code, nil)
		},
		IdentityHandler: claimHandlerFun,
		IdentityKey:     middlewares.KeyUserID,
		TokenLookup:     "header: Authorization:, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
	})
	return auth.NewJWTStrategy(*gjwt)
}

func claimHandlerFun(c *gin.Context) interface{} {
	claims := ginjwt.ExtractClaims(c)
	c.Set(middlewares.KeyUserID, claims[middlewares.KeyUserID])
	return claims[ginjwt.IdentityKey]
}
