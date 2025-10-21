package tasks

import (
    "context"
    "fmt"
    "time"

    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/internal/app/coupon/srv/domain/do"
    "emshop/internal/app/coupon/srv/consumer"
    redisClient "github.com/go-redis/redis/v8"
    "emshop/pkg/log"
)

// Reconciler 定时对账任务：比对 Redis 与 DB 的库存/销量一致性
type Reconciler struct {
    data   interfaces.DataFactory
    redis  *redisClient.Client
    quit   chan struct{}
    tick   *time.Ticker
    producer consumer.FlashSaleEventProducer
    // 控制参数
    threshold int
    maxPerRun int
    cooldown  time.Duration
    lastComp  map[int64]time.Time // 每活动上次补偿时间
}

func NewReconciler(data interfaces.DataFactory, redis *redisClient.Client, interval time.Duration, producer consumer.FlashSaleEventProducer, threshold int, maxPerRun int, cooldown time.Duration) *Reconciler {
    if interval <= 0 { interval = 30 * time.Second }
    if maxPerRun <= 0 { maxPerRun = 100 }
    if cooldown <= 0 { cooldown = 60 * time.Second }
    return &Reconciler{data: data, redis: redis, quit: make(chan struct{}), tick: time.NewTicker(interval), producer: producer, threshold: threshold, maxPerRun: maxPerRun, cooldown: cooldown, lastComp: make(map[int64]time.Time)}
}

func (r *Reconciler) Start() { go r.loop() }
func (r *Reconciler) Stop()  { if r.tick != nil { r.tick.Stop() }; close(r.quit) }

func (r *Reconciler) loop() {
    for {
        select {
        case <-r.quit:
            return
        case <-r.tick.C:
            if reconcileRuns != nil { reconcileRuns.Inc() }
            start := time.Now()
            r.checkOnce()
            if reconcileDuration != nil { reconcileDuration.Observe(time.Since(start).Seconds()) }
        }
    }
}

func (r *Reconciler) checkOnce() {
    ctx := context.Background()
    acts, err := r.data.FlashSales().GetByStatus(ctx, r.data.DB(), do.FlashSaleStatusActive)
    if err != nil {
        log.Warnf("获取进行中的活动失败: %v", err)
        return
    }
    for _, act := range acts {
        // Redis 剩余与成功数
        stockKey := fmt.Sprintf("coupon:stock:%d", act.CouponTemplateID)
        activityKey := fmt.Sprintf("coupon:activity:%d", act.ID)
        stock, err1 := r.redis.Get(ctx, stockKey).Int()
        succ, err2 := r.redis.HGet(ctx, activityKey, "success_count").Int()
        if err1 != nil || err2 != nil {
            log.Debugf("跳过对账，Redis缺少活动数据: act=%d err1=%v err2=%v", act.ID, err1, err2)
            continue
        }
        // DB 已售
        dbSold := int(act.SoldCount)
        // 期望：总量 = 剩余 + 成功
        expectedTotal := int(act.FlashSaleCount)
        actualTotal := stock + succ
        if expectedTotal != actualTotal || succ != dbSold {
            log.Warnf("对账不一致: act=%d total(exp=%d,act=%d) succ(redis=%d,db=%d)", act.ID, expectedTotal, actualTotal, succ, dbSold)
            if reconcileInconsistent != nil { reconcileInconsistent.Inc() }
            // 当 Redis 成功数大于 DB 已售时，尝试补发（受阈值/冷却/单次上限限制）
            missing := succ - dbSold
            if missing > r.threshold && r.producer != nil {
                if t, ok := r.lastComp[act.ID]; !ok || time.Since(t) >= r.cooldown {
                    toComp := missing
                    if toComp > r.maxPerRun { toComp = r.maxPerRun }
                    r.compensateMissing(ctx, act, toComp)
                    r.lastComp[act.ID] = time.Now()
                } else {
                    log.Infof("补偿冷却中: act=%d remaining=%d", act.ID, int((r.cooldown - time.Since(r.lastComp[act.ID]))/time.Second))
                    if reconcileCompSkipped != nil { reconcileCompSkipped.WithLabelValues("cooldown").Inc() }
                }
            } else if missing <= r.threshold {
                if reconcileCompSkipped != nil { reconcileCompSkipped.WithLabelValues("below_threshold").Inc() }
            } else if r.producer == nil {
                if reconcileCompSkipped != nil { reconcileCompSkipped.WithLabelValues("no_producer").Inc() }
            }
        } else {
            log.Debugf("对账一致: act=%d", act.ID)
        }
    }
}

// compensateMissing 根据日志补发缺失成功事件（幂等安全）
func (r *Reconciler) compensateMissing(ctx context.Context, act *do.FlashSaleActivityDO, missing int) {
    // 近1小时日志
    since := time.Now().Add(-1 * time.Hour).Unix()
    // 读取较多日志，逐条筛选未落库的 request_id
    logs, err := r.listLogsByActivitySince(ctx, act.ID, since, 1000)
    if err != nil {
        log.Errorf("读取扣减日志失败: %v", err)
        return
    }
    count := 0
    for _, lg := range logs {
        if lg.RequestID == "" { continue }
        existed, e := r.data.UserCoupons().GetByRequestID(ctx, r.data.DB(), lg.RequestID)
        if e != nil {
            log.Warnf("检查补偿幂等失败: %v", e)
            continue
        }
        if existed != nil {
            continue
        }
        evt := &consumer.FlashSaleSuccessEvent{
            ActivityID: act.ID,
            CouponID:   act.CouponTemplateID,
            UserID:     lg.UserID,
            CouponSn:   fmt.Sprintf("CPN%d", time.Now().UnixNano()%100000000),
            Timestamp:  time.Now().Unix(),
            RequestID:  lg.RequestID,
        }
        if err := r.producer.SendFlashSaleSuccessEvent(evt); err != nil {
            log.Errorf("补偿事件发送失败: %v", err)
            continue
        }
        count++
        if reconcileCompSent != nil { reconcileCompSent.Inc() }
        if count >= missing {
            break
        }
    }
    if count > 0 {
        log.Warnf("已补偿发送缺失成功事件: act=%d count=%d", act.ID, count)
    }
}

// listLogsByActivitySince 简化的日志读取：直接扫描 MySQL 日志表
func (r *Reconciler) listLogsByActivitySince(ctx context.Context, activityID int64, sinceTS int64, limit int) ([]*do.FlashSaleStockLogDO, error) {
    // 直接查询 DB（无专门接口，简单实现通过工厂 DB）
    var rows []*do.FlashSaleStockLogDO
    if err := r.data.DB().WithContext(ctx).
        Where("activity_id = ? AND ts >= ?", activityID, sinceTS).
        Order("ts DESC").
        Limit(limit).
        Find(&rows).Error; err != nil {
        return nil, err
    }
    return rows, nil
}
