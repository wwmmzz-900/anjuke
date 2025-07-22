# WebSocket 测试指南

## 1. 启动服务器

确保您的服务器已经启动并运行在 8080 端口。

## 2. 测试 WebSocket 连接

### 方法一：使用 websocat 工具

```bash
# 安装 websocat（如果尚未安装）
# Ubuntu/Debian: sudo apt install websocat
# macOS: brew install websocat
# Windows: 下载二进制文件

# 连接 WebSocket（注意使用引号包围 URL）
websocat "ws://localhost:8080/ws/house?house_id=101&user_id=1001"
```

### 方法二：使用提供的 Go 客户端

```bash
# 编译并运行 WebSocket 客户端
go run websocket_client.go 101 1001
```

### 方法三：使用 ApiPost

1. 创建新的 WebSocket 请求
2. 设置 URL: `ws://localhost:8080/ws/house?house_id=101&user_id=1001`
3. 点击连接

## 3. 验证连接状态

### 检查连接统计信息

```bash
curl "http://localhost:8080/api/websocket/stats"
```

预期响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "house_101": 1,
    "total_connections": 1,
    "total_houses": 1
  }
}
```

## 4. 测试消息推送

### 发送预约请求

在保持 WebSocket 连接的同时，发送预约请求：

```bash
curl -X POST "http://localhost:8080/house/reserve" \
  -H "Content-Type: application/json" \
  -d '{
    "landlord_id": 2001,
    "user_id": 1001,
    "user_name": "张三",
    "house_id": 101,
    "house_title": "精装修两室一厅",
    "reserve_time": "2025-07-25 14:00:00"
  }'
```

### 预期的 WebSocket 消息

连接成功后，您应该收到：

1. **连接确认消息**：
```json
{
  "type": "connection",
  "message": "WebSocket 连接成功",
  "house_id": 101,
  "user_id": 1001
}
```

2. **预约创建消息**（发送预约请求后）：
```json
{
  "type": "reservation_created",
  "title": "预约成功",
  "message": "您已成功预约房源《精装修两室一厅》，请等待房东确认",
  "house_id": 101,
  "landlord_id": 2001,
  "reserve_time": "2025-07-25 14:00:00",
  "timestamp": 1627123456789
}
```

## 5. 多用户测试

### 测试房东和租客同时在线

1. 打开两个终端或 ApiPost 窗口
2. 第一个连接作为租客：
   ```
   ws://localhost:8080/ws/house?house_id=101&user_id=1001
   ```
3. 第二个连接作为房东：
   ```
   ws://localhost:8080/ws/house?house_id=101&user_id=2001
   ```
4. 发送预约请求，观察两个连接是否都收到消息

## 6. 故障排除

### 连接失败的常见原因

1. **服务器未启动**：确保服务器正在运行
2. **端口错误**：检查服务器是否在 8080 端口监听
3. **URL 格式错误**：确保 URL 格式正确，参数用 & 分隔
4. **参数缺失**：确保提供了 house_id 和 user_id 参数

### 检查服务器日志

连接成功时，服务器应该输出：
```
WebSocket connected successfully: houseID=101, userID=1001, remoteAddr=127.0.0.1:xxxxx
```

断开连接时，服务器应该输出：
```
WebSocket disconnected: houseID=101, userID=1001
```

### 使用测试脚本

运行提供的测试脚本：
```bash
chmod +x test_websocket.sh
./test_websocket.sh
```

## 7. 性能测试

### 测试多个并发连接

```bash
# 使用 Go 客户端测试多个连接
for i in {1..10}; do
  go run websocket_client.go 101 $((1000 + i)) &
done
```

然后检查连接统计：
```bash
curl "http://localhost:8080/api/websocket/stats"
```

这应该显示 10 个活跃连接。