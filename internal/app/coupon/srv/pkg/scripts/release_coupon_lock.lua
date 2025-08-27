-- 释放优惠券锁定Redis Lua脚本
-- 用于释放优惠券使用后的锁定状态
-- 
-- 参数说明:
-- KEYS[1]: 优惠券锁定key，格式: coupon:lock:{user_coupon_id}
-- KEYS[2]: 优惠券状态key，格式: coupon:status:{user_coupon_id}
-- ARGV[1]: 用户ID
-- ARGV[2]: 操作类型: 'use'表示使用成功, 'release'表示释放锁定

local lock_key = KEYS[1]
local status_key = KEYS[2]

local user_id = ARGV[1]
local operation = ARGV[2]

-- 返回码定义
-- 1: 成功释放
-- 0: 锁不存在或不属于当前用户

-- 检查锁的所有者
local lock_owner = redis.call('GET', lock_key)
if not lock_owner or lock_owner ~= user_id then
    return 0  -- 锁不存在或不属于当前用户
end

-- 删除锁
redis.call('DEL', lock_key)

-- 根据操作类型更新状态
if operation == 'use' then
    -- 标记为已使用
    redis.call('SET', status_key, 2)  -- 2表示已使用
elseif operation == 'release' then
    -- 保持原状态（回滚到未使用状态）
    redis.call('SET', status_key, 1)  -- 1表示未使用
end

return 1  -- 成功释放