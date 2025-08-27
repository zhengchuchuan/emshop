-- 秒杀抢购Redis Lua脚本
-- 该脚本确保秒杀操作的原子性，防止超卖
-- 
-- 参数说明:
-- KEYS[1]: 秒杀活动库存key，格式: flashsale:stock:{flash_sale_id}
-- KEYS[2]: 用户限购key，格式: flashsale:user_limit:{flash_sale_id}:{user_id}
-- KEYS[3]: 秒杀活动状态key，格式: flashsale:status:{flash_sale_id}
-- ARGV[1]: 用户ID
-- ARGV[2]: 当前时间戳
-- ARGV[3]: 秒杀活动开始时间
-- ARGV[4]: 秒杀活动结束时间
-- ARGV[5]: 每用户限购数量
-- ARGV[6]: 秒杀活动状态 (2表示进行中)

local stock_key = KEYS[1]
local user_limit_key = KEYS[2]
local status_key = KEYS[3]

local user_id = ARGV[1]
local current_time = tonumber(ARGV[2])
local start_time = tonumber(ARGV[3])
local end_time = tonumber(ARGV[4])
local per_user_limit = tonumber(ARGV[5])
local active_status = tonumber(ARGV[6])

-- 返回码定义
-- 1: 成功
-- -1: 秒杀活动未开始
-- -2: 秒杀活动已结束
-- -3: 秒杀活动已暂停或无效
-- -4: 库存不足
-- -5: 用户限购超出

-- 检查活动时间
if current_time < start_time then
    return -1  -- 秒杀未开始
end

if current_time > end_time then
    return -2  -- 秒杀已结束
end

-- 检查活动状态
local activity_status = redis.call('GET', status_key)
if not activity_status or tonumber(activity_status) ~= active_status then
    return -3  -- 活动已暂停或无效
end

-- 检查库存
local current_stock = redis.call('GET', stock_key)
if not current_stock or tonumber(current_stock) <= 0 then
    return -4  -- 库存不足
end

-- 检查用户限购
local user_purchase_count = redis.call('GET', user_limit_key)
if user_purchase_count and tonumber(user_purchase_count) >= per_user_limit then
    return -5  -- 用户限购超出
end

-- 扣减库存
redis.call('DECR', stock_key)

-- 增加用户购买计数
redis.call('INCR', user_limit_key)

-- 设置用户限购key的过期时间(秒杀结束后1小时)
redis.call('EXPIRE', user_limit_key, end_time - current_time + 3600)

return 1  -- 成功