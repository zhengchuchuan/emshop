# EMShop Components

EMShop 项目的 Docker 容器化组件集合，包含微服务架构所需的所有基础设施服务。

## 架构概览

本项目使用 Docker Compose 管理多个基础设施服务：

- **数据存储**: MySQL, Redis
- **服务发现**: Consul, Nacos  
- **消息队列**: RocketMQ
- **搜索引擎**: Elasticsearch + Kibana
- **监控**: Prometheus + Grafana
- **日志**: Logstash
- **API 网关**: Kong
- **分布式事务**: DTM
- **链路追踪**: Jaeger
- **CI/CD**: Jenkins
- **数据同步**: Canal

## 命名规范

所有服务统一使用 `emshop-` 前缀：

- **容器名**: `emshop-mysql`, `emshop-redis`, `emshop-consul` 等
- **卷名**: `emshop-mysql-data`, `emshop-redis-data` 等
- **网络**: `emshop-net`

## 快速开始

### 1. 初始化配置
```bash
# 初始化必需的配置文件
./scripts/init-configs.sh
```

### 2. 数据迁移（可选）
如果你有现有的绑定挂载数据需要迁移：
```bash
# 迁移现有数据到命名卷
./scripts/migrate-to-volumes.sh
```

### 3. 启动服务
```bash
# 启动所有服务
docker-compose up -d
```

## 服务访问地址

| 服务 | 地址 | 用途 |
|------|------|------|
| Nacos | http://localhost:8848 | 配置中心和服务发现 |
| Consul | http://localhost:8500 | 服务发现 |
| Kibana | http://localhost:5601 | 日志分析 |
| Prometheus | http://localhost:9090 | 监控指标 |
| Grafana | http://localhost:3000 | 监控仪表板 |
| RocketMQ Console | http://localhost:8080 | 消息队列管理 |
| Canal Admin | http://localhost:18089 | 数据同步管理 |
