-- 秒杀回滚Redis Lua脚本
-- 当秒杀成功后，由于其他原因需要回滚时使用
-- 
-- 参数说明:
-- KEYS[1]: 秒杀活动库存key，格式: flashsale:stock:{flash_sale_id}
-- KEYS[2]: 用户限购key，格式: flashsale:user_limit:{flash_sale_id}:{user_id}
-- ARGV[1]: 用户ID

local stock_key = KEYS[1]
local user_limit_key = KEYS[2]
local user_id = ARGV[1]

-- 返回码定义
-- 1: 成功回滚
-- 0: 无需回滚（用户未购买过）

-- 检查用户是否有购买记录
local user_purchase_count = redis.call('GET', user_limit_key)
if not user_purchase_count or tonumber(user_purchase_count) <= 0 then
    return 0  -- 无需回滚
end

-- 回滚库存
redis.call('INCR', stock_key)

-- 减少用户购买计数
redis.call('DECR', user_limit_key)

-- 如果用户购买计数归零，删除key
if tonumber(redis.call('GET', user_limit_key)) <= 0 then
    redis.call('DEL', user_limit_key)
end

return 1  -- 成功回滚