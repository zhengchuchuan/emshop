package data

import (
	"context"
	cpb "emshop/api/coupon/v1"
	gpb "emshop/api/goods/v1"
	ipb "emshop/api/inventory/v1"
	lpb "emshop/api/logistics/v1"
	opb "emshop/api/order/v1"
	ppb "emshop/api/payment/v1"
	upb "emshop/api/user/v1"
	uoppb "emshop/api/userop/v1"
)

type GoodsData interface {
	GoodsList(ctx context.Context, request *gpb.GoodsFilterRequest) (*gpb.GoodsListResponse, error)
	CreateGoods(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)
	SyncGoodsData(ctx context.Context, request *gpb.SyncDataRequest) (*gpb.SyncDataResponse, error)
	GetGoodsDetail(ctx context.Context, request *gpb.GoodInfoRequest) (*gpb.GoodsInfoResponse, error)
	DeleteGoods(ctx context.Context, info *gpb.DeleteGoodsInfo) (*gpb.GoodsInfoResponse, error)
	UpdateGoods(ctx context.Context, info *gpb.CreateGoodsInfo) (*gpb.GoodsInfoResponse, error)

	// 分类管理
	GetAllCategorysList(ctx context.Context) (*gpb.CategoryListResponse, error)
	GetSubCategory(ctx context.Context, request *gpb.CategoryListRequest) (*gpb.SubCategoryListResponse, error)
	CreateCategory(ctx context.Context, request *gpb.CategoryInfoRequest) (*gpb.CategoryInfoResponse, error)
	UpdateCategory(ctx context.Context, request *gpb.CategoryInfoRequest) (*gpb.CategoryInfoResponse, error)
	DeleteCategory(ctx context.Context, request *gpb.DeleteCategoryRequest) (*gpb.CategoryInfoResponse, error)

	// 品牌管理
	BrandList(ctx context.Context, request *gpb.BrandFilterRequest) (*gpb.BrandListResponse, error)
	CreateBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error)
	UpdateBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error)
	DeleteBrand(ctx context.Context, request *gpb.BrandRequest) (*gpb.BrandInfoResponse, error)

	// 轮播图管理
	BannerList(ctx context.Context) (*gpb.BannerListResponse, error)
	CreateBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error)
	UpdateBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error)
	DeleteBanner(ctx context.Context, request *gpb.BannerRequest) (*gpb.BannerResponse, error)
}

type OrderData interface {
	// 订单管理
	OrderList(ctx context.Context, request *opb.OrderFilterRequest) (*opb.OrderListResponse, error)
	CreateOrder(ctx context.Context, request *opb.OrderRequest) (*opb.OrderInfoResponse, error)
	OrderDetail(ctx context.Context, request *opb.OrderRequest) (*opb.OrderInfoDetailResponse, error)
	UpdateOrderStatus(ctx context.Context, request *opb.OrderStatus) (*opb.OrderInfoResponse, error)

	// 购物车管理
	CartItemList(ctx context.Context, request *opb.UserInfo) (*opb.CartItemListResponse, error)
	CreateCartItem(ctx context.Context, request *opb.CartItemRequest) (*opb.ShopCartInfoResponse, error)
	UpdateCartItem(ctx context.Context, request *opb.CartItemRequest) (*opb.ShopCartInfoResponse, error)
	DeleteCartItem(ctx context.Context, request *opb.CartItemRequest) (*opb.ShopCartInfoResponse, error)
}

type UserOpData interface {
	// 用户收藏管理
	UserFavList(ctx context.Context, request *uoppb.UserFavListRequest) (*uoppb.UserFavListResponse, error)
	CreateUserFav(ctx context.Context, request *uoppb.UserFavRequest) (*uoppb.UserFavResponse, error)
	DeleteUserFav(ctx context.Context, request *uoppb.UserFavRequest) (*uoppb.UserFavResponse, error)
	GetUserFavDetail(ctx context.Context, request *uoppb.UserFavRequest) (*uoppb.UserFavResponse, error)

	// 用户地址管理
	GetAddressList(ctx context.Context, request *uoppb.AddressRequest) (*uoppb.AddressListResponse, error)
	CreateAddress(ctx context.Context, request *uoppb.AddressRequest) (*uoppb.AddressResponse, error)
	UpdateAddress(ctx context.Context, request *uoppb.AddressRequest) (*uoppb.AddressResponse, error)
	DeleteAddress(ctx context.Context, request *uoppb.DeleteAddressRequest) (*uoppb.AddressResponse, error)

	// 用户留言管理
	MessageList(ctx context.Context, request *uoppb.MessageRequest) (*uoppb.MessageListResponse, error)
	CreateMessage(ctx context.Context, request *uoppb.MessageRequest) (*uoppb.MessageResponse, error)
}

