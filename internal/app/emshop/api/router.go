package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/api/config"
	"emshop/internal/app/emshop/api/controller/goods/v1"
	v12 "emshop/internal/app/emshop/api/controller/sms/v1"
	"emshop/internal/app/emshop/api/controller/user/v1"
	"emshop/internal/app/emshop/api/data/rpc"
	"emshop/internal/app/emshop/api/service"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	
	v1 := g.Group("/v1")


	// 用户服务api
	ugroup := v1.Group("/user")
	data, err := rpc.GetDataFactoryOr(cfg.Registry)
	if err != nil {
		panic(err)
	}
	// 创建服务工厂
	serviceFactory := service.NewService(data, cfg.Sms, cfg.Jwt)
	uController := user.NewUserController(g.Translator(), serviceFactory)
	{
		ugroup.POST("pwd_login", uController.Login)
		ugroup.POST("register", uController.Register)

		jwtAuth := newJWTAuth(cfg.Jwt)
		ugroup.GET("detail", jwtAuth.AuthFunc(), uController.GetUserDetail)
		ugroup.PATCH("update", jwtAuth.AuthFunc(), uController.GetUserDetail)
	}


	// 基础服务api
	baseRouter := v1.Group("base")
	{
		smsCtl := v12.NewSmsController(serviceFactory, g.Translator())
		baseRouter.POST("send_sms", smsCtl.SendSms)
		baseRouter.GET("captcha", user.GetCaptcha)
	}

	//商品服务api
	goodsRouter := v1.Group("goods")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		goodsRouter.GET("", goodsController.List)
	}
}
