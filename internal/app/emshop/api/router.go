package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/api/config"
	"emshop/internal/app/emshop/api/controller/goods/v1"
	"emshop/internal/app/emshop/api/controller/order/v1"
	v12 "emshop/internal/app/emshop/api/controller/sms/v1"
	"emshop/internal/app/emshop/api/controller/user/v1"
	"emshop/internal/app/emshop/api/controller/userop/v1"
	"emshop/internal/app/emshop/api/data/rpc"
	"emshop/internal/app/emshop/api/service"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	
	v1 := g.Group("/v1")


	// 用户服务api（C端用户自助功能）
	ugroup := v1.Group("/user")
	// 创建数据工厂
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
		ugroup.PATCH("update", jwtAuth.AuthFunc(), uController.UpdateUser)

		// 注意：管理员功能已迁移到Admin应用
		// ugroup.GET("list", jwtAuth.AuthFunc(), uController.GetUserList) // 已迁移到 /v1/admin/users
		// ugroup.GET("mobile",jwtAuth.AuthFunc(), uController.GetByMobile) // 已迁移到 /v1/admin/users/by-mobile  
		// ugroup.GET("id",jwtAuth.AuthFunc(), uController.GetById)        // 已迁移到 /v1/admin/users/:id
	}


	// 基础服务api
	baseRouter := v1.Group("base")
	{
		smsCtl := v12.NewSmsController(serviceFactory, g.Translator())
		baseRouter.POST("send_sms", smsCtl.SendSms)
		baseRouter.GET("captcha", user.GetCaptcha)
	}

	//商品服务api（C端用户浏览功能）
	goodsRouter := v1.Group("goods")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		goodsRouter.GET("", goodsController.List)                                       //商品列表（前端展示）
		goodsRouter.GET("/:id", goodsController.Detail)                                 //获取商品的详情
		goodsRouter.GET("/:id/stocks", goodsController.Stocks)                          //获取商品的库存
		
		// 注意：管理员功能已迁移到Admin应用
		// goodsRouter.POST("", jwtAuth.AuthFunc(), goodsController.New)          		// 已迁移到 /v1/admin/goods
		// goodsRouter.POST("/sync", jwtAuth.AuthFunc(), goodsController.Sync)    		// 已迁移到 /v1/admin/goods/sync
		// goodsRouter.DELETE("/:id", jwtAuth.AuthFunc(), goodsController.Delete) 		// 已迁移到 /v1/admin/goods/:id
		// goodsRouter.PUT("/:id", jwtAuth.AuthFunc(), goodsController.Update)			// 已迁移到 /v1/admin/goods/:id
		// goodsRouter.PATCH("/:id", jwtAuth.AuthFunc(), goodsController.UpdateStatus)	// 已迁移到 /v1/admin/goods/:id/status
	}

	//商品分类api（C端用户浏览功能）
	categorysRouter := v1.Group("categorys")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		categorysRouter.GET("", goodsController.CategoryList)                                    // 分类列表（前端展示）
		categorysRouter.GET("/:id", goodsController.CategoryDetail)                              // 获取分类详情
		
		// 注意：管理员功能已迁移到Admin应用
		// categorysRouter.DELETE("/:id", jwtAuth.AuthFunc(), goodsController.DeleteCategory)   // 已迁移到 /v1/admin/categories/:id
		// categorysRouter.POST("", jwtAuth.AuthFunc(), goodsController.CreateCategory)         // 已迁移到 /v1/admin/categories
		// categorysRouter.PUT("/:id", jwtAuth.AuthFunc(), goodsController.UpdateCategory)      // 已迁移到 /v1/admin/categories/:id
	}

	//品牌api（C端用户浏览功能）
	brandsRouter := v1.Group("brands")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		brandsRouter.GET("", goodsController.BrandList)                                         // 品牌列表（前端展示）
		
		// 注意：管理员功能已迁移到Admin应用
		// brandsRouter.POST("", jwtAuth.AuthFunc(), goodsController.CreateBrand)              // 已迁移到 /v1/admin/brands
		// brandsRouter.PUT("/:id", jwtAuth.AuthFunc(), goodsController.UpdateBrand)           // 已迁移到 /v1/admin/brands/:id
		// brandsRouter.DELETE("/:id", jwtAuth.AuthFunc(), goodsController.DeleteBrand)        // 已迁移到 /v1/admin/brands/:id
	}

	//轮播图api（C端用户浏览功能）
	bannersRouter := v1.Group("banners")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		bannersRouter.GET("", goodsController.BannerList)                                       // 轮播图列表（前端展示）
		
		// 注意：管理员功能已迁移到Admin应用
		// bannersRouter.POST("", jwtAuth.AuthFunc(), goodsController.CreateBanner)            // 已迁移到 /v1/admin/banners
		// bannersRouter.PUT("/:id", jwtAuth.AuthFunc(), goodsController.UpdateBanner)         // 已迁移到 /v1/admin/banners/:id
		// bannersRouter.DELETE("/:id", jwtAuth.AuthFunc(), goodsController.DeleteBanner)      // 已迁移到 /v1/admin/banners/:id
	}

	//订单管理api
	ordersRouter := v1.Group("orders")
	{
		orderController := order.NewOrderController(serviceFactory, g.Translator())
		ordersRouter.GET("", jwtAuth.AuthFunc(), orderController.OrderList)                    // 订单列表
		ordersRouter.POST("", jwtAuth.AuthFunc(), orderController.CreateOrder)                 // 创建订单
		ordersRouter.GET("/:id", jwtAuth.AuthFunc(), orderController.OrderDetail)              // 订单详情
	}

	//购物车管理api
	cartRouter := v1.Group("shopcarts")
	{
		orderController := order.NewOrderController(serviceFactory, g.Translator())
		cartRouter.GET("", jwtAuth.AuthFunc(), orderController.CartList)                       // 购物车列表
		cartRouter.POST("", jwtAuth.AuthFunc(), orderController.AddToCart)                     // 添加到购物车
		cartRouter.PATCH("/:id", jwtAuth.AuthFunc(), orderController.UpdateCartItem)           // 更新购物车商品
		cartRouter.DELETE("/:id", jwtAuth.AuthFunc(), orderController.DeleteCartItem)          // 删除购物车商品
	}

	//用户收藏管理api
	favRouter := v1.Group("userfavs")
	{
		userOpController := userop.NewUserOpController(serviceFactory, g.Translator())
		favRouter.GET("", jwtAuth.AuthFunc(), userOpController.UserFavList)                    // 收藏列表
		favRouter.POST("", jwtAuth.AuthFunc(), userOpController.CreateUserFav)                 // 添加收藏
		favRouter.DELETE("/:id", jwtAuth.AuthFunc(), userOpController.DeleteUserFav)           // 删除收藏
		favRouter.GET("/:id", jwtAuth.AuthFunc(), userOpController.GetUserFavDetail)           // 查看是否收藏
	}

	//用户地址管理api
	addressRouter := v1.Group("address")
	{
		userOpController := userop.NewUserOpController(serviceFactory, g.Translator())
		addressRouter.GET("", jwtAuth.AuthFunc(), userOpController.GetAddressList)             // 地址列表
		addressRouter.POST("", jwtAuth.AuthFunc(), userOpController.CreateAddress)             // 创建地址
		addressRouter.PUT("/:id", jwtAuth.AuthFunc(), userOpController.UpdateAddress)          // 更新地址
		addressRouter.DELETE("/:id", jwtAuth.AuthFunc(), userOpController.DeleteAddress)       // 删除地址
	}

	//用户留言管理api
	messageRouter := v1.Group("message")
	{
		userOpController := userop.NewUserOpController(serviceFactory, g.Translator())
		messageRouter.GET("", jwtAuth.AuthFunc(), userOpController.MessageList)                // 留言列表
		messageRouter.POST("", jwtAuth.AuthFunc(), userOpController.CreateMessage)             // 创建留言
	}
}
