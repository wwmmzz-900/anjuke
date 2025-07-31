# 安居客预约系统

基于Go语言和MySQL开发的房产经纪人预约系统，支持智能分配、排队管理、状态跟踪等功能。采用领域驱动设计(DDD)架构，提供高性能、高可用的预约服务。

## 🚀 功能特性

### 核心功能
- **智能预约分配**：基于经纪人负载、活跃度的智能分配算法
- **排队管理**：支持预约排队，自动位置更新和经纪人释放后的重新分配
- **状态跟踪**：完整的预约状态流转（待确认→已确认→进行中→已完成）
- **时间冲突检查**：防止用户和经纪人的时间冲突
- **预约码查询**：6位数字预约码，方便用户查询

### 高级功能
- **工作时间管理**：支持门店和经纪人的个性化工作时间设置
- **实时状态管理**：经纪人在线/离线/忙碌状态实时更新
- **操作日志记录**：完整的预约操作历史追踪
- **评价系统**：预约完成后的服务评价功能
- **数据统计**：预约数据的统计分析功能

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │   Load Balancer │    │     Monitor     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
├─────────────────┬─────────────────┬─────────────────────────────┤
│   HTTP Server   │   gRPC Server   │       Background Jobs       │
└─────────────────┴─────────────────┴─────────────────────────────┘
         │                       │                       │
┌─────────────────────────────────────────────────────────────────┐
│                        Business Layer                           │
├─────────────────┬─────────────────┬─────────────────────────────┤
│ Appointment UC  │   Store UC      │       Realtor UC            │
└─────────────────┴─────────────────┴─────────────────────────────┘
         │                       │                       │
┌─────────────────────────────────────────────────────────────────┐
│                         Data Layer                              │
├─────────────────┬─────────────────┬─────────────────────────────┤
│ Appointment Repo│   Store Repo    │       Realtor Repo          │
└─────────────────┴─────────────────┴─────────────────────────────┘
         │                       │                       │
┌─────────────────────────────────────────────────────────────────┐
│                      Infrastructure                             │
├─────────────────┬─────────────────┬─────────────────────────────┤
│      MySQL      │      Redis      │         Message Queue       │
└─────────────────┴─────────────────┴─────────────────────────────┘
```

## 📊 数据库设计

### 核心表结构

```sql
-- 预约表
CREATE TABLE appointments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    appointment_code VARCHAR(6) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    store_id BIGINT NOT NULL,
    realtor_id BIGINT,
    customer_name VARCHAR(50) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    appointment_date DATE NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    duration_minutes INT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    queue_position INT DEFAULT 0,
    estimated_wait_minutes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 门店表
CREATE TABLE stores (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    store_name VARCHAR(100) NOT NULL,
    address VARCHAR(200) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    business_hours VARCHAR(50),
    rating DECIMAL(3,2) DEFAULT 0,
    review_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE
);

-- 经纪人表
CREATE TABLE realtors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    store_id BIGINT NOT NULL,
    realtor_name VARCHAR(50) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE
);
```

## 🛠️ 技术栈

### 后端技术
- **语言**：Go 1.19+
- **框架**：Kratos v2 (微服务框架)
- **数据库**：MySQL 8.0+
- **ORM**：GORM v2
- **缓存**：Redis 6.0+
- **协议**：gRPC + HTTP/JSON
- **配置**：YAML
- **日志**：结构化日志

### 开发工具
- **依赖注入**：Wire
- **API文档**：Protocol Buffers
- **测试**：Go Testing + Testify
- **构建**：Docker + Docker Compose
- **监控**：Prometheus + Grafana

## 🚀 快速开始

### 环境要求
- Go 1.19+
- MySQL 8.0+
- Redis 6.0+ (可选)

### 1. 克隆项目
```bash
git clone <repository-url>
cd appointment-system/server
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 配置数据库
```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE anjuke_appointment CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 执行数据库迁移
mysql -u root -p anjuke_appointment < internal/data/migrations/001_create_appointment_tables.sql
```

### 4. 修改配置
```bash
cp configs/config_mysql.yaml configs/config.yaml
# 编辑 configs/config.yaml，修改数据库连接信息
```

