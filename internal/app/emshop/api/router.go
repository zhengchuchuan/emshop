package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/api/config"
	"emshop/internal/app/emshop/api/controller/user/v1"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	
	v1 := g.Group("/v1")
	ugroup := v1.Group("/user")

	data, err := rpc.GetDataFactoryOr(cfg.Registry)
	if err != nil {
		panic(err)
	}

	// //原来的过程其实很复杂
	serviceFactory := service.NewService(data, cfg.Sms, cfg.Jwt)
	uController := user.NewUserController(g.Translator(), serviceFactory)
	{
		ugroup.POST("pwd_login", uController.Login)
		ugroup.POST("register", uController.Register)

	// 	jwtAuth := newJWTAuth(cfg.Jwt)
	// 	ugroup.GET("detail", jwtAuth.AuthFunc(), uController.GetUserDetail)
	// 	ugroup.PATCH("update", jwtAuth.AuthFunc(), uController.GetUserDetail)
	// }

	// baseRouter := v1.Group("base")
	// {
	// 	smsCtl := v12.NewSmsController(serviceFactory, g.Translator())
	// 	baseRouter.POST("send_sms", smsCtl.SendSms)
	// 	baseRouter.GET("captcha", user.GetCaptcha)
	// }

	// //商品相关的api
	// goodsRouter := v1.Group("goods")
	// {
	// 	goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
	// 	goodsRouter.GET("", goodsController.List)
	}
}
