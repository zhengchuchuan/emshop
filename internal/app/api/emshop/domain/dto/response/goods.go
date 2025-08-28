package response

// GoodsItemResponse 商品列表项响应
type GoodsItemResponse struct {
	ID         int32            `json:"id"`
	Name       string           `json:"name"`
	GoodsBrief string           `json:"goodsBrief"`
	Desc       string           `json:"desc"`
	ShipFree   bool             `json:"shipFree"`
	Images     []string         `json:"images"`
	DescImages []string         `json:"descImages"`
	FrontImage string           `json:"frontImage"`
	ShopPrice  float32          `json:"shopPrice"`
	Category   CategoryResponse `json:"category"`
	Brand      BrandResponse    `json:"brand"`
	IsHot      bool             `json:"isHot"`
	IsNew      bool             `json:"isNew"`
	OnSale     bool             `json:"onSale"`
}

// GoodsListResponse 商品列表响应
type GoodsListResponse struct {
	Total int64               `json:"total"`
	Items []GoodsItemResponse `json:"data"`
}

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// BrandResponse 品牌响应
type BrandResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// CategoryItemResponse 分类项响应
type CategoryItemResponse struct {
	ID             int32  `json:"id"`
	Name           string `json:"name"`
	Level          int32  `json:"level"`
	ParentCategory int32  `json:"parentCategory"`
	IsTab          bool   `json:"isTab"`
}

// CategoryDetailResponse 分类详情响应
type CategoryDetailResponse struct {
	ID             int32                  `json:"id"`
	Name           string                 `json:"name"`
	Level          int32                  `json:"level"`
	ParentCategory int32                  `json:"parentCategory"`
	IsTab          bool                   `json:"isTab"`
	SubCategories  []CategoryItemResponse `json:"subCategories"`
}

// BrandListResponse 品牌列表响应
type BrandListResponse struct {
	Total int64           `json:"total"`
	Items []BrandResponse `json:"data"`
}

// BannerResponse 轮播图响应
type BannerResponse struct {
	ID    int32  `json:"id"`
	Index int32  `json:"index"`
	Image string `json:"image"`
	Url   string `json:"url"`
}

// BannerListResponse 轮播图列表响应
type BannerListResponse struct {
	Total int64            `json:"total"`
	Items []BannerResponse `json:"data"`
}

// SyncResponse 同步响应
type SyncResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	SyncedCount int32    `json:"syncedCount"`
	FailedCount int32    `json:"failedCount"`
	Errors      []string `json:"errors"`
}

// MessageResponse 通用消息响应
type MessageResponse struct {
	Message string `json:"msg"`
}
