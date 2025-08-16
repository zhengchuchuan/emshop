# 库存管理重构方案

## 概述

当前项目存在库存数据冗余问题：`goods`表和`inventory`表都维护了`stocks`字段，导致数据不一致风险。本方案提出统一使用独立的库存数据库进行库存管理。

## 问题分析

### 1. 当前架构问题
- **数据冗余**: goods表和inventory表都存储stocks字段，造成数据重复
- **同步问题**: 两个表的库存数据没有自动同步机制，容易出现数据不一致
- **职责混乱**: 商品表承担了库存管理职责，违反单一职责原则

### 2. 代码证据
- `internal/app/goods/srv/domain/do/goods.go:55`: goods表定义了`Stocks int32`
- `internal/app/inventory/srv/domain/do/inventory.go:32`: inventory表也定义了`Stocks int32`
- `internal/app/emshop/api/controller/goods/v1/goods.go`: 商品详情已通过独立库存服务获取库存

## 重构方案

### 阶段1: 数据库结构调整

#### 1.1 删除goods表的stocks字段
```sql
ALTER TABLE goods DROP COLUMN stocks;
```

#### 1.2 确保inventory表作为唯一库存数据源
保留现有inventory表结构：
```go
type InventoryDO struct {
    bgorm.BaseModel
    Goods   int32 `gorm:"type:int;index"`
    Stocks  int32 `gorm:"type:int"`
    Version int32 `gorm:"type:int"` // 分布式锁的乐观锁
}
```

### 阶段2: 代码重构

#### 2.1 修改goods相关结构体
- 删除 `internal/app/emshop/api/domain/request/goods.go:19,43` 中的stocks字段验证
- 删除 `api/goods/v1/goods.proto:185` 中的stocks字段定义
- 更新goods创建/更新逻辑，移除stocks处理

#### 2.2 更新API接口
- 修改商品创建接口，不再接收stocks参数
- 修改商品列表/详情接口，通过inventory服务获取库存信息
- 保持现有库存服务接口不变

#### 2.3 完善库存服务集成
- 确保所有需要库存信息的接口都通过inventory服务获取
- 优化库存服务调用性能
- 添加库存服务异常处理

### 阶段3: 数据迁移和验证

#### 3.1 数据迁移脚本
1. 对比goods表和inventory表的库存数据差异
2. 以inventory表为准，修正数据不一致问题
3. 备份原有数据
4. 执行删除goods.stocks字段操作

#### 3.2 功能验证
- 商品创建功能测试
- 商品列表/详情展示测试
- 库存扣减/恢复功能测试
- 并发库存操作测试

## 实施计划

### 第一步: 代码重构 (不涉及数据库变更)
1. 修改goods相关代码，移除stocks字段依赖
2. 确保所有库存操作都通过inventory服务
3. 完善单元测试和集成测试

### 第二步: 数据一致性检查
1. 编写脚本检查goods和inventory表库存数据差异
2. 修正数据不一致问题
3. 确保inventory表数据完整性

### 第三步: 数据库结构调整
1. 在测试环境执行删除goods.stocks字段
2. 验证所有功能正常运行
3. 生产环境执行变更

### 第四步: 监控和优化
1. 监控库存服务性能
2. 优化库存查询缓存策略
3. 完善库存异常告警机制

## 预期收益

1. **数据一致性**: 单一数据源，避免同步问题
2. **职责清晰**: 商品服务专注商品信息，库存服务专注库存管理
3. **并发安全**: inventory服务已有完善的乐观锁和分布式事务支持
4. **扩展性**: 便于后续库存策略优化和分库分表

## 风险评估

### 潜在风险
1. 库存服务故障导致商品信息显示异常
2. 数据库变更过程中的服务中断
3. 现有代码依赖stocks字段的未发现逻辑

### 风险缓解
1. 增加库存服务降级策略
2. 分阶段执行，充分测试
3. 全面代码review，确保清理完整

## 时间安排

- 代码重构: 2-3天
- 测试验证: 1-2天  
- 数据迁移: 1天
- 生产部署: 1天
- **总计**: 约1周

## 结论

删除goods表中的stocks字段，统一使用inventory表进行库存管理是必要且可行的。这将显著提升系统的数据一致性和架构清晰度，为后续功能扩展奠定良好基础。