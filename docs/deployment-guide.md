# EMShop 部署指南

## 🚀 快速部署

### 先决条件
- Docker & Docker Compose
- Go 1.24.3+
- Git

### 1. 获取项目
```bash
git clone <repository-url>
cd emshop
```

### 2. 启动基础设施服务
```bash
# 启动所有基础设施服务
docker-compose up -d

# 验证服务状态
docker-compose ps
```

#### 服务端口映射
- **MySQL**: 3306
- **Redis**: 6379
- **Consul**: 8500 (UI: http://localhost:8500)
- **DTM**: 36789 (HTTP), 36790 (gRPC)
- **Elasticsearch**: 9200
- **Kibana**: 5601 (UI: http://localhost:5601)
- **Prometheus**: 9090 (UI: http://localhost:9090)
- **Grafana**: 3000 (UI: http://localhost:3000, admin/admin)

### 3. 初始化数据库
```bash
# 自动创建所有数据库和表结构
./scripts/init-database.sh
```

### 4. 启动微服务

#### 方式1: 使用Go直接运行
```bash
# 支付服务
go run cmd/payment/main.go -c configs/payment.yaml

# 订单服务  
go run cmd/order/main.go -c configs/order.yaml

# 物流服务
go run cmd/logistics/main.go -c configs/logistics.yaml

# 库存服务
go run cmd/inventory/main.go -c configs/inventory.yaml
```

#### 方式2: 编译后运行
```bash
# 编译所有服务
make build

# 运行服务
./bin/payment-srv -c configs/payment.yaml &
./bin/order-srv -c configs/order.yaml &
./bin/logistics-srv -c configs/logistics.yaml &
./bin/inventory-srv -c configs/inventory.yaml &
```

### 5. 验证部署

#### 检查服务注册
```bash
# 查看Consul中的服务注册
curl http://localhost:8500/v1/catalog/services

# 查看健康检查
curl http://localhost:8500/v1/health/checks
```

#### 测试gRPC服务
```bash
# 安装grpcurl (如果未安装)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 测试支付服务
grpcurl -plaintext localhost:50051 list

# 测试订单服务
grpcurl -plaintext localhost:50052 list

# 测试物流服务
grpcurl -plaintext localhost:50053 list
```

## 🔧 配置说明

### 环境变量配置
创建 `.env` 文件:
```env
# 数据库配置
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=emshop
MYSQL_PASSWORD=emshop123

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379

# Consul配置
CONSUL_HOST=localhost
CONSUL_PORT=8500

# DTM配置
DTM_GRPC=localhost:36790
DTM_HTTP=localhost:36789
```

### 服务配置文件
每个服务的配置文件位于 `configs/` 目录:
- `payment.yaml` - 支付服务配置
- `order.yaml` - 订单服务配置
- `logistics.yaml` - 物流服务配置
- `inventory.yaml` - 库存服务配置

## 🧪 测试与验证

### 集成测试
```bash
# 运行集成测试
go test -v ./test/integration/...

# 运行DTM分布式事务测试
./scripts/test-dtm-integration.sh
```

### 功能测试
```bash
# 测试完整的订单流程
./scripts/test-order-flow.sh

# 测试支付流程
./scripts/test-payment-flow.sh
```

## 📊 监控与维护

### 日志查看
```bash
# 查看服务日志
docker-compose logs -f payment-srv
docker-compose logs -f order-srv

# 查看基础设施日志
docker-compose logs -f mysql
docker-compose logs -f dtm
```

### 性能监控
- **Grafana Dashboard**: http://localhost:3000
  - 默认用户: admin/admin
  - 预配置了微服务监控面板

- **Prometheus Metrics**: http://localhost:9090
  - 查看各服务的指标数据

### 健康检查
```bash
# 检查所有服务健康状态
./scripts/health-check.sh
```

## 🛠️ 开发环境配置

### 开发工具安装
```bash
# protobuf编译器
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 代码生成工具
go install github.com/spf13/cobra@latest

# gRPC测试工具
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### 代码生成
```bash
# 生成protobuf代码
make proto

# 生成错误码
make codegen

# 格式化代码
make fmt
```

## 🐛 故障排除

### 常见问题

#### 1. 服务启动失败
```bash
# 检查端口占用
netstat -tuln | grep :50051

# 检查配置文件
./bin/payment-srv -c configs/payment.yaml --validate-config
```

#### 2. 数据库连接失败
```bash
# 测试数据库连接
mysql -h localhost -u emshop -pemshop123

# 检查数据库是否存在
mysql -h localhost -u emshop -pemshop123 -e "SHOW DATABASES;"
```

#### 3. DTM事务失败
```bash
# 检查DTM服务状态
curl http://localhost:36789/health

# 查看DTM日志
docker-compose logs dtm
```

#### 4. 服务发现问题
```bash
# 检查Consul状态
curl http://localhost:8500/v1/status/leader

# 重新注册服务
curl -X PUT http://localhost:8500/v1/agent/service/register \
  -d @configs/consul/payment-service.json
```

## 🔒 安全配置

### 生产环境建议
1. **更改默认密码**: 修改MySQL、Redis等默认密码
2. **启用TLS**: 配置gRPC服务使用TLS
3. **网络隔离**: 使用Docker网络隔离服务
4. **访问控制**: 配置防火墙规则
5. **日志脱敏**: 确保敏感信息不出现在日志中

### SSL/TLS配置
```yaml
# 在服务配置中启用TLS
server:
  tls:
    enabled: true
    cert_file: /etc/ssl/server.crt
    key_file: /etc/ssl/server.key
```

## 📞 技术支持

### 问题报告
如遇问题，请收集以下信息:
1. 错误日志
2. 服务配置
3. 系统环境信息
4. 复现步骤

### 性能调优
- **数据库连接池**: 根据负载调整连接数
- **gRPC连接复用**: 启用keepalive
- **缓存策略**: 合理使用Redis缓存
- **资源限制**: 设置合适的Docker资源限制

---

**更新日期**: 2025-01-25  
**文档版本**: v2.0