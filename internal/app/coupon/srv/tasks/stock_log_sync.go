package tasks

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "time"
    "sync"

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
    workers int
    refreshEvery time.Duration
    activeActIDs []int64
    lastRefresh  time.Time
    mu     sync.RWMutex
}

// NewStockLogSyncer 创建日志落库任务
func NewStockLogSyncer(data interfaces.DataFactory, redis *redisClient.Client, batch int, interval time.Duration, workers int, refreshEvery time.Duration) *StockLogSyncer {
    if batch <= 0 { batch = 100 }
    if interval <= 0 { interval = 100 * time.Millisecond }
    if workers <= 0 { workers = 8 }
    if refreshEvery <= 0 { refreshEvery = 2 * time.Second }
    initTaskMetrics()
    return &StockLogSyncer{data: data, redis: redis, quit: make(chan struct{}), tick: time.NewTicker(interval), batch: batch, workers: workers, refreshEvery: refreshEvery}
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
    // 周期性刷新活跃活动列表，避免每个tick都查DB
    needRefresh := time.Since(s.lastRefresh) >= s.refreshEvery
    if needRefresh {
        activities, err := s.data.FlashSales().GetByStatus(ctx, s.data.DB(), do.FlashSaleStatusActive)
        if err != nil {
            log.Warnf("获取进行中的活动失败: %v", err)
        } else {
            ids := make([]int64, 0, len(activities))
            for _, a := range activities { ids = append(ids, a.ID) }
            s.mu.Lock()
            s.activeActIDs = ids
            s.lastRefresh = time.Now()
            s.mu.Unlock()
        }
    }
    s.mu.RLock()
    ids := append([]int64(nil), s.activeActIDs...)
    s.mu.RUnlock()
    if len(ids) == 0 { return }
    // 使用工作池并发处理每个活动的日志落库
    actCh := make(chan int64, len(ids))
    for _, id := range ids { actCh <- id }
    close(actCh)
    var wg sync.WaitGroup
    for w := 0; w < s.workers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for actID := range actCh {
                s.drainAndPersist(ctx, actID)
            }
        }()
    }
    wg.Wait()
}

func (s *StockLogSyncer) drainAndPersist(ctx context.Context, actID int64) {
    key := fmt.Sprintf("coupon:log:%d", actID)
    var logs []*do.FlashSaleStockLogDO
    for i := 0; i < s.batch; i++ {
        v, err := s.redis.RPop(ctx, key).Result()
        if err == redisClient.Nil {
            break
        } else if err != nil {
            log.Warnf("读取扣减日志失败: %v", err)
            break
        }
        parts := strings.Split(v, ":")
        if len(parts) < 4 { continue }
        uid, _ := strconv.ParseInt(parts[0], 10, 64)
        actID2, _ := strconv.ParseInt(parts[1], 10, 64)
        decr64, _ := strconv.ParseInt(parts[2], 10, 32)
        ts, _ := strconv.ParseInt(parts[3], 10, 64)
        reqID := ""
        if len(parts) >= 5 {
            reqID = parts[4]
        }
        // 兜底：如果 request_id 为空，基于行内容派生一个稳定ID，避免写入''触发唯一键冲突
        if strings.TrimSpace(reqID) == "" {
            reqID = fmt.Sprintf("r_%d_%d_%d_%d", uid, actID2, decr64, ts)
        }
        logs = append(logs, &do.FlashSaleStockLogDO{
            ActivityID: actID2,
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
                rid := item.RequestID
                if strings.TrimSpace(rid) == "" {
                    rid = fmt.Sprintf("r_%d_%d_%d_%d", item.UserID, item.ActivityID, item.Decr, item.TS)
                }
                raw := fmt.Sprintf("%d:%d:%d:%d:%s", item.UserID, item.ActivityID, item.Decr, item.TS, rid)
                if _, e := s.redis.LPush(ctx, key, raw).Result(); e != nil {
                    log.Errorf("回推扣减日志失败: %v", e)
                }
            }
        } else {
            log.Debugf("扣减日志入库: activity=%d, 条数=%d", actID, len(logs))
            stockLogRows.Add(float64(len(logs)))
        }
    }
}
