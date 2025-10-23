package redis

import (
    "context"
    "fmt"
    "strings"
    "strconv"
    "time"

	"emshop/pkg/log"
	"github.com/go-redis/redis/v8"
)

// StockManager 高性能库存管理器
// 基于Redis Lua脚本实现原子操作，确保零超卖
type StockManager struct {
    redis                 *redis.Client
    flashSaleScript       *redis.Script
    reserveScript         *redis.Script
    prewarmScript         *redis.Script
    userLimitScript       *redis.Script
    rollbackScript        *redis.Script
    activityStatusScript  *redis.Script
    suppressLogs          bool
}

// FlashSaleRequest 秒杀请求
type FlashSaleRequest struct {
    ActivityID   int64  `json:"activity_id"`
    CouponID     int64  `json:"coupon_id"`
    UserID       int64  `json:"user_id"`
    RequestCount int32  `json:"request_count"` // 请求数量，通常为1
    ClientIP     string `json:"client_ip"`
    UserAgent    string `json:"user_agent"`
    RequestID    string `json:"request_id"`
    PerUserLimitOverride int32 `json:"per_user_limit_override"`
}

// FlashSaleResult 秒杀结果
type FlashSaleResult struct {
	Code         int    `json:"code"`          // 1:成功 -1:库存不足 -2:用户限制 -3:活动异常
	Success      bool   `json:"success"`       // 是否成功
	Message      string `json:"message"`       // 结果消息
	RemainStock  int    `json:"remain_stock"`  // 剩余库存
	CouponSn     string `json:"coupon_sn"`     // 优惠券编号（成功时生成）
	Timestamp    int64  `json:"timestamp"`     // 操作时间戳
}

// ActivityInfo 活动信息
type ActivityInfo struct {
	ID           int64     `json:"id"`
	CouponID     int64     `json:"coupon_id"`
	Status       int32     `json:"status"`        // 1:待开始 2:进行中 3:已结束
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	TotalCount   int32     `json:"total_count"`
	SuccessCount int32     `json:"success_count"`
	PerUserLimit int32     `json:"per_user_limit"`
}

// NewStockManager 创建库存管理器（默认不抑制日志）
func NewStockManager(redisClient *redis.Client) *StockManager {
    return &StockManager{
        redis:                redisClient,
        flashSaleScript:      redis.NewScript(FlashSaleLuaScript),
        reserveScript:        redis.NewScript(FlashSaleReserveLuaScript),
        prewarmScript:        redis.NewScript(StockPrewarmLuaScript),
        userLimitScript:      redis.NewScript(UserLimitCheckLuaScript),
        rollbackScript:       redis.NewScript(StockRollbackLuaScript),
        activityStatusScript: redis.NewScript(ActivityStatusLuaScript),
        suppressLogs:         false,
    }
}

// NewStockManagerWithOptions 创建库存管理器（可抑制日志）
func NewStockManagerWithOptions(redisClient *redis.Client, suppressLogs bool) *StockManager {
    sm := NewStockManager(redisClient)
    sm.suppressLogs = suppressLogs
    return sm
}

// FlashSale 执行秒杀操作
func (sm *StockManager) FlashSale(ctx context.Context, req *FlashSaleRequest) (*FlashSaleResult, error) {
    if !sm.suppressLogs {
        log.Infof("开始执行秒杀: activityID=%d, userID=%d, couponID=%d", req.ActivityID, req.UserID, req.CouponID)
    }

	// 构建Redis keys
    keys := []string{
        fmt.Sprintf("coupon:stock:%d", req.CouponID),           // 库存key
        fmt.Sprintf("coupon:user:%d:%d", req.ActivityID, req.UserID), // 用户参与记录key
        fmt.Sprintf("coupon:log:%d", req.ActivityID),          // 日志key
        fmt.Sprintf("coupon:activity:%d", req.ActivityID),     // 活动信息key
        fmt.Sprintf("coupon:reserve:%d:%d:%s", req.ActivityID, req.UserID, req.RequestID), // 预留幂等key
    }

	// 构建参数
	currentTime := time.Now().Unix()
    args := []interface{}{
        req.UserID,
        req.ActivityID,
        req.RequestCount,
        300, // 5分钟TTL 用于用户参与计数与请求幂等
        currentTime,
        req.RequestID,
        req.PerUserLimitOverride,
    }

	// 执行Lua脚本
	result, err := sm.flashSaleScript.Run(ctx, sm.redis, keys, args...).Result()
	if err != nil {
		log.Errorf("秒杀脚本执行失败: %v", err)
		return nil, fmt.Errorf("秒杀执行失败: %v", err)
	}

	// 解析结果
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 3 {
		return nil, fmt.Errorf("秒杀脚本返回格式错误")
	}

	code := resultSlice[0].(int64)
	stock := resultSlice[1].(int64)
	message := resultSlice[2].(string)

	flashSaleResult := &FlashSaleResult{
		Code:        int(code),
		Success:     code == 1,
		Message:     message,
		RemainStock: int(stock),
		Timestamp:   currentTime,
	}

    // 如果秒杀成功，生成优惠券编号（引入 request_id 派生，防止并发碰撞）
    if flashSaleResult.Success {
        flashSaleResult.CouponSn = sm.generateCouponSn(req.ActivityID, req.UserID, currentTime, req.RequestID)
        if !sm.suppressLogs {
            log.Infof("秒杀成功: userID=%d, couponSn=%s, remainStock=%d", req.UserID, flashSaleResult.CouponSn, flashSaleResult.RemainStock)
        }
    } else {
        if !sm.suppressLogs {
            log.Warnf("秒杀失败: userID=%d, code=%d, message=%s", req.UserID, flashSaleResult.Code, flashSaleResult.Message)
        }
    }

	return flashSaleResult, nil
}

