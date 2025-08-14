package data

import (
	"context"
	gpb "emshop/api/goods/v1"
	opb "emshop/api/order/v1"
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

type DataFactory interface {
	Goods() GoodsData
	Users() UserData
	Order() OrderData
	UserOp() UserOpData
}
