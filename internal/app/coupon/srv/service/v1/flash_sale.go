package v1

import (
	"context"
	"fmt"
	"time"

	"emshop/internal/app/coupon/srv/data/v1/interfaces"
	"emshop/internal/app/coupon/srv/domain/do"
	"emshop/internal/app/coupon/srv/domain/dto"
	"emshop/internal/app/coupon/srv/pkg/cache"
	"emshop/internal/app/coupon/srv/pkg/scripts"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	v1 "emshop/pkg/common/meta/v1"
	"github.com/go-redis/redis/v8"
)

// FlashSaleSrv 秒杀服务接口
type FlashSaleSrv interface {
	// 秒杀活动管理
	CreateFlashSaleActivity(ctx context.Context, req *dto.CreateFlashSaleActivityDTO) (*dto.FlashSaleActivityDTO, error)
	GetFlashSaleActivity(ctx context.Context, id int64) (*dto.FlashSaleActivityDTO, error)
	ListFlashSaleActivities(ctx context.Context, req *dto.ListFlashSaleActivitiesDTO) (*dto.FlashSaleActivityListDTO, error)
	GetActiveFlashSales(ctx context.Context) (*dto.FlashSaleActivityListDTO, error)
	
	// 秒杀参与
	ParticipateFlashSale(ctx context.Context, req *dto.ParticipateFlashSaleDTO) (*dto.ParticipateFlashSaleResultDTO, error)
	GetFlashSaleStock(ctx context.Context, flashSaleID int64) (*dto.FlashSaleStockDTO, error)
	GetUserFlashSaleRecords(ctx context.Context, req *dto.GetUserFlashSaleRecordsDTO) (*dto.FlashSaleRecordListDTO, error)
}

type flashSaleService struct {
	data         interfaces.DataFactory
	redisClient  *redis.Client
	keyFormatter *scripts.RedisKeyFormatter
	cacheManager interface {
		GetFlashSaleActivity(ctx context.Context, activityID int64) (*cache.FlashSaleActivity, error)
		InvalidateCache(keys ...string)
	}
}

// NewFlashSaleService 创建秒杀服务实例
func NewFlashSaleService(data interfaces.DataFactory, redisClient *redis.Client, cacheManager interface {
	GetFlashSaleActivity(ctx context.Context, activityID int64) (*cache.FlashSaleActivity, error)
	InvalidateCache(keys ...string)
}) FlashSaleSrv {
	return &flashSaleService{
		data:         data,
		redisClient:  redisClient,
		keyFormatter: scripts.NewRedisKeyFormatter(),
		cacheManager: cacheManager,
	}
}

// CreateFlashSaleActivity 创建秒杀活动
func (fss *flashSaleService) CreateFlashSaleActivity(ctx context.Context, req *dto.CreateFlashSaleActivityDTO) (*dto.FlashSaleActivityDTO, error) {
	log.Infof("创建秒杀活动: %s", req.Name)
	
	// 验证时间范围
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.WithCode(code.ErrInvalidRequest, "结束时间不能早于开始时间")
	}
	
	// 验证优惠券模板存在
	templateDO, err := fss.data.CouponTemplates().Get(ctx, fss.data.DB(), req.CouponTemplateID)
	if err != nil || templateDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "优惠券模板不存在")
	}
	
	// 构建DO对象
	activityDO := &do.FlashSaleActivityDO{
		CouponTemplateID: req.CouponTemplateID,
		Name:             req.Name,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		FlashSaleCount:   req.FlashSaleCount,
		SoldCount:        0,
		PerUserLimit:     req.PerUserLimit,
		Status:           do.FlashSaleStatusPending,
		SortOrder:        0,
	}
	
	// 保存到数据库
	if err := fss.data.FlashSales().Create(ctx, fss.data.DB(), activityDO); err != nil {
		log.Errorf("创建秒杀活动失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "创建秒杀活动失败")
	}
	
	// 如果活动即将开始或已开始，初始化Redis
	currentTime := time.Now()
	if req.StartTime.Before(currentTime) || req.StartTime.Sub(currentTime) < time.Minute*10 {
		if err := fss.initFlashSaleRedis(ctx, activityDO); err != nil {
			log.Warnf("初始化秒杀活动Redis失败: %v", err)
		}
	}
	
	// 转换为DTO
	return fss.convertFlashSaleToDTO(activityDO, templateDO), nil
}

