package tasks

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "time"

    "emshop/internal/app/coupon/srv/data/v1/interfaces"
    "emshop/internal/app/coupon/srv/domain/do"
    redisClient "github.com/go-redis/redis/v8"
    "emshop/pkg/log"
)

// StockLogSyncer 从 Redis 扣减日志落库到 MySQL
type StockLogSyncer struct {
    data   interfaces.DataFactory
    redis  *redisClient.Client
    quit   chan struct{}
    tick   *time.Ticker
    batch  int
}

// NewStockLogSyncer 创建日志落库任务
func NewStockLogSyncer(data interfaces.DataFactory, redis *redisClient.Client, batch int, interval time.Duration) *StockLogSyncer {
    if batch <= 0 { batch = 100 }
    if interval <= 0 { interval = 2 * time.Second }
    initTaskMetrics()
    return &StockLogSyncer{data: data, redis: redis, quit: make(chan struct{}), tick: time.NewTicker(interval), batch: batch}
}

// Start 启动任务
func (s *StockLogSyncer) Start() {
    go s.loop()
}

// Stop 停止任务
func (s *StockLogSyncer) Stop() {
    if s.tick != nil { s.tick.Stop() }
    close(s.quit)
}

func (s *StockLogSyncer) loop() {
    for {
        select {
        case <-s.quit:
            return
        case <-s.tick.C:
            s.syncOnce()
        }
    }
}

func (s *StockLogSyncer) syncOnce() {
    ctx := context.Background()
    // 仅处理进行中的活动，避免 KEYS 扫描
    activities, err := s.data.FlashSales().GetByStatus(ctx, s.data.DB(), do.FlashSaleStatusActive)
    if err != nil {
        log.Warnf("获取进行中的活动失败: %v", err)
        return
    }
    for _, act := range activities {
        key := fmt.Sprintf("coupon:log:%d", act.ID)
        var logs []*do.FlashSaleStockLogDO
        for i := 0; i < s.batch; i++ {
            // RPOP 一条日志，格式 userId:activityId:decr:ts
            v, err := s.redis.RPop(ctx, key).Result()
            if err == redisClient.Nil {
                break
            } else if err != nil {
                log.Warnf("读取扣减日志失败: %v", err)
                break
            }
            parts := strings.Split(v, ":")
            if len(parts) < 4 {
                continue
            }
            uid, _ := strconv.ParseInt(parts[0], 10, 64)
            actID, _ := strconv.ParseInt(parts[1], 10, 64)
            decr64, _ := strconv.ParseInt(parts[2], 10, 32)
            ts, _ := strconv.ParseInt(parts[3], 10, 64)
            reqID := ""
            if len(parts) >= 5 {
                reqID = parts[4]
            }
            logs = append(logs, &do.FlashSaleStockLogDO{
                ActivityID: actID,
                UserID:     uid,
                Decr:       int32(decr64),
                TS:         ts,
                RequestID:  reqID,
            })
        }
        if len(logs) > 0 {
            if err := s.data.FlashSaleStockLogs().BatchCreate(ctx, s.data.DB(), logs); err != nil {
                log.Errorf("批量写入扣减日志失败: %v", err)
                stockLogErrors.Inc()
                // 回退：将日志重新压回列表左侧，避免丢失
                for i := len(logs)-1; i >= 0; i-- {
                    item := logs[i]
                    raw := fmt.Sprintf("%d:%d:%d:%d", item.UserID, item.ActivityID, item.Decr, item.TS)
                    if _, e := s.redis.LPush(ctx, key, raw).Result(); e != nil {
                        log.Errorf("回推扣减日志失败: %v", e)
                    }
                }
            } else {
                log.Debugf("扣减日志入库: activity=%d, 条数=%d", act.ID, len(logs))
                stockLogRows.Add(float64(len(logs)))
            }
        }
    }
}
