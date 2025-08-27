package do

import (
	"emshop/pkg/db"
	"time"
)

// FlashSaleStatus 秒杀活动状态定义
type FlashSaleStatus int32

const (
	FlashSaleStatusPending FlashSaleStatus = 1 // 待开始
	FlashSaleStatusActive  FlashSaleStatus = 2 // 进行中
	FlashSaleStatusFinished FlashSaleStatus = 3 // 已结束
	FlashSaleStatusPaused  FlashSaleStatus = 4 // 已暂停
)

// FlashSaleRecordStatus 秒杀记录状态定义
type FlashSaleRecordStatus int32

const (
	FlashSaleRecordStatusSuccess FlashSaleRecordStatus = 1 // 成功
	FlashSaleRecordStatusFailed  FlashSaleRecordStatus = 2 // 失败
	FlashSaleRecordStatusTimeout FlashSaleRecordStatus = 3 // 超时
)

// FlashSaleActivityDO 秒杀活动数据对象
type FlashSaleActivityDO struct {
	db.BaseModel
	ID        int64     `gorm:"primarykey" json:"id"`
	CouponTemplateID int64           `json:"coupon_template_id" gorm:"column:coupon_template_id;type:bigint;not null;index:idx_coupon_template_id;comment:关联的优惠券模板ID"`
	Name             string          `json:"name" gorm:"column:name;type:varchar(100);not null;comment:秒杀活动名称"`
	StartTime        time.Time       `json:"start_time" gorm:"column:start_time;type:timestamp;not null;index:idx_flash_sale_time;comment:秒杀开始时间"`
	EndTime          time.Time       `json:"end_time" gorm:"column:end_time;type:timestamp;not null;index:idx_flash_sale_time;comment:秒杀结束时间"`
	FlashSaleCount   int32           `json:"flash_sale_count" gorm:"column:flash_sale_count;type:int;not null;comment:秒杀数量"`
	SoldCount        int32           `json:"sold_count" gorm:"column:sold_count;type:int;not null;default:0;comment:已售数量"`
	PerUserLimit     int32           `json:"per_user_limit" gorm:"column:per_user_limit;type:int;not null;default:1;comment:每用户限抢数量"`
	Status           FlashSaleStatus `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_status;comment:状态"`
	SortOrder        int32           `json:"sort_order" gorm:"column:sort_order;type:int;default:0;index:idx_sort_order;comment:排序权重"`
}

// TableName 指定表名
func (FlashSaleActivityDO) TableName() string {
	return "flash_sale_activities"
}

// FlashSaleRecordDO 秒杀参与记录数据对象
type FlashSaleRecordDO struct {
	db.BaseModel
	ID        int64     `gorm:"primarykey" json:"id"`
	FlashSaleID    int64                 `json:"flash_sale_id" gorm:"column:flash_sale_id;type:bigint;not null;index:idx_flash_sale_id;uniqueIndex:uk_flash_sale_user;comment:秒杀活动ID"`
	UserID         int64                 `json:"user_id" gorm:"column:user_id;type:bigint;not null;index:idx_user_id;uniqueIndex:uk_flash_sale_user;comment:用户ID"`
	UserCouponID   *int64                `json:"user_coupon_id" gorm:"column:user_coupon_id;type:bigint;comment:生成的用户优惠券ID"`
	Status         FlashSaleRecordStatus `json:"status" gorm:"column:status;type:tinyint;not null;index:idx_status;comment:状态"`
	FailReason     string                `json:"fail_reason" gorm:"column:fail_reason;type:varchar(200);comment:失败原因"`
	CreatedAt      time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;index:idx_created_at"`
}

// TableName 指定表名
func (FlashSaleRecordDO) TableName() string {
	return "flash_sale_records"
}

// FlashSaleActivityDOList 秒杀活动列表
type FlashSaleActivityDOList struct {
	TotalCount int64                  `json:"totalCount"`
	Items      []*FlashSaleActivityDO `json:"items"`
}

// FlashSaleRecordDOList 秒杀记录列表
type FlashSaleRecordDOList struct {
	TotalCount int64                `json:"totalCount"`
	Items      []*FlashSaleRecordDO `json:"items"`
}

// FlashSaleStockInfo 秒杀库存信息
type FlashSaleStockInfo struct {
	FlashSaleID    int64 `json:"flash_sale_id"`
	TotalStock     int32 `json:"total_stock"`
	RemainingStock int32 `json:"remaining_stock"`
	SoldCount      int32 `json:"sold_count"`
}

// FlashSaleUserInfo 用户秒杀参与信息
type FlashSaleUserInfo struct {
	FlashSaleID      int64 `json:"flash_sale_id"`
	UserID           int64 `json:"user_id"`
	ParticipateCount int32 `json:"participate_count"`
	SuccessCount     int32 `json:"success_count"`
	LastParticipateTime *time.Time `json:"last_participate_time"`
}

// FlashSaleStatistics 秒杀统计信息
type FlashSaleStatistics struct {
	FlashSaleID       int64   `json:"flash_sale_id"`
	TotalParticipants int64   `json:"total_participants"`
	SuccessParticipants int64 `json:"success_participants"`
	SuccessRate       float64 `json:"success_rate"`
	PeakQPS           int64   `json:"peak_qps"`
	AverageResponseTime float64 `json:"average_response_time"`
}