// GetFlashSaleActivity 获取秒杀活动
func (fss *flashSaleService) GetFlashSaleActivity(ctx context.Context, id int64) (*dto.FlashSaleActivityDTO, error) {
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), id)
	if err != nil {
		log.Errorf("获取秒杀活动失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取秒杀活动失败")
	}
	
	if activityDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "秒杀活动不存在")
	}
	
	// 获取关联的优惠券模板
	templateDO, err := fss.data.CouponTemplates().Get(ctx, fss.data.DB(), activityDO.CouponTemplateID)
	if err != nil {
		log.Warnf("获取秒杀活动关联的优惠券模板失败: %v", err)
	}
	
	return fss.convertFlashSaleToDTO(activityDO, templateDO), nil
}

// ListFlashSaleActivities 获取秒杀活动列表
func (fss *flashSaleService) ListFlashSaleActivities(ctx context.Context, req *dto.ListFlashSaleActivitiesDTO) (*dto.FlashSaleActivityListDTO, error) {
	var status do.FlashSaleStatus
	if req.Status != nil {
		status = do.FlashSaleStatus(*req.Status)
	}
	
	meta := v1.ListMeta{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	
	activityListDO, err := fss.data.FlashSales().List(ctx, fss.data.DB(), status, meta, []string{"sort_order DESC", "start_time DESC"})
	if err != nil {
		log.Errorf("获取秒杀活动列表失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取秒杀活动列表失败")
	}
	
	// 转换为DTO
	items := make([]*dto.FlashSaleActivityDTO, 0, len(activityListDO.Items))
	for _, activityDO := range activityListDO.Items {
		items = append(items, fss.convertFlashSaleToDTO(activityDO, nil))
	}
	
	return &dto.FlashSaleActivityListDTO{
		TotalCount: activityListDO.TotalCount,
		Items:      items,
	}, nil
}

// GetActiveFlashSales 获取进行中的秒杀活动
func (fss *flashSaleService) GetActiveFlashSales(ctx context.Context) (*dto.FlashSaleActivityListDTO, error) {
	currentTime := time.Now()
	
	activityDOs, err := fss.data.FlashSales().GetActiveActivities(ctx, fss.data.DB(), currentTime)
	if err != nil {
		log.Errorf("获取进行中的秒杀活动失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取进行中的秒杀活动失败")
	}
	
	// 转换为DTO
	items := make([]*dto.FlashSaleActivityDTO, 0, len(activityDOs))
	for _, activityDO := range activityDOs {
		// 确保Redis状态已初始化
		if err := fss.initFlashSaleRedis(ctx, activityDO); err != nil {
			log.Warnf("初始化秒杀活动Redis失败: %v", err)
		}
		items = append(items, fss.convertFlashSaleToDTO(activityDO, nil))
	}
	
	return &dto.FlashSaleActivityListDTO{
		TotalCount: int64(len(items)),
		Items:      items,
	}, nil
}

// ParticipateFlashSale 参与秒杀
func (fss *flashSaleService) ParticipateFlashSale(ctx context.Context, req *dto.ParticipateFlashSaleDTO) (*dto.ParticipateFlashSaleResultDTO, error) {
	log.Infof("用户参与秒杀: userID=%d, flashSaleID=%d", req.UserID, req.FlashSaleID)
	
	// 获取秒杀活动
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), req.FlashSaleID)
	if err != nil || activityDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "秒杀活动不存在")
	}
	
	currentTime := time.Now()
	
	// 确保Redis状态已初始化
	if err := fss.initFlashSaleRedis(ctx, activityDO); err != nil {
		log.Errorf("初始化秒杀活动Redis失败: %v", err)
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: stringPtr("系统错误，请稍后重试"),
		}, nil
	}
	
	// 使用Redis Lua脚本进行原子性秒杀
	stockKey := fss.keyFormatter.FlashSaleStockKey(req.FlashSaleID)
	userLimitKey := fss.keyFormatter.FlashSaleUserLimitKey(req.FlashSaleID, req.UserID)
	statusKey := fss.keyFormatter.FlashSaleStatusKey(req.FlashSaleID)
	
	result, err := fss.redisClient.Eval(ctx, scripts.FlashSaleLua, 
		[]string{stockKey, userLimitKey, statusKey},
		req.UserID, 
		currentTime.Unix(),
		activityDO.StartTime.Unix(),
		activityDO.EndTime.Unix(),
		activityDO.PerUserLimit,
		int(do.FlashSaleStatusActive),
	).Result()
	
	if err != nil {
		log.Errorf("执行秒杀Lua脚本失败: %v", err)
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: stringPtr("系统繁忙，请稍后重试"),
		}, nil
	}
	
	flashSaleResult := result.(int64)
	
	// 秒杀失败
	if flashSaleResult != scripts.FlashSaleSuccess {
		reason := scripts.GetFlashSaleResultMessage(flashSaleResult)
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: &reason,
		}, nil
	}
	
	// 秒杀成功，创建用户优惠券和记录
	tx := fss.data.Begin()
	
	// 创建秒杀记录
	recordDO := &do.FlashSaleRecordDO{
		FlashSaleID: req.FlashSaleID,
		UserID:      req.UserID,
		Status:      do.FlashSaleRecordStatusSuccess,
		CreatedAt:   currentTime,
	}
	
	if err := fss.data.FlashSaleRecords().Create(ctx, tx, recordDO); err != nil {
		tx.Rollback()
		// 回滚Redis状态
		fss.rollbackFlashSaleRedis(ctx, req.FlashSaleID, req.UserID)
		
		log.Errorf("创建秒杀记录失败: %v", err)
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: stringPtr("创建记录失败"),
		}, nil
	}
	
	// 获取优惠券模板
	templateDO, err := fss.data.CouponTemplates().Get(ctx, fss.data.DB(), activityDO.CouponTemplateID)
	if err != nil || templateDO == nil {
		tx.Rollback()
		fss.rollbackFlashSaleRedis(ctx, req.FlashSaleID, req.UserID)
		
		log.Errorf("获取优惠券模板失败: %v", err)
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: stringPtr("优惠券模板不存在"),
		}, nil
	}
	
	// 创建用户优惠券
	expiredAt := templateDO.ValidEndTime
	if templateDO.ValidDays > 0 {
		expiredAt = currentTime.AddDate(0, 0, int(templateDO.ValidDays))
	}
	
	userCouponDO := &do.UserCouponDO{
		CouponTemplateID: activityDO.CouponTemplateID,
		UserID:           req.UserID,
		CouponCode:       generateCouponCode(),
		Status:           do.UserCouponStatusUnused,
		ReceivedAt:       currentTime,
		ExpiredAt:        expiredAt,
	}
	
	if err := fss.data.UserCoupons().Create(ctx, tx, userCouponDO); err != nil {
		tx.Rollback()
		fss.rollbackFlashSaleRedis(ctx, req.FlashSaleID, req.UserID)
		
		log.Errorf("创建用户优惠券失败: %v", err)
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: stringPtr("创建优惠券失败"),
		}, nil
	}
	
	// 更新秒杀记录的优惠券ID
	recordDO.UserCouponID = &userCouponDO.ID
	if err := fss.data.FlashSaleRecords().UpdateUserCouponID(ctx, tx, recordDO.ID, userCouponDO.ID); err != nil {
		log.Warnf("更新秒杀记录优惠券ID失败: %v", err)
	}
	
	// 更新活动已售数量
	if err := fss.data.FlashSales().IncrementSoldCount(ctx, tx, req.FlashSaleID); err != nil {
		log.Warnf("更新秒杀活动已售数量失败: %v", err)
	}
	
	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交秒杀事务失败: %v", err)
		fss.rollbackFlashSaleRedis(ctx, req.FlashSaleID, req.UserID)
		
		return &dto.ParticipateFlashSaleResultDTO{
			Status:     2,
			FailReason: stringPtr("系统错误"),
		}, nil
	}
	
	return &dto.ParticipateFlashSaleResultDTO{
		Status:       1,
		UserCouponID: &userCouponDO.ID,
	}, nil
}

