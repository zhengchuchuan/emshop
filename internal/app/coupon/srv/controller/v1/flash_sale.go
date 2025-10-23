package v1

import (
	"context"
	"errors"
	"time"

	couponpb "emshop/api/coupon/v1"
	"emshop/internal/app/coupon/srv/domain/dto"
	"emshop/pkg/log"

	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateFlashSaleActivity 创建秒杀活动
func (cs *couponServer) CreateFlashSaleActivity(ctx context.Context, req *couponpb.CreateFlashSaleActivityRequest) (*couponpb.FlashSaleActivityResponse, error) {
	log.Infof("CreateFlashSaleActivity: %s", req.Name)

	dto := &dto.CreateFlashSaleActivityDTO{
		CouponTemplateID: req.CouponTemplateId,
		Name:             req.Name,
		StartTime:        time.Unix(req.StartTime, 0),
		EndTime:          time.Unix(req.EndTime, 0),
		FlashSaleCount:   req.FlashSaleCount,
		PerUserLimit:     req.PerUserLimit,
	}

	// 统一使用异步核心服务
	result, err := cs.srv.FlashSaleCore.CreateFlashSaleActivity(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return cs.convertFlashSaleToProto(result), nil
}

// GetFlashSaleActivity 获取秒杀活动
func (cs *couponServer) GetFlashSaleActivity(ctx context.Context, req *couponpb.GetFlashSaleActivityRequest) (*couponpb.FlashSaleActivityResponse, error) {
	result, err := cs.srv.FlashSaleCore.GetFlashSaleActivity(ctx, req.Id)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return cs.convertFlashSaleToProto(result), nil
}

// ListFlashSaleActivities 获取秒杀活动列表
func (cs *couponServer) ListFlashSaleActivities(ctx context.Context, req *couponpb.ListFlashSaleActivitiesRequest) (*couponpb.ListFlashSaleActivitiesResponse, error) {
	dto := &dto.ListFlashSaleActivitiesDTO{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.Status != nil {
		dto.Status = req.Status
	}

	result, err := cs.srv.FlashSaleCore.ListFlashSaleActivities(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换响应
	items := make([]*couponpb.FlashSaleActivityResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, cs.convertFlashSaleToProto(item))
	}

	return &couponpb.ListFlashSaleActivitiesResponse{
		TotalCount: result.TotalCount,
		Items:      items,
	}, nil
}

// GetActiveFlashSales 获取进行中的秒杀活动
func (cs *couponServer) GetActiveFlashSales(ctx context.Context, req *emptypb.Empty) (*couponpb.ListFlashSaleActivitiesResponse, error) {
	// 等价：按状态过滤
	status := int32(2) // Active
	result, err := cs.srv.FlashSaleCore.ListFlashSaleActivities(ctx, &dto.ListFlashSaleActivitiesDTO{Status: &status})
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换响应
	items := make([]*couponpb.FlashSaleActivityResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, cs.convertFlashSaleToProto(item))
	}

	return &couponpb.ListFlashSaleActivitiesResponse{
		TotalCount: result.TotalCount,
		Items:      items,
	}, nil
}

// ParticipateFlashSale 参与秒杀
func (cs *couponServer) ParticipateFlashSale(ctx context.Context, req *couponpb.ParticipateFlashSaleRequest) (*couponpb.ParticipateFlashSaleResponse, error) {
	// 压测模式下抑制成功路径日志，避免IO瓶颈
	// log.Infof("ParticipateFlashSale: userID=%d, flashSaleID=%d", req.UserId, req.FlashSaleId)

	// 简化触发条件：按活动async_enabled+全局可用性自动分流
	if cs.srv.ShouldUseAsync(ctx, req.FlashSaleId) && cs.srv.FlashSaleCore != nil {
		// log.Debug("异步秒杀")
		flashReq := &dto.FlashSaleRequestDTO{
			ActivityID: req.FlashSaleId,
			UserID:     req.UserId,
		}
		coreResult, err := cs.srv.FlashSaleCore.FlashSaleCoupon(ctx, flashReq)
		if err != nil {
			return nil, cs.handleError(err)
		}

		resp := &couponpb.ParticipateFlashSaleResponse{}
		if coreResult != nil && coreResult.Success {
			resp.Status = 1
		} else {
			resp.Status = 2
			if coreResult != nil {
				failMsg := coreResult.Message
				resp.FailReason = &failMsg
			}
		}
		return resp, nil
	}

	return nil, errors.New("当前秒杀活动参与方式不可用，请稍后重试")
}

// GetFlashSaleStock 获取秒杀库存
func (cs *couponServer) GetFlashSaleStock(ctx context.Context, req *couponpb.GetFlashSaleStockRequest) (*couponpb.FlashSaleStockResponse, error) {
	result, err := cs.srv.FlashSaleCore.GetFlashSaleStock(ctx, req.FlashSaleId)
	if err != nil {
		return nil, cs.handleError(err)
	}

	return &couponpb.FlashSaleStockResponse{
		FlashSaleId:    result.FlashSaleID,
		TotalStock:     result.TotalStock,
		RemainingStock: result.RemainingStock,
		SoldCount:      result.SoldCount,
	}, nil
}

// GetUserFlashSaleRecord 获取用户秒杀记录
func (cs *couponServer) GetUserFlashSaleRecord(ctx context.Context, req *couponpb.GetUserFlashSaleRecordRequest) (*couponpb.ListFlashSaleRecordsResponse, error) {
	dto := &dto.GetUserFlashSaleRecordsDTO{
		UserID:   req.UserId,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.FlashSaleId != nil {
		dto.FlashSaleID = req.FlashSaleId
	}

	result, err := cs.srv.FlashSaleCore.GetUserFlashSaleRecords(ctx, dto)
	if err != nil {
		return nil, cs.handleError(err)
	}

	// 转换响应
	items := make([]*couponpb.FlashSaleRecordResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, cs.convertFlashSaleRecordToProto(item))
	}

	return &couponpb.ListFlashSaleRecordsResponse{
		TotalCount: result.TotalCount,
		Items:      items,
	}, nil
}

// convertFlashSaleToProto 转换秒杀活动DTO为Protobuf
func (cs *couponServer) convertFlashSaleToProto(dto *dto.FlashSaleActivityDTO) *couponpb.FlashSaleActivityResponse {
	resp := &couponpb.FlashSaleActivityResponse{
		Id:               dto.ID,
		CouponTemplateId: dto.CouponTemplateID,
		Name:             dto.Name,
		StartTime:        dto.StartTime.Unix(),
		EndTime:          dto.EndTime.Unix(),
		FlashSaleCount:   dto.FlashSaleCount,
		SoldCount:        dto.SoldCount,
		PerUserLimit:     dto.PerUserLimit,
		Status:           dto.Status,
		CreatedAt:        dto.CreatedAt.Unix(),
	}

	if dto.Template != nil {
		resp.Template = cs.convertTemplateToProto(dto.Template)
	}

	return resp
}

// convertFlashSaleRecordToProto 转换秒杀记录DTO为Protobuf
func (cs *couponServer) convertFlashSaleRecordToProto(dto *dto.FlashSaleRecordDTO) *couponpb.FlashSaleRecordResponse {
	resp := &couponpb.FlashSaleRecordResponse{
		Id:          dto.ID,
		FlashSaleId: dto.FlashSaleID,
		UserId:      dto.UserID,
		Status:      dto.Status,
		CreatedAt:   dto.CreatedAt.Unix(),
	}

	if !dto.OrderCreatedAt.IsZero() {
		resp.OrderCreatedAt = dto.OrderCreatedAt.UnixMilli()
	}

	if dto.UserCouponID != nil {
		resp.UserCouponId = dto.UserCouponID
	}
	if dto.FailReason != nil {
		resp.FailReason = dto.FailReason
	}
	if dto.Activity != nil {
		resp.Activity = cs.convertFlashSaleToProto(dto.Activity)
	}

	return resp
}

// 注意：RocketMQ生产者已经在service层集成完毕
// 现有的ParticipateFlashSale方法可以通过FlashSaleCore替换实现更高性能的秒杀功能
