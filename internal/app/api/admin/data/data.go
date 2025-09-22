package data

import (
    "context"
    upbv1 "emshop/api/user/v1"
    gpbv1 "emshop/api/goods/v1"
    ipbv1 "emshop/api/inventory/v1"
    opbv1 "emshop/api/order/v1"
    cpbv1 "emshop/api/coupon/v1"
)

// DataFactory 数据访问工厂接口
type DataFactory interface {
    Users() UserData
    Goods() GoodsData
    Inventory() InventoryData
    Order() OrderData
    UserOp() UserOpData
    Coupon() CouponData
}

// UserData 用户数据访问接口
type UserData interface {
	CheckPassWord(ctx context.Context, request *upbv1.PasswordCheckInfo) (*upbv1.CheckResponse, error)
	CreateUser(ctx context.Context, request *upbv1.CreateUserInfo) (*upbv1.UserInfoResponse, error)
	UpdateUser(ctx context.Context, request *upbv1.UpdateUserInfo) (*upbv1.UserInfoResponse, error)
	GetUserById(ctx context.Context, request *upbv1.IdRequest) (*upbv1.UserInfoResponse, error)
	GetUserByMobile(ctx context.Context, request *upbv1.MobileRequest) (*upbv1.UserInfoResponse, error)
	GetUserList(ctx context.Context, request *upbv1.PageInfo) (*upbv1.UserListResponse, error)
}

// GoodsData 商品数据访问接口
type GoodsData interface {
	// 商品管理
	GoodsList(ctx context.Context, request *gpbv1.GoodsFilterRequest) (*gpbv1.GoodsListResponse, error)
	CreateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error)
	SyncGoodsData(ctx context.Context, request *gpbv1.SyncDataRequest) (*gpbv1.SyncDataResponse, error)
	GetGoodsDetail(ctx context.Context, request *gpbv1.GoodInfoRequest) (*gpbv1.GoodsInfoResponse, error)
	DeleteGoods(ctx context.Context, info *gpbv1.DeleteGoodsInfo) (*gpbv1.GoodsInfoResponse, error)
	UpdateGoods(ctx context.Context, info *gpbv1.CreateGoodsInfo) (*gpbv1.GoodsInfoResponse, error)
	
	// 分类管理
	GetCategoriesList(ctx context.Context) (*gpbv1.CategoryListResponse, error)
	GetCategoriesByLevel(ctx context.Context, level int32) (*gpbv1.CategoryListResponse, error)
	GetCategoryTree(ctx context.Context) (*gpbv1.CategoryTreeResponse, error)
	GetSubCategory(ctx context.Context, request *gpbv1.CategoryListRequest) (*gpbv1.SubCategoryListResponse, error)
	CreateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error)
	UpdateCategory(ctx context.Context, request *gpbv1.CategoryInfoRequest) (*gpbv1.CategoryInfoResponse, error)
	DeleteCategory(ctx context.Context, request *gpbv1.DeleteCategoryRequest) (*gpbv1.CategoryInfoResponse, error)
	
	// 品牌管理
	BrandList(ctx context.Context, request *gpbv1.BrandFilterRequest) (*gpbv1.BrandListResponse, error)
	CreateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error)
	UpdateBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error)
	DeleteBrand(ctx context.Context, request *gpbv1.BrandRequest) (*gpbv1.BrandInfoResponse, error)
	
	// 轮播图管理
	BannerList(ctx context.Context) (*gpbv1.BannerListResponse, error)
	CreateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error)
	UpdateBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error)
	DeleteBanner(ctx context.Context, request *gpbv1.BannerRequest) (*gpbv1.BannerResponse, error)
	
	// 批量操作
	BatchDeleteGoods(ctx context.Context, request *gpbv1.BatchDeleteGoodsRequest) (*gpbv1.BatchOperationResponse, error)
	BatchUpdateGoodsStatus(ctx context.Context, request *gpbv1.BatchUpdateGoodsStatusRequest) (*gpbv1.BatchOperationResponse, error)
}

// InventoryData 库存数据访问接口
type InventoryData interface {
	// 获取商品库存信息
	GetInventory(ctx context.Context, goodsId int32) (*ipbv1.GoodsInvInfo, error)
	// 批量获取商品库存信息
	BatchGetInventory(ctx context.Context, goodsIds []int32) (map[int32]*ipbv1.GoodsInvInfo, error)
	// 设置商品库存
	SetInventory(ctx context.Context, request *ipbv1.GoodsInvInfo) error
	// 批量设置商品库存
	BatchSetInventory(ctx context.Context, inventories []*ipbv1.GoodsInvInfo) error
}

// OrderData 订单数据访问接口
type OrderData interface {
	// 管理员查看所有订单列表（支持多维度筛选）
	AdminOrderList(ctx context.Context, request *opbv1.OrderFilterRequest) (*opbv1.OrderListResponse, error)
	// 管理员查看订单详情
	AdminOrderDetail(ctx context.Context, request *opbv1.OrderRequest) (*opbv1.OrderInfoDetailResponse, error)
	// 管理员更新订单状态
	UpdateOrderStatus(ctx context.Context, request *opbv1.OrderStatus) error
	// 按订单号查询订单
	GetOrderByOrderSn(ctx context.Context, orderSn string) (*opbv1.OrderInfoDetailResponse, error)
	// 按用户ID查询订单列表
	GetOrdersByUserId(ctx context.Context, userId int32, pages, pagePerNums int32) (*opbv1.OrderListResponse, error)
}

// UserOpData 用户操作数据访问接口
type UserOpData interface {
    // 管理员查看用户操作相关方法
}

// CouponData 优惠券数据访问接口
type CouponData interface {
    // 优惠券模板
    ListCouponTemplates(ctx context.Context, req *cpbv1.ListCouponTemplatesRequest) (*cpbv1.ListCouponTemplatesResponse, error)
    CreateCouponTemplate(ctx context.Context, req *cpbv1.CreateCouponTemplateRequest) (*cpbv1.CouponTemplateResponse, error)
}