// GetFlashSaleStock 获取秒杀库存
func (fss *flashSaleService) GetFlashSaleStock(ctx context.Context, flashSaleID int64) (*dto.FlashSaleStockDTO, error) {
	// 先从Redis获取
	stockKey := fss.keyFormatter.FlashSaleStockKey(flashSaleID)
	stockInt, err := fss.redisClient.Get(ctx, stockKey).Int()
	
	var remainingStock int32
	if err == redis.Nil {
		// Redis中没有，从数据库获取
		stockInfo, err := fss.data.FlashSales().CheckStock(ctx, fss.data.DB(), flashSaleID)
		if err != nil {
			return nil, errors.WithCode(code.ErrDatabase, "获取秒杀库存失败")
		}
		remainingStock = stockInfo.RemainingStock
	} else if err != nil {
		return nil, errors.WithCode(code.ErrRedis, "获取Redis库存失败")
	} else {
		// 从Redis获取库存
		remainingStock = int32(stockInt)
	}
	
	// 获取活动信息
	activityDO, err := fss.data.FlashSales().Get(ctx, fss.data.DB(), flashSaleID)
	if err != nil || activityDO == nil {
		return nil, errors.WithCode(code.ErrResourceNotFound, "秒杀活动不存在")
	}
	
	return &dto.FlashSaleStockDTO{
		FlashSaleID:    flashSaleID,
		TotalStock:     activityDO.FlashSaleCount,
		RemainingStock: remainingStock,
		SoldCount:      activityDO.SoldCount,
	}, nil
}

