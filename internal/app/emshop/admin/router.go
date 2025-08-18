package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/admin/config"
	"emshop/internal/app/emshop/admin/controller/user/v1"
	"emshop/internal/app/emshop/admin/controller/goods/v1"
	"emshop/internal/app/emshop/admin/controller/order/v1"
	"emshop/internal/app/emshop/admin/data/rpc"
	"emshop/internal/app/emshop/admin/service"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	v1 := g.Group("/v1")

	// 创建数据工厂
	data, err := rpc.GetDataFactoryOr(cfg.Registry)
	if err != nil {
		panic(err)
	}
	
	// 创建服务工厂
	serviceFactory := service.NewService(data, cfg.Jwt)
	
	// TODO: 添加JWT认证中间件
	// jwtAuth := newJWTAuth(cfg.Jwt)

	// 管理员API
	adminGroup := v1.Group("/admin")
	{
		// 用户管理
		userGroup := adminGroup.Group("/users")
		userController := user.NewUserController(g.Translator(), serviceFactory)
		{
			userGroup.GET("", userController.GetUserList)                    // GET /v1/admin/users?pn=页码&psize=每页数量
			userGroup.GET("/by-mobile", userController.GetUserByMobile)      // GET /v1/admin/users/by-mobile?mobile=手机号
			userGroup.GET("/:id", userController.GetUserById)                // GET /v1/admin/users/:id
			userGroup.PATCH("/:id/status", userController.UpdateUserStatus)  // PATCH /v1/admin/users/:id/status
		}

		// 商品管理
		goodsGroup := adminGroup.Group("/goods")
		goodsController := goods.NewGoodsController(serviceFactory, g.Translator())
		{
			goodsGroup.GET("", goodsController.List)                         // GET /v1/admin/goods 商品列表
			goodsGroup.POST("", goodsController.Create)                      // POST /v1/admin/goods 创建商品
			goodsGroup.GET("/:id", goodsController.Detail)                   // GET /v1/admin/goods/:id 商品详情
			goodsGroup.PUT("/:id", goodsController.Update)                   // PUT /v1/admin/goods/:id 更新商品
			goodsGroup.DELETE("/:id", goodsController.Delete)                // DELETE /v1/admin/goods/:id 删除商品
			goodsGroup.POST("/sync", goodsController.Sync)                   // POST /v1/admin/goods/sync 同步商品数据
		}

		// 分类管理
		categoriesGroup := adminGroup.Group("/categories")
		{
			categoriesGroup.GET("", goodsController.CategoryList)            // GET /v1/admin/categories 分类列表
			categoriesGroup.POST("", goodsController.CreateCategory)         // POST /v1/admin/categories 创建分类
			categoriesGroup.PUT("/:id", goodsController.UpdateCategory)      // PUT /v1/admin/categories/:id 更新分类
			categoriesGroup.DELETE("/:id", goodsController.DeleteCategory)   // DELETE /v1/admin/categories/:id 删除分类
		}

		// 品牌管理
		brandsGroup := adminGroup.Group("/brands")
		{
			brandsGroup.GET("", goodsController.BrandList)                   // GET /v1/admin/brands 品牌列表
			brandsGroup.POST("", goodsController.CreateBrand)                // POST /v1/admin/brands 创建品牌
			brandsGroup.PUT("/:id", goodsController.UpdateBrand)             // PUT /v1/admin/brands/:id 更新品牌
			brandsGroup.DELETE("/:id", goodsController.DeleteBrand)          // DELETE /v1/admin/brands/:id 删除品牌
		}

		// 轮播图管理
		bannersGroup := adminGroup.Group("/banners")
		{
			bannersGroup.GET("", goodsController.BannerList)                 // GET /v1/admin/banners 轮播图列表
			bannersGroup.POST("", goodsController.CreateBanner)              // POST /v1/admin/banners 创建轮播图
			bannersGroup.PUT("/:id", goodsController.UpdateBanner)           // PUT /v1/admin/banners/:id 更新轮播图
			bannersGroup.DELETE("/:id", goodsController.DeleteBanner)        // DELETE /v1/admin/banners/:id 删除轮播图
		}

		// 订单管理
		orderController := order.NewOrderController(serviceFactory, g.Translator())
		ordersGroup := adminGroup.Group("/orders")
		{
			ordersGroup.GET("", orderController.AdminOrderList)              // GET /v1/admin/orders 订单列表（支持多维度筛选）
			ordersGroup.GET("/:id", orderController.AdminOrderDetail)        // GET /v1/admin/orders/:id 订单详情
			ordersGroup.PATCH("/:id/status", orderController.UpdateOrderStatus) // PATCH /v1/admin/orders/:id/status 更新订单状态
			ordersGroup.GET("/by-sn/:order_sn", orderController.GetOrderByOrderSn) // GET /v1/admin/orders/by-sn/:order_sn 按订单号查询
			ordersGroup.GET("/by-user/:user_id", orderController.GetOrdersByUserId) // GET /v1/admin/orders/by-user/:user_id 按用户ID查询
		}
	}
}
