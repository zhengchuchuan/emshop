package v1

import (
	"context"
	"emshop/internal/app/logistics/srv/domain/dto"
	"emshop/internal/app/logistics/srv/service/v1"
	logisticspb "emshop/api/logistics/v1"
	"emshop/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// LogisticsController 物流服务gRPC控制器
type LogisticsController struct {
	logisticspb.UnimplementedLogisticsServer
	logisticsSrv v1.LogisticsSrv
}

// NewLogisticsController 创建物流控制器实例
func NewLogisticsController(logisticsSrv v1.LogisticsSrv) *LogisticsController {
	return &LogisticsController{
		logisticsSrv: logisticsSrv,
	}
}

// CreateLogisticsOrder 创建物流订单
func (lc *LogisticsController) CreateLogisticsOrder(ctx context.Context, req *logisticspb.CreateLogisticsOrderRequest) (*logisticspb.CreateLogisticsOrderResponse, error) {
	log.Infof("收到创建物流订单请求: 订单号=%s", req.OrderSn)

	// 转换为DTO
	items := make([]dto.OrderItemDTO, len(req.Items))
	for i, item := range req.Items {
		items[i] = dto.OrderItemDTO{
			GoodsID:  item.GoodsId,
			Name:     item.GoodsName,
			Quantity: item.Quantity,
			Weight:   item.Weight,
			Volume:   item.Volume,
		}
	}

	reqDTO := &dto.CreateLogisticsOrderDTO{
		OrderSn:          req.OrderSn,
		UserID:           req.UserId,
		LogisticsCompany: req.LogisticsCompany,
		ShippingMethod:   req.ShippingMethod,
		SenderName:       req.SenderName,
		SenderPhone:      req.SenderPhone,
		SenderAddress:    req.SenderAddress,
		ReceiverName:     req.ReceiverName,
		ReceiverPhone:    req.ReceiverPhone,
		ReceiverAddress:  req.ReceiverAddress,
		Items:            items,
		Remark:           req.Remark,
	}

	// 调用服务层
	resp, err := lc.logisticsSrv.CreateLogisticsOrder(ctx, reqDTO)
	if err != nil {
		log.Errorf("创建物流订单失败: %v", err)
		return nil, err
	}

	return &logisticspb.CreateLogisticsOrderResponse{
		LogisticsSn:         resp.LogisticsSn,
		TrackingNumber:      resp.TrackingNumber,
		ShippingFee:         resp.ShippingFee,
		EstimatedDeliveryAt: resp.EstimatedDeliveryAt.Unix(),
	}, nil
}

// GetLogisticsInfo 查询物流信息
func (lc *LogisticsController) GetLogisticsInfo(ctx context.Context, req *logisticspb.GetLogisticsInfoRequest) (*logisticspb.GetLogisticsInfoResponse, error) {
	// 转换查询条件
	reqDTO := &dto.GetLogisticsInfoDTO{}
	
	switch q := req.Query.(type) {
	case *logisticspb.GetLogisticsInfoRequest_LogisticsSn:
		reqDTO.LogisticsSn = &q.LogisticsSn
	case *logisticspb.GetLogisticsInfoRequest_OrderSn:
		reqDTO.OrderSn = &q.OrderSn
	case *logisticspb.GetLogisticsInfoRequest_TrackingNumber:
		reqDTO.TrackingNumber = &q.TrackingNumber
	}

	// 调用服务层
	info, err := lc.logisticsSrv.GetLogisticsInfo(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	// 构建响应
	resp := &logisticspb.GetLogisticsInfoResponse{
		LogisticsSn:      info.LogisticsSn,
		OrderSn:          info.OrderSn,
		TrackingNumber:   info.TrackingNumber,
		LogisticsCompany: info.LogisticsCompany,
		ShippingMethod:   info.ShippingMethod,
		LogisticsStatus:  info.LogisticsStatus,
		SenderName:       info.SenderName,
		SenderPhone:      info.SenderPhone,
		SenderAddress:    info.SenderAddress,
		ReceiverName:     info.ReceiverName,
		ReceiverPhone:    info.ReceiverPhone,
		ReceiverAddress:  info.ReceiverAddress,
		ShippingFee:      info.ShippingFee,
		Remark:           info.Remark,
	}

	if info.ShippedAt != nil {
		resp.ShippedAt = info.ShippedAt.Unix()
	}
	if info.DeliveredAt != nil {
		resp.DeliveredAt = info.DeliveredAt.Unix()
	}
	if info.EstimatedDeliveryAt != nil {
		resp.EstimatedDeliveryAt = info.EstimatedDeliveryAt.Unix()
	}

	return resp, nil
}

// GetLogisticsTracks 查询物流轨迹
func (lc *LogisticsController) GetLogisticsTracks(ctx context.Context, req *logisticspb.GetLogisticsTracksRequest) (*logisticspb.GetLogisticsTracksResponse, error) {
	// 转换查询条件
	reqDTO := &dto.GetLogisticsTracksDTO{}
	
	switch q := req.Query.(type) {
	case *logisticspb.GetLogisticsTracksRequest_LogisticsSn:
		reqDTO.LogisticsSn = &q.LogisticsSn
	case *logisticspb.GetLogisticsTracksRequest_TrackingNumber:
		reqDTO.TrackingNumber = &q.TrackingNumber
	}

	// 调用服务层
	tracks, err := lc.logisticsSrv.GetLogisticsTracks(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	// 转换轨迹数据
	pbTracks := make([]*logisticspb.LogisticsTrack, len(tracks.Tracks))
	for i, track := range tracks.Tracks {
		pbTracks[i] = &logisticspb.LogisticsTrack{
			Location:     track.Location,
			Description:  track.Description,
			TrackTime:    track.TrackTime.Unix(),
			OperatorName: track.OperatorName,
		}
	}

	return &logisticspb.GetLogisticsTracksResponse{
		LogisticsSn:    tracks.LogisticsSn,
		TrackingNumber: tracks.TrackingNumber,
		Tracks:         pbTracks,
	}, nil
}

// UpdateLogisticsStatus 更新物流状态
func (lc *LogisticsController) UpdateLogisticsStatus(ctx context.Context, req *logisticspb.UpdateLogisticsStatusRequest) (*emptypb.Empty, error) {
	reqDTO := &dto.UpdateLogisticsStatusDTO{
		LogisticsSn: req.LogisticsSn,
		NewStatus:   req.NewStatus,
		Remark:      req.Remark,
	}

	err := lc.logisticsSrv.UpdateLogisticsStatus(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SimulateShipment 模拟发货
func (lc *LogisticsController) SimulateShipment(ctx context.Context, req *logisticspb.SimulateShipmentRequest) (*emptypb.Empty, error) {
	reqDTO := &dto.SimulateShipmentDTO{
		LogisticsSn:  req.LogisticsSn,
		CourierName:  req.CourierName,
		CourierPhone: req.CourierPhone,
	}

	err := lc.logisticsSrv.SimulateShipment(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SimulateDelivery 模拟签收
func (lc *LogisticsController) SimulateDelivery(ctx context.Context, req *logisticspb.SimulateDeliveryRequest) (*emptypb.Empty, error) {
	reqDTO := &dto.SimulateDeliveryDTO{
		LogisticsSn:    req.LogisticsSn,
		ReceiverName:   req.ReceiverName,
		DeliveryRemark: req.DeliveryRemark,
	}

	err := lc.logisticsSrv.SimulateDelivery(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CalculateShippingFee 计算运费
func (lc *LogisticsController) CalculateShippingFee(ctx context.Context, req *logisticspb.CalculateShippingFeeRequest) (*logisticspb.CalculateShippingFeeResponse, error) {
	reqDTO := &dto.CalculateShippingFeeDTO{
		SenderAddress:   req.SenderAddress,
		ReceiverAddress: req.ReceiverAddress,
		ShippingMethod:  req.ShippingMethod,
		TotalWeight:     req.TotalWeight,
		TotalVolume:     req.TotalVolume,
		GoodsValue:      req.GoodsValue,
		NeedInsurance:   req.NeedInsurance,
	}

	fee, err := lc.logisticsSrv.CalculateShippingFee(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	return &logisticspb.CalculateShippingFeeResponse{
		ShippingFee:   fee.ShippingFee,
		InsuranceFee:  fee.InsuranceFee,
		TotalFee:      fee.TotalFee,
		EstimatedDays: fee.EstimatedDays,
	}, nil
}

// GetLogisticsCompanies 获取物流公司列表
func (lc *LogisticsController) GetLogisticsCompanies(ctx context.Context, req *emptypb.Empty) (*logisticspb.LogisticsCompaniesResponse, error) {
	companies, err := lc.logisticsSrv.GetLogisticsCompanies(ctx)
	if err != nil {
		return nil, err
	}

	pbCompanies := make([]*logisticspb.LogisticsCompany, len(companies))
	for i, company := range companies {
		pbCompanies[i] = &logisticspb.LogisticsCompany{
			CompanyId:   company.CompanyID,
			CompanyName: company.CompanyName,
			CompanyCode: company.CompanyCode,
		}
	}

	return &logisticspb.LogisticsCompaniesResponse{
		Companies: pbCompanies,
	}, nil
}

// GetCouriers 获取配送员列表
func (lc *LogisticsController) GetCouriers(ctx context.Context, req *logisticspb.GetCouriersRequest) (*logisticspb.GetCouriersResponse, error) {
	reqDTO := &dto.GetCouriersDTO{}
	if req.LogisticsCompany != 0 {
		reqDTO.LogisticsCompany = &req.LogisticsCompany
	}
	if req.ServiceArea != "" {
		reqDTO.ServiceArea = &req.ServiceArea
	}

	couriers, err := lc.logisticsSrv.GetCouriers(ctx, reqDTO)
	if err != nil {
		return nil, err
	}

	pbCouriers := make([]*logisticspb.Courier, len(couriers))
	for i, courier := range couriers {
		pbCouriers[i] = &logisticspb.Courier{
			CourierCode:      courier.CourierCode,
			CourierName:      courier.CourierName,
			Phone:            courier.Phone,
			LogisticsCompany: courier.LogisticsCompany,
			ServiceArea:      courier.ServiceArea,
		}
	}

	return &logisticspb.GetCouriersResponse{
		Couriers: pbCouriers,
	}, nil
}