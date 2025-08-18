package response

// GoodsOverviewResponse 商品概览响应
type GoodsOverviewResponse struct {
	TotalGoods       int     `json:"totalGoods"`
	OnSaleGoods      int     `json:"onSaleGoods"`
	OffSaleGoods     int     `json:"offSaleGoods"`
	NewGoods         int     `json:"newGoods"`
	HotGoods         int     `json:"hotGoods"`
	LowStockGoods    int     `json:"lowStockGoods"`
	OutOfStockGoods  int     `json:"outOfStockGoods"`
	TotalValue       float32 `json:"totalValue"`
	AvgPrice         float32 `json:"avgPrice"`
	MaxPrice         float32 `json:"maxPrice"`
	MinPrice         float32 `json:"minPrice"`
}

// TopSellingGoodsItem 热销商品项
type TopSellingGoodsItem struct {
	Rank    int     `json:"rank"`
	GoodsId int32   `json:"goodsId"`
	Name    string  `json:"name"`
	SoldNum int32   `json:"soldNum"`
	Price   float32 `json:"price"`
}

// TopSellingGoodsResponse 热销商品响应
type TopSellingGoodsResponse struct {
	TopSellingGoods []TopSellingGoodsItem `json:"topSellingGoods"`
	TotalFound      int                   `json:"totalFound"`
}

// CategoryStatsItem 分类统计项
type CategoryStatsItem struct {
	CategoryId   int32   `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	GoodsCount   int     `json:"goodsCount"`
	OnSaleCount  int     `json:"onSaleCount"`
	OffSaleCount int     `json:"offSaleCount"`
	TotalSold    int32   `json:"totalSold"`
	TotalRevenue float32 `json:"totalRevenue"`
	AvgPrice     float32 `json:"avgPrice"`
	TotalStocks  int32   `json:"totalStocks"`
}

// CategoryStatsResponse 分类统计响应
type CategoryStatsResponse struct {
	CategoryStats   []CategoryStatsItem `json:"categoryStats"`
	TotalCategories int                 `json:"totalCategories"`
}

// StockAlertGoodsItem 库存预警商品项
type StockAlertGoodsItem struct {
	GoodsId  int32  `json:"goodsId"`
	Name     string `json:"name"`
	GoodsSn  string `json:"goodsSn"`
	Stocks   int32  `json:"stocks"`
	Category string `json:"category"`
	Brand    string `json:"brand"`
	OnSale   bool   `json:"onSale"`
	SoldNum  int32  `json:"soldNum"`
}

// StockAlertSummary 库存预警汇总
type StockAlertSummary struct {
	OutOfStockCount int   `json:"outOfStockCount"`
	LowStockCount   int   `json:"lowStockCount"`
	TotalAlerts     int   `json:"totalAlerts"`
	LowThreshold    int32 `json:"lowThreshold"`
}

// StockAlertResponse 库存预警响应
type StockAlertResponse struct {
	OutOfStock []StockAlertGoodsItem `json:"outOfStock"`
	LowStock   []StockAlertGoodsItem `json:"lowStock"`
	Summary    StockAlertSummary     `json:"summary"`
}