// Reserve 预留库存（用于事务消息方案A：预留成功后，本地事务建单，失败则回滚）
func (sm *StockManager) Reserve(ctx context.Context, req *FlashSaleRequest) (*FlashSaleResult, error) {
    if !sm.suppressLogs {
        log.Infof("开始预留库存: activityID=%d, userID=%d, couponID=%d", req.ActivityID, req.UserID, req.CouponID)
    }

    // 构建Redis keys（注意与脚本参数一致）
    keys := []string{
        fmt.Sprintf("coupon:stock:%d", req.CouponID),              // 库存key
        fmt.Sprintf("coupon:user:%d:%d", req.ActivityID, req.UserID), // 用户参与记录key
        fmt.Sprintf("coupon:activity:%d", req.ActivityID),         // 活动信息key
        fmt.Sprintf("coupon:reserve:%d:%d:%s", req.ActivityID, req.UserID, req.RequestID), // 预留幂等key
    }

    currentTime := time.Now().Unix()
    args := []interface{}{
        req.UserID,
        req.ActivityID,
        req.RequestCount,
        10, // 10秒 TTL：避免预留长时间占用，事务确认会落库；失败由回查或本地回滚处理
        currentTime,
        req.RequestID,
        req.PerUserLimitOverride,
    }

    result, err := sm.reserveScript.Run(ctx, sm.redis, keys, args...).Result()
    if err != nil {
        log.Errorf("预留脚本执行失败: %v", err)
        return nil, fmt.Errorf("预留执行失败: %v", err)
    }

    resultSlice, ok := result.([]interface{})
    if !ok || len(resultSlice) < 3 {
        return nil, fmt.Errorf("预留脚本返回格式错误")
    }

    code := resultSlice[0].(int64)
    stock := resultSlice[1].(int64)
    message := resultSlice[2].(string)

    fs := &FlashSaleResult{
        Code:        int(code),
        Success:     code == 1,
        Message:     message,
        RemainStock: int(stock),
        Timestamp:   currentTime,
    }
    if fs.Success {
        fs.CouponSn = sm.generateCouponSn(req.ActivityID, req.UserID, currentTime, req.RequestID)
        if !sm.suppressLogs {
            log.Infof("预留成功: userID=%d, couponSn=%s, remainStock=%d", req.UserID, fs.CouponSn, fs.RemainStock)
        }
    } else {
        if !sm.suppressLogs {
            log.Warnf("预留失败: userID=%d, code=%d, message=%s", req.UserID, fs.Code, fs.Message)
        }
    }
    return fs, nil
}

// PrewarmStock 预热库存到Redis
func (sm *StockManager) PrewarmStock(ctx context.Context, stockMaps map[int64]int32) error {
	if len(stockMaps) == 0 {
		return nil
	}

    if !sm.suppressLogs { log.Infof("开始库存预热，预热%d个优惠券", len(stockMaps)) }

	// 构建keys数组（key, value交替）
	keys := make([]string, 0, len(stockMaps)*2)
	for couponID, stock := range stockMaps {
		stockKey := fmt.Sprintf("coupon:stock:%d", couponID)
		keys = append(keys, stockKey, strconv.Itoa(int(stock)))
	}

	// 执行预热脚本（TTL 1小时）
	args := []interface{}{3600}
	result, err := sm.prewarmScript.Run(ctx, sm.redis, keys, args...).Result()
	if err != nil {
		log.Errorf("库存预热失败: %v", err)
		return fmt.Errorf("库存预热失败: %v", err)
	}

	// 解析结果
	resultSlice := result.([]interface{})
	code := resultSlice[0].(int64)
	successCount := resultSlice[1].(int64)
	
    if code == 1 && !sm.suppressLogs { log.Infof("库存预热成功，成功设置%d个库存", successCount) }

	return nil
}

