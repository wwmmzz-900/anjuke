# 📚 安居客系统接口详细说明

---

## 1. 用户服务（User Service）

### 1.1 创建用户
- **接口地址**：`POST /user/create`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段      | 类型   | 必填 | 说明     |
  |-----------|--------|------|----------|
  | Mobile    | string | 是   | 手机号   |
  | NickName  | string | 是   | 昵称     |
  | Password  | string | 是   | 密码     |
- **请求示例**：
  ```json
  {
    "Mobile": "13800138000",
    "NickName": "张三",
    "Password": "password123"
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "用户创建成功",
    "data": {
      "user_id": "user_123456",
      "mobile": "13800138000",
      "nick_name": "张三"
    }
  }
  ```
- **后端能力说明**：
  - 校验手机号唯一性、密码加密存储。
  - 支持统一响应格式，错误时返回详细原因。
- **典型调用流程**：
  1. 前端表单收集用户信息，POST 到 `/user/create`。
  2. 后端校验、入库，返回用户ID。
- **FAQ**：
  - Q: 手机号已注册会怎样？  
    A: 返回 `code!=0`，msg 提示手机号已存在。

### 1.2 实名认证
- **接口地址**：`POST /user/realname`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段   | 类型   | 必填 | 说明   |
  |--------|--------|------|--------|
  | UserId | int    | 是   | 用户ID |
  | Name   | string | 是   | 姓名   |
  | IdCard | string | 是   | 身份证 |
- **请求示例**：
  ```json
  {
    "UserId": 123456,
    "Name": "张三",
    "IdCard": "110101199001011234"
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "实名认证成功",
    "data": {
      "user_id": 123456,
      "name": "张三",
      "status": "verified"
    }
  }
  ```
- **后端能力说明**：
  - 支持第三方实名认证服务对接。
  - 实名状态写入用户表。
- **典型调用流程**：
  1. 前端收集实名信息，POST 到 `/user/realname`。
  2. 后端校验、调用实名服务，返回认证结果。
- **FAQ**：
  - Q: 身份证格式不对会怎样？  
    A: 返回 `code!=0`，msg 提示格式错误。

### 1.3 发送短信验证码
- **接口地址**：`POST /user/sendSms`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段      | 类型   | 必填 | 说明         |
  |-----------|--------|------|--------------|
  | phone     | string | 是   | 手机号       |
  | device_id | string | 否   | 设备ID       |
  | scene     | string | 是   | 场景（如register、login）|
- **请求示例**：
  ```json
  {
    "phone": "13800138000",
    "device_id": "device123",
    "scene": "register"
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "短信发送成功",
    "data": {
      "phone": "13800138000",
      "scene": "register",
      "expire_time": 1704067200
    }
  }
  ```
- **后端能力说明**：
  - 支持短信风控（频率、次数、IP、设备限制）。
  - 支持多场景模板。
- **典型调用流程**：
  1. 前端请求发送验证码，后端校验风控，调用短信服务。
- **FAQ**：
  - Q: 频繁请求会怎样？  
    A: 返回 `code!=0`，msg 提示操作频繁。

### 1.4 验证短信验证码
- **接口地址**：`POST /user/verifySms`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段   | 类型   | 必填 | 说明   |
  |--------|--------|------|--------|
  | phone  | string | 是   | 手机号 |
  | code   | string | 是   | 验证码 |
  | scene  | string | 是   | 场景   |
- **请求示例**：
  ```json
  {
    "phone": "13800138000",
    "code": "123456",
    "scene": "register"
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "验证成功",
    "data": {
      "phone": "13800138000",
      "success": true,
      "scene": "register"
    }
  }
  ```
- **后端能力说明**：
  - 校验验证码有效性、过期时间。
  - 验证成功后验证码失效。
- **典型调用流程**：
  1. 前端提交验证码，后端校验，返回结果。
- **FAQ**：
  - Q: 验证码错误/过期会怎样？  
    A: 返回 `code!=0`，msg 提示错误或过期。

---

## 2. 文件上传服务（File Upload Service）

（详见前述“文件上传接口与后端能力说明”）

---

## 3. 积分服务（Points Service）