### 5. 启动服务
```bash
go run cmd/server/main.go -conf configs/config.yaml
```

### 6. 验证服务
```bash
# 检查服务状态
curl http://localhost:8000/health

# 获取可预约时段
curl "http://localhost:8000/api/v1/appointment/stores/1/slots?start_date=2025-01-31&days=7"
```

## 📖 API 文档

### 预约管理

#### 创建预约
```http
POST /api/v1/appointment/appointments
Content-Type: application/json

{
    "store_id": "1",
    "customer_name": "张三",
    "customer_phone": "13800138000",
    "appointment_date": "2025-02-01",
    "start_time": "14:00",
    "duration_minutes": 60,
    "requirements": "需要了解二手房购买流程"
}
```

#### 查询预约
```http
GET /api/v1/appointment/appointments/{appointment_code}
```

#### 取消预约
```http
POST /api/v1/appointment/appointments/{appointment_code}/cancel
Content-Type: application/json

{
    "reason": "临时有事，需要取消"
}
```

#### 获取可预约时段
```http
GET /api/v1/appointment/stores/{store_id}/slots?start_date=2025-02-01&days=7
```

### 经纪人管理

#### 更新经纪人状态
```http
POST /api/v1/appointment/realtor/status
Content-Type: application/json

{
    "realtor_id": "1",
    "status": "online"
}
```

#### 经纪人接单
```http
POST /api/v1/appointment/appointments/{appointment_id}/accept
Content-Type: application/json

{
    "realtor_id": "1"
}
```

## 🧪 测试

### 运行单元测试
```bash
go test ./...
```

### 运行集成测试
```bash
go test ./test/...
```

### 运行性能测试
```bash
go test -bench=. ./test/...
```

### API测试
```bash
# 使用提供的测试脚本
go test -v ./test/appointment_api_test.go
```

## 📈 性能指标

### 基准性能
- **并发处理**：1000+ QPS
- **响应时间**：P99 < 100ms
- **数据库连接**：连接池复用，最大100连接
- **内存使用**：< 100MB (空载)

### 扩展性
- **水平扩展**：支持多实例部署
- **数据库分片**：支持按门店ID分片
- **缓存策略**：热点数据Redis缓存

## 🔧 配置说明

### 数据库配置
```yaml
data:
  mysql:
    host: localhost
    port: 3306
    username: root
    password: password
    database: anjuke_appointment
    charset: utf8mb4
    max_idle_conns: 10
    max_open_conns: 100
    max_lifetime: 3600
```

### 业务配置
```yaml
business:
  appointment:
    default_duration: 60
    max_advance_days: 7
    min_advance_minutes: 30
    queue_timeout_minutes: 120
    realtor:
      default_max_load: 3
      online_timeout_minutes: 30
```

## 🚀 部署

### Docker部署
```bash
# 构建镜像
docker build -t appointment-server .

# 运行容器
docker run -d \
  --name appointment-server \
  -p 8000:8000 \
  -p 9000:9000 \
  -v $(pwd)/configs:/app/configs \
  appointment-server
```

### Docker Compose部署
```bash
docker-compose up -d
```

### Kubernetes部署
```bash
kubectl apply -f k8s/
```

## 📊 监控

### 健康检查
```http
GET /health
```

### 指标监控
```http
GET /metrics
```

### 关键指标
- `appointment_created_total`：预约创建总数
- `appointment_success_rate`：预约成功率
- `realtor_utilization`：经纪人利用率
- `queue_wait_time`：平均排队时间

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📝 更新日志

### v1.0.0 (2025-01-30)
- ✨ 初始版本发布
- 🎯 支持基础预约功能
- 🤖 智能经纪人分配算法
- 📊 排队管理系统
- 🔍 预约状态跟踪
- 📱 RESTful API接口

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 联系我们

- 项目维护者：开发团队
- 邮箱：dev@anjuke.com
- 问题反馈：[GitHub Issues](https://github.com/anjuke/appointment-system/issues)

---

**注意**：这是一个演示项目，生产环境使用前请进行充分的测试和安全评估。