// CheckUserLimit 检查用户频率限制
func (sm *StockManager) CheckUserLimit(ctx context.Context, userID int64, timeWindow int, maxLimit int) (bool, error) {
	keys := []string{fmt.Sprintf("coupon:limit")}
	args := []interface{}{
		userID,
		timeWindow,
		maxLimit,
		time.Now().Unix(),
	}

	result, err := sm.userLimitScript.Run(ctx, sm.redis, keys, args...).Result()
	if err != nil {
		return false, fmt.Errorf("用户限制检查失败: %v", err)
	}

	resultSlice := result.([]interface{})
	code := resultSlice[0].(int64)
	
	return code == 1, nil
}

// RollbackStock 回滚库存（用于异步处理失败）
func (sm *StockManager) RollbackStock(ctx context.Context, activityID, userID int64, couponID int64, rollbackCount int32) error {
    if !sm.suppressLogs { log.Infof("开始回滚库存: activityID=%d, userID=%d, couponID=%d, count=%d", activityID, userID, couponID, rollbackCount) }

	keys := []string{
		fmt.Sprintf("coupon:stock:%d", couponID),
		fmt.Sprintf("coupon:user:%d:%d", activityID, userID),
	}

	args := []interface{}{rollbackCount, userID}

	result, err := sm.rollbackScript.Run(ctx, sm.redis, keys, args...).Result()
	if err != nil {
		log.Errorf("库存回滚失败: %v", err)
		return fmt.Errorf("库存回滚失败: %v", err)
	}

	resultSlice := result.([]interface{})
	code := resultSlice[0].(int64)
	message := resultSlice[2].(string)

    if !sm.suppressLogs {
        if code == 1 { log.Infof("库存回滚成功: %s", message) } else { log.Warnf("库存回滚跳过: %s", message) }
    }

	return nil
}

// StartActivity 启动秒杀活动
func (sm *StockManager) StartActivity(ctx context.Context, activityInfo *ActivityInfo) error {
	// 先预设活动信息到Redis
	activityKey := fmt.Sprintf("coupon:activity:%d", activityInfo.ID)
	
	activityData := map[string]interface{}{
		"id":             activityInfo.ID,
		"coupon_id":      activityInfo.CouponID,
		"status":         1, // 待开始
		"start_time":     activityInfo.StartTime.Unix(),
		"end_time":       activityInfo.EndTime.Unix(),
		"total_count":    activityInfo.TotalCount,
		"success_count":  0,
		"per_user_limit": activityInfo.PerUserLimit,
		"created_at":     time.Now().Unix(),
	}

	err := sm.redis.HMSet(ctx, activityKey, activityData).Err()
	if err != nil {
		return fmt.Errorf("设置活动信息失败: %v", err)
	}

	// 预热库存
	stockMaps := map[int64]int32{
		activityInfo.CouponID: activityInfo.TotalCount,
	}
	if err := sm.PrewarmStock(ctx, stockMaps); err != nil {
		return fmt.Errorf("预热库存失败: %v", err)
	}

	// 更新活动状态为进行中
	keys := []string{activityKey}
	args := []interface{}{2, "start", time.Now().Unix()} // 状态2表示进行中

	_, err = sm.activityStatusScript.Run(ctx, sm.redis, keys, args...).Result()
	if err != nil {
		return fmt.Errorf("启动活动失败: %v", err)
	}

	log.Infof("秒杀活动启动成功: activityID=%d, couponID=%d, totalCount=%d", 
		activityInfo.ID, activityInfo.CouponID, activityInfo.TotalCount)

	return nil
}

// StopActivity 停止秒杀活动
func (sm *StockManager) StopActivity(ctx context.Context, activityID int64) error {
	keys := []string{fmt.Sprintf("coupon:activity:%d", activityID)}
	args := []interface{}{3, "end", time.Now().Unix()} // 状态3表示已结束

	result, err := sm.activityStatusScript.Run(ctx, sm.redis, keys, args...).Result()
	if err != nil {
		return fmt.Errorf("停止活动失败: %v", err)
	}

	resultSlice := result.([]interface{})
	code := resultSlice[0].(int64)
	
	if code == 1 {
		log.Infof("秒杀活动停止成功: activityID=%d", activityID)
	}

	return nil
}

