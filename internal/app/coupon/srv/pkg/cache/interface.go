package cache

import (
	"context"
	"fmt"
)

// CacheManager 缓存管理器接口
type CacheManager interface {
	// 优惠券模板缓存
	GetCouponTemplate(ctx context.Context, couponID int64) (*CouponTemplate, error)
	
	// 用户优惠券缓存
	GetUserCoupon(ctx context.Context, userCouponID int64) (*UserCoupon, error)
	
	// 秒杀活动缓存
	GetFlashSaleActivity(ctx context.Context, activityID int64) (*FlashSaleActivity, error)
	
	// 缓存失效
	InvalidateCache(keys ...string)
	InvalidateCacheByPattern(ctx context.Context, pattern string) error
	
	// 缓存预热
	WarmupCache(ctx context.Context) error
	
	// 统计信息
	GetCacheStats() map[string]interface{}
	
	// 资源清理
	Close()
}

// CacheKey 缓存键生成器
type CacheKey struct{}

// 缓存键生成方法
func (ck *CacheKey) CouponTemplate(couponID int64) string {
	return fmt.Sprintf("coupon:template:%d", couponID)
}

func (ck *CacheKey) UserCoupon(userCouponID int64) string {
	return fmt.Sprintf("coupon:user:%d", userCouponID)
}

func (ck *CacheKey) UserCouponList(userID int64) string {
	return fmt.Sprintf("coupon:user:list:%d", userID)
}

func (ck *CacheKey) UserAvailableCoupons(userID int64) string {
	return fmt.Sprintf("coupon:user:available:%d", userID)
}

func (ck *CacheKey) FlashSaleActivity(activityID int64) string {
	return fmt.Sprintf("flashsale:activity:%d", activityID)
}

func (ck *CacheKey) FlashSaleStatus(activityID int64) string {
	return fmt.Sprintf("flashsale:status:%d", activityID)
}

func (ck *CacheKey) CouponStock(couponID int64) string {
	return fmt.Sprintf("coupon:stock:%d", couponID)
}

func (ck *CacheKey) UserFlashSaleLimit(activityID, userID int64) string {
	return fmt.Sprintf("coupon:user:%d:%d", activityID, userID)
}

// 全局缓存键生成器实例
var CacheKeys = &CacheKey{}

// CacheWithSyncManager 带Canal同步的缓存管理器接口
type CacheWithSyncManager interface {
	CacheManager
	
	// Canal同步相关方法
	StartCanalSync() error
	StopCanalSync() error
	GetSyncStats() map[string]interface{}
}

// IntegratedCacheManager 集成缓存管理器（缓存+Canal同步）
type IntegratedCacheManager struct {
	CacheManager
	canalSync *CanalSyncManager
}

// NewIntegratedCacheManager 创建集成缓存管理器
func NewIntegratedCacheManager(cacheManager CacheManager, canalSync *CanalSyncManager) CacheWithSyncManager {
	return &IntegratedCacheManager{
		CacheManager: cacheManager,
		canalSync:    canalSync,
	}
}

// StartCanalSync 启动Canal同步
func (icm *IntegratedCacheManager) StartCanalSync() error {
	if icm.canalSync != nil {
		return icm.canalSync.Start()
	}
	return nil
}

// StopCanalSync 停止Canal同步
func (icm *IntegratedCacheManager) StopCanalSync() error {
	if icm.canalSync != nil {
		return icm.canalSync.Stop()
	}
	return nil
}

// GetSyncStats 获取同步统计信息
func (icm *IntegratedCacheManager) GetSyncStats() map[string]interface{} {
	if icm.canalSync != nil {
		return icm.canalSync.GetSyncStats()
	}
	return map[string]interface{}{
		"canal_sync_enabled": false,
	}
}

// Close 关闭集成缓存管理器
func (icm *IntegratedCacheManager) Close() {
	if icm.canalSync != nil {
		icm.canalSync.Stop()
	}
	if icm.CacheManager != nil {
		icm.CacheManager.Close()
	}
}