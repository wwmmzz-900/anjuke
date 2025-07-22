# API 身份验证与错误处理指南

## 错误码 10206: "验证码不正确"

当你看到错误码 10206 和错误消息"验证码不正确"时，这表示API请求的身份验证环节出现了问题。本文档将帮助你理解和解决这个问题。

## 问题原因

这个错误通常由以下几种原因导致：

1. **验证码缺失或过期**：API请求中没有提供验证码，或者提供的验证码已经过期
2. **验证码格式错误**：提供的验证码格式不符合要求
3. **验证码与验证码ID不匹配**：提供的验证码与验证码ID不对应
4. **身份验证令牌(Token)无效**：提供的身份验证令牌已过期或无效
5. **请求头缺少必要的身份验证信息**：没有在请求头中包含必要的Authorization字段

## 解决方案

### 方案1: 获取并使用正确的验证码

1. 首先调用验证码获取接口：
   ```
   GET /api/captcha/get
   ```

2. 从响应中获取验证码ID和图片：
   ```json
   {
     "code": 0,
     "message": "success",
     "data": {
       "captcha_id": "abc123",
       "captcha_image": "base64编码的图片..."
     }
   }
   ```

3. 识别图片中的验证码

4. 验证验证码：
   ```
   POST /api/captcha/verify
   Content-Type: application/json
   
   {
     "captcha_id": "abc123",
     "captcha_code": "识别的验证码"
   }
   ```

5. 使用验证通过的验证码获取身份验证令牌：
   ```
   POST /api/user/login
   Content-Type: application/json
   
   {
     "user_id": "1001",
     "captcha_id": "abc123",
     "captcha_code": "识别的验证码"
   }
   ```

6. 从响应中获取身份验证令牌：
   ```json
   {
     "code": 0,
     "message": "success",
     "data": {
       "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
       "expires_in": 3600
     }
   }
   ```

7. 在后续请求中使用身份验证令牌：
   ```
   GET /api/house/recommend?user_id=1001
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

### 方案2: 使用测试脚本自动处理身份验证

我们提供了一个测试脚本 `test_recommendation.sh`，它可以自动处理验证码和身份验证流程：

```bash
# 使用基本参数
./test_recommendation.sh

# 指定参数
./test_recommendation.sh --host api.example.com --port 443 --user 1001
```

脚本会引导你完成以下步骤：
1. 获取验证码
2. 输入识别的验证码
3. 获取身份验证令牌
4. 使用令牌请求个性化推荐

## 身份验证流程图

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  获取验证码  │────>│  验证验证码  │────>│  获取令牌   │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
                                        ┌─────────────┐
                                        │  API请求    │
                                        │(带身份验证) │
                                        └─────────────┘
```

## 常见问题

### Q: 验证码总是提示不正确
A: 可能原因：
   - 验证码已过期（通常有效期为5分钟）
   - 验证码识别错误
   - 验证码ID与验证码不匹配
   
   解决方法：重新获取验证码并确保正确输入

### Q: 身份验证令牌无效
A: 可能原因：
   - 令牌已过期（通常有效期为1小时）
   - 令牌格式错误
   
   解决方法：重新获取身份验证令牌

### Q: API请求返回401或403错误
A: 可能原因：
   - 请求头中缺少Authorization字段
   - Authorization格式错误（应为"Bearer {token}"）
   - 令牌权限不足
   
   解决方法：检查请求头格式并确保使用正确的令牌

## 调试技巧

1. **检查请求头**：确保包含正确的Authorization头
   ```
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

2. **检查令牌格式**：JWT令牌通常由三部分组成，用点(.)分隔
   ```
   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
   ```

3. **使用测试脚本**：`test_recommendation.sh` 脚本可以帮助你自动处理身份验证流程

4. **查看令牌内容**：可以在 [jwt.io](https://jwt.io/) 上解码JWT令牌查看其内容（不要粘贴生产环境的令牌）

## 安全最佳实践

1. 不要在客户端代码中硬编码令牌
2. 令牌过期后及时刷新
3. 使用HTTPS保护API请求
4. 不要在URL参数中传递令牌，应使用Authorization请求头
5. 实现令牌撤销机制，以便在安全事件发生时撤销令牌

## 结论

正确处理身份验证是API调用成功的关键。通过遵循本文档中的步骤，你应该能够解决"验证码不正确"的错误，并成功调用个性化推荐接口。

如果问题仍然存在，请联系API提供方获取更多支持。