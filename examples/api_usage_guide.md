# API 使用指南

## 发起在线聊天

### 接口说明

该接口用于发起在线聊天，连接用户和房东进行实时沟通。

### 请求方法

```
POST /house/chat/start
```

### 请求头

```
Content-Type: application/json
```

### 请求体

```json
{
  "reservation_id": 12345,  // 预约ID
  "user_id": 1001,          // 用户ID
  "landlord_id": 2001,      // 房东ID
  "initial_message": "您好，我想咨询一下这套房子的情况"  // 初始消息（可选）
}
```

### 响应

```json
{
  "code": 0,                // 状态码，0表示成功
  "msg": "聊天发起成功",      // 消息
  "data": {
    "chat_id": "chat_12345_1001_1627372800",  // 聊天ID
    "success": true         // 是否成功
  }
}
```

### 错误码

| 错误码 | 说明 |
|-------|------|
| 400   | 请求参数错误 |
| 401   | 未授权 |
| 404   | 资源不存在 |
| 500   | 服务器内部错误 |

### 常见问题

1. **404 错误**：确保URL路径正确，应为 `/house/chat/start`
2. **400 错误**：检查请求体格式是否正确，确保包含必要的字段
3. **Content-Type 错误**：确保设置了正确的 `Content-Type: application/json` 请求头

### 示例

#### cURL

```bash
curl -X POST "http://localhost:8000/house/chat/start" \
  -H "Content-Type: application/json" \
  -d '{
    "reservation_id": 12345,
    "user_id": 1001,
    "landlord_id": 2001,
    "initial_message": "您好，我想咨询一下这套房子的情况"
  }'
```

#### APIPost

1. 设置请求方法为 `POST`
2. 设置URL为 `http://localhost:8000/house/chat/start`
3. 在Headers标签页添加 `Content-Type: application/json`
4. 在Body标签页选择 `JSON` 格式，并输入以下内容：
```json
{
  "reservation_id": 12345,
  "user_id": 1001,
  "landlord_id": 2001,
  "initial_message": "您好，我想咨询一下这套房子的情况"
}
```
5. 点击发送按钮