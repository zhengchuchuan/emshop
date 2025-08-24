# Canal + RocketMQ + Elasticsearch 数据同步实现方案

## 项目概述

本方案为 `emshop` 项目实现从 MySQL `emshop_goods_srv` 数据库到 Elasticsearch 的实时数据同步，使用 Alibaba Canal + Apache RocketMQ 作为中间件。

### 技术架构

```
MySQL (emshop_goods_srv) → Canal Server → RocketMQ → Canal Consumer → Elasticsearch
```

## 已有实现分析

### ✅ 已完成组件

1. **Canal Consumer 实现** (`/internal/app/goods/srv/consumer/canal_consumer.go`)
   - 完整的 RocketMQ 消费者实现
   - 支持商品、品牌、分类、轮播图等多表变更处理
   - 包含 Prometheus 监控指标

2. **数据同步管理器** (`/internal/app/goods/srv/data/v1/sync/sync_manager.go`)
   - 实现了完整的 MySQL → Elasticsearch 同步逻辑
   - 支持批量同步和增量同步

3. **Elasticsearch 搜索层** (`/internal/app/goods/srv/data/v1/elasticsearch/goods.go`)
   - 完整的商品搜索 CRUD 操作
   - 支持复杂查询过滤

4. **配置结构** (`/configs/goods-canal.yaml`)
   - RocketMQ 配置已定义
   - Elasticsearch 配置已定义

5. **Docker Compose 基础设施** (`/components/mysql-canal/docker-compose.yml`)
   - MySQL、Canal Admin、Canal Server 配置
   - RocketMQ 集成配置

## 实施计划

### 阶段一：配置优化 (预计时间: 30分钟)

#### 1.1 Canal Server 配置优化
- **文件**: `components/mysql-canal/docker-compose.yml`
- **任务**: 
  - 确认数据库过滤规则：`emshop_goods_srv\\..*`
  - 验证 RocketMQ 连接配置

#### 1.2 应用配置验证
- **文件**: `configs/goods-canal.yaml`
- **任务**:
  - 验证 MySQL 连接参数
  - 确认 RocketMQ nameservers 配置
  - 检查 Elasticsearch 连接配置

### 阶段二：服务集成验证 (预计时间: 45分钟)

#### 2.1 Canal Consumer 集成验证
- **文件**: `internal/app/goods/srv/app.go:53-69`
- **任务**:
  - 验证 Canal Consumer 自动启动逻辑
  - 确保正确的配置参数传递
  - 添加启动状态检查

#### 2.2 数据库权限和连接
- **任务**:
  - 确认 MySQL binlog 开启状态
  - 验证 Canal 用户权限
  - 检查数据库表结构

### 阶段三：端到端测试 (预计时间: 60分钟)

#### 3.1 基础连接测试
- **工具**: 使用现有测试脚本 `scripts/test-canal-consumer.go`
- **任务**:
  - 验证 RocketMQ 连接
  - 测试 Canal 消息接收
  - 检查消息格式解析

#### 3.2 数据同步测试
- **测试表**: `emshop_goods_srv.goods`
- **操作**:
  - INSERT: 新增商品记录
  - UPDATE: 更新商品信息
  - DELETE: 删除商品记录

#### 3.3 Elasticsearch 同步验证
- **任务**:
  - 验证商品数据正确同步到 ES
  - 检查搜索功能正常性
  - 测试批量同步功能

### 阶段四：监控和日志 (预计时间: 30分钟)

#### 4.1 监控指标验证
- **组件**: Canal Consumer Prometheus 指标
- **指标**:
  - `canal_message_total`: 处理消息总数
  - `canal_sync_latency_seconds`: 同步延迟
  - `canal_error_total`: 错误总数

#### 4.2 日志系统
- **任务**:
  - 确认详细日志输出
  - 验证错误日志记录
  - 检查性能日志

### 阶段五：生产优化 (预计时间: 45分钟)

#### 5.1 性能调优
- **配置项**:
  - RocketMQ 消费者并发度
  - Elasticsearch 批量操作大小
  - Canal 消费者重试策略

#### 5.2 故障处理
- **实现**:
  - 死信队列处理机制
  - 自动重连逻辑
  - 数据一致性检查

## 技术实现细节

### Canal Consumer 消息处理流程

