package redis

// FlashSaleLuaScript 优惠券秒杀原子操作Lua脚本
// 确保库存扣减和用户记录的原子性，防止超卖
const FlashSaleLuaScript = `
-- 优惠券秒杀原子操作脚本
-- KEYS[1]: 库存key - coupon:stock:{coupon_id}
-- KEYS[2]: 用户key - coupon:user:{activity_id}:{user_id}  
-- KEYS[3]: 日志key - coupon:log:{activity_id}
-- KEYS[4]: 活动信息key - coupon:activity:{activity_id}

-- ARGV[1]: 用户ID
-- ARGV[2]: 活动ID
-- ARGV[3]: 扣减数量 (通常为1)
-- ARGV[4]: TTL秒数
-- ARGV[5]: 当前时间戳

local stockKey = KEYS[1]
local userKey = KEYS[2] 
local logKey = KEYS[3]
local activityKey = KEYS[4]

local userId = ARGV[1]
local activityId = ARGV[2]
local decreNum = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])
local currentTime = tonumber(ARGV[5])

-- 1. 检查活动是否存在和有效
local activityInfo = redis.call('HMGET', activityKey, 'status', 'start_time', 'end_time', 'per_user_limit')
if not activityInfo[1] then
    return {-3, 0, "活动不存在"}
end

local status = tonumber(activityInfo[1])
local startTime = tonumber(activityInfo[2]) 
local endTime = tonumber(activityInfo[3])
local perUserLimit = tonumber(activityInfo[4]) or 1

-- 检查活动状态
if status ~= 2 then  -- 2表示进行中
    return {-3, 0, "活动未开始或已结束"}
end

-- 检查活动时间
if currentTime < startTime or currentTime > endTime then
    return {-3, 0, "不在活动时间内"}
end

-- 2. 检查用户是否已参与（防重复抢购）
local userParticipated = redis.call('GET', userKey)
if userParticipated then
    local participatedCount = tonumber(userParticipated)
    if participatedCount >= perUserLimit then
        return {-2, 0, "用户已达到参与上限"}
    end
end

-- 3. 检查库存
local stock = redis.call('GET', stockKey)
if not stock then
    return {-1, 0, "库存信息不存在"}
end

stock = tonumber(stock)
if stock < decreNum then
    return {-1, stock, "库存不足"}
end

-- 4. 执行原子操作：扣库存 + 记录用户参与 + 写日志
local remainStock = stock - decreNum

-- 扣减库存
redis.call('SET', stockKey, remainStock)

-- 记录用户参与次数（累加）
local newParticipatedCount = (tonumber(userParticipated) or 0) + decreNum
redis.call('SETEX', userKey, ttl, newParticipatedCount)

-- 5. 记录抢购成功日志
local logData = string.format("%s:%s:%d:%d", userId, activityId, decreNum, currentTime)
redis.call('LPUSH', logKey, logData)
redis.call('EXPIRE', logKey, ttl)

-- 6. 更新活动统计信息
redis.call('HINCRBY', activityKey, 'success_count', decreNum)

-- 7. 如果库存为0，设置活动状态为已结束
if remainStock <= 0 then
    redis.call('HSET', activityKey, 'status', '3')  -- 3表示已结束
end

return {1, remainStock, "秒杀成功"}
`

// StockPrewarmLuaScript 库存预热Lua脚本
// 批量设置多个优惠券的库存信息
const StockPrewarmLuaScript = `
-- 库存预热脚本
-- KEYS为偶数个，格式：key1, value1, key2, value2, ...
-- ARGV[1]: TTL秒数

local ttl = tonumber(ARGV[1])
local keyCount = #KEYS

if keyCount % 2 ~= 0 then
    return {-1, "KEYS数量必须为偶数"}
end

local successCount = 0
for i = 1, keyCount, 2 do
    local key = KEYS[i]
    local value = KEYS[i + 1]
    
    -- 只有当key不存在时才设置（避免覆盖正在进行的秒杀）
    if redis.call('EXISTS', key) == 0 then
        redis.call('SETEX', key, ttl, value)
        successCount = successCount + 1
    end
end

return {1, successCount, "预热完成"}
`

