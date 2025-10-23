package tasks

import (
    "context"
    "fmt"
    "time"

    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/pkg/log"
    redisClient "github.com/go-redis/redis/v8"
)

// StatsMergeSyncer 将 Redis 中的统计增量合并落库
// Redis 结构：
//  - Hash coupon:stats:template  field=templateID value=deltaUsed
//  - Hash coupon:stats:flashsale field=activityID value=deltaSold
// 每次合并：原子获取所有字段并清空，然后对 DB 累加。

type StatsMergeSyncer struct {
    data   interfaces.DataFactory
    redis  *redisClient.Client
    quit   chan struct{}
    tick   *time.Ticker
}

func NewStatsMergeSyncer(data interfaces.DataFactory, redis *redisClient.Client, interval time.Duration) *StatsMergeSyncer {
    if interval <= 0 { interval = 1 * time.Second }
    return &StatsMergeSyncer{data: data, redis: redis, quit: make(chan struct{}), tick: time.NewTicker(interval)}
}

func (s *StatsMergeSyncer) Start() { go s.loop() }

func (s *StatsMergeSyncer) Stop() { if s.tick != nil { s.tick.Stop() }; close(s.quit) }

func (s *StatsMergeSyncer) loop() {
    for {
        select {
        case <-s.quit:
            return
        case <-s.tick.C:
            s.mergeOnce()
        }
    }
}

var luaPopAll = redisClient.NewScript(`
local key = KEYS[1]
local res = {}
local fields = redis.call('HKEYS', key)
if fields and #fields > 0 then
  for i, f in ipairs(fields) do
    local v = redis.call('HGET', key, f)
    table.insert(res, f)
    table.insert(res, v)
  end
  redis.call('DEL', key)
end
return res
`)

func (s *StatsMergeSyncer) mergeOnce() {
    ctx := context.Background()
    // 模板 used_count
    tPairs, err := luaPopAll.Run(ctx, s.redis, []string{"coupon:stats:template"}).Result()
    if err != nil { log.Warnf("读取模板统计失败: %v", err); return }
    s.applyTemplatePairs(ctx, tPairs)
    // 活动 remaining_count（由 deltaSold 反向更新）
    aPairs, err := luaPopAll.Run(ctx, s.redis, []string{"coupon:stats:flashsale"}).Result()
    if err != nil { log.Warnf("读取活动统计失败: %v", err); return }
    s.applyActivityPairs(ctx, aPairs)
}

func (s *StatsMergeSyncer) applyTemplatePairs(ctx context.Context, pairs interface{}) {
    arr, ok := pairs.([]interface{})
    if !ok || len(arr) == 0 { return }
    for i := 0; i+1 < len(arr); i += 2 {
        idStr := fmt.Sprintf("%v", arr[i])
        deltaStr := fmt.Sprintf("%v", arr[i+1])
        var id int64
        var delta int32
        fmt.Sscanf(idStr, "%d", &id)
        var d64 int64
        fmt.Sscanf(deltaStr, "%d", &d64)
        delta = int32(d64)
        if delta == 0 || id == 0 { continue }
        if err := s.data.CouponTemplates().UpdateUsedCount(ctx, s.data.DB(), id, delta); err != nil {
            log.Warnf("合并模板统计失败: template=%d, delta=%d, err=%v", id, delta, err)
            // 回写增量，避免丢失
            _ = s.redis.HIncrBy(ctx, "coupon:stats:template", idStr, int64(delta)).Err()
        }
    }
}

func (s *StatsMergeSyncer) applyActivityPairs(ctx context.Context, pairs interface{}) {
    arr, ok := pairs.([]interface{})
    if !ok || len(arr) == 0 { return }
    for i := 0; i+1 < len(arr); i += 2 {
        idStr := fmt.Sprintf("%v", arr[i])
        deltaStr := fmt.Sprintf("%v", arr[i+1])
        var id int64
        var delta int32
        fmt.Sscanf(idStr, "%d", &id)
        var d64 int64
        fmt.Sscanf(deltaStr, "%d", &d64)
        delta = int32(d64)
        if delta == 0 || id == 0 { continue }
        // 语义调整：DB按剩余数量持久化，sold 的增量需要反向更新 remaining_count
        if err := s.data.FlashSales().UpdateSoldCount(ctx, s.data.DB(), id, -delta); err != nil {
            log.Warnf("合并活动统计失败: activity=%d, deltaSold=%d, err=%v", id, delta, err)
            _ = s.redis.HIncrBy(ctx, "coupon:stats:flashsale", idStr, int64(delta)).Err()
        }
    }
}