```go
// 消息流转: RocketMQ → Canal Consumer → Data Sync Manager → Elasticsearch
1. RocketMQ 接收 Canal binlog 消息
2. CanalConsumer.handleMessage() 解析消息
3. 根据表类型调用对应处理器 (handleGoodsChange)
4. DataSyncManager.SyncToSearch() 同步到 ES
5. Elasticsearch 索引更新完成
```

### 支持的数据表

- **goods**: 商品主表 ✅ (已实现完整同步)
- **brands**: 品牌表 ⚠️ (预留接口)
- **category**: 分类表 ⚠️ (预留接口)  
- **category_brand**: 分类品牌关联 ⚠️ (预留接口)
- **banner**: 轮播图 ⚠️ (预留接口)

### 关键代码位置

- **Canal Consumer**: `internal/app/goods/srv/consumer/canal_consumer.go`
- **Sync Manager**: `internal/app/goods/srv/data/v1/sync/sync_manager.go`
- **ES Search**: `internal/app/goods/srv/data/v1/elasticsearch/goods.go`
- **App 集成**: `internal/app/goods/srv/app.go:53-69`

## 配置参数详解

### RocketMQ 配置
```yaml
rocketmq:
  nameservers: ["localhost:9876"]
  consumer_group: "goods-sync-consumer-group"
  topic: "goods-binlog-topic"
  max_reconsume: 3
```

### Canal Server 关键配置
```yaml
# Docker Compose 环境变量
canal.serverMode: rocketMQ
canal.mq.servers: rmqnamesrv:9876
canal.mq.topic: goods-binlog-topic
canal.instance.filter.regex: emshop_goods_srv\\..*
```

### MySQL binlog 要求
```sql
-- 必需的 binlog 配置
log-bin = mysql-bin
binlog-format = ROW
server-id = 1
```

## 部署和启动流程

### 1. 基础设施启动
```bash
# 启动 MySQL, Canal, RocketMQ
cd /home/zcc/project/golang/emshop/emshop/components/mysql-canal
docker-compose up -d

# 启动 ELK (Elasticsearch)
cd ../elk
docker-compose up -d elasticsearch
```

### 2. 应用启动
```bash
# 启动商品服务 (包含 Canal Consumer)
cd /home/zcc/project/golang/emshop/emshop
go run cmd/goods/goods.go -c configs/goods-canal.yaml
```

### 3. 验证流程
```bash


# 2. 验证 RocketMQ
# 使用测试消费者: scripts/test-canal-consumer.go

# 3. 测试数据同步
# 在 emshop_goods_srv.goods 表执行 INSERT/UPDATE/DELETE
```

## 监控和维护

### 健康检查
- RocketMQ Console: 需要额外部署
- Elasticsearch: `http://localhost:9200/_cluster/health`

### 日志位置
- Canal Server: `components/mysql-canal/canal-server/logs/`
- 应用日志: `logs/emshop-goods-srv.log`

### 故障排查

#### 常见问题
1. **RocketMQ 无消息**: 检查 Canal Server 配置和网络连通性
2. **ES 同步失败**: 验证 Elasticsearch 连接和索引配置
3. **Canal 连接失败**: 检查 MySQL binlog 配置和用户权限

#### 调试工具
- RocketMQ 消费者测试工具 (`scripts/test-canal-consumer.go`)
- Elasticsearch 查询验证

## 性能指标

### 预期性能
- **同步延迟**: < 1秒 (正常网络条件下)
- **吞吐量**: 支持 1000+ TPS 的数据变更
- **资源消耗**: 
  - Canal Server: ~200MB 内存
  - Canal Consumer: ~100MB 内存

### 扩展性
- RocketMQ 支持水平扩展
- Canal 支持多实例部署
- Elasticsearch 集群模式

## 下一步优化

1. **多表关联同步**: 实现品牌、分类变更对商品搜索的影响
2. **缓存层集成**: 添加 Redis 缓存同步
3. **监控告警**: 集成 Prometheus + Grafana
4. **自动化测试**: 完善端到端测试用例

---

## 总结

本方案基于项目现有的完善实现，主要工作是配置验证、集成测试和性能优化。预计总实施时间约 3.5 小时，包含完整的测试和验证流程。

**关键优势**:
- 实时同步 (< 1秒延迟)
- 高可用架构 (Canal + RocketMQ)
- 完整监控体系 (Prometheus 指标)
- 生产就绪 (错误处理、重试机制)