// GetActivityStatus 获取活动状态
func (sm *StockManager) GetActivityStatus(ctx context.Context, activityID int64) (*ActivityInfo, error) {
	activityKey := fmt.Sprintf("coupon:activity:%d", activityID)
	
	result, err := sm.redis.HMGet(ctx, activityKey, 
		"id", "coupon_id", "status", "start_time", "end_time", 
		"total_count", "success_count", "per_user_limit").Result()
	if err != nil {
		return nil, fmt.Errorf("获取活动状态失败: %v", err)
	}

	// 检查活动是否存在
	if result[0] == nil {
		return nil, fmt.Errorf("活动不存在")
	}

	// 解析活动信息
	activityInfo := &ActivityInfo{}
	if id, ok := result[0].(string); ok {
		activityInfo.ID, _ = strconv.ParseInt(id, 10, 64)
	}
	if couponID, ok := result[1].(string); ok {
		activityInfo.CouponID, _ = strconv.ParseInt(couponID, 10, 64)
	}
	if status, ok := result[2].(string); ok {
		statusInt, _ := strconv.ParseInt(status, 10, 32)
		activityInfo.Status = int32(statusInt)
	}
	if startTime, ok := result[3].(string); ok {
		timestamp, _ := strconv.ParseInt(startTime, 10, 64)
		activityInfo.StartTime = time.Unix(timestamp, 0)
	}
	if endTime, ok := result[4].(string); ok {
		timestamp, _ := strconv.ParseInt(endTime, 10, 64)
		activityInfo.EndTime = time.Unix(timestamp, 0)
	}
	if totalCount, ok := result[5].(string); ok {
		count, _ := strconv.ParseInt(totalCount, 10, 32)
		activityInfo.TotalCount = int32(count)
	}
	if successCount, ok := result[6].(string); ok {
		count, _ := strconv.ParseInt(successCount, 10, 32)
		activityInfo.SuccessCount = int32(count)
	}
	if perUserLimit, ok := result[7].(string); ok {
		limit, _ := strconv.ParseInt(perUserLimit, 10, 32)
		activityInfo.PerUserLimit = int32(limit)
	}

	return activityInfo, nil
}

// GetCurrentStock 获取当前库存
func (sm *StockManager) GetCurrentStock(ctx context.Context, couponID int64) (int32, error) {
	stockKey := fmt.Sprintf("coupon:stock:%d", couponID)
	result, err := sm.redis.Get(ctx, stockKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("获取库存失败: %v", err)
	}

	stock, err := strconv.ParseInt(result, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("库存格式错误: %v", err)
	}

	return int32(stock), nil
}

// generateCouponSn 生成优惠券编号
func (sm *StockManager) generateCouponSn(activityID, userID, timestamp int64, requestID string) string {
    // 基于 request_id 派生短码，避免并发下同秒碰撞
    // 结构: FLASH{活动ID}{用户ID后4位}{base36(request_id哈希,5)}
    // 长度示例: FLASH1 0793 ABCDE => 总长 <= 5+1+4+5 = 15
    h := fnv32a(requestID)
    base := toBase36(uint64(h % 60466176)) // 36^5
    // 左填充至5位
    for len(base) < 5 { base = "0" + base }
    return fmt.Sprintf("FLASH%d%04d%s", activityID, userID%10000, strings.ToUpper(base))
}

// fnv32a 计算FNV-1a哈希
func fnv32a(s string) uint32 {
    var h uint32 = 2166136261
    for i := 0; i < len(s); i++ {
        h ^= uint32(s[i])
        h *= 16777619
    }
    return h
}

// toBase36 将数字转换为base36字符串
func toBase36(n uint64) string {
    const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
    if n == 0 { return "0" }
    b := make([]byte, 0, 16)
    for n > 0 {
        b = append([]byte{chars[n%36]}, b...)
        n /= 36
    }
    return string(b)
}

// GetUserParticipationCount 获取用户参与次数
func (sm *StockManager) GetUserParticipationCount(ctx context.Context, activityID, userID int64) (int32, error) {
	userKey := fmt.Sprintf("coupon:user:%d:%d", activityID, userID)
	result, err := sm.redis.Get(ctx, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("获取用户参与次数失败: %v", err)
	}

	count, err := strconv.ParseInt(result, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("参与次数格式错误: %v", err)
	}

	return int32(count), nil
}

// ClearActivityData 清理活动数据（活动结束后调用）
func (sm *StockManager) ClearActivityData(ctx context.Context, activityID int64, couponID int64) error {
	keys := []string{
		fmt.Sprintf("coupon:stock:%d", couponID),
		fmt.Sprintf("coupon:log:%d", activityID),
		fmt.Sprintf("coupon:activity:%d", activityID),
	}

	// 删除用户参与记录（使用模式匹配）
	userPattern := fmt.Sprintf("coupon:user:%d:*", activityID)
	userKeys, err := sm.redis.Keys(ctx, userPattern).Result()
	if err == nil {
		keys = append(keys, userKeys...)
	}

	if len(keys) > 0 {
		_, err := sm.redis.Del(ctx, keys...).Result()
		if err != nil {
			log.Errorf("清理活动数据失败: %v", err)
			return fmt.Errorf("清理活动数据失败: %v", err)
		}
	}

	log.Infof("活动数据清理完成: activityID=%d, 清理key数量=%d", activityID, len(keys))
	return nil
}
