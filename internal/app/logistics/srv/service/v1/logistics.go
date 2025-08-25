package v1

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"emshop/internal/app/logistics/srv/data/v1/interfaces"
	"emshop/internal/app/logistics/srv/domain/do"
	"emshop/internal/app/logistics/srv/domain/dto"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

// LogisticsSrv 物流服务接口
type LogisticsSrv interface {
	// 核心物流管理功能
	CreateLogisticsOrder(ctx context.Context, req *dto.CreateLogisticsOrderDTO) (*dto.LogisticsOrderDTO, error)
	GetLogisticsInfo(ctx context.Context, req *dto.GetLogisticsInfoDTO) (*dto.LogisticsInfoDTO, error)
	GetLogisticsTracks(ctx context.Context, req *dto.GetLogisticsTracksDTO) (*dto.LogisticsTracksDTO, error)
	UpdateLogisticsStatus(ctx context.Context, req *dto.UpdateLogisticsStatusDTO) error
	
	// 模拟操作功能
	SimulateShipment(ctx context.Context, req *dto.SimulateShipmentDTO) error
	SimulateDelivery(ctx context.Context, req *dto.SimulateDeliveryDTO) error
	
	// 工具功能
	CalculateShippingFee(ctx context.Context, req *dto.CalculateShippingFeeDTO) (*dto.ShippingFeeDTO, error)
	GetLogisticsCompanies(ctx context.Context) ([]dto.LogisticsCompanyDTO, error)
	GetCouriers(ctx context.Context, req *dto.GetCouriersDTO) ([]dto.CourierDTO, error)
}

type logisticsService struct {
	data         interfaces.DataFactory
	redisOptions *options.RedisOptions
}

// NewLogisticsService 创建物流服务实例
func NewLogisticsService(data interfaces.DataFactory, redisOpts *options.RedisOptions) LogisticsSrv {
	return &logisticsService{
		data:         data,
		redisOptions: redisOpts,
	}
}

// 物流公司信息映射
var logisticsCompanies = map[int32]dto.LogisticsCompanyDTO{
	1: {CompanyID: 1, CompanyName: "圆通速递", CompanyCode: "YTO"},
	2: {CompanyID: 2, CompanyName: "申通快递", CompanyCode: "STO"},
	3: {CompanyID: 3, CompanyName: "中通快递", CompanyCode: "ZTO"},
	4: {CompanyID: 4, CompanyName: "韵达速递", CompanyCode: "YD"},
	5: {CompanyID: 5, CompanyName: "顺丰速运", CompanyCode: "SF"},
	6: {CompanyID: 6, CompanyName: "京东物流", CompanyCode: "JD"},
	7: {CompanyID: 7, CompanyName: "中国邮政", CompanyCode: "EMS"},
}

// generateLogisticsSn 生成物流单号
func (ls *logisticsService) generateLogisticsSn() string {
	timestamp := time.Now().Format("20060102150405")
	n, _ := rand.Int(rand.Reader, big.NewInt(9999))
	return fmt.Sprintf("LG%s%04d", timestamp, n.Int64())
}

// generateTrackingNumber 生成快递单号
func (ls *logisticsService) generateTrackingNumber(company int32) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(999999999999))
	switch company {
	case 5: // 顺丰
		return fmt.Sprintf("SF%012d", n.Int64())
	case 6: // 京东
		return fmt.Sprintf("JD%013d", n.Int64())
	case 7: // 邮政
		return fmt.Sprintf("EMS%011d", n.Int64())
	default:
		return fmt.Sprintf("EX%012d", n.Int64())
	}
}