// UserLimitCheckLuaScript 用户限制检查脚本
// 检查用户在指定时间窗口内的参与次数
const UserLimitCheckLuaScript = `
-- 用户限制检查脚本
-- KEYS[1]: 用户限制key前缀
-- ARGV[1]: 用户ID
-- ARGV[2]: 时间窗口(秒)
-- ARGV[3]: 最大次数限制
-- ARGV[4]: 当前时间戳

local keyPrefix = KEYS[1]
local userId = ARGV[1]
local timeWindow = tonumber(ARGV[2])
local maxLimit = tonumber(ARGV[3])
local currentTime = tonumber(ARGV[4])

-- 构建滑动窗口key
local windowStart = currentTime - timeWindow
local limitKey = keyPrefix .. ":" .. userId

-- 清理过期记录
redis.call('ZREMRANGEBYSCORE', limitKey, 0, windowStart)

-- 获取当前窗口内的记录数量
local currentCount = redis.call('ZCARD', limitKey)

if currentCount >= maxLimit then
    return {-1, currentCount, "超过频率限制"}
end

-- 添加当前记录
redis.call('ZADD', limitKey, currentTime, currentTime)
redis.call('EXPIRE', limitKey, timeWindow)

return {1, currentCount + 1, "检查通过"}
`

// StockRollbackLuaScript 库存回滚脚本
// 用于异步处理失败时回滚库存
const StockRollbackLuaScript = `
-- 库存回滚脚本
-- KEYS[1]: 库存key
-- KEYS[2]: 用户key
-- ARGV[1]: 回滚数量
-- ARGV[2]: 用户ID

local stockKey = KEYS[1]
local userKey = KEYS[2]
local rollbackNum = tonumber(ARGV[1])
local userId = ARGV[2]

-- 检查用户记录是否存在
local userRecord = redis.call('GET', userKey)
if not userRecord then
    return {-1, 0, "用户记录不存在，无需回滚"}
end

-- 增加库存
local currentStock = redis.call('GET', stockKey)
if currentStock then
    local newStock = tonumber(currentStock) + rollbackNum
    redis.call('SET', stockKey, newStock)
else
    redis.call('SET', stockKey, rollbackNum)
end

-- 删除用户记录
redis.call('DEL', userKey)

return {1, rollbackNum, "库存回滚成功"}
`

// ActivityStatusLuaScript 活动状态管理脚本
// 原子性更新活动状态和相关统计信息
const ActivityStatusLuaScript = `
-- 活动状态管理脚本
-- KEYS[1]: 活动信息key
-- ARGV[1]: 新状态
-- ARGV[2]: 操作类型 (start|end|update)
-- ARGV[3]: 当前时间戳

local activityKey = KEYS[1]
local newStatus = tonumber(ARGV[1])
local operation = ARGV[2]
local currentTime = tonumber(ARGV[3])

-- 检查活动是否存在
if redis.call('EXISTS', activityKey) == 0 then
    return {-1, 0, "活动不存在"}
end

local oldStatus = redis.call('HGET', activityKey, 'status')
if not oldStatus then
    return {-1, 0, "活动状态异常"}
end

oldStatus = tonumber(oldStatus)

-- 更新状态
redis.call('HSET', activityKey, 'status', newStatus)

-- 根据操作类型更新时间戳
if operation == 'start' then
    redis.call('HSET', activityKey, 'actual_start_time', currentTime)
elseif operation == 'end' then
    redis.call('HSET', activityKey, 'actual_end_time', currentTime)
end

-- 更新最后修改时间
redis.call('HSET', activityKey, 'updated_at', currentTime)

return {1, oldStatus, "状态更新成功"}
`