type UserData interface {
	// 用户管理
	GetUserList(ctx context.Context, request *upb.PageInfo) (*upb.UserListResponse, error)
	GetUserByMobile(ctx context.Context, request *upb.MobileRequest) (*upb.UserInfoResponse, error)
	GetUserById(ctx context.Context, request *upb.IdRequest) (*upb.UserInfoResponse, error)
	CreateUser(ctx context.Context, request *upb.CreateUserInfo) (*upb.UserInfoResponse, error)
	UpdateUser(ctx context.Context, request *upb.UpdateUserInfo) (*upb.UserInfoResponse, error)
	CheckPassWord(ctx context.Context, request *upb.PasswordCheckInfo) (*upb.CheckResponse, error)
}

type InventoryData interface {
	// 库存管理
	InvDetail(ctx context.Context, request *ipb.GoodsInvInfo) (*ipb.GoodsInvInfo, error)
}

type CouponData interface {
	// 优惠券模板管理（用于展示可领取的优惠券）
	ListCouponTemplates(ctx context.Context, request *cpb.ListCouponTemplatesRequest) (*cpb.ListCouponTemplatesResponse, error)
	GetCouponTemplate(ctx context.Context, request *cpb.GetCouponTemplateRequest) (*cpb.CouponTemplateResponse, error)

	// 用户优惠券操作
	ReceiveCoupon(ctx context.Context, request *cpb.ReceiveCouponRequest) (*cpb.UserCouponResponse, error)
	GetUserCoupons(ctx context.Context, request *cpb.GetUserCouponsRequest) (*cpb.ListUserCouponsResponse, error)
	GetAvailableCoupons(ctx context.Context, request *cpb.GetAvailableCouponsRequest) (*cpb.ListUserCouponsResponse, error)

	// 优惠券计算和使用
	CalculateCouponDiscount(ctx context.Context, request *cpb.CalculateCouponDiscountRequest) (*cpb.CalculateCouponDiscountResponse, error)
}

type PaymentData interface {
	// 支付订单管理
	CreatePayment(ctx context.Context, request *ppb.CreatePaymentRequest) (*ppb.CreatePaymentResponse, error)
	GetPaymentStatus(ctx context.Context, request *ppb.GetPaymentStatusRequest) (*ppb.PaymentStatusResponse, error)

	// 模拟支付操作（测试用）
	SimulatePaymentSuccess(ctx context.Context, request *ppb.SimulatePaymentRequest) error
	SimulatePaymentFailure(ctx context.Context, request *ppb.SimulatePaymentRequest) error
}

type LogisticsData interface {
	// 物流信息查询
	GetLogisticsInfo(ctx context.Context, request *lpb.GetLogisticsInfoRequest) (*lpb.GetLogisticsInfoResponse, error)
	GetLogisticsTracks(ctx context.Context, request *lpb.GetLogisticsTracksRequest) (*lpb.GetLogisticsTracksResponse, error)

	// 运费计算
	CalculateShippingFee(ctx context.Context, request *lpb.CalculateShippingFeeRequest) (*lpb.CalculateShippingFeeResponse, error)

	// 基础数据获取
	GetLogisticsCompanies(ctx context.Context) (*lpb.LogisticsCompaniesResponse, error)
}

type DataFactory interface {
	Goods() GoodsData
	Users() UserData
	Order() OrderData
	UserOp() UserOpData
	Inventory() InventoryData
	Coupon() CouponData
	Payment() PaymentData
	Logistics() LogisticsData
}
