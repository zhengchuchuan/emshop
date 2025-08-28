package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/emshop/config"
	"emshop/internal/app/api/emshop/controller/coupon/v1"
	"emshop/internal/app/api/emshop/controller/goods/v1"
	"emshop/internal/app/api/emshop/controller/logistics/v1"
	"emshop/internal/app/api/emshop/controller/order/v1"
	"emshop/internal/app/api/emshop/controller/payment/v1"
	v12 "emshop/internal/app/api/emshop/controller/sms/v1"
	"emshop/internal/app/api/emshop/controller/user/v1"
	"emshop/internal/app/api/emshop/controller/userop/v1"
	"emshop/internal/app/api/emshop/data/rpc"
	"emshop/internal/app/api/emshop/service"
	"emshop/internal/app/pkg/middleware"
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

	jwtAuth := middleware.JWTAuth(cfg.Jwt)

	// 基础服务api
	baseRouter := v1.Group("base")
	{
		smsCtl := v12.NewSmsController(serviceFactory, g.Translator())
		baseRouter.POST("send_sms", smsCtl.SendSms)
		baseRouter.GET("captcha", user.GetCaptcha)
	}
	// 用户服务
	uController := user.NewUserController(g.Translator(), serviceFactory)
	{
		ugroup.POST("pwd_login", uController.Login)
		ugroup.POST("register", uController.Register)

		ugroup.GET("detail", jwtAuth, uController.GetUserDetail)
		ugroup.PATCH("update", jwtAuth, uController.UpdateUser)

	}



	//商品服务api（C端用户浏览功能）
	goodsRouter := v1.Group("goods")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		goodsRouter.GET("", goodsController.List)              //商品列表（前端展示）
		goodsRouter.GET("/:id", goodsController.Detail)        //获取商品的详情
		goodsRouter.GET("/:id/stocks", goodsController.Stocks) //获取商品的库存

	}

	//商品分类api（C端用户浏览功能）
	categorysRouter := v1.Group("categorys")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		categorysRouter.GET("", goodsController.CategoryList)       // 分类列表（前端展示）
		categorysRouter.GET("/:id", goodsController.CategoryDetail) // 获取分类详情

	}

	//品牌api（C端用户浏览功能）
	brandsRouter := v1.Group("brands")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		brandsRouter.GET("", goodsController.BrandList) // 品牌列表（前端展示）

	}

	//轮播图api（C端用户浏览功能）
	bannersRouter := v1.Group("banners")
	{
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		bannersRouter.GET("", goodsController.BannerList) // 轮播图列表（前端展示）

	}

	//订单管理api
	ordersRouter := v1.Group("orders")
	{
		orderController := order.NewOrderController(serviceFactory, g.Translator())
		ordersRouter.GET("", jwtAuth, orderController.OrderList)       // 订单列表
		ordersRouter.POST("", jwtAuth, orderController.CreateOrder)    // 创建订单
		ordersRouter.GET("/:id", jwtAuth, orderController.OrderDetail) // 订单详情
	}

	//购物车管理api
	cartRouter := v1.Group("shopcarts")
	{
		orderController := order.NewOrderController(serviceFactory, g.Translator())
		cartRouter.GET("", jwtAuth, orderController.CartList)              // 购物车列表
		cartRouter.POST("", jwtAuth, orderController.AddToCart)            // 添加到购物车
		cartRouter.PATCH("/:id", jwtAuth, orderController.UpdateCartItem)  // 更新购物车商品
		cartRouter.DELETE("/:id", jwtAuth, orderController.DeleteCartItem) // 删除购物车商品
	}

	//用户收藏管理api
	favRouter := v1.Group("userfavs")
	{
		userOpController := userop.NewUserOpController(serviceFactory, g.Translator())
		favRouter.GET("", jwtAuth, userOpController.UserFavList)          // 收藏列表
		favRouter.POST("", jwtAuth, userOpController.CreateUserFav)       // 添加收藏
		favRouter.DELETE("/:id", jwtAuth, userOpController.DeleteUserFav) // 删除收藏
		favRouter.GET("/:id", jwtAuth, userOpController.GetUserFavDetail) // 查看是否收藏
	}

	//用户地址管理api
	addressRouter := v1.Group("address")
	{
		userOpController := userop.NewUserOpController(serviceFactory, g.Translator())
		addressRouter.GET("", jwtAuth, userOpController.GetAddressList)       // 地址列表
		addressRouter.POST("", jwtAuth, userOpController.CreateAddress)       // 创建地址
		addressRouter.PUT("/:id", jwtAuth, userOpController.UpdateAddress)    // 更新地址
		addressRouter.DELETE("/:id", jwtAuth, userOpController.DeleteAddress) // 删除地址
	}

	//用户留言管理api
	messageRouter := v1.Group("message")
	{
		userOpController := userop.NewUserOpController(serviceFactory, g.Translator())
		messageRouter.GET("", jwtAuth, userOpController.MessageList)    // 留言列表
		messageRouter.POST("", jwtAuth, userOpController.CreateMessage) // 创建留言
	}

	//优惠券管理api
	couponRouter := v1.Group("coupons")
	{
		couponController := coupon.NewCouponController(g.Translator(), serviceFactory)
		couponRouter.GET("templates", couponController.ListTemplates)                        // 获取优惠券模板列表
		couponRouter.POST("receive", jwtAuth, couponController.ReceiveCoupon)                // 用户领取优惠券
		couponRouter.GET("user", jwtAuth, couponController.GetUserCoupons)                   // 获取用户优惠券列表
		couponRouter.GET("available", jwtAuth, couponController.GetAvailableCoupons)         // 获取用户可用优惠券
		couponRouter.POST("calculate-discount", jwtAuth, couponController.CalculateDiscount) // 计算优惠券折扣
	}

	//支付管理api
	paymentRouter := v1.Group("payments")
	{
		paymentController := payment.NewPaymentController(g.Translator(), serviceFactory)
		paymentRouter.POST("", jwtAuth, paymentController.CreatePayment)              // 创建支付订单
		paymentRouter.GET("/:paymentSN/status", paymentController.GetPaymentStatus)   // 获取支付状态
		paymentRouter.POST("/:paymentSN/simulate", paymentController.SimulatePayment) // 模拟支付
	}

	//物流管理api
	logisticsRouter := v1.Group("logistics")
	{
		logisticsController := logistics.NewLogisticsController(g.Translator(), serviceFactory)
		logisticsRouter.GET("info", logisticsController.GetLogisticsInfo)              // 获取物流信息
		logisticsRouter.GET("tracks", logisticsController.GetLogisticsTracks)          // 获取物流轨迹
		logisticsRouter.POST("shipping-fee", logisticsController.CalculateShippingFee) // 计算运费
		logisticsRouter.GET("companies", logisticsController.GetLogisticsCompanies)    // 获取物流公司列表
	}
}
