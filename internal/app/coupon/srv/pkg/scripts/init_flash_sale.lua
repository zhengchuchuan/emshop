-- 初始化秒杀活动Redis Lua脚本
-- 用于设置秒杀活动的初始库存和状态
-- 
-- 参数说明:
-- KEYS[1]: 秒杀活动库存key，格式: flashsale:stock:{flash_sale_id}
-- KEYS[2]: 秒杀活动状态key，格式: flashsale:status:{flash_sale_id}
-- ARGV[1]: 秒杀活动ID
-- ARGV[2]: 初始库存数量
-- ARGV[3]: 活动状态 (2表示进行中)
-- ARGV[4]: 活动结束时间戳
-- ARGV[5]: 当前时间戳

local stock_key = KEYS[1]
local status_key = KEYS[2]

local flash_sale_id = ARGV[1]
local initial_stock = tonumber(ARGV[2])
local activity_status = tonumber(ARGV[3])
local end_time = tonumber(ARGV[4])
local current_time = tonumber(ARGV[5])

-- 返回码定义
-- 1: 成功初始化
-- 0: 活动已存在，跳过初始化

-- 检查是否已经初始化
local existing_stock = redis.call('GET', stock_key)
if existing_stock then
    return 0  -- 已经初始化，跳过
end

-- 设置库存
redis.call('SET', stock_key, initial_stock)

-- 设置活动状态
redis.call('SET', status_key, activity_status)

-- 设置过期时间（活动结束后1小时自动清理）
local expire_time = end_time - current_time + 3600
redis.call('EXPIRE', stock_key, expire_time)
redis.call('EXPIRE', status_key, expire_time)

return 1  -- 成功初始化