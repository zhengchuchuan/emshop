package export

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	restserver "emshop/gin-micro/server/rest-server"
	gpbv1 "emshop/api/goods/v1"
	"emshop/internal/app/emshop/admin/service"
	"emshop/pkg/common/core"
	"emshop/pkg/log"
)

type exportController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

func NewExportController(sf service.ServiceFactory, trans restserver.I18nTranslator) *exportController {
	return &exportController{
		sf:    sf,
		trans: trans,
	}
}

// ExportGoods 导出商品数据为CSV（管理员专用）
func (ec *exportController) ExportGoods(ctx *gin.Context) {
	// 解析查询参数
	var r struct {
		IsNew       *bool   `form:"isNew"`
		IsHot       *bool   `form:"isHot"`
		PriceMax    *int32  `form:"priceMax"`
		PriceMin    *int32  `form:"priceMin"`
		TopCategory *int32  `form:"topCategory"`
		Brand       *int32  `form:"brand"`
		KeyWords    *string `form:"keyWords"`
	}

	if err := ctx.ShouldBindQuery(&r); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid query parameters"})
		return
	}

	// 构建查询请求
	pages := int32(1)
	pagePerNums := int32(10000)
	request := &gpbv1.GoodsFilterRequest{
		Pages:       &pages,
		PagePerNums: &pagePerNums, // 导出时设置较大的分页数
	}
	
	if r.IsNew != nil {
		request.IsNew = r.IsNew
	}
	if r.IsHot != nil {
		request.IsHot = r.IsHot
	}
	if r.PriceMax != nil {
		request.PriceMax = r.PriceMax
	}
	if r.PriceMin != nil {
		request.PriceMin = r.PriceMin
	}
	if r.TopCategory != nil {
		request.TopCategory = r.TopCategory
	}
	if r.Brand != nil {
		request.Brand = r.Brand
	}
	if r.KeyWords != nil {
		request.KeyWords = r.KeyWords
	}

	// 获取商品数据
	response, err := ec.sf.Goods().GetGoodsList(ctx, request)
	if err != nil {
		log.Errorf("Failed to get goods list for export: %v", err)
		core.WriteResponse(ctx, err, nil)
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("goods_export_%s.csv", time.Now().Format("20060102_150405"))
	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 创建CSV writer
	writer := csv.NewWriter(ctx.Writer)
	defer writer.Flush()

	// 写入表头
	headers := []string{
		"商品ID", "商品名称", "商品编号", "分类", "品牌", "市场价", "销售价", 
		"库存", "点击数", "销量", "收藏数", "是否新品", "是否热销", "是否上架", 
		"是否免邮", "商品简介", "创建时间",
	}
	if err := writer.Write(headers); err != nil {
		log.Errorf("Failed to write CSV headers: %v", err)
		return
	}

	// 写入数据
	for _, goods := range response.Data {
		record := []string{
			strconv.Itoa(int(goods.Id)),
			goods.Name,
			goods.GoodsSn,
			goods.Category.Name,
			goods.Brand.Name,
			fmt.Sprintf("%.2f", goods.MarketPrice),
			fmt.Sprintf("%.2f", goods.ShopPrice),
			strconv.Itoa(int(goods.Stocks)),
			strconv.Itoa(int(goods.ClickNum)),
			strconv.Itoa(int(goods.SoldNum)),
			strconv.Itoa(int(goods.FavNum)),
			strconv.FormatBool(goods.IsNew),
			strconv.FormatBool(goods.IsHot),
			strconv.FormatBool(goods.OnSale),
			strconv.FormatBool(goods.ShipFree),
			goods.GoodsBrief,
			time.Unix(goods.AddTime, 0).Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(record); err != nil {
			log.Errorf("Failed to write CSV record: %v", err)
			return
		}
	}

	log.Infof("Exported %d goods records to CSV", len(response.Data))
}

// ExportGoodsTemplate 导出商品模板文件（管理员专用）
func (ec *exportController) ExportGoodsTemplate(ctx *gin.Context) {
	// 设置响应头
	filename := "goods_import_template.csv"
	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// 创建CSV writer
	writer := csv.NewWriter(ctx.Writer)
	defer writer.Flush()

	// 写入模板表头
	headers := []string{
		"商品名称*", "商品编号*", "分类ID*", "品牌ID*", "市场价*", "销售价*", 
		"库存*", "是否新品(true/false)", "是否热销(true/false)", "是否上架(true/false)", 
		"是否免邮(true/false)", "商品简介", "商品描述", "商品主图URL", "商品图片URLs(逗号分隔)",
	}
	if err := writer.Write(headers); err != nil {
		log.Errorf("Failed to write template headers: %v", err)
		return
	}

	// 写入示例数据
	examples := [][]string{
		{
			"iPhone 14", "IP14001", "1", "1", "6999.00", "6499.00", 
			"100", "true", "true", "true", "true", 
			"Apple iPhone 14 智能手机", "详细的商品描述", 
			"/uploads/goods/images/iphone14.jpg", 
			"/uploads/goods/images/iphone14_1.jpg,/uploads/goods/images/iphone14_2.jpg",
		},
		{
			"华为P50", "HWP50001", "1", "2", "4999.00", "4599.00", 
			"50", "false", "true", "true", "false", 
			"华为P50 智能手机", "详细的商品描述", 
			"/uploads/goods/images/huaweip50.jpg", 
			"/uploads/goods/images/huaweip50_1.jpg",
		},
	}

	for _, example := range examples {
		if err := writer.Write(example); err != nil {
			log.Errorf("Failed to write template example: %v", err)
			return
		}
	}

	log.Info("Exported goods import template")
}