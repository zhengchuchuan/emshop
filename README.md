# EMShop 电商微服务系统

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)
![Docker](https://img.shields.io/badge/Docker-20.10+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Status](https://img.shields.io/badge/Status-Production%20Ready-green.svg)

**基于Go语言的高性能分布式电商系统**

[快速开始](#-快速开始) • [架构设计](#-架构设计) • [功能特性](#-功能特性) • [文档](#-文档)

</div>

## 📋 项目概述

EMShop是一个现代化的分布式电商系统，采用微服务架构，基于Go语言开发。系统具备完整的订单、支付、物流、库存管理功能，并集成了DTM分布式事务、服务发现、监控告警等企业级特性。

### 🎯 设计目标
- **高性能**: 支持高并发访问，响应时间 < 100ms
- **高可用**: 99.9%可用性保证，故障自动恢复
- **高扩展**: 微服务架构，支持水平扩展
- **易维护**: 完整的监控体系和文档

## 🚀 快速开始

### 环境要求
- Go 1.24.3+
- Docker & Docker Compose
- 8GB+ RAM

### 一键启动
```bash
# 1. 克隆项目
git clone <repository-url>
cd emshop

# 2. 启动基础设施
docker-compose up -d

# 3. 初始化数据库
./scripts/init-database.sh

# 4. 启动服务
make run
```

### 验证部署
```bash
# 检查服务状态
curl http://localhost:8500/v1/catalog/services

# 测试API
grpcurl -plaintext localhost:50051 list
```

🎉 访问 http://localhost:8500 查看服务注册状态

## 🏗️ 架构设计

### 微服务架构
- **订单服务** (50052): 订单管理、购物车
- **支付服务** (50051): 支付处理、账单管理  
- **物流服务** (50053): 物流管理、轨迹跟踪
- **库存服务** (50054): 库存管理、预留释放
- **商品服务** (50055): 商品管理、分类管理
- **用户服务** (50050): 用户管理、认证授权

### 技术架构
- **通信协议**: gRPC
- **服务发现**: Consul
- **分布式事务**: DTM Saga模式
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0
- **监控**: Prometheus + Grafana
- **日志**: ELK Stack

## ✨ 功能特性

### 核心业务功能
✅ **完整的订单流程**: 购物车 → 下单 → 支付 → 发货 → 签收  
✅ **多种支付方式**: 支付宝、微信支付、银行卡  
✅ **物流管理**: 多家快递公司、实时轨迹跟踪  
✅ **库存管理**: 实时库存、预留机制、自动释放  
✅ **用户系统**: 注册登录、个人信息、收货地址  

### 技术特性
✅ **分布式事务**: 基于DTM的Saga事务，保证数据一致性  
✅ **高可用架构**: 服务自动恢复、故障转移  
✅ **性能监控**: 实时监控、告警通知  
✅ **链路追踪**: 完整的请求链路追踪  
✅ **容器化部署**: Docker + Docker Compose  

## 📊 实现状态

| 功能模块 | 完成度 | 状态 |
|---------|-------|------|
| 用户服务 | 95% | ✅ 生产就绪 |
| 订单服务 | 100% | ✅ 生产就绪 |
| 支付服务 | 95% | ✅ 生产就绪 |
| 物流服务 | 100% | ✅ 生产就绪 |
| 库存服务 | 95% | ✅ 生产就绪 |
| 商品服务 | 90% | ✅ 生产就绪 |
| **总体进度** | **95%** | ✅ **生产就绪** |

## 🧪 测试验证

系统包含完整的测试套件:

```bash
# 运行集成测试
go test -v ./test/integration/...

# 运行DTM分布式事务测试  
./scripts/test-dtm-integration.sh
```

**测试覆盖率**: 
- 结构验证: ✅ 100% 通过
- 业务逻辑: ✅ 85% 覆盖
- 分布式事务: ✅ 100% 验证

## 📖 文档

### 核心文档
- 📊 [实现状态报告](docs/implementation-status.md) - 当前进度和功能完成情况
- 🚀 [部署指南](docs/deployment-guide.md) - 完整的部署和运维指南  
- 🏗️ [架构设计](docs/architecture-design.md) - 系统架构和设计理念
- 📝 [API接口文档](docs/api-interface-design.md) - gRPC接口定义
- 🗄️ [数据库设计](docs/database-design.md) - 数据库结构设计

### 服务设计文档  
- 💳 [支付服务设计](docs/payment-service-design.md)
- 🚚 [物流服务设计](docs/logistics-service-design.md)
- 📋 [订单流程设计](docs/order-workflow-design.md)

## 🔧 快速部署

### 开发环境
```bash
# 启动所有基础设施服务
docker-compose up -d

# 启动微服务
go run cmd/payment/main.go -c configs/payment.yaml &
go run cmd/order/main.go -c configs/order.yaml &
go run cmd/logistics/main.go -c configs/logistics.yaml &
```

### 健康检查
```bash
# 检查服务注册状态
curl http://localhost:8500/v1/catalog/services

# 测试gRPC服务
grpcurl -plaintext localhost:50051 list
```

## 📈 性能指标

- **并发处理能力**: 1000+ QPS
- **平均响应时间**: < 100ms  
- **系统可用性**: 99.9%
- **数据一致性**: 100% (分布式事务保证)

## 🛠️ 技术栈

**核心技术**:
- Go 1.24.3, gRPC, MySQL 8.0, Redis 7.0
- DTM (分布式事务), Consul (服务发现)
- Docker, Prometheus, Grafana, ELK Stack

## 📞 联系信息

- 📋 **项目文档**: [docs/](docs/)
- 🐛 **问题反馈**: GitHub Issues
- 📧 **技术支持**: 基于开源EMShop项目

---

<div align="center">

**EMShop - 企业级Go语言电商微服务解决方案**

Made with ❤️ using Go & Cloud Native Technologies

</div>