// CreateLogisticsOrder 创建物流订单
func (ls *logisticsService) CreateLogisticsOrder(ctx context.Context, req *dto.CreateLogisticsOrderDTO) (*dto.LogisticsOrderDTO, error) {
	log.Infof("创建物流订单: 订单号=%s, 用户ID=%d", req.OrderSn, req.UserID)

	// 检查该订单是否已经创建物流订单
	existingOrder, err := ls.data.LogisticsOrders().GetByOrderSn(ctx, ls.data.DB(), req.OrderSn)
	if err != nil && !errors.IsCode(err, code.ErrLogisticsOrderNotFound) {
		return nil, err
	}
	if existingOrder != nil {
		return nil, errors.WithCode(code.ErrLogisticsOrderExists, "该订单已存在物流订单")
	}

	// 生成物流单号和快递单号
	logisticsSn := ls.generateLogisticsSn()
	trackingNumber := ls.generateTrackingNumber(req.LogisticsCompany)

	// 计算运费
	feeReq := &dto.CalculateShippingFeeDTO{
		SenderAddress:   req.SenderAddress,
		ReceiverAddress: req.ReceiverAddress,
		ShippingMethod:  req.ShippingMethod,
		TotalWeight:     ls.calculateTotalWeight(req.Items),
		TotalVolume:     ls.calculateTotalVolume(req.Items),
	}
	feeResp, err := ls.CalculateShippingFee(ctx, feeReq)
	if err != nil {
		return nil, err
	}

	// 计算预计送达时间
	estimatedDelivery := ls.calculateEstimatedDelivery(req.ShippingMethod)

	// 序列化商品信息
	itemsJSON, _ := json.Marshal(req.Items)

	// 创建物流订单
	order := &do.LogisticsOrderDO{
		LogisticsSn:         logisticsSn,
		OrderSn:             req.OrderSn,
		UserID:              req.UserID,
		LogisticsCompany:    req.LogisticsCompany,
		ShippingMethod:      req.ShippingMethod,
		TrackingNumber:      trackingNumber,
		LogisticsStatus:     int32(do.LogisticsStatusPending),
		SenderName:          req.SenderName,
		SenderPhone:         req.SenderPhone,
		SenderAddress:       req.SenderAddress,
		ReceiverName:        req.ReceiverName,
		ReceiverPhone:       req.ReceiverPhone,
		ReceiverAddress:     req.ReceiverAddress,
		EstimatedDeliveryAt: &estimatedDelivery,
		ShippingFee:         feeResp.ShippingFee,
		InsuranceFee:        feeResp.InsuranceFee,
		ItemsInfo:           string(itemsJSON),
		Remark:              req.Remark,
	}

	// 开启事务
	tx := ls.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 创建物流订单
	if err := ls.data.LogisticsOrders().Create(ctx, tx, order); err != nil {
		tx.Rollback()
		return nil, errors.WithCode(code.ErrCreateLogisticsOrderFailed, "创建物流订单失败")
	}

	// 创建初始轨迹记录
	initialTrack := &do.LogisticsTrackDO{
		LogisticsSn:    logisticsSn,
		TrackingNumber: trackingNumber,
		Location:       "商家仓库",
		Description:    "商家正在准备发货",
		TrackTime:      time.Now(),
		OperatorName:   "系统",
	}
	if err := ls.data.LogisticsTracks().Create(ctx, tx, initialTrack); err != nil {
		log.Warnf("创建初始物流轨迹失败: %v", err)
		// 轨迹失败不影响主流程
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	log.Infof("物流订单创建成功: 物流单号=%s, 快递单号=%s", logisticsSn, trackingNumber)

	return &dto.LogisticsOrderDTO{
		LogisticsSn:         logisticsSn,
		TrackingNumber:      trackingNumber,
		ShippingFee:         feeResp.TotalFee,
		EstimatedDeliveryAt: estimatedDelivery,
	}, nil
}

// GetLogisticsInfo 查询物流信息
func (ls *logisticsService) GetLogisticsInfo(ctx context.Context, req *dto.GetLogisticsInfoDTO) (*dto.LogisticsInfoDTO, error) {
	var order *do.LogisticsOrderDO
	var err error

	// 根据不同条件查询
	if req.LogisticsSn != nil {
		order, err = ls.data.LogisticsOrders().GetByLogisticsSn(ctx, ls.data.DB(), *req.LogisticsSn)
	} else if req.OrderSn != nil {
		order, err = ls.data.LogisticsOrders().GetByOrderSn(ctx, ls.data.DB(), *req.OrderSn)
	} else if req.TrackingNumber != nil {
		order, err = ls.data.LogisticsOrders().GetByTrackingNumber(ctx, ls.data.DB(), *req.TrackingNumber)
	} else {
		return nil, errors.WithCode(code.ErrInvalidShippingAddress, "查询条件不能为空")
	}

	if err != nil {
		return nil, err
	}

	return &dto.LogisticsInfoDTO{
		LogisticsSn:         order.LogisticsSn,
		OrderSn:             order.OrderSn,
		TrackingNumber:      order.TrackingNumber,
		LogisticsCompany:    order.LogisticsCompany,
		ShippingMethod:      order.ShippingMethod,
		LogisticsStatus:     order.LogisticsStatus,
		SenderName:          order.SenderName,
		SenderPhone:         order.SenderPhone,
		SenderAddress:       order.SenderAddress,
		ReceiverName:        order.ReceiverName,
		ReceiverPhone:       order.ReceiverPhone,
		ReceiverAddress:     order.ReceiverAddress,
		ShippingFee:         order.ShippingFee,
		ShippedAt:           order.ShippedAt,
		DeliveredAt:         order.DeliveredAt,
		EstimatedDeliveryAt: order.EstimatedDeliveryAt,
		Remark:              order.Remark,
	}, nil
}

// GetLogisticsTracks 查询物流轨迹
func (ls *logisticsService) GetLogisticsTracks(ctx context.Context, req *dto.GetLogisticsTracksDTO) (*dto.LogisticsTracksDTO, error) {
	var tracks []*do.LogisticsTrackDO
	var err error
	var logisticsSn, trackingNumber string

	// 根据不同条件查询
	if req.LogisticsSn != nil {
		tracks, err = ls.data.LogisticsTracks().GetByLogisticsSn(ctx, ls.data.DB(), *req.LogisticsSn)
		logisticsSn = *req.LogisticsSn
		// 获取快递单号
		if order, _ := ls.data.LogisticsOrders().GetByLogisticsSn(ctx, ls.data.DB(), *req.LogisticsSn); order != nil {
			trackingNumber = order.TrackingNumber
		}
	} else if req.TrackingNumber != nil {
		tracks, err = ls.data.LogisticsTracks().GetByTrackingNumber(ctx, ls.data.DB(), *req.TrackingNumber)
		trackingNumber = *req.TrackingNumber
		// 获取物流单号
		if order, _ := ls.data.LogisticsOrders().GetByTrackingNumber(ctx, ls.data.DB(), *req.TrackingNumber); order != nil {
			logisticsSn = order.LogisticsSn
		}
	} else {
		return nil, errors.WithCode(code.ErrInvalidShippingAddress, "查询条件不能为空")
	}

	if err != nil {
		return nil, err
	}

	// 转换为DTO
	trackDTOs := make([]dto.LogisticsTrackDTO, len(tracks))
	for i, track := range tracks {
		trackDTOs[i] = dto.LogisticsTrackDTO{
			Location:     track.Location,
			Description:  track.Description,
			TrackTime:    track.TrackTime,
			OperatorName: track.OperatorName,
		}
	}

	return &dto.LogisticsTracksDTO{
		LogisticsSn:    logisticsSn,
		TrackingNumber: trackingNumber,
		Tracks:         trackDTOs,
	}, nil
}

// UpdateLogisticsStatus 更新物流状态
func (ls *logisticsService) UpdateLogisticsStatus(ctx context.Context, req *dto.UpdateLogisticsStatusDTO) error {
	log.Infof("更新物流状态: 物流单号=%s, 新状态=%d", req.LogisticsSn, req.NewStatus)
	
	return ls.data.LogisticsOrders().UpdateStatus(ctx, ls.data.DB(), req.LogisticsSn, req.NewStatus)
}

// SimulateShipment 模拟发货
func (ls *logisticsService) SimulateShipment(ctx context.Context, req *dto.SimulateShipmentDTO) error {
	log.Infof("模拟发货: 物流单号=%s", req.LogisticsSn)

	// 获取物流订单
	order, err := ls.data.LogisticsOrders().GetByLogisticsSn(ctx, ls.data.DB(), req.LogisticsSn)
	if err != nil {
		return err
	}

	if order.LogisticsStatus != int32(do.LogisticsStatusPending) {
		return errors.WithCode(code.ErrLogisticsStatusTransitionInvalid, "物流状态不允许发货操作")
	}

	// 开启事务
	tx := ls.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 更新发货信息
	now := time.Now()
	if err := ls.data.LogisticsOrders().UpdateShipmentInfo(ctx, tx, req.LogisticsSn, &now); err != nil {
		tx.Rollback()
		return err
	}

	// 创建发货轨迹
	courier := req.CourierName
	if courier == "" {
		courier = "配送员"
	}
	
	shipmentTrack := &do.LogisticsTrackDO{
		LogisticsSn:    req.LogisticsSn,
		TrackingNumber: order.TrackingNumber,
		Location:       "商家仓库",
		Description:    fmt.Sprintf("快件已发出，配送员：%s", courier),
		TrackTime:      now,
		OperatorName:   courier,
	}
	
	if err := ls.data.LogisticsTracks().Create(ctx, tx, shipmentTrack); err != nil {
		log.Warnf("创建发货轨迹失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	// 异步生成轨迹
	go ls.generateLogisticsTracks(req.LogisticsSn)

	return nil
}

// SimulateDelivery 模拟签收
func (ls *logisticsService) SimulateDelivery(ctx context.Context, req *dto.SimulateDeliveryDTO) error {
	log.Infof("模拟签收: 物流单号=%s", req.LogisticsSn)

	// 获取物流订单
	order, err := ls.data.LogisticsOrders().GetByLogisticsSn(ctx, ls.data.DB(), req.LogisticsSn)
	if err != nil {
		return err
	}

	if order.LogisticsStatus == int32(do.LogisticsStatusDelivered) {
		return nil // 已经签收，幂等性
	}

	// 开启事务
	tx := ls.data.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 更新签收信息
	now := time.Now()
	if err := ls.data.LogisticsOrders().UpdateDeliveryInfo(ctx, tx, req.LogisticsSn, &now); err != nil {
		tx.Rollback()
		return err
	}

	// 创建签收轨迹
	receiverName := req.ReceiverName
	if receiverName == "" {
		receiverName = order.ReceiverName
	}
	
	deliveryTrack := &do.LogisticsTrackDO{
		LogisticsSn:    req.LogisticsSn,
		TrackingNumber: order.TrackingNumber,
		Location:       order.ReceiverAddress,
		Description:    fmt.Sprintf("快件已签收，签收人：%s", receiverName),
		TrackTime:      now,
		OperatorName:   receiverName,
	}
	
	if err := ls.data.LogisticsTracks().Create(ctx, tx, deliveryTrack); err != nil {
		log.Warnf("创建签收轨迹失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithCode(code.ErrConnectDB, "提交事务失败")
	}

	return nil
}

// CalculateShippingFee 计算运费
func (ls *logisticsService) CalculateShippingFee(ctx context.Context, req *dto.CalculateShippingFeeDTO) (*dto.ShippingFeeDTO, error) {
	// 模拟运费计算逻辑
	distance := ls.calculateDistance(req.SenderAddress, req.ReceiverAddress)
	baseFee := ls.calculateBaseFee(distance, req.TotalWeight)
	
	// 配送方式加成
	methodMultiplier := ls.getMethodMultiplier(req.ShippingMethod)
	shippingFee := baseFee * methodMultiplier
	
	// 保价费计算
	var insuranceFee float64
	if req.NeedInsurance {
		insuranceFee = req.GoodsValue * 0.005 // 0.5%保价费率
	}
	
	// 预计天数
	estimatedDays := ls.getEstimatedDays(req.ShippingMethod, distance)
	
	return &dto.ShippingFeeDTO{
		ShippingFee:   shippingFee,
		InsuranceFee:  insuranceFee,
		TotalFee:      shippingFee + insuranceFee,
		EstimatedDays: estimatedDays,
	}, nil
}

// GetLogisticsCompanies 获取物流公司列表
func (ls *logisticsService) GetLogisticsCompanies(ctx context.Context) ([]dto.LogisticsCompanyDTO, error) {
	companies := make([]dto.LogisticsCompanyDTO, 0, len(logisticsCompanies))
	for _, company := range logisticsCompanies {
		companies = append(companies, company)
	}
	return companies, nil
}

// GetCouriers 获取配送员列表
func (ls *logisticsService) GetCouriers(ctx context.Context, req *dto.GetCouriersDTO) ([]dto.CourierDTO, error) {
	var couriers []*do.LogisticsCourierDO
	var err error

	if req.LogisticsCompany != nil {
		couriers, err = ls.data.LogisticsCouriers().GetByCompany(ctx, ls.data.DB(), *req.LogisticsCompany)
	} else if req.ServiceArea != nil {
		couriers, err = ls.data.LogisticsCouriers().GetByServiceArea(ctx, ls.data.DB(), *req.ServiceArea)
	} else {
		// 获取所有配送员
		couriers, _, err = ls.data.LogisticsCouriers().List(ctx, ls.data.DB(), 0, 100, nil, nil)
	}

	if err != nil {
		return nil, err
	}

	// 转换为DTO
	courierDTOs := make([]dto.CourierDTO, len(couriers))
	for i, courier := range couriers {
		courierDTOs[i] = dto.CourierDTO{
			CourierCode:      courier.CourierCode,
			CourierName:      courier.CourierName,
			Phone:            courier.Phone,
			LogisticsCompany: courier.LogisticsCompany,
			ServiceArea:      courier.ServiceArea,
		}
	}

	return courierDTOs, nil
}

// 辅助方法

func (ls *logisticsService) calculateTotalWeight(items []dto.OrderItemDTO) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Weight * float64(item.Quantity)
	}
	if total == 0 {
		return 1.0 // 默认重量
	}
	return total
}

func (ls *logisticsService) calculateTotalVolume(items []dto.OrderItemDTO) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Volume * float64(item.Quantity)
	}
	return total
}

