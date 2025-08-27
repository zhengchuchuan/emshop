package scripts

import (
	_ "embed"
	"fmt"
)

// Lua脚本内容嵌入
//go:embed flash_sale.lua
var FlashSaleLua string

//go:embed rollback_flash_sale.lua  
var RollbackFlashSaleLua string

//go:embed init_flash_sale.lua
var InitFlashSaleLua string

//go:embed check_coupon_usage.lua
var CheckCouponUsageLua string

//go:embed release_coupon_lock.lua
var ReleaseCouponLockLua string

// 脚本返回值常量
const (
	// 秒杀脚本返回码
	FlashSaleSuccess        = 1
	FlashSaleNotStarted     = -1
	FlashSaleEnded          = -2
	FlashSaleInactive       = -3
	FlashSaleOutOfStock     = -4
	FlashSaleUserLimitExceed = -5

	// 优惠券检查脚本返回码
	CouponLockSuccess   = 1
	CouponExpired      = -1
	CouponUsed         = -2
	CouponLocked       = -3

	// 通用返回码
	OperationSuccess = 1
	OperationSkipped = 0
)

// 秒杀脚本返回码说明
func GetFlashSaleResultMessage(result int64) string {
	switch result {
	case FlashSaleSuccess:
		return "秒杀成功"
	case FlashSaleNotStarted:
		return "秒杀活动未开始"
	case FlashSaleEnded:
		return "秒杀活动已结束"
	case FlashSaleInactive:
		return "秒杀活动已暂停或无效"
	case FlashSaleOutOfStock:
		return "库存不足"
	case FlashSaleUserLimitExceed:
		return "超出个人限购数量"
	default:
		return "未知错误"
	}
}

// 优惠券检查脚本返回码说明
func GetCouponCheckResultMessage(result int64) string {
	switch result {
	case CouponLockSuccess:
		return "优惠券锁定成功"
	case CouponExpired:
		return "优惠券已过期"
	case CouponUsed:
		return "优惠券已被使用"
	case CouponLocked:
		return "优惠券正在被使用"
	default:
		return "未知错误"
	}
}

// Redis键格式化函数
type RedisKeyFormatter struct{}

func NewRedisKeyFormatter() *RedisKeyFormatter {
	return &RedisKeyFormatter{}
}

// 秒杀相关键格式
func (r *RedisKeyFormatter) FlashSaleStockKey(flashSaleID int64) string {
	return fmt.Sprintf("flashsale:stock:%d", flashSaleID)
}

func (r *RedisKeyFormatter) FlashSaleUserLimitKey(flashSaleID, userID int64) string {
	return fmt.Sprintf("flashsale:user_limit:%d:%d", flashSaleID, userID)
}

func (r *RedisKeyFormatter) FlashSaleStatusKey(flashSaleID int64) string {
	return fmt.Sprintf("flashsale:status:%d", flashSaleID)
}

// 优惠券相关键格式
func (r *RedisKeyFormatter) CouponLockKey(userCouponID int64) string {
	return fmt.Sprintf("coupon:lock:%d", userCouponID)
}

func (r *RedisKeyFormatter) CouponStatusKey(userCouponID int64) string {
	return fmt.Sprintf("coupon:status:%d", userCouponID)
}