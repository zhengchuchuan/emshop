package request

type GoodsFilter struct {
	PriceMin    *int32  `form:"priceMin"`
	PriceMax    *int32  `form:"priceMax"`
	IsHot       *bool   `form:"isHot"`
	IsNew       *bool   `form:"isNew"`
	IsTab       *bool   `form:"isTab"`
	TopCategory *int32  `form:"topCategory"`
	Pages       *int32  `form:"pages"`
	PagePerNums *int32  `form:"pagePerNums"`
	KeyWords    *string `form:"keyWords"`
	Brand       *int32  `form:"brand"`
}

type CreateGoods struct {
	Name            string   `json:"name" binding:"required,min=2,max=20"`
	GoodsSn         string   `json:"goods_sn"`
	MarketPrice     float32  `json:"market_price" binding:"required,min=0"`
	ShopPrice       float32  `json:"shop_price" binding:"required,min=0"`
	GoodsBrief      string   `json:"goods_brief" binding:"required,min=3"`
	GoodsDesc       string   `json:"goods_desc"`
	ShipFree        bool     `json:"ship_free"`
	Images          []string `json:"images"`
	DescImages      []string `json:"desc_images"`
	GoodsFrontImage string   `json:"goods_front_image" binding:"required,url"`
	IsNew           bool     `json:"is_new"`
	IsHot           bool     `json:"is_hot"`
	OnSale          bool     `json:"on_sale"`
	CategoryId      int32    `json:"category_id" binding:"required"`
	BrandId         int32    `json:"brand_id" binding:"required"`
}

type SyncData struct {
	ForceSync bool    `json:"force_sync"`
	GoodsIds  []int32 `json:"goods_ids"`
}

type UpdateGoods struct {
	Name            string   `json:"name" binding:"required,min=2,max=20"`
	GoodsSn         string   `json:"goods_sn"`
	MarketPrice     float32  `json:"market_price" binding:"required,min=0"`
	ShopPrice       float32  `json:"shop_price" binding:"required,min=0"`
	GoodsBrief      string   `json:"goods_brief" binding:"required,min=3"`
	GoodsDesc       string   `json:"goods_desc"`
	ShipFree        bool     `json:"ship_free"`
	Images          []string `json:"images"`
	DescImages      []string `json:"desc_images"`
	GoodsFrontImage string   `json:"goods_front_image" binding:"required,url"`
	IsNew           bool     `json:"is_new"`
	IsHot           bool     `json:"is_hot"`
	OnSale          bool     `json:"on_sale"`
	CategoryId      int32    `json:"category_id" binding:"required"`
	BrandId         int32    `json:"brand_id" binding:"required"`
}

type UpdateGoodsStatus struct {
	IsNew  *bool `json:"is_new" binding:"required"`
	IsHot  *bool `json:"is_hot" binding:"required"`
	OnSale *bool `json:"on_sale" binding:"required"`
}

// 分类管理相关结构体
type CreateCategory struct {
	Name           string `json:"name" binding:"required,min=2,max=20"`
	ParentCategory int32  `json:"parent_category"`
	Level          int32  `json:"level" binding:"required,min=1,max=3"`
	IsTab          *bool  `json:"is_tab" binding:"required"`
}

type UpdateCategory struct {
	Name  string `json:"name" binding:"required,min=2,max=20"`
	IsTab *bool  `json:"is_tab"`
}

// 品牌管理相关结构体
type CreateBrand struct {
	Name string `json:"name" binding:"required,min=2,max=20"`
	Logo string `json:"logo" binding:"url"`
}

type UpdateBrand struct {
	Name string `json:"name" binding:"required,min=2,max=20"`
	Logo string `json:"logo" binding:"url"`
}

type BrandFilter struct {
	Pages       *int32 `form:"pages"`
	PagePerNums *int32 `form:"pagePerNums"`
}

// 轮播图管理相关结构体
type CreateBanner struct {
	Index int32  `json:"index"`
	Image string `json:"image" binding:"required,url"`
	Url   string `json:"url" binding:"required,url"`
}

type UpdateBanner struct {
	Index int32  `json:"index"`
	Image string `json:"image" binding:"required,url"`
	Url   string `json:"url" binding:"required,url"`
}

// ==================== 订单管理相关结构体 ====================

// 订单列表查询参数
type OrderFilter struct {
	Pages       *int32 `form:"pages"`
	PagePerNums *int32 `form:"pagePerNums"`
}

// 创建订单请求
type CreateOrder struct {
	Address string `json:"address" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Mobile  string `json:"mobile" binding:"required"`
	Post    string `json:"post" binding:"required"`
}

// ==================== 购物车管理相关结构体 ====================

// 添加商品到购物车
type AddToCart struct {
	GoodsId int32 `json:"goods" binding:"required"`
	Nums    int32 `json:"nums" binding:"required,min=1"`
}

// 更新购物车商品
type UpdateCartItem struct {
	Nums    int32 `json:"nums" binding:"required,min=1"`
	Checked *bool `json:"checked"`
}

// ==================== 用户操作相关结构体 ====================

// 用户收藏
type UserFav struct {
	GoodsId int32 `json:"goods" binding:"required"`
}

// 用户地址
type CreateAddress struct {
	Province     string `json:"province" binding:"required"`
	City         string `json:"city" binding:"required"`
	District     string `json:"district" binding:"required"`
	Address      string `json:"address" binding:"required"`
	SignerName   string `json:"signer_name" binding:"required"`
	SignerMobile string `json:"signer_mobile" binding:"required"`
}

type UpdateAddress struct {
	Province     string `json:"province" binding:"required"`
	City         string `json:"city" binding:"required"`
	District     string `json:"district" binding:"required"`
	Address      string `json:"address" binding:"required"`
	SignerName   string `json:"signer_name" binding:"required"`
	SignerMobile string `json:"signer_mobile" binding:"required"`
}

// 用户留言
type CreateMessage struct {
	MessageType int32  `json:"type" binding:"required,oneof=1 2 3 4 5"`
	Subject     string `json:"subject" binding:"required"`
	Message     string `json:"message" binding:"required"`
	File        string `json:"file"`
}