func (ls *logisticsService) calculateEstimatedDelivery(method int32) time.Time {
	hours := 72 // 默认3天
	switch method {
	case 2: // 急速配送
		hours = 24
	case 3: // 经济配送
		hours = 120
	case 4: // 自提
		hours = 48
	}
	return time.Now().Add(time.Duration(hours) * time.Hour)
}

func (ls *logisticsService) calculateDistance(senderAddr, receiverAddr string) float64 {
	// 简单的距离计算模拟（根据地址关键词判断）
	if strings.Contains(senderAddr, "北京") && strings.Contains(receiverAddr, "北京") {
		return 50 // 同城
	} else if strings.Contains(senderAddr, "北京") || strings.Contains(receiverAddr, "北京") {
		return 800 // 跨省
	}
	return 200 // 默认距离
}

func (ls *logisticsService) calculateBaseFee(distance, weight float64) float64 {
	baseFee := 8.0 // 首重费用
	if weight > 1.0 {
		baseFee += (weight - 1.0) * 2.0 // 续重费用
	}
	if distance > 100 {
		baseFee += (distance - 100) * 0.01 // 远距离加费
	}
	return baseFee
}

func (ls *logisticsService) getMethodMultiplier(method int32) float64 {
	switch method {
	case 2: // 急速配送
		return 2.0
	case 3: // 经济配送
		return 0.8
	default:
		return 1.0
	}
}

func (ls *logisticsService) getEstimatedDays(method int32, distance float64) int32 {
	baseDays := int32(1)
	if distance > 500 {
		baseDays = 3
	} else if distance > 100 {
		baseDays = 2
	}
	
	switch method {
	case 2: // 急速配送
		return baseDays
	case 3: // 经济配送
		return baseDays + 2
	default:
		return baseDays + 1
	}
}

// generateLogisticsTracks 异步生成模拟物流轨迹
func (ls *logisticsService) generateLogisticsTracks(logisticsSn string) {
	// 这是一个简化的轨迹生成，实际项目中可以使用更复杂的模拟逻辑
	time.Sleep(2 * time.Hour) // 模拟2小时后到达集散中心
	
	// 可以在这里添加更多轨迹生成逻辑
	log.Infof("异步生成物流轨迹: %s", logisticsSn)
}