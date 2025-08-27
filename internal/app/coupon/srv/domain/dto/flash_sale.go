package dto

import (
	"time"
	v1 "emshop/pkg/common/meta/v1"
)

// CreateFlashSaleActivityDTO 创建秒杀活动DTO
type CreateFlashSaleActivityDTO struct {
	CouponTemplateID int64     `json:"coupon_template_id" validate:"required"`
	Name             string    `json:"name" validate:"required,max=100"`
	StartTime        time.Time `json:"start_time" validate:"required"`
	EndTime          time.Time `json:"end_time" validate:"required"`
	FlashSaleCount   int32     `json:"flash_sale_count" validate:"required,min=1"`
	PerUserLimit     int32     `json:"per_user_limit" validate:"required,min=1"`
}

// FlashSaleActivityDTO 秒杀活动DTO
type FlashSaleActivityDTO struct {
	ID               int64              `json:"id"`
	CouponTemplateID int64              `json:"coupon_template_id"`
	CouponID         int64              `json:"coupon_id"`         // 兼容字段
	Name             string             `json:"name"`
	StartTime        time.Time          `json:"start_time"`
	EndTime          time.Time          `json:"end_time"`
	FlashSaleCount   int32              `json:"flash_sale_count"`
	SoldCount        int32              `json:"sold_count"`
	RemainStock      int32              `json:"remain_stock"`      // 剩余库存
	PerUserLimit     int32              `json:"per_user_limit"`
	Status           int32              `json:"status"`
	Template         *CouponTemplateDTO `json:"template,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	
	// 优惠券模板相关字段（冗余便于展示）
	CouponName    string  `json:"coupon_name,omitempty"`
	CouponType    int32   `json:"coupon_type,omitempty"`
	DiscountValue float64 `json:"discount_value,omitempty"`
}

// ListFlashSaleActivitiesDTO 秒杀活动列表DTO
type ListFlashSaleActivitiesDTO struct {
	Status   *int32      `json:"status,omitempty" validate:"omitempty,min=1,max=4"`
	ListMeta v1.ListMeta `json:",inline"`
	
	// 兼容字段
	Page     int32 `json:"page" validate:"required,min=1"`
	PageSize int32 `json:"page_size" validate:"required,min=1,max=100"`
}

// FlashSaleActivityListDTO 秒杀活动列表响应DTO
type FlashSaleActivityListDTO struct {
	TotalCount int64                    `json:"total_count"`
	Items      []*FlashSaleActivityDTO  `json:"items"`
	ListMeta   v1.ListMeta              `json:",inline"`
}

// ParticipateFlashSaleDTO 参与秒杀DTO
type ParticipateFlashSaleDTO struct {
	UserID      int64 `json:"user_id" validate:"required"`
	FlashSaleID int64 `json:"flash_sale_id" validate:"required"`
}

// ParticipateFlashSaleResultDTO 参与秒杀结果DTO
type ParticipateFlashSaleResultDTO struct {
	Status        int32   `json:"status"` // 1-成功, 2-失败
	FailReason    *string `json:"fail_reason,omitempty"`
	UserCouponID  *int64  `json:"user_coupon_id,omitempty"`
}

// FlashSaleStockDTO 秒杀库存信息DTO
type FlashSaleStockDTO struct {
	FlashSaleID    int64 `json:"flash_sale_id"`
	TotalStock     int32 `json:"total_stock"`
	RemainingStock int32 `json:"remaining_stock"`
	SoldCount      int32 `json:"sold_count"`
}

// GetUserFlashSaleRecordsDTO 获取用户秒杀记录DTO
type GetUserFlashSaleRecordsDTO struct {
	UserID      int64  `json:"user_id" validate:"required"`
	FlashSaleID *int64 `json:"flash_sale_id,omitempty"`
	Page        int32  `json:"page" validate:"required,min=1"`
	PageSize    int32  `json:"page_size" validate:"required,min=1,max=100"`
}

// FlashSaleRecordDTO 秒杀记录DTO
type FlashSaleRecordDTO struct {
	ID           int64                  `json:"id"`
	FlashSaleID  int64                  `json:"flash_sale_id"`
	UserID       int64                  `json:"user_id"`
	UserCouponID *int64                 `json:"user_coupon_id,omitempty"`
	Status       int32                  `json:"status"`
	FailReason   *string                `json:"fail_reason,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	Activity     *FlashSaleActivityDTO  `json:"activity,omitempty"`
}

// FlashSaleRecordListDTO 秒杀记录列表响应DTO
type FlashSaleRecordListDTO struct {
	TotalCount int64                 `json:"total_count"`
	Items      []*FlashSaleRecordDTO `json:"items"`
}

// ===== 新增的秒杀相关DTO =====

// StartFlashSaleDTO 启动秒杀活动DTO
type StartFlashSaleDTO struct {
	ActivityID int64 `json:"activity_id" validate:"required"`
}

// StopFlashSaleDTO 停止秒杀活动DTO  
type StopFlashSaleDTO struct {
	ActivityID  int64 `json:"activity_id" validate:"required"`
	CleanupData bool  `json:"cleanup_data,omitempty"` // 是否清理数据
}

// FlashSaleRequestDTO 秒杀请求DTO
type FlashSaleRequestDTO struct {
	ActivityID int64  `json:"activity_id" validate:"required"`
	UserID     int64  `json:"user_id" validate:"required"`
	ClientIP   string `json:"client_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// FlashSaleResultDTO 秒杀结果DTO
type FlashSaleResultDTO struct {
	Success     bool   `json:"success"`
	Code        int    `json:"code"`         // 1:成功 -1:库存不足 -2:用户限制 -3:活动异常 -4:系统错误
	Message     string `json:"message"`
	CouponSn    string `json:"coupon_sn,omitempty"`    // 优惠券编号（成功时）
	RemainStock int32  `json:"remain_stock"`           // 剩余库存
	Timestamp   int64  `json:"timestamp"`              // 操作时间戳
}

// FlashSaleStatusDTO 查询秒杀状态DTO
type FlashSaleStatusDTO struct {
	ActivityID int64 `json:"activity_id" validate:"required"`
	UserID     int64 `json:"user_id,omitempty"` // 可选，查询用户参与状态
}

// FlashSaleStatusResultDTO 秒杀状态结果DTO
type FlashSaleStatusResultDTO struct {
	ActivityID             int64     `json:"activity_id"`
	CouponID               int64     `json:"coupon_id"`
	Status                 int32     `json:"status"`                   // 1:待开始 2:进行中 3:已结束
	TotalCount             int32     `json:"total_count"`              // 总投放数量
	SuccessCount           int32     `json:"success_count"`            // 成功抢购数量
	RemainStock            int32     `json:"remain_stock"`             // 剩余库存
	StartTime              time.Time `json:"start_time"`
	EndTime                time.Time `json:"end_time"`
	UserParticipated       bool      `json:"user_participated"`        // 用户是否已参与
	UserParticipationCount int32     `json:"user_participation_count"` // 用户参与次数
}

// UpdateFlashSaleActivityDTO 更新秒杀活动DTO
type UpdateFlashSaleActivityDTO struct {
	ID             int64     `json:"id" validate:"required"`
	Name           string    `json:"name,omitempty" validate:"omitempty,max=100"`
	StartTime      time.Time `json:"start_time,omitempty"`
	EndTime        time.Time `json:"end_time,omitempty"`
	FlashSaleCount int32     `json:"flash_sale_count,omitempty" validate:"omitempty,min=1"`
	PerUserLimit   int32     `json:"per_user_limit,omitempty" validate:"omitempty,min=1"`
}