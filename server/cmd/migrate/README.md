# 数据库迁移工具

## 概述

这是一个独立的数据库迁移工具，用于手动执行数据库表结构的创建和更新。

## 为什么需要这个工具？

在生产环境中，每次应用启动时自动执行数据库迁移存在以下问题：

1. **性能问题**：每次启动都检查和迁移表结构会增加启动时间
2. **生产环境风险**：在生产环境中自动迁移可能不安全
3. **并发问题**：多个实例同时启动时可能产生冲突
4. **权限控制**：生产环境的应用账户通常不应该有DDL权限

## 使用方法

### 基本用法

```bash
cd server
go run cmd/migrate/main.go -host=localhost -port=3306 -username=root -password=your_password -database=anjuke
```

### 参数说明

- `-host`: 数据库主机地址（默认：localhost）
- `-port`: 数据库端口（默认：3306）
- `-username`: 数据库用户名（默认：root）
- `-password`: 数据库密码（如果不提供，会提示输入）
- `-database`: 数据库名称（默认：anjuke）
- `-charset`: 字符集（默认：utf8mb4）
- `-help`: 显示帮助信息

### 示例

```bash
# 显示帮助信息
go run cmd/migrate/main.go -help

# 连接本地数据库
go run cmd/migrate/main.go -password=123456

# 连接远程数据库
go run cmd/migrate/main.go -host=192.168.1.100 -port=3306 -username=anjuke_user -password=secure_password -database=anjuke_prod
```

## 配置应用程序

在应用程序的配置文件中，设置 `auto_migrate` 为 `false` 来禁用自动迁移：

```json
{
  "database": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "your_password",
    "database": "anjuke",
    "charset": "utf8mb4",
    "max_idle_conns": 10,
    "max_open_conns": 100,
    "max_lifetime": 3600,
    "auto_migrate": false
  }
}
```

## 迁移的表

该工具会迁移以下数据表：

### 预约相关表
- `appointments` - 预约表
- `appointment_logs` - 预约日志表
- `store_working_hours` - 门店工作时间表
- `realtor_working_hours` - 经纪人工作时间表
- `realtor_status` - 经纪人状态表
- `appointment_reviews` - 预约评价表

### 公司相关表
- `companies` - 公司表
- `stores` - 门店表
- `realtors` - 经纪人表

## 最佳实践

### 开发环境
- 可以启用 `auto_migrate: true` 以便快速开发
- 或者使用迁移工具进行手动迁移

### 测试环境
- 建议使用迁移工具进行手动迁移
- 确保测试环境与生产环境的迁移流程一致

### 生产环境
- **必须**禁用自动迁移（`auto_migrate: false`）
- 使用专门的数据库管理账户执行迁移
- 在部署新版本前先执行数据库迁移
- 建议在维护窗口期间执行迁移

## 安全注意事项

1. **权限分离**：应用程序账户和迁移账户应该分离
2. **备份**：执行迁移前务必备份数据库
3. **测试**：在生产环境执行前，先在测试环境验证迁移
4. **监控**：迁移过程中监控数据库性能和锁等待情况

## 故障排除

### 连接失败
- 检查数据库服务是否运行
- 验证主机地址和端口是否正确
- 确认用户名和密码是否正确
- 检查网络连接和防火墙设置

### 权限错误
- 确保数据库用户有CREATE、ALTER、DROP等DDL权限
- 检查用户是否有目标数据库的访问权限

### 迁移失败
- 查看具体的错误信息
- 检查表结构定义是否正确
- 确认数据库版本兼容性