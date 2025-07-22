# 预约看房 WebSocket 实时消息推送系统

基于 Go + gRPC + WebSocket 实现的预约看房实时消息推送系统，支持房东和租客之间的实时消息通知。

## 功能特性

### 核心功能
- **预约看房**: 租客可以预约看房，系统自动通知房东
- **预约确认**: 房东可以确认或拒绝预约请求
- **预约取消**: 租客和房东都可以取消预约
- **实时通知**: 基于 WebSocket 的实时消息推送
- **预约提醒**: 定时任务发送预约提醒消息

### 技术特性
- **gRPC 服务**: 高性能的 RPC 通信
- **WebSocket 连接管理**: 支持多用户并发连接
- **消息类型**: 支持多种消息类型（创建、确认、取消、提醒）
- **数据持久化**: MySQL 数据库存储
- **定时任务**: 自动发送预约提醒
- **连接管理**: 自动清理无效连接和心跳检测

## 项目结构

```
├── api/house/v3/           # Proto 文件和生成的代码
│   ├── house.proto         # gRPC 服务定义
│   ├── house.pb.go         # 生成的 protobuf 代码
│   ├── house_grpc.pb.go    # 生成的 gRPC 代码
│   └── house_http.pb.go    # 生成的 HTTP 代码
├── cmd/server/             # 服务器启动入口
│   └── main.go             # 主程序
├── internal/
│   ├── data/               # 数据访问层
│   │   └── reservation.go  # 预约数据模型和仓库
│   ├── service/            # 业务逻辑层
│   │   ├── websocket_manager.go    # WebSocket 连接管理
│   │   ├── reservation_service.go  # 预约业务服务
│   │   ├── house_service.go        # 房源服务
│   │   └── scheduler_service.go    # 定时任务服务
│   └── server/             # HTTP 服务器
│       └── websocket_handler.go    # WebSocket 处理器
├── scripts/                # 数据库脚本
│   └── init.sql            # 数据库初始化脚本
├── web/                    # 前端测试页面
│   └── index.html          # WebSocket 测试页面
├── go.mod                  # Go 模块依赖
└── README.md               # 项目文档
```

## 快速开始

### 1. 环境准备

确保已安装以下软件：
- Go 1.24+
- MySQL 8.0+
- Protocol Buffers 编译器

### 2. 数据库初始化

```bash
# 连接到 MySQL
mysql -u root -p

# 执行初始化脚本
source scripts/init.sql
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 配置数据库连接

修改 `cmd/server/main.go` 中的数据库连接字符串：

```go
dsn := "root:password@tcp(localhost:3306)/anjuke?charset=utf8mb4&parseTime=True&loc=Local"
```

### 5. 启动服务

```bash
go run cmd/server/main.go
```

服务启动后：
- gRPC 服务运行在 `:9000` 端口
- HTTP/WebSocket 服务运行在 `:8080` 端口

### 6. 测试功能

打开浏览器访问 `http://localhost:8080/web/index.html` 进行功能测试。

## API 接口

### gRPC 接口

```protobuf
service House {
  // 普通推荐列表
  rpc RecommendList (HouseRecommendRequest) returns (HouseRecommendReply);
  
  // 个性化推荐列表
  rpc PersonalRecommendList (PersonalRecommendRequest) returns (HouseRecommendReply);
  
  // 预约看房
  rpc ReserveHouse (ReserveHouseRequest) returns (ReserveHouseReply);
}
```

### HTTP 接口

```
GET  /house/recommend                    # 获取推荐房源
GET  /house/personal-recommend           # 获取个性化推荐
POST /house/reserve                      # 创建预约看房

POST /api/reservations/{id}/confirm      # 确认预约
POST /api/reservations/{id}/cancel       # 取消预约
POST /api/reservations/{id}/remind       # 发送提醒

GET  /api/websocket/online-users         # 获取在线用户
GET  /api/websocket/stats                # 获取 WebSocket 统计
```

### WebSocket 接口

```
ws://localhost:8080/ws?user_id={用户ID}
```

## WebSocket 消息格式

### 消息类型

```go
enum MessageType {
  UNKNOWN = 0;
  RESERVATION_CREATED = 1;    // 预约创建
  RESERVATION_CONFIRMED = 2;  // 预约确认
  RESERVATION_CANCELLED = 3;  // 预约取消
  RESERVATION_REMINDER = 4;   // 预约提醒
}
```

### 消息结构

```json
{
  "type": 1,
  "user_id": 1001,
  "reservation_id": 12345,
  "title": "新的看房预约",
  "content": "张三 预约了您的房源《精装修两室一厅》",
  "timestamp": 1642838400,
  "detail": {
    "reservation_id": 12345,
    "house_id": 3001,
    "house_title": "精装修两室一厅",
    "landlord_id": 2001,
    "user_id": 1001,
    "user_name": "张三",
    "reserve_time": "2025-01-22 14:00:00",
    "status": "pending",
    "created_at": 1642838400
  }
}
```

## 使用示例

### 1. 建立 WebSocket 连接

```javascript
const websocket = new WebSocket('ws://localhost:8080/ws?user_id=1001');

websocket.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('收到消息:', message);
};
```

### 2. 创建预约

```bash
curl -X POST http://localhost:8080/house/reserve \
  -H "Content-Type: application/json" \
  -d '{
    "landlord_id": 2001,
    "user_id": 1001,
    "user_name": "张三",
    "house_id": 3001,
    "house_title": "精装修两室一厅",
    "reserve_time": "2025-01-22 14:00:00"
  }'
```

### 3. 确认预约

```bash
curl -X POST http://localhost:8080/api/reservations/1/confirm \
  -H "Content-Type: application/json" \
  -d '{"landlord_id": 2001}'
```

## 核心组件说明

### WebSocket 管理器 (websocket_manager.go)
- **连接管理**: 维护用户ID到WebSocket连接的映射
- **消息推送**: 支持单用户和多用户消息推送
- **心跳检测**: 定期发送ping消息检测连接状态
- **连接清理**: 自动清理无效连接

### 预约服务 (reservation_service.go)
- **预约创建**: 创建预约并发送通知消息
- **状态管理**: 处理预约确认、取消等状态变更
- **消息通知**: 根据不同操作发送相应的WebSocket消息
- **数据验证**: 验证预约请求参数的有效性

### 定时任务服务 (scheduler_service.go)
- **预约提醒**: 定时检查并发送预约提醒消息
- **连接清理**: 定期清理无效的WebSocket连接
- **任务调度**: 支持多种定时任务的统一管理

## 扩展功能

### 1. 消息持久化
可以扩展消息存储功能，将WebSocket消息保存到数据库中，支持离线消息推送。

### 2. 用户认证
集成JWT或其他认证机制，确保WebSocket连接的安全性。

### 3. 消息推送
集成第三方推送服务（如极光推送、友盟推送），支持APP推送通知。

### 4. 集群部署
使用Redis等中间件支持多实例部署和消息广播。

## 注意事项

1. **数据库连接**: 请根据实际环境修改数据库连接配置
2. **端口配置**: 确保8080和9000端口未被占用
3. **跨域设置**: 生产环境请配置合适的CORS策略
4. **错误处理**: 建议添加更完善的错误处理和日志记录
5. **性能优化**: 大量连接时考虑连接池和消息队列优化

## 许可证

MIT License