# 🏠 安居客 (Anjuke) - 房产交易平台后端系统

[![Go Version](https://img.shields.io/badge/Go-1.24.4-blue.svg)](https://golang.org/)
[![Kratos](https://img.shields.io/badge/Kratos-v2.8.4-green.svg)](https://go-kratos.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-91.8%25-brightgreen.svg)](TEST_README.md)

一个基于 Go + Kratos 框架构建的现代化房产交易平台后端系统，采用微服务架构和领域驱动设计（DDD），提供完整的用户管理、房源管理、交易处理、积分系统等核心功能。

## 🚀 项目特色

- **🏗️ 微服务架构**: 基于 Kratos 框架的现代化微服务设计
- **📐 领域驱动设计**: 采用 DDD 分层架构，业务逻辑清晰
- **🔄 双协议支持**: 同时支持 gRPC 和 HTTP/RESTful API
- **📊 完整积分系统**: 签到、消费获得积分、积分抵扣等功能
- **📱 短信验证**: 集成短信服务，支持多场景验证码
- **🔐 实名认证**: 对接第三方实名认证服务
- **📁 智能文件上传**: 支持分片上传、断点续传、进度回调
- **🧪 高测试覆盖**: 91.8% 的单元测试覆盖率
- **🐳 容器化部署**: 完整的 Docker 容器化方案

## 🛠️ 技术栈

### 核心框架
- **[Kratos v2.8.4](https://go-kratos.dev/)** - Go 微服务框架
- **[gRPC](https://grpc.io/)** - 高性能 RPC 框架
- **[Protocol Buffers](https://protobuf.dev/)** - 接口定义语言

### 数据存储
- **[MySQL 8.0](https://www.mysql.com/)** - 主数据库
- **[Redis 6.2](https://redis.io/)** - 缓存和会话存储
- **[MinIO](https://min.io/)** - 对象存储服务

### 开发工具
- **[GORM](https://gorm.io/)** - ORM 框架
- **[Wire](https://github.com/google/wire)** - 依赖注入
- **[Testify](https://github.com/stretchr/testify)** - 测试框架

### 第三方服务
- **腾讯云实名认证** - 身份验证服务
- **数脉短信服务** - 短信验证码发送

## 📋 功能模块

### 👤 用户服务 (User Service)
- 用户注册、登录（密码/短信验证码）
- 实名认证（对接第三方服务）
- 短信验证码发送与验证
- 用户状态管理

### 🏠 房源服务 (House Service)
- 房源信息管理
- 房源搜索与筛选
- 房源图片上传

### 💰 交易服务 (Transaction Service)
- 交易订单管理
- 支付流程处理
- 交易状态跟踪

### 🎯 积分服务 (Points Service)
- 每日签到获得积分
- 消费获得积分（1元=1积分）
- 积分抵扣（10积分=1元）
- 积分明细查询
- 连续签到奖励

### 📁 文件上传服务 (Upload Service)
- 智能上传（自动选择普通/分片上传）
- 大文件分片上传
- 断点续传
- 上传进度回调
- 多文件批量上传

### 🎧 客服服务 (Customer Service)
- 客户信息管理
- 服务记录跟踪

## 🏗️ 项目架构

```
server/
├── api/                    # API 定义 (Protocol Buffers)
│   ├── common/            # 通用消息定义
│   ├── user/              # 用户服务 API
│   ├── house/             # 房源服务 API
│   ├── points/            # 积分服务 API
│   └── ...
├── internal/              # 内部代码
│   ├── biz/              # 业务逻辑层 (Use Cases)
│   ├── data/             # 数据访问层 (Repository 实现)
│   ├── domain/           # 领域模型层 (Entities & Interfaces)
│   ├── service/          # 服务层 (gRPC/HTTP 适配器)
│   └── server/           # 服务器配置
├── configs/              # 配置文件
├── migrations/           # 数据库迁移
└── docs/                 # 项目文档
```

### 分层架构说明

- **Service Layer**: 处理 gRPC/HTTP 请求，参数验证和响应格式化
- **Business Layer**: 核心业务逻辑，用例编排
- **Data Layer**: 数据访问实现，外部服务集成
- **Domain Layer**: 领域模型和接口定义，不依赖任何外部框架

## 🚀 快速开始

### 环境要求

- Go 1.24.4+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 6.2+

### 1. 克隆项目

```bash
git clone <repository-url>
cd anjuke
```

### 2. 环境配置

```bash
# 复制环境变量文件
cp .env.example .env

# 编辑配置文件
vim .env
```

### 3. 启动服务

#### 方式一：Docker Compose（推荐）

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f api
```

#### 方式二：本地开发

```bash
# 启动依赖服务
docker-compose up -d mysql redis minio

# 安装依赖
cd server
go mod tidy

# 生成代码
make generate

# 运行服务
make run
```

### 4. 验证服务

```bash
# 健康检查
curl http://localhost:8001/health

# 测试 API
curl http://localhost:8001/helloworld/world
```

## 🧪 测试

项目包含完整的单元测试套件，覆盖率达到 91.8%。

### 运行测试

```bash
# 运行所有测试
./run_tests.ps1  # Windows
./run_tests.sh   # Linux/Mac

# 运行特定模块测试
cd server
go test ./internal/biz/... -v
go test ./internal/service/... -v
go test ./internal/data/... -v

# 查看覆盖率
go test ./internal/... -cover
```

### 测试监控

```bash
# 使用测试监控脚本
./test_monitor.ps1
```

详细测试说明请参考 [TEST_README.md](TEST_README.md)

## 📖 API 文档

### 接口概览

- **用户服务**: `/user/*` - 用户注册、登录、实名认证
- **积分服务**: `/points/*` - 积分查询、签到、使用
- **文件上传**: `/user/uploadFile` - 文件上传服务
- **房源服务**: `/house/*` - 房源管理
- **交易服务**: `/transaction/*` - 交易处理

### 详细文档

- [API 快速参考](API-QUICK-REFERENCE.md)
- [API 详细文档](server/README.md)
- [OpenAPI 规范](server/openapi.yaml)
- [Postman 集合](server/docs/Postman测试集合.json)

## 🔧 开发指南

### 代码生成

```bash
cd server
make generate  # 生成 protobuf 代码
make wire      # 生成依赖注入代码
```

### 数据库迁移

```bash
# 运行迁移
go run migrate.go

# 或使用 Docker
docker-compose exec api go run migrate.go
```

### 添加新服务

1. 在 `api/` 目录定义 protobuf 接口
2. 在 `internal/domain/` 定义领域模型
3. 在 `internal/data/` 实现数据访问
4. 在 `internal/biz/` 实现业务逻辑
5. 在 `internal/service/` 实现服务接口
6. 更新 `wire.go` 依赖注入配置

## 📊 监控和运维

### 健康检查

```bash
# API 健康检查
curl http://localhost:8001/health

# 数据库连接检查
curl http://localhost:8001/health/db

# Redis 连接检查
curl http://localhost:8001/health/redis
```

### 日志查看

```bash
# 查看应用日志
docker-compose logs -f api

# 查看数据库日志
docker-compose logs -f mysql

# 查看 Redis 日志
docker-compose logs -f redis
```

### 性能监控

项目集成了基础的性能监控，可通过以下方式查看：

- 应用指标：`http://localhost:9003/metrics`
- 健康状态：`http://localhost:8001/health`

## 🚀 部署

### 生产环境部署

```bash
# 构建生产镜像
docker build -f Dockerfile.prod -t anjuke-api:latest .

# 使用生产配置启动
docker-compose -f docker-compose.prod.yml up -d
```

### 部署检查清单

详细部署步骤请参考 [DEPLOYMENT-CHECKLIST.md](DEPLOYMENT-CHECKLIST.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的单元测试
- 更新相关文档

## 📝 更新日志

### v1.0.0 (2025-01-28)

- ✨ 初始版本发布
- 🏗️ 完整的微服务架构
- 👤 用户管理系统
- 🎯 积分系统
- 📁 文件上传服务
- 🧪 完整的测试套件
- 🐳 Docker 容器化支持

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

- [Kratos](https://go-kratos.dev/) - 优秀的 Go 微服务框架
- [GORM](https://gorm.io/) - 强大的 Go ORM 库
- [MinIO](https://min.io/) - 高性能对象存储
- 所有贡献者和开源社区

## 📞 联系方式

- 项目维护者: [Your Name]
- 邮箱: [your.email@example.com]
- 项目地址: [https://github.com/your-username/anjuke]

---

⭐ 如果这个项目对你有帮助，请给它一个星标！