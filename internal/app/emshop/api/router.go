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
	jwtAuth := newJWTAuth(cfg.Jwt)
	uController := user.NewUserController(g.Translator(), serviceFactory)
	{
		ugroup.POST("pwd_login", uController.Login)
		ugroup.POST("register", uController.Register)

		ugroup.GET("detail", jwtAuth.AuthFunc(), uController.GetUserDetail)
		ugroup.PATCH("update", jwtAuth.AuthFunc(), uController.GetUserDetail)

		ugroup.GET("list", jwtAuth.AuthFunc(), uController.GetUserList) // GET /v1/user/list?pn=页码&psize=每页数量
		ugroup.GET("mobile",jwtAuth.AuthFunc(), uController.GetByMobile)  // GET /v1/user/mobile?mobile=手机号
  		ugroup.GET("id",jwtAuth.AuthFunc(), uController.GetById)          // GET /v1/user/id?id=用户ID
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
		goodsRouter.GET("", goodsController.List)                                                            //商品列表
		// TODO: 待实现以下接口
		goodsRouter.POST("", jwtAuth.AuthFunc(), goodsController.New)          //该接口需要管理员权限
		goodsRouter.POST("/sync", jwtAuth.AuthFunc(), goodsController.Sync)    //数据同步接口，需要管理员权限
		// goodsRouter.GET("/:id", goodsController.Detail)                                                      //获取商品的详情
		// goodsRouter.DELETE("/:id", jwtAuth.AuthFunc(), goodsController.Delete) //删除商品
		// goodsRouter.GET("/:id/stocks", goodsController.Stocks)                                               //获取商品的库存
		// goodsRouter.PUT("/:id", jwtAuth.AuthFunc(), goodsController.Update)
		// goodsRouter.PATCH("/:id", jwtAuth.AuthFunc(), goodsController.UpdateStatus)
	}
}
