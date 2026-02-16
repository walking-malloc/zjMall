# zjMall - 电商 B2C 微服务系统

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)
![Vue Version](https://img.shields.io/badge/Vue-3.3-green.svg)
![License](https://img.shields.io/badge/license-MIT-orange.svg)

一个基于 Go 微服务架构的电商 B2C 平台，支持完整的购物流程：浏览→加购→下单→支付

[功能特性](#-功能特性) • [技术栈](#️-技术栈) • [快速开始](#-快速开始) • [项目结构](#-项目结构) • [文档](#-文档)

</div>

---

## 📋 项目简介

zjMall 是一个面向 B2C 零售的电商微服务系统，完成了从商品浏览到售后服务的完整电商闭环，并预留了大促扩展能力（限流、降级、秒杀隔离）。

### 核心特点

- 🏗️ **微服务架构**：采用 gRPC + REST 双协议，服务间通过 gRPC 通信，对外提供 REST API
- 🔐 **统一鉴权**：基于 JWT 的认证授权，支持 RBAC 权限控制
- 📦 **领域驱动**：按业务域拆分服务，职责清晰，易于扩展
- 🚀 **高性能**：Redis 缓存、Elasticsearch 搜索、RabbitMQ 异步消息处理
- 🐳 **容器化部署**：Docker Compose 一键启动所有基础设施

## ✨ 功能特性

### 已实现功能

- ✅ **用户服务**：注册登录（短信/密码）、JWT 认证、用户信息管理、收货地址管理、头像上传
- ✅ **商品服务**：SPU/SKU 管理、类目管理、品牌管理、商品搜索、商品详情
- ✅ **购物车服务**：加购、修改、删除、合并、结算预览
- ✅ **订单服务**：下单、订单查询、订单状态管理、超时取消
- ✅ **支付服务**：支付单创建、支付回调、退款处理
- ✅ **库存服务**：库存查询、预占、扣减、防超卖

### 规划中功能

- 🔄 促销/优惠券服务
- 🔄 搜索服务（Elasticsearch 集成）
- 🔄 履约/物流服务
- 🔄 售后服务
- 🔄 评价服务

## 🛠️ 技术栈

### 后端技术

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.24+ | 主要开发语言 |
| gRPC | 1.77+ | 服务间通信 |
| gRPC-Gateway | 2.27+ | REST API 网关 |
| GORM | 1.31+ | ORM 框架 |
| JWT | 5.3+ | 身份认证 |
| Protocol Buffers | 3.x | API 契约定义 |

### 前端技术

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.3+ | 前端框架 |
| Element Plus | 2.4+ | UI 组件库 |
| Pinia | 2.1+ | 状态管理 |
| Vite | 5.0+ | 构建工具 |
| Axios | 1.6+ | HTTP 客户端 |

### 基础设施

| 组件 | 版本 | 用途 |
|------|------|------|
| MySQL | 8.0 | 主数据库 |
| Redis | 7.2 | 缓存/会话存储 |
| Elasticsearch | 8.17 | 搜索引擎 |
| RabbitMQ | 3.13+ | 消息队列 |
| Nacos | 2.3+ | 服务注册/配置中心 |

## 🚀 快速开始

### 前置要求

- Go 1.24+
- Node.js 18+
- Docker & Docker Compose
- MySQL 8.0+（或使用 Docker）
- Redis 7.2+（或使用 Docker）

### 1. 克隆项目

```bash
git clone git@github.com:walking-malloc/zjMall.git
cd zjMall
```

### 2. 启动基础设施

使用 Docker Compose 一键启动所有基础设施：

```bash
docker-compose up -d
```

这将启动以下服务：
- MySQL (端口: 3307)
- Redis (端口: 6380)
- Elasticsearch (端口: 9200)
- RabbitMQ (端口: 5673, 管理界面: 15673)
- Nacos (端口: 8848)

### 3. 初始化数据库

数据库初始化脚本位于 `deploy/mysql/init/` 目录，Docker Compose 会自动执行。

### 4. 配置服务

配置文件：

```bash
configs/config.yaml
```

根据实际情况修改配置（数据库连接、Redis、Nacos 等）。

### 5. 启动后端服务

```bash
#生成proto代码
 .\scripts\generate-all.bat

# 启动用户服务
go run cmd/user-service/main.go

# 启动商品服务
go run cmd/product-service/main.go

# 启动库存服务
go run cmd/inventory-service/main.go

# 启动购物车服务
go run cmd/cart-service/main.go

# 启动订单服务
go run cmd/order-service/main.go

# 启动支付服务
go run cmd/payment-service/main.go

```

### 6. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认运行在 `http://localhost:3000`

## 📁 项目结构

```
zjMall/
├── api/                    # API 定义
│   └── proto/              # Protocol Buffers 定义文件
├── cmd/                    # 服务入口
│   ├── user-service/       # 用户服务
│   ├── product-service/    # 商品服务
│   ├── cart-service/       # 购物车服务
│   ├── order-service/      # 订单服务
│   ├── payment-service/    # 支付服务
│   └── inventory-service/  # 库存服务
├── internal/               # 内部代码
│   ├── common/            # 公共组件
│   │   ├── middleware/    # 中间件（认证、RBAC等）
│   │   ├── cache/         # 缓存封装
│   │   └── server/        # gRPC 服务器封装
│   ├── user-service/      # 用户服务实现
│   ├── product-service/   # 商品服务实现
│   ├── cart-service/      # 购物车服务实现
│   ├── order-service/     # 订单服务实现
│   ├── payment-service/   # 支付服务实现
│   └── inventory-service/ # 库存服务实现
├── pkg/                    # 公共包
│   ├── jwt.go            # JWT 工具
│   └── validator.go      # 参数校验
├── frontend/              # 前端项目
│   ├── src/
│   │   ├── api/          # API 接口
│   │   ├── views/        # 页面组件
│   │   ├── stores/       # 状态管理
│   │   └── router/       # 路由配置
│   └── package.json
├── configs/               # 配置文件
├── deploy/                # 部署相关
│   └── mysql/            # 数据库初始化脚本
├── docs/                  # 项目文档
├── scripts/               # 工具脚本
├── docker-compose.yml     # Docker Compose 配置
└── go.mod                 # Go 依赖管理
```

## 🏗️ 架构设计

### 微服务架构图

```
┌─────────────┐
│   Client    │ (Web/APP/小程序)
└──────┬──────┘
       │ HTTP/REST
       ▼
┌─────────────┐
│ API Gateway │ (gRPC-Gateway)
└──────┬──────┘
       │ gRPC
       ▼
┌─────────────────────────────────────┐
│         Microservices               │
│  ┌──────────┐  ┌──────────┐        │
│  │   User   │  │ Product  │  ...   │
│  └──────────┘  └──────────┘        │
└──────┬──────────────────┬───────────┘
       │                  │
       ▼                  ▼
┌──────────┐      ┌──────────┐
│  MySQL   │      │  Redis   │
└──────────┘      └──────────┘
```

### 核心服务说明

| 服务 | 端口 | 职责 |
|------|------|------|
| user-service | 50051 | 用户注册登录、信息管理、地址管理 |
| product-service | 50052 | 商品管理、类目管理、品牌管理 |
| cart-service | 50053 | 购物车管理、结算预览 |
| order-service | 50054 | 订单创建、查询、状态管理 |
| payment-service | 50055 | 支付单管理、支付回调、退款 |
| inventory-service | 50056 | 库存查询、预占、扣减 |


## 🔧 开发指南

### 代码生成

使用 Buf 生成 gRPC 代码：

```bash
# 安装 buf
go install github.com/bufbuild/buf/cmd/buf@latest

# 生成代码
buf generate
```

### 添加新服务

1. 在 `api/proto/` 下定义 `.proto` 文件
2. 运行 `buf generate` 生成代码
3. 在 `cmd/` 下创建服务入口
4. 在 `internal/` 下实现服务逻辑
5. 更新 `docker-compose.yml` 和配置文件

### API 测试

项目提供了 OpenAPI 文档，位于 `docs/openapi/` 目录，可以使用 Postman 或 Swagger UI 导入测试。


______________________________________________

<div align="center">

**如果这个项目对你有帮助，请给一个 ⭐ Star！**

Made with ❤️ by walking-malloc 

</div>
