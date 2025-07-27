# API使用指南 - 三范式响应格式

## 📋 概述

本系统所有API接口均采用统一的三范式响应格式：`{code, msg, data}`

- **code**: 响应码，0表示成功，非0表示失败
- **msg**: 响应消息，描述操作结果
- **data**: 响应数据，包含具体的业务数据

## 🔧 统一响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "操作成功",
  "data": {
    // 具体的业务数据
  }
}
```

### 错误响应
```json
{
  "code": 1,
  "msg": "错误描述信息",
  "data": null
}
```

### 分页响应
```json
{
  "code": 0,
  "msg": "查询成功",
  "data": {
    "records": [...],
    "page_info": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

## 📚 接口详细说明

### 1. 用户服务 (User Service)

#### 1.1 创建用户
- **接口**: `POST /user/create`
- **描述**: 创建新用户账户

**请求参数**:
```json
{
  "Mobile": "13800138000",
  "NickName": "张三",
  "Password": "password123"
}
```

**成功响应**:
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

#### 1.2 实名认证
- **接口**: `POST /user/realname`
- **描述**: 用户实名认证

**请求参数**:
```json
{
  "UserId": 123456,
  "Name": "张三",
  "IdCard": "110101199001011234"
}
```

**成功响应**:
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

#### 1.3 发送短信验证码
- **接口**: `POST /user/sendSms`
- **描述**: 发送短信验证码

**请求参数**:
```json
{
  "phone": "13800138000",
  "device_id": "device123",
  "scene": "register"
}
```

**成功响应**:
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

#### 1.4 验证短信验证码
- **接口**: `POST /user/verifySms`
- **描述**: 验证短信验证码

**请求参数**:
```json
{
  "phone": "13800138000",
  "code": "123456",
  "scene": "register"
}
```

**成功响应**:
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

### 2. 积分服务 (Points Service)

#### 2.1 查询用户积分余额
- **接口**: `GET /points/balance/{user_id}`
- **描述**: 查询指定用户的积分余额

**成功响应**:
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

#### 2.2 查询积分明细记录
- **接口**: `GET /points/history/{user_id}?page=1&page_size=20&type=earn`
- **描述**: 查询用户积分明细记录，支持分页和类型筛选

**查询参数**:
- `page`: 页码，从1开始
- `page_size`: 每页数量，默认20
- `type`: 类型筛选，earn(获取) 或 use(消费)，空表示全部

**成功响应**:
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

#### 2.3 签到获取积分
- **接口**: `POST /points/checkin`
- **描述**: 用户每日签到获取积分

**请求参数**:
```json
{
  "user_id": 123456
}
```

**成功响应**:
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

#### 2.4 消费获取积分
- **接口**: `POST /points/earn/consume`
- **描述**: 用户消费获取积分

**请求参数**:
```json
{
  "user_id": 123456,
  "order_id": "order_789",
  "amount": 10000
}
```

**成功响应**:
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

#### 2.5 使用积分抵扣
- **接口**: `POST /points/use`
- **描述**: 使用积分进行抵扣

**请求参数**:
```json
{
  "user_id": 123456,
  "points": 100,
  "order_id": "order_890",
  "description": "商品抵扣"
}
```

**成功响应**:
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

### 3. 其他服务

#### 3.1 HelloWorld服务
- **接口**: `GET /helloworld/{name}`
- **描述**: 问候服务

**成功响应**:
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

#### 3.2 交易服务
- **接口**: `POST /transaction/create`
- **描述**: 创建交易记录

**请求参数**:
```json
{
  "user_id": 123456,
  "amount": 10000,
  "type": "payment",
  "description": "商品购买"
}
```

**成功响应**:
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

## 🚨 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误（参数错误、业务逻辑错误等） |
| 1001 | 用户不存在 |
| 1002 | 积分不足 |
| 1003 | 验证码错误 |
| 1004 | 验证码已过期 |

## 📝 调用示例

### JavaScript/TypeScript
```javascript
// 查询用户积分
async function getUserPoints(userId) {
  try {
    const response = await fetch(`/points/balance/${userId}`);
    const result = await response.json();
    
    if (result.code === 0) {
      console.log('积分查询成功:', result.data);
      return result.data;
    } else {
      console.error('积分查询失败:', result.msg);
      throw new Error(result.msg);
    }
  } catch (error) {
    console.error('请求失败:', error);
    throw error;
  }
}

// 用户签到
async function checkIn(userId) {
  try {
    const response = await fetch('/points/checkin', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ user_id: userId })
    });
    
    const result = await response.json();
    
    if (result.code === 0) {
      console.log('签到成功:', result.data);
      return result.data;
    } else {
      console.error('签到失败:', result.msg);
      throw new Error(result.msg);
    }
  } catch (error) {
    console.error('签到请求失败:', error);
    throw error;
  }
}
```

### Python
```python
import requests
import json

def get_user_points(user_id):
    """查询用户积分"""
    try:
        response = requests.get(f'/points/balance/{user_id}')
        result = response.json()
        
        if result['code'] == 0:
            print('积分查询成功:', result['data'])
            return result['data']
        else:
            print('积分查询失败:', result['msg'])
            raise Exception(result['msg'])
    except Exception as e:
        print('请求失败:', str(e))
        raise

def check_in(user_id):
    """用户签到"""
    try:
        response = requests.post('/points/checkin', 
                               json={'user_id': user_id})
        result = response.json()
        
        if result['code'] == 0:
            print('签到成功:', result['data'])
            return result['data']
        else:
            print('签到失败:', result['msg'])
            raise Exception(result['msg'])
    except Exception as e:
        print('签到请求失败:', str(e))
        raise
```

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type BaseResponse struct {
    Code int32       `json:"code"`
    Msg  string      `json:"msg"`
    Data interface{} `json:"data"`
}

func getUserPoints(userID int64) error {
    resp, err := http.Get(fmt.Sprintf("/points/balance/%d", userID))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    var result BaseResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }
    
    if result.Code == 0 {
        fmt.Println("积分查询成功:", result.Data)
        return nil
    } else {
        return fmt.Errorf("积分查询失败: %s", result.Msg)
    }
}

func checkIn(userID int64) error {
    reqBody := map[string]int64{"user_id": userID}
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := http.Post("/points/checkin", "application/json", 
                         bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    var result BaseResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }
    
    if result.Code == 0 {
        fmt.Println("签到成功:", result.Data)
        return nil
    } else {
        return fmt.Errorf("签到失败: %s", result.Msg)
    }
}
```

## 🔍 调试建议

1. **检查响应码**: 始终先检查 `code` 字段，0表示成功
2. **错误处理**: 非0响应码时，`msg` 字段包含错误描述
3. **数据解析**: 成功时，业务数据在 `data` 字段中
4. **分页处理**: 分页接口的 `data.page_info` 包含分页信息
5. **类型转换**: 注意数字类型的字段可能以字符串形式传输

## 📞 技术支持

如有问题，请联系开发团队或查看项目文档。

---
**文档版本**: v1.0  
**更新时间**: 2024年1月  
**适用版本**: 三范式响应格式改造后