### 3.1 查询用户积分余额
- **接口地址**：`GET /points/balance/{user_id}`
- **请求类型**：`GET`
- **参数说明**：
  - `user_id`（路径参数）：用户ID
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "查询成功",
    "data": {
      "user_id": 123456,
      "total_points": 1500
    }
  }
  ```
- **后端能力说明**：
  - 实时查询积分表，返回当前积分。
- **典型调用流程**：
  1. 前端 GET 请求，后端查表返回。
- **FAQ**：
  - Q: 用户不存在会怎样？  
    A: 返回 `code!=0`，msg 提示用户不存在。

### 3.2 查询积分明细记录
- **接口地址**：`GET /points/history/{user_id}`
- **请求类型**：`GET`
- **参数说明**：
  - `user_id`（路径参数）：用户ID
  - `page`（查询参数）：页码
  - `page_size`（查询参数）：每页数量
  - `type`（查询参数）：类型筛选 earn/use
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "查询成功",
    "data": {
      "records": [
        {
          "id": 1,
          "user_id": 123456,
          "type": "checkin",
          "points": 10,
          "description": "每日签到",
          "order_id": "",
          "amount": 0,
          "created_at": "2024-01-01 12:00:00"
        }
      ],
      "page_info": {
        "page": 1,
        "page_size": 20,
        "total": 50,
        "total_pages": 3
      }
    }
  }
  ```
- **后端能力说明**：
  - 支持分页、类型筛选。
- **典型调用流程**：
  1. 前端 GET 请求，带分页参数，后端查表返回。
- **FAQ**：
  - Q: 没有记录会怎样？  
    A: 返回空数组，分页信息正常。

### 3.3 签到获取积分
- **接口地址**：`POST /points/checkin`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段    | 类型   | 必填 | 说明   |
  |---------|--------|------|--------|
  | user_id | int    | 是   | 用户ID |
- **请求示例**：
  ```json
  {
    "user_id": 123456
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "签到成功",
    "data": {
      "points_earned": 10,
      "total_points": 1510,
      "consecutive_days": 5
    }
  }
  ```
- **后端能力说明**：
  - 支持每日签到、连续签到奖励。
- **典型调用流程**：
  1. 前端 POST 请求，后端校验并发放积分。
- **FAQ**：
  - Q: 当天已签到会怎样？  
    A: 返回 `code!=0`，msg 提示已签到。

### 3.4 消费获取积分
- **接口地址**：`POST /points/earn/consume`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段     | 类型   | 必填 | 说明   |
  |----------|--------|------|--------|
  | user_id  | int    | 是   | 用户ID |
  | order_id | string | 是   | 订单ID |
  | amount   | int    | 是   | 金额   |
- **请求示例**：
  ```json
  {
    "user_id": 123456,
    "order_id": "order_789",
    "amount": 10000
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "消费获得积分成功，获得100积分",
    "data": {
      "points_earned": 100,
      "total_points": 1610
    }
  }
  ```
- **后端能力说明**：
  - 按消费金额自动计算积分。
- **典型调用流程**：
  1. 前端 POST 请求，后端校验并发放积分。
- **FAQ**：
  - Q: 金额为0会怎样？  
    A: 返回 `code!=0`，msg 提示金额无效。

### 3.5 使用积分抵扣
- **接口地址**：`POST /points/use`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段        | 类型   | 必填 | 说明   |
  |-------------|--------|------|--------|
  | user_id     | int    | 是   | 用户ID |
  | points      | int    | 是   | 使用积分|
  | order_id    | string | 是   | 订单ID |
  | description | string | 否   | 说明   |
- **请求示例**：
  ```json
  {
    "user_id": 123456,
    "points": 100,
    "order_id": "order_890",
    "description": "商品抵扣"
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "积分使用成功，抵扣1.00元",
    "data": {
      "points_used": 100,
      "amount_deducted": 100,
      "total_points": 1510
    }
  }
  ```
- **后端能力说明**：
  - 校验积分余额，自动扣减。
- **典型调用流程**：
  1. 前端 POST 请求，后端校验并扣减积分。
- **FAQ**：
  - Q: 积分不足会怎样？  
    A: 返回 `code!=0`，msg 提示积分不足。

---

## 4. 房源服务（House Service）

