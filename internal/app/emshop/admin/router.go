package admin

import (
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/emshop/admin/config"
	"emshop/internal/app/emshop/admin/controller/user/v1"
	"emshop/internal/app/emshop/admin/controller/goods/v1"
	"emshop/internal/app/emshop/admin/controller/order/v1"
	"emshop/internal/app/emshop/admin/controller/upload/v1"
	"emshop/internal/app/emshop/admin/controller/export/v1"
	import_controller "emshop/internal/app/emshop/admin/controller/import/v1"
	"emshop/internal/app/emshop/admin/controller/analytics/v1"
	"emshop/internal/app/emshop/admin/data/rpc"
	"emshop/internal/app/emshop/admin/service"
	"emshop/internal/app/pkg/middleware"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	// 静态文件服务
	g.Static("/uploads", "./uploads")
	
	v1 := g.Group("/v1")

	// 创建数据工厂
	data, err := rpc.GetDataFactoryOr(cfg.Registry)
	if err != nil {
		panic(err)
	}
	
	// 创建服务工厂
	serviceFactory := service.NewService(data, cfg.Jwt)
	
	// 创建管理员认证中间件
	adminAuth := middleware.AdminAuth(cfg.Jwt)

	// 管理员认证接口（无需认证）
	authGroup := v1.Group("/admin/auth")
	{
		authController := user.NewUserController(g.Translator(), serviceFactory)
		authGroup.POST("/pwd_login", authController.AdminLogin)  // 管理员登录
		authGroup.GET("/captcha", user.GetCaptcha)              // 管理员验证码
		// authGroup.POST("/logout", authController.AdminLogout)  // 管理员退出（可选）
		// authGroup.GET("/profile", adminAuth, authController.AdminProfile)  // 管理员信息（需认证）
	}

	// 管理员API（需要管理员权限）
	adminGroup := v1.Group("/admin", adminAuth)
	{
		// 用户管理
		userGroup := adminGroup.Group("/users")
		userController := user.NewUserController(g.Translator(), serviceFactory)
		{
			userGroup.GET("", userController.GetUserList)                    // GET /v1/admin/users?pn=页码&psize=每页数量
			userGroup.GET("/by-mobile", userController.GetUserByMobile)      // GET /v1/admin/users/by-mobile?mobile=手机号
			userGroup.GET("/:id", userController.GetUserById)                // GET /v1/admin/users/:id
			userGroup.PATCH("/:id", userController.UpdateUser)                // PATCH /v1/admin/users/:id 更新用户信息
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
			goodsGroup.DELETE("/batch", goodsController.BatchDeleteGoods)    // DELETE /v1/admin/goods/batch 批量删除商品
			goodsGroup.PATCH("/batch/status", goodsController.BatchUpdateGoodsStatus) // PATCH /v1/admin/goods/batch/status 批量更新状态
		}

		// 分类管理
		categoriesGroup := adminGroup.Group("/categories")
		{
			categoriesGroup.GET("", goodsController.CategoryList)              // GET /v1/admin/categories 获取分类列表，支持?level=1,2,3参数
			categoriesGroup.GET("/flat", goodsController.CategoryListFlat)     // GET /v1/admin/categories/flat 获取所有层级的扁平分类列表
			categoriesGroup.GET("/tree", goodsController.CategoryListTree)     // GET /v1/admin/categories/tree 获取嵌套的分类树结构(JSON字符串)
			categoriesGroup.GET("/hierarchy", goodsController.CategoryHierarchy) // GET /v1/admin/categories/hierarchy 获取强类型的分类树结构
			categoriesGroup.POST("", goodsController.CreateCategory)           // POST /v1/admin/categories 创建分类
			categoriesGroup.PUT("/:id", goodsController.UpdateCategory)        // PUT /v1/admin/categories/:id 更新分类
			categoriesGroup.DELETE("/:id", goodsController.DeleteCategory)     // DELETE /v1/admin/categories/:id 删除分类
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

		// 库存管理
		inventoryGroup := adminGroup.Group("/inventory")
		{
			inventoryGroup.GET("/:id", goodsController.GetInventory)         // GET /v1/admin/inventory/:id 获取商品库存
			inventoryGroup.PUT("/:id", goodsController.SetInventory)         // PUT /v1/admin/inventory/:id 设置商品库存
			inventoryGroup.POST("/batch", goodsController.BatchSetInventory) // POST /v1/admin/inventory/batch 批量设置库存
		}

		// 文件上传管理
		uploadController := upload.NewUploadController(g.Translator())
		uploadGroup := adminGroup.Group("/upload")
		{
			uploadGroup.POST("/image", uploadController.UploadImage)         // POST /v1/admin/upload/image 上传单张图片
			uploadGroup.POST("/images", uploadController.BatchUploadImages)  // POST /v1/admin/upload/images 批量上传图片
		}

		// 数据导出管理
		exportController := export.NewExportController(serviceFactory, g.Translator())
		exportGroup := adminGroup.Group("/export")
		{
			exportGroup.GET("/goods", exportController.ExportGoods)           // GET /v1/admin/export/goods 导出商品数据
			exportGroup.GET("/goods/template", exportController.ExportGoodsTemplate) // GET /v1/admin/export/goods/template 下载导入模板
		}

		// 数据导入管理
		importController := import_controller.NewImportController(serviceFactory, g.Translator())
		importGroup := adminGroup.Group("/import")
		{
			importGroup.POST("/goods", importController.ImportGoods)         // POST /v1/admin/import/goods 导入商品数据
			importGroup.POST("/goods/validate", importController.ValidateImportFile) // POST /v1/admin/import/goods/validate 验证导入文件
		}

		// 数据分析统计
		analyticsController := analytics.NewAnalyticsController(serviceFactory, g.Translator())
		analyticsGroup := adminGroup.Group("/analytics")
		{
			analyticsGroup.GET("/goods/overview", analyticsController.GetGoodsOverview)    // GET /v1/admin/analytics/goods/overview 商品概览统计
			analyticsGroup.GET("/goods/top-selling", analyticsController.GetTopSellingGoods) // GET /v1/admin/analytics/goods/top-selling 热销商品排行
			analyticsGroup.GET("/category/stats", analyticsController.GetCategoryStats)   // GET /v1/admin/analytics/category/stats 分类统计
			analyticsGroup.GET("/inventory/alerts", analyticsController.GetInventoryAlerts) // GET /v1/admin/analytics/inventory/alerts 库存预警
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
