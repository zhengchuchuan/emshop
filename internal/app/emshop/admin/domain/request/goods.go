package request

// AdminGoodsFilter 管理员商品列表查询参数
type AdminGoodsFilter struct {
	PriceMin    *int32  `form:"priceMin"`    // 最小价格筛选
	PriceMax    *int32  `form:"priceMax"`    // 最大价格筛选
	IsHot       *bool   `form:"isHot"`       // 热门商品筛选
	IsNew       *bool   `form:"isNew"`       // 新品筛选
	IsTab       *bool   `form:"isTab"`       // Tab推荐筛选
	TopCategory *int32  `form:"topCategory"` // 顶级分类筛选
	Pages       *int32  `form:"pages"`       // 页码
	PagePerNums *int32  `form:"pagePerNums"` // 每页数量
	KeyWords    *string `form:"keyWords"`    // 关键词搜索
	Brand       *int32  `form:"brand"`       // 品牌筛选
}