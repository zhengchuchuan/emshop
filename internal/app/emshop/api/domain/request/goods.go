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
