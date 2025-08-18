package request

// AdminOrderFilter 管理员订单列表查询参数
type AdminOrderFilter struct {
	Pages       int32  `form:"p"`        // 页码
	PagePerNums int32  `form:"pnum"`     // 每页数量
	UserId      int32  `form:"userId"`  // 用户ID筛选（可选）
	Status      string `form:"status"`   // 订单状态筛选（可选）
	OrderSn     string `form:"orderSn"` // 订单号模糊查询（可选）
	StartDate   string `form:"startDate"` // 开始日期 YYYY-MM-DD（可选）
	EndDate     string `form:"endDate"`   // 结束日期 YYYY-MM-DD（可选）
}

// UpdateOrderStatusRequest 更新订单状态请求
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"` // 新状态
}

// OrderSearchBySnRequest 按订单号查询请求
type OrderSearchBySnRequest struct {
	OrderSn string `form:"orderSn" binding:"required"` // 订单号
}

// OrderSearchByUserIdRequest 按用户ID查询请求
type OrderSearchByUserIdRequest struct {
	UserId      int32 `form:"userId" binding:"required,min=1"` // 用户ID
	Pages       int32 `form:"p"`                                // 页码
	PagePerNums int32 `form:"pnum"`                             // 每页数量
}