### 4.1 创建房源
- **接口地址**：`POST /house/create`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段      | 类型   | 必填 | 说明   |
  |-----------|--------|------|--------|
  | ...       | ...    | ...  | ...    |
- **请求示例**：
  ```json
  {
    // 具体字段视业务而定
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "房源创建成功",
    "data": {
      "house_id": "house_123456"
    }
  }
  ```
- **后端能力说明**：
  - 校验房源信息，入库。
- **典型调用流程**：
  1. 前端 POST 请求，后端校验并入库。
- **FAQ**：
  - Q: 信息不全会怎样？  
    A: 返回 `code!=0`，msg 提示缺少字段。

---

## 5. 交易服务（Transaction Service）

### 5.1 创建交易
- **接口地址**：`POST /transaction/create`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段        | 类型   | 必填 | 说明   |
  |-------------|--------|------|--------|
  | user_id     | int    | 是   | 用户ID |
  | amount      | int    | 是   | 金额   |
  | type        | string | 是   | 类型   |
  | description | string | 否   | 说明   |
- **请求示例**：
  ```json
  {
    "user_id": 123456,
    "amount": 10000,
    "type": "payment",
    "description": "商品购买"
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "交易创建成功",
    "data": {
      "transaction_id": "txn_123456",
      "user_id": 123456,
      "amount": 10000,
      "type": "payment",
      "status": "created",
      "created_at": "2024-01-01 12:00:00"
    }
  }
  ```
- **后端能力说明**：
  - 校验用户、金额，入库。
- **典型调用流程**：
  1. 前端 POST 请求，后端校验并入库。
- **FAQ**：
  - Q: 金额为0会怎样？  
    A: 返回 `code!=0`，msg 提示金额无效。

---

## 6. 客服服务（Customer Service）

### 6.1 创建客户
- **接口地址**：`POST /customer/create`
- **请求类型**：`application/json`
- **参数说明**：
  | 字段      | 类型   | 必填 | 说明   |
  |-----------|--------|------|--------|
  | ...       | ...    | ...  | ...    |
- **请求示例**：
  ```json
  {
    // 具体字段视业务而定
  }
  ```
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "客户创建成功",
    "data": {
      "customer_id": "cust_123456"
    }
  }
  ```
- **后端能力说明**：
  - 校验客户信息，入库。
- **典型调用流程**：
  1. 前端 POST 请求，后端校验并入库。
- **FAQ**：
  - Q: 信息不全会怎样？  
    A: 返回 `code!=0`，msg 提示缺少字段。

---

## 7. HelloWorld 测试服务

### 7.1 问候
- **接口地址**：`GET /helloworld/{name}`
- **请求类型**：`GET`
- **参数说明**：
  - `name`（路径参数）：姓名
- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "问候成功",
    "data": {
      "message": "Hello 张三",
      "name": "张三"
    }
  }
  ```
- **后端能力说明**：
  - 简单测试接口。
- **典型调用流程**：
  1. 前端 GET 请求，后端返回问候语。
- **FAQ**：
  - Q: 传空会怎样？  
    A: 返回默认问候。

---


# 📦 文件上传接口与后端能力说明

## 1. 对外 HTTP 接口

### 1.1 单文件上传（推荐）
- **接口地址**：`POST /user/uploadFile`
- **请求类型**：`multipart/form-data`
- **参数**：`file`（form-data 文件字段）
- **说明**：自动适配小文件和大文件，底层采用智能上传（SmartUploadWithProgress），支持分片上传和进度回调（但当前 HTTP 接口未直接暴露进度）。
- **响应**：
  ```json
  {
    "code": 0,
    "msg": "上传成功",
    "data": {
      "url": "http://minio-endpoint/mybucket/1700000000.png"
    }
  }
  ```

### 1.2 多文件上传
- **接口地址**：`POST /user/uploadFiles`
- **请求类型**：`multipart/form-data`
- **参数**：`files`（form-data 多文件字段）
- **说明**：每个文件都走智能上传（SmartUploadWithProgress），支持大文件分片，返回每个文件的 url、文件名、大小、类型。
- **响应**：
  ```json
  {
    "code": 0,
    "msg": "上传成功",
    "data": [
      {
        "url": "...",
        "filename": "image1.png",
        "size": 123456,
        "content_type": "image/png"
      }
    ]
  }
  ```

