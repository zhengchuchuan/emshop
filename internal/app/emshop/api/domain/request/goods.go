package request

type GoodsFilter struct {
	PriceMin    int32  `form:"pmin"`	// 请求参数，请求 ...?pmin=100 时，PriceMin字段会赋值为 100。
	PriceMax    int32  `form:"pmax"`
	IsHot       bool   `form:"ih"`
	IsNew       bool   `form:"in"`
	IsTab       bool   `form:"it"`
	TopCategory int32  `form:"c"`
	Pages       int32  `form:"p"`
	PagePerNums int32  `form:"pnum"`
	KeyWords    string `form:"q"`
	Brand       int32  `form:"b"`
}

type CreateGoods struct {
	Name            string   `json:"name" binding:"required,min=2,max=20"`
	GoodsSn         string   `json:"goods_sn"`
	Stocks          int32    `json:"stocks" binding:"required,min=1"`
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
