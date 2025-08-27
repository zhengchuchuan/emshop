-- 检查优惠券使用条件Redis Lua脚本
-- 用于在使用优惠券前进行原子性检查和锁定
-- 
-- 参数说明:
-- KEYS[1]: 优惠券锁定key，格式: coupon:lock:{user_coupon_id}
-- KEYS[2]: 优惠券状态key，格式: coupon:status:{user_coupon_id}
-- ARGV[1]: 用户优惠券ID
-- ARGV[2]: 用户ID
-- ARGV[3]: 当前时间戳
-- ARGV[4]: 过期时间戳
-- ARGV[5]: 锁定超时时间(秒)

local lock_key = KEYS[1]
local status_key = KEYS[2]

local user_coupon_id = ARGV[1]
local user_id = ARGV[2]
local current_time = tonumber(ARGV[3])
local expired_at = tonumber(ARGV[4])
local lock_timeout = tonumber(ARGV[5])

-- 返回码定义
-- 1: 成功锁定
-- -1: 优惠券已过期
-- -2: 优惠券已被使用
-- -3: 优惠券已被锁定（正在被其他请求使用）

-- 检查优惠券是否过期
if current_time > expired_at then
    return -1  -- 已过期
end

-- 检查优惠券状态
local coupon_status = redis.call('GET', status_key)
if coupon_status and tonumber(coupon_status) ~= 1 then
    return -2  -- 已被使用或冻结（1表示未使用）
end

-- 尝试获取锁
local lock_result = redis.call('SET', lock_key, user_id, 'NX', 'EX', lock_timeout)
if not lock_result then
    return -3  -- 已被锁定
end

return 1  -- 成功锁定