---

## 2. 后端能力说明

### 2.1 智能上传（SmartUpload/SmartUploadWithProgress）

- **SmartUpload**：自动判断文件大小，选择普通上传或分片上传，适合所有文件类型。
- **SmartUploadWithProgress**：在智能上传基础上，增加进度回调参数，适合需要实时进度反馈的场景（如 CLI、WebSocket、前端轮询等）。

#### 代码示例

```go
// 智能上传（无进度）
url, err := minioClient.SmartUpload(ctx, filename, file, size, contentType)

// 智能上传（带进度回调）
url, err := minioClient.SmartUploadWithProgress(ctx, filename, file, size, contentType, func(uploaded, total int64) {
    fmt.Printf("上传进度: %d/%d\n", uploaded, total)
})
```

- **说明**：HTTP 层目前统一调用 `SmartUploadWithProgress`，如果不需要进度回调，传 `nil` 即可。

---

### 2.2 分片上传与断点续传（高级场景）

- **StartMultipartUpload**：手动初始化分片上传，返回分片信息。
- **ResumeMultipartUpload**：断点续传，恢复未完成的分片上传。
- **用途**：适合大文件、断点续传、前端直传等高级需求。

#### 代码示例

```go
// 初始化分片上传
info, err := minioClient.StartMultipartUpload(ctx, filename, contentType, totalSize)
// ...前端/客户端分片上传...
// 断点续传
url, err := minioClient.ResumeMultipartUpload(ctx, info, fileReader)
```

---

## 3. 进度回调与智能上传的关系

- **所有 HTTP 上传接口底层都用带进度回调的智能上传（SmartUploadWithProgress）**，只是当前未将进度直接暴露给前端。
- 后端如需扩展前端进度条，可通过 WebSocket、轮询等方式将进度回调结果推送给前端。

---

## 4. 典型调用流程

1. 前端调用 `/user/uploadFile` 或 `/user/uploadFiles`，上传文件。
2. 后端 handler 调用 `user.UploadToMinioWithProgress`，底层统一用 `SmartUploadWithProgress`，自动适配分片上传。
3. 进度回调参数当前为 `nil`，如需进度可扩展。

---

## 5. FAQ

- **Q: 智能上传和带进度回调的智能上传有何区别？**
  - A: 功能完全一致，后者多了进度通知能力。你可以统一用带进度回调的接口，传 `nil` 也不会有副作用。

- **Q: HTTP 接口能否直接获取上传进度？**
  - A: 目前未直接暴露。如需支持，可通过 WebSocket 或轮询方式扩展。

- **Q: 分片上传/断点续传如何用？**
  - A: 需用后端高级接口（StartMultipartUpload/ResumeMultipartUpload），适合大文件和断点续传场景。

---

### 2.5 管理员清理未完成分片
- **接口地址**：`POST /admin/cleanupIncompleteUploads`
- **描述**：清理 MinIO 未完成分片（管理员/内部使用）
- **请求参数**（Query）：
  - `prefix`（可选）：只清理指定前缀的分片
  - `older_than`（可选）：清理多久以前的分片（如 `24h`、`48h`），默认 24 小时
- **请求示例**：
  ```bash
  curl -X POST "http://localhost:8000/admin/cleanupIncompleteUploads?prefix=&older_than=24h"
  ```
- **成功响应**：
  ```json
  {
    "code": 0,
    "msg": "清理成功"
  }
  ```
- **失败响应**：
  ```json
  {
    "code": 1,
    "msg": "清理失败: 错误信息"
  }
  ```
- **后端能力说明**：
  - 通过 `minioClient.CleanupIncompleteUploads` 实现，支持定时任务和手动触发。
  - 仅建议管理员或内部系统调用，防止误删。
- **典型调用流程**：
  1. 管理员或定时任务调用该接口。
  2. 后端清理所有超过指定时间未完成的分片上传。
- **FAQ**：
  - Q: 该接口是否有权限控制？  
    A: 建议加权限校验，仅限内部或管理员使用。
  - Q: 会影响正常上传吗？  
    A: 只会清理长时间未完成的分片，不影响正常上传。