// GetUserFlashSaleRecords 获取用户秒杀记录
func (fss *flashSaleService) GetUserFlashSaleRecords(ctx context.Context, req *dto.GetUserFlashSaleRecordsDTO) (*dto.FlashSaleRecordListDTO, error) {
	meta := v1.ListMeta{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	
	recordListDO, err := fss.data.FlashSaleRecords().GetUserRecords(ctx, fss.data.DB(), req.UserID, meta)
	if err != nil {
		log.Errorf("获取用户秒杀记录失败: %v", err)
		return nil, errors.WithCode(code.ErrDatabase, "获取用户秒杀记录失败")
	}
	
	// 转换为DTO
	items := make([]*dto.FlashSaleRecordDTO, 0, len(recordListDO.Items))
	for _, recordDO := range recordListDO.Items {
		items = append(items, fss.convertFlashSaleRecordToDTO(recordDO))
	}
	
	return &dto.FlashSaleRecordListDTO{
		TotalCount: recordListDO.TotalCount,
		Items:      items,
	}, nil
}

// initFlashSaleRedis 初始化秒杀活动的Redis状态
func (fss *flashSaleService) initFlashSaleRedis(ctx context.Context, activityDO *do.FlashSaleActivityDO) error {
	stockKey := fss.keyFormatter.FlashSaleStockKey(activityDO.ID)
	statusKey := fss.keyFormatter.FlashSaleStatusKey(activityDO.ID)
	
	currentTime := time.Now()
	remainingStock := activityDO.FlashSaleCount - activityDO.SoldCount
	
	_, err := fss.redisClient.Eval(ctx, scripts.InitFlashSaleLua,
		[]string{stockKey, statusKey},
		activityDO.ID,
		remainingStock,
		int(do.FlashSaleStatusActive),
		activityDO.EndTime.Unix(),
		currentTime.Unix(),
	).Result()
	
	return err
}

// rollbackFlashSaleRedis 回滚秒杀Redis状态
func (fss *flashSaleService) rollbackFlashSaleRedis(ctx context.Context, flashSaleID, userID int64) {
	stockKey := fss.keyFormatter.FlashSaleStockKey(flashSaleID)
	userLimitKey := fss.keyFormatter.FlashSaleUserLimitKey(flashSaleID, userID)
	
	_, err := fss.redisClient.Eval(ctx, scripts.RollbackFlashSaleLua,
		[]string{stockKey, userLimitKey},
		userID,
	).Result()
	
	if err != nil {
		log.Errorf("回滚秒杀Redis状态失败: %v", err)
	}
}

// convertFlashSaleToDTO 转换秒杀活动DO为DTO
func (fss *flashSaleService) convertFlashSaleToDTO(activityDO *do.FlashSaleActivityDO, templateDO *do.CouponTemplateDO) *dto.FlashSaleActivityDTO {
	result := &dto.FlashSaleActivityDTO{
		ID:               activityDO.ID,
		CouponTemplateID: activityDO.CouponTemplateID,
		Name:             activityDO.Name,
		StartTime:        activityDO.StartTime,
		EndTime:          activityDO.EndTime,
		FlashSaleCount:   activityDO.FlashSaleCount,
		SoldCount:        activityDO.SoldCount,
		PerUserLimit:     activityDO.PerUserLimit,
		Status:           int32(activityDO.Status),
		CreatedAt:        activityDO.CreatedAt,
	}
	
	if templateDO != nil {
		result.Template = &dto.CouponTemplateDTO{
			ID:                templateDO.ID,
			Name:              templateDO.Name,
			Type:              int32(templateDO.Type),
			DiscountType:      int32(templateDO.DiscountType),
			DiscountValue:     templateDO.DiscountValue,
			MinOrderAmount:    templateDO.MinOrderAmount,
			MaxDiscountAmount: templateDO.MaxDiscountAmount,
			Description:       templateDO.Description,
		}
	}
	
	return result
}

// convertFlashSaleRecordToDTO 转换秒杀记录DO为DTO
func (fss *flashSaleService) convertFlashSaleRecordToDTO(recordDO *do.FlashSaleRecordDO) *dto.FlashSaleRecordDTO {
	result := &dto.FlashSaleRecordDTO{
		ID:           recordDO.ID,
		FlashSaleID:  recordDO.FlashSaleID,
		UserID:       recordDO.UserID,
		UserCouponID: recordDO.UserCouponID,
		Status:       int32(recordDO.Status),
		CreatedAt:    recordDO.CreatedAt,
	}
	
	// 处理失败原因 - 只有非空字符串才转换为指针
	if recordDO.FailReason != "" {
		result.FailReason = &recordDO.FailReason
	}
	
	return result
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// generateCouponCode 生成优惠券码 (简化版)
func generateCouponCode() string {
	return fmt.Sprintf("CPN%d", time.Now().UnixNano()%100000000)
}