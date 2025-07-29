# 权限管理功能部署指南

## 1. 数据库初始化

首先执行 SQL 脚本创建权限相关表：

```bash
mysql -u root -p anjuke < sql/user_permission.sql
```

## 2. 生成 Protobuf 代码

```bash
# 给脚本执行权限
chmod +x generate_permission_proto.sh

# 生成 protobuf 代码
./generate_permission_proto.sh
```

## 3. 重新生成依赖注入代码

```bash
cd cmd/anjuke
wire
```

## 4. 编译和运行服务

```bash
# 编译
go build -o ./bin/anjuke ./cmd/anjuke

# 运行服务
./bin/anjuke -conf ./configs
```

## 5. 测试权限功能

```bash
# 给测试脚本执行权限
chmod +x test_permission.sh

# 运行测试
./test_permission.sh
```

## API 接口说明

### 1. 更新用户权限
```
PUT /permission/user/{user_id}
```

请求体：
```json
{
  "user_id": 1001,
  "permissions": ["READ", "WRITE", "PUBLISH_HOUSE"],
  "role": "ROLE_LANDLORD",
  "reason": "升级为房东用户",
  "operator_id": 1
}
```

### 2. 获取用户权限
```
GET /permission/user/{user_id}
```

### 3. 批量更新用户权限
```
PUT /permission/batch
```

请求体：
```json
{
  "updates": [
    {
      "user_id": 1002,
      "permissions": ["READ", "WRITE"],
      "role": "ROLE_NORMAL_USER",
      "reason": "普通用户权限",
      "operator_id": 1
    }
  ]
}
```

### 4. 获取权限列表
```
GET /permission/list?page=1&page_size=10&role_filter=ROLE_AGENT
```

### 5. 获取角色权限
```
GET /permission/role/{role_id}
```

## 权限类型说明

- `READ`: 读取权限
- `WRITE`: 写入权限
- `DELETE`: 删除权限
- `ADMIN`: 管理员权限
- `PUBLISH_HOUSE`: 发布房源权限
- `MANAGE_USER`: 用户管理权限
- `MANAGE_TRANSACTION`: 交易管理权限
- `CUSTOMER_SERVICE`: 客服权限

## 用户角色说明

- `ROLE_GUEST`: 游客用户，只能浏览基本信息
- `ROLE_NORMAL_USER`: 普通用户，可以浏览和基本操作
- `ROLE_VIP_USER`: VIP用户，享有更多特权
- `ROLE_LANDLORD`: 房东用户，可以发布和管理房源
- `ROLE_AGENT`: 经纪人，可以管理房源和用户
- `ROLE_ADMIN`: 管理员，拥有大部分管理权限
- `ROLE_SUPER_ADMIN`: 超级管理员，拥有所有权限

## 注意事项

1. 确保数据库中存在用户数据，否则权限更新会失败
2. 权限验证会根据角色进行，例如游客用户只能拥有读取权限
3. 只有超级管理员才能拥有管理员权限
4. 权限变更会记录操作者信息，便于审计

## 故障排除

1. 如果 protobuf 生成失败，请确保安装了 kratos 工具链
2. 如果依赖注入失败，请检查 wire.go 文件是否正确配置
3. 如果数据库连接失败，请检查 configs/config.yaml 中的数据库配置