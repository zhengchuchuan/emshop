package import_controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	gpbv1 "emshop/api/goods/v1"
	ipbv1 "emshop/api/inventory/v1"
	restserver "emshop/gin-micro/server/rest-server"
	"emshop/internal/app/api/admin/service"
	"emshop/pkg/common/core"
	"emshop/pkg/log"

	"github.com/gin-gonic/gin"
)

type importController struct {
	trans restserver.I18nTranslator
	sf    service.ServiceFactory
}

func NewImportController(sf service.ServiceFactory, trans restserver.I18nTranslator) *importController {
	return &importController{
		sf:    sf,
		trans: trans,
	}
}

// ImportGoods 导入商品数据（管理员专用）
func (ic *importController) ImportGoods(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to get file from request"})
		return
	}
	defer file.Close()

	// 验证文件类型
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "only CSV files are supported"})
		return
	}

	// 解析CSV文件
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Errorf("Failed to parse CSV file: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse CSV file"})
		return
	}

	if len(records) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "CSV file must contain header and at least one data row"})
		return
	}

	// 验证表头（简化验证，实际应该更严格）
	headers := records[0]
	if len(headers) < 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "CSV file format invalid, missing required columns"})
		return
	}

	var results []map[string]interface{}
	var errors []string
	successCount := 0

	// 处理数据行（跳过表头）
	for i, record := range records[1:] {
		rowNum := i + 2 // CSV行号（从1开始，跳过表头）
		
		if len(record) < 6 {
			errors = append(errors, fmt.Sprintf("Row %d: insufficient columns", rowNum))
			continue
		}

		// 解析商品数据
		goodsInfo, err := ic.parseGoodsRecord(record, rowNum)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %v", rowNum, err))
			continue
		}

		// 创建商品
		goodsResp, err := ic.sf.Goods().CreateGoods(ctx, goodsInfo)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: failed to create goods - %v", rowNum, err))
			continue
		}

		// 设置库存（如果提供了库存数据）
		if len(record) > 6 && record[6] != "" {
			stocks, err := strconv.ParseInt(record[6], 10, 32)
			if err == nil && stocks > 0 {
				invInfo := &ipbv1.GoodsInvInfo{
					GoodsId: goodsResp.Id,
					Num:     int32(stocks),
				}
				if err := ic.sf.Goods().SetGoodsInventory(ctx, invInfo); err != nil {
					log.Warnf("Failed to set inventory for goods %d: %v", goodsResp.Id, err)
				}
			}
		}

		results = append(results, map[string]interface{}{
			"row":      rowNum,
			"goodsId": goodsResp.Id,
			"name":     goodsInfo.Name,
		})
		successCount++
	}

	response := map[string]interface{}{
		"success":       results,
		"errors":        errors,
		"total":         len(records) - 1, // 减去表头行
		"successCount": successCount,
		"errorCount":   len(errors),
	}

	core.WriteResponse(ctx, nil, response)
	log.Infof("Imported %d goods successfully, %d errors", successCount, len(errors))
}

// parseGoodsRecord 解析CSV记录为商品信息
func (ic *importController) parseGoodsRecord(record []string, rowNum int) (*gpbv1.CreateGoodsInfo, error) {
	// 必填字段验证
	if record[0] == "" {
		return nil, fmt.Errorf("goods name is required")
	}
	if record[1] == "" {
		return nil, fmt.Errorf("goods SN is required")
	}

	// 解析分类ID
	categoryId, err := strconv.ParseInt(record[2], 10, 32)
	if err != nil || categoryId <= 0 {
		return nil, fmt.Errorf("invalid category ID")
	}

	// 解析品牌ID
	brandId, err := strconv.ParseInt(record[3], 10, 32)
	if err != nil || brandId <= 0 {
		return nil, fmt.Errorf("invalid brand ID")
	}

	// 解析市场价
	marketPrice, err := strconv.ParseFloat(record[4], 32)
	if err != nil || marketPrice <= 0 {
		return nil, fmt.Errorf("invalid market price")
	}

	// 解析销售价
	shopPrice, err := strconv.ParseFloat(record[5], 32)
	if err != nil || shopPrice <= 0 {
		return nil, fmt.Errorf("invalid shop price")
	}

	// 构建商品信息
	goodsInfo := &gpbv1.CreateGoodsInfo{
		Name:        record[0],
		GoodsSn:     record[1],
		CategoryId:  int32(categoryId),
		BrandId:     int32(brandId),
		MarketPrice: float32(marketPrice),
		ShopPrice:   float32(shopPrice),
		OnSale:      true, // 默认上架
	}

	// 解析可选字段
	if len(record) > 7 && record[7] != "" {
		isNew, err := strconv.ParseBool(record[7])
		if err == nil {
			goodsInfo.IsNew = isNew
		}
	}

	if len(record) > 8 && record[8] != "" {
		isHot, err := strconv.ParseBool(record[8])
		if err == nil {
			goodsInfo.IsHot = isHot
		}
	}

	if len(record) > 9 && record[9] != "" {
		onSale, err := strconv.ParseBool(record[9])
		if err == nil {
			goodsInfo.OnSale = onSale
		}
	}

	if len(record) > 10 && record[10] != "" {
		shipFree, err := strconv.ParseBool(record[10])
		if err == nil {
			goodsInfo.ShipFree = shipFree
		}
	}

	if len(record) > 11 && record[11] != "" {
		goodsInfo.GoodsBrief = record[11]
	}

	if len(record) > 12 && record[12] != "" {
		goodsInfo.GoodsDesc = record[12]
	}

	if len(record) > 13 && record[13] != "" {
		goodsInfo.GoodsFrontImage = record[13]
	}

	if len(record) > 14 && record[14] != "" {
		images := strings.Split(record[14], ",")
		for i, img := range images {
			images[i] = strings.TrimSpace(img)
		}
		goodsInfo.Images = images
	}

	return goodsInfo, nil
}

// ValidateImportFile 验证导入文件格式（管理员专用）
func (ic *importController) ValidateImportFile(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to get file from request"})
		return
	}
	defer file.Close()

	// 验证文件类型
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "only CSV files are supported"})
		return
	}

	// 解析CSV文件
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "failed to parse CSV file"})
		return
	}

	if len(records) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "CSV file must contain header and at least one data row"})
		return
	}

	var errors []string
	validRows := 0

	// 验证数据行
	for i, record := range records[1:] {
		rowNum := i + 2
		
		_, err := ic.parseGoodsRecord(record, rowNum)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %v", rowNum, err))
		} else {
			validRows++
		}
	}

	response := map[string]interface{}{
		"valid":       len(errors) == 0,
		"totalRows":  len(records) - 1,
		"validRows":  validRows,
		"errorRows":  len(errors),
		"errors":      errors,
		"filename":    header.Filename,
		"fileSize":   header.Size,
	}

	core.WriteResponse(ctx, nil, response)
}