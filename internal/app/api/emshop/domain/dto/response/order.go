package response

// OrderItemResponse 订单项响应
type OrderItemResponse struct {
	ID      int32   `json:"id"`
	OrderSn string  `json:"orderSn"`
	Status  string  `json:"status"`
	PayType string  `json:"payType"`
	Total   float32 `json:"total"`
	Address string  `json:"address"`
	Name    string  `json:"name"`
	Mobile  string  `json:"mobile"`
	AddTime int64   `json:"addTime"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Total int64               `json:"total"`
	Items []OrderItemResponse `json:"data"`
}

// OrderGoodsItem 订单商品项
type OrderGoodsItem struct {
	ID         int32   `json:"id"`
	GoodsId    int32   `json:"goodsId"`
	GoodsName  string  `json:"goodsName"`
	GoodsImage string  `json:"goodsImage"`
	GoodsPrice float32 `json:"goodsPrice"`
	Nums       int32   `json:"nums"`
}

// OrderDetailResponse 订单详情响应
type OrderDetailResponse struct {
	ID      int32            `json:"id"`
	OrderSn string           `json:"orderSn"`
	Status  string           `json:"status"`
	PayType string           `json:"payType"`
	Total   float32          `json:"total"`
	Address string           `json:"address"`
	Name    string           `json:"name"`
	Mobile  string           `json:"mobile"`
	AddTime int64            `json:"addTime"`
	Goods   []OrderGoodsItem `json:"goods"`
}

// CartItemResponse 购物车项响应
type CartItemResponse struct {
	ID      int32 `json:"id"`
	GoodsId int32 `json:"goodsId"`
	Nums    int32 `json:"nums"`
	Checked bool  `json:"checked"`
}

// CartListResponse 购物车列表响应
type CartListResponse struct {
	Total int64              `json:"total"`
	Items []CartItemResponse `json:"data"`
}
