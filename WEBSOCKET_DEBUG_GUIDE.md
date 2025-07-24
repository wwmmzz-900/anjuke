# WebSocket 接口调试指南

本文档提供了使用各种工具进行WebSocket接口调试的详细说明。

## 可用的调试工具

1. **命令行测试脚本** (`test_websocket.sh`)
2. **Go语言WebSocket客户端** (`websocket_client.go`)
3. **浏览器WebSocket测试页面** (`web/websocket_test.html`)

## 1. 命令行测试脚本

`test_websocket.sh`是一个交互式命令行工具，提供了多种WebSocket测试功能。

### 使用方法

```bash
# 基本用法
./test_websocket.sh

# 指定参数
./test_websocket.sh --host localhost --port 8080 --house 101 --user 1001 --landlord 2001
```

### 可用选项

- `-h, --host HOST`: 指定主机名 (默认: localhost)
- `-p, --port PORT`: 指定端口号 (默认: 8080)
- `--house ID`: 指定房源ID (默认: 101)
- `--user ID`: 指定用户ID (默认: 1001)
- `--landlord ID`: 指定房东ID (默认: 2001)
- `--help`: 显示帮助信息

### 功能菜单

脚本启动后会显示一个交互式菜单，提供以下功能：

1. **检查WebSocket连接状态**: 获取当前WebSocket连接的统计信息
2. **连接WebSocket**: 使用websocat工具连接到WebSocket服务器
3. **发送预约请求**: 发送HTTP请求来触发WebSocket消息
4. **发送聊天消息**: 发送聊天消息到指定用户
5. **运行Go WebSocket客户端**: 启动Go语言编写的WebSocket客户端

## 2. Go语言WebSocket客户端

`websocket_client.go`是一个功能丰富的命令行WebSocket客户端，支持发送各种类型的消息。

### 使用方法

```bash
# 基本用法
go run websocket_client.go <house_id> <user_id> [host] [port]

# 示例
go run websocket_client.go 101 1001 localhost 8080
```

### 客户端命令

连接后，可以使用以下命令：

- `/help`: 显示帮助信息
- `/quit`: 关闭连接并退出
- `/to <user_id> <message>`: 发送私聊消息
- `/json <json_string>`: 发送原始JSON消息
- 其他任何输入将作为普通消息发送

### 特点

- 彩色输出，区分不同类型的消息
- 自动格式化接收到的JSON消息
- 定期发送心跳消息保持连接
- 支持发送各种格式的消息

## 3. 浏览器WebSocket测试页面

`web/websocket_test.html`是一个功能全面的浏览器WebSocket测试工具，提供图形界面进行WebSocket调试。

### 访问方式

1. 启动服务器
2. 访问 `http://localhost:8000/ws/test` 或直接打开HTML文件

### 主要功能

- **连接管理**: 连接/断开WebSocket
- **消息发送**: 发送文本消息、预定义动作和自定义JSON
- **日志查看**: 查看接收到的消息和系统日志
- **API测试**: 测试相关的HTTP API
- **设置选项**: 配置心跳、自动重连等功能

### 标签页说明

1. **连接**: 配置连接参数并管理连接状态
2. **消息**: 发送简单消息和预定义动作
3. **JSON**: 编辑和发送自定义JSON消息
4. **设置**: 配置客户端行为和显示选项

## 调试技巧

1. **逐步测试**:
   - 先测试基本连接
   - 然后测试简单消息
   - 最后测试复杂交互

2. **监控双向通信**:
   - 使用浏览器测试页面查看完整的消息日志
   - 注意检查发送和接收的消息格式

3. **常见问题排查**:
   - 连接失败: 检查主机名、端口和路径
   - 消息未收到: 检查消息格式和接收者ID
   - 服务器未响应: 检查心跳设置和网络连接

4. **性能测试**:
   - 使用Go客户端可以进行简单的负载测试
   - 监控服务器日志查看消息处理情况

## 示例场景

### 场景1: 基本连接测试

```bash
# 使用命令行脚本
./test_websocket.sh
# 选择选项1检查连接状态
# 选择选项2连接WebSocket
```

### 场景2: 发送预约并观察WebSocket消息

1. 打开浏览器测试页面
2. 连接WebSocket
3. 点击"发送预约请求"按钮
4. 观察日志中接收到的消息

### 场景3: 测试聊天功能

```bash
# 使用Go客户端
go run websocket_client.go 101 1001
# 连接成功后发送消息
/to 2001 你好，这是一条测试消息
```

## 故障排除

如果遇到问题，请检查：

1. 服务器是否正在运行
2. 端口是否正确
3. 网络连接是否正常
4. 消息格式是否符合服务器要求
5. 服务器日志中是否有错误信息

## 结论

这些工具提供了全面的WebSocket接口调试能力，从简单的命令行测试到复杂的浏览器交互，满足不同场景的需求。根据具体情况选择合适的工具，可以大大提高WebSocket接口的调试效率。