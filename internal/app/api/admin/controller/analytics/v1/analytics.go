package analytics

import (
	"sort"
	"strconv"

	gpbv1 "emshop/api/goods/v1"
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/admin/service"
	"emshop/pkg/common/core"
	"emshop/pkg/log"

	"github.com/gin-gonic/gin"
)

type analyticsController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

func NewAnalyticsController(sf service.ServiceFactory, trans restserver.I18nTranslator) *analyticsController {
	return &analyticsController{
		sf:    sf,
		trans: trans,
	}
}

// GetGoodsOverview 获取商品概览统计（管理员专用）
func (ac *analyticsController) GetGoodsOverview(ctx *gin.Context) {
	// 获取所有商品数据进行统计
	pages := int32(1)
	pagePerNums := int32(10000)
	request := &gpbv1.GoodsFilterRequest{
		Pages:       &pages,
		PagePerNums: &pagePerNums, // 获取大量数据进行统计
	}

	response, err := ac.sf.Goods().GetGoodsList(ctx, request)
	if err != nil {
		log.Errorf("Failed to get goods list for analytics: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 统计数据
	overview := map[string]interface{}{
		"totalGoods":    len(response.Data),
		"onSaleGoods":  0,
		"offSaleGoods": 0,
		"newGoods":      0,
		"hotGoods":      0,
		"lowStockGoods": 0, // 库存小于10的商品
		"outOfStockGoods": 0, // 库存为0的商品
		"totalValue":    float32(0), // 总价值（库存 * 销售价）
		"avgPrice":      float32(0),
		"maxPrice":      float32(0),
		"minPrice":      float32(0),
	}

	var totalPrice float32
	var prices []float32
	var lowStockCount int
	var outOfStockCount int

	for _, goods := range response.Data {
		// 基本统计
		if goods.OnSale {
			overview["on_sale_goods"] = overview["on_sale_goods"].(int) + 1
		} else {
			overview["off_sale_goods"] = overview["off_sale_goods"].(int) + 1
		}

		if goods.IsNew {
			overview["new_goods"] = overview["new_goods"].(int) + 1
		}

		if goods.IsHot {
			overview["hot_goods"] = overview["hot_goods"].(int) + 1
		}

		// 库存统计
		if goods.Stocks == 0 {
			outOfStockCount++
		} else if goods.Stocks < 10 {
			lowStockCount++
		}

		// 价格统计
		totalPrice += goods.ShopPrice
		prices = append(prices, goods.ShopPrice)
		overview["total_value"] = overview["total_value"].(float32) + (goods.ShopPrice * float32(goods.Stocks))
	}

	overview["low_stock_goods"] = lowStockCount
	overview["out_of_stock_goods"] = outOfStockCount

	// 计算平均价格和价格范围
	if len(response.Data) > 0 {
		overview["avg_price"] = totalPrice / float32(len(response.Data))
		
		sort.Slice(prices, func(i, j int) bool {
			return prices[i] < prices[j]
		})
		overview["min_price"] = prices[0]
		overview["max_price"] = prices[len(prices)-1]
	}

	core.WriteResponse(ctx, nil, overview)
}

// GetTopSellingGoods 获取热销商品排行（管理员专用）
func (ac *analyticsController) GetTopSellingGoods(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	// 获取商品数据
	pages2 := int32(1)
	pagePerNums2 := int32(1000)
	request := &gpbv1.GoodsFilterRequest{
		Pages:       &pages2,
		PagePerNums: &pagePerNums2,
	}

	response, err := ac.sf.Goods().GetGoodsList(ctx, request)
	if err != nil {
		log.Errorf("Failed to get goods list for top selling analytics: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 按销量排序
	sort.Slice(response.Data, func(i, j int) bool {
		return response.Data[i].SoldNum > response.Data[j].SoldNum
	})

	// 限制返回数量
	if len(response.Data) > limit {
		response.Data = response.Data[:limit]
	}

	// 构建结果
	result := make([]map[string]interface{}, 0, len(response.Data))
	for i, goods := range response.Data {
		result = append(result, map[string]interface{}{
			"rank":       i + 1,
			"goodsId":   goods.Id,
			"name":       goods.Name,
			"soldNum":   goods.SoldNum,
			"price":      goods.ShopPrice,
			"stocks":     goods.Stocks,
			"category":   goods.Category.Name,
			"brand":      goods.Brand.Name,
			"revenue":    float32(goods.SoldNum) * goods.ShopPrice,
		})
	}

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"topSellingGoods": result,
		"totalFound":       len(response.Data),
	})
}

// GetCategoryStats 获取分类统计（管理员专用）
func (ac *analyticsController) GetCategoryStats(ctx *gin.Context) {
	// 获取所有商品数据
	pages3 := int32(1)
	pagePerNums3 := int32(10000)
	goodsRequest := &gpbv1.GoodsFilterRequest{
		Pages:       &pages3,
		PagePerNums: &pagePerNums3,
	}

	goodsResponse, err := ac.sf.Goods().GetGoodsList(ctx, goodsRequest)
	if err != nil {
		log.Errorf("Failed to get goods list for category analytics: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 统计分类数据
	categoryStats := make(map[string]map[string]interface{})

	for _, goods := range goodsResponse.Data {
		categoryName := goods.Category.Name
		if categoryName == "" {
			categoryName = "未分类"
		}

		if _, exists := categoryStats[categoryName]; !exists {
			categoryStats[categoryName] = map[string]interface{}{
				"categoryId":     goods.CategoryId,
				"categoryName":   categoryName,
				"goodsCount":     0,
				"onSaleCount":   0,
				"offSaleCount":  0,
				"totalSold":      int32(0),
				"totalRevenue":   float32(0),
				"avgPrice":       float32(0),
				"totalStocks":    int32(0),
			}
		}

		stats := categoryStats[categoryName]
		stats["goods_count"] = stats["goods_count"].(int) + 1

		if goods.OnSale {
			stats["on_sale_count"] = stats["on_sale_count"].(int) + 1
		} else {
			stats["off_sale_count"] = stats["off_sale_count"].(int) + 1
		}

		stats["total_sold"] = stats["total_sold"].(int32) + goods.SoldNum
		stats["total_revenue"] = stats["total_revenue"].(float32) + (float32(goods.SoldNum) * goods.ShopPrice)
		stats["total_stocks"] = stats["total_stocks"].(int32) + goods.Stocks
	}

	// 计算平均价格并转换为数组
	result := make([]map[string]interface{}, 0, len(categoryStats))
	for _, stats := range categoryStats {
		if stats["goods_count"].(int) > 0 {
			// 重新计算平均价格
			totalPrice := float32(0)
			count := 0
			for _, goods := range goodsResponse.Data {
				if goods.Category.Name == stats["category_name"].(string) {
					totalPrice += goods.ShopPrice
					count++
				}
			}
			if count > 0 {
				stats["avg_price"] = totalPrice / float32(count)
			}
		}
		result = append(result, stats)
	}

	// 按商品数量排序
	sort.Slice(result, func(i, j int) bool {
		return result[i]["goods_count"].(int) > result[j]["goods_count"].(int)
	})

	core.WriteResponse(ctx, nil, map[string]interface{}{
		"categoryStats": result,
		"totalCategories": len(result),
	})
}

// GetInventoryAlerts 获取库存预警（管理员专用）
func (ac *analyticsController) GetInventoryAlerts(ctx *gin.Context) {
	// 获取库存阈值参数
	lowThresholdStr := ctx.DefaultQuery("low_threshold", "10")
	lowThreshold, err := strconv.ParseInt(lowThresholdStr, 10, 32)
	if err != nil || lowThreshold < 0 {
		lowThreshold = 10
	}

	// 获取商品数据
	pages4 := int32(1)
	pagePerNums4 := int32(10000)
	request := &gpbv1.GoodsFilterRequest{
		Pages:       &pages4,
		PagePerNums: &pagePerNums4,
	}

	response, err := ac.sf.Goods().GetGoodsList(ctx, request)
	if err != nil {
		log.Errorf("Failed to get goods list for inventory alerts: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	var outOfStock []map[string]interface{}
	var lowStock []map[string]interface{}

	for _, goods := range response.Data {
		goodsInfo := map[string]interface{}{
			"goodsId":   goods.Id,
			"name":       goods.Name,
			"goodsSn":   goods.GoodsSn,
			"stocks":     goods.Stocks,
			"price":      goods.ShopPrice,
			"category":   goods.Category.Name,
			"brand":      goods.Brand.Name,
			"onSale":    goods.OnSale,
			"soldNum":   goods.SoldNum,
		}

		if goods.Stocks == 0 {
			goodsInfo["alert_level"] = "critical"
			goodsInfo["alert_message"] = "商品已缺货"
			outOfStock = append(outOfStock, goodsInfo)
		} else if goods.Stocks <= int32(lowThreshold) {
			goodsInfo["alert_level"] = "warning"
			goodsInfo["alert_message"] = "库存不足，建议补货"
			lowStock = append(lowStock, goodsInfo)
		}
	}

	// 按库存数量排序（从低到高）
	sort.Slice(outOfStock, func(i, j int) bool {
		return outOfStock[i]["stocks"].(int32) < outOfStock[j]["stocks"].(int32)
	})
	sort.Slice(lowStock, func(i, j int) bool {
		return lowStock[i]["stocks"].(int32) < lowStock[j]["stocks"].(int32)
	})

	result := map[string]interface{}{
		"outOfStock": outOfStock,
		"lowStock":    lowStock,
		"summary": map[string]interface{}{
			"outOfStockCount": len(outOfStock),
			"lowStockCount":    len(lowStock),
			"totalAlerts":       len(outOfStock) + len(lowStock),
			"lowThreshold":      lowThreshold,
		},
	}

	core.WriteResponse(ctx, nil, result)
}