# 预约系统MySQL迁移指南

## 概述

本文档介绍如何将预约系统从MongoDB迁移到MySQL，以及MySQL版本的优势和使用方法。

## 为什么选择MySQL

### 1. 事务支持更强
- **ACID特性**：MySQL提供完整的ACID事务支持，确保数据一致性
- **并发控制**：更好的锁机制，适合高并发预约场景
- **死锁检测**：自动检测和处理死锁情况

### 2. 查询能力更强
- **复杂JOIN**：支持复杂的多表关联查询
- **聚合统计**：强大的GROUP BY和聚合函数支持
- **索引优化**：丰富的索引类型和优化策略

### 3. 生态更成熟
- **监控工具**：丰富的监控和性能分析工具
- **备份恢复**：成熟的备份和恢复方案
- **运维经验**：团队对MySQL运维更熟悉

## 数据库设计

### 核心表结构

1. **stores** - 门店表
   - 存储门店基本信息
   - 支持门店状态管理

2. **realtors** - 经纪人表
   - 存储经纪人基本信息
   - 关联门店信息

3. **appointments** - 预约表
   - 核心预约数据
   - 支持状态流转
   - 包含排队信息

4. **appointment_logs** - 预约日志表
   - 记录所有操作历史
   - 便于问题追踪

5. **realtor_status** - 经纪人状态表
   - 实时状态管理
   - 负载均衡支持

### 关键设计特点

- **外键约束**：确保数据完整性
- **索引优化**：针对查询场景优化索引
- **分区支持**：支持按时间分区（可选）

## 部署步骤

### 1. 环境准备

```bash
# 安装MySQL 8.0+
sudo apt-get install mysql-server-8.0

# 创建数据库
mysql -u root -p
CREATE DATABASE anjuke_appointment CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 配置文件

复制 `configs/config_mysql.yaml` 并修改数据库连接信息：

```yaml
data:
  mysql:
    host: localhost
    port: 3306
    username: your_username
    password: your_password
    database: anjuke_appointment
    charset: utf8mb4
    max_idle_conns: 10
    max_open_conns: 100
    max_lifetime: 3600
```

### 3. 数据库初始化

```bash
# 执行数据库迁移脚本
mysql -u your_username -p anjuke_appointment < migrations/005_create_appointment_tables.sql
```

### 4. 启动服务

```bash
# 编译并启动服务
go build -o appointment-server ./cmd/server
./appointment-server -conf configs/config_mysql.yaml
```

## 主要功能特性

### 1. 智能经纪人分配

```go
// 综合评分算法
func (uc *AppointmentMySQLUsecase) calculateRealtorScore(realtor *domain.RealtorStatusInfo) float64 {
    // 负载评分（40%权重）
    loadScore := float64(realtor.MaxLoad-realtor.CurrentLoad) / float64(realtor.MaxLoad) * 100
    
    // 活跃评分（30%权重）
    timeSinceActive := time.Since(realtor.LastActiveAt).Minutes()
    activeScore := 100.0
    if timeSinceActive > 60 {
        activeScore = 100 - timeSinceActive/60*10
    }
    
    // 经验评分（30%权重）
    experienceScore := 80.0
    
    return loadScore*0.4 + activeScore*0.3 + experienceScore*0.3
}
```

### 2. 事务保证

```go
// 使用数据库事务确保数据一致性
func (r *appointmentMySQLRepo) CreateAppointment(ctx context.Context, appointment *domain.AppointmentInfo) (*domain.AppointmentInfo, error) {
    tx := r.data.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // 创建预约记录
    if err := tx.Create(appointmentModel).Error; err != nil {
        tx.Rollback()
        return nil, err
    }
    
    // 创建操作日志
    if err := tx.Create(logModel).Error; err != nil {
        tx.Rollback()
        return nil, err
    }
    
    return appointmentModel.ToAppointmentInfo(), tx.Commit().Error
}
```

### 3. 排队机制

- **自动排队**：经纪人不足时自动进入排队
- **位置更新**：取消预约后自动更新排队位置
- **智能分配**：有经纪人空闲时自动分配排队预约

### 4. 状态管理

- **实时状态**：经纪人在线/离线状态实时更新
- **负载均衡**：根据当前负载智能分配
- **超时处理**：自动处理长时间未活跃的经纪人

## 性能优化

### 1. 索引策略

```sql
-- 预约表关键索引
CREATE INDEX idx_appointments_user_time ON appointments(user_id, start_time);
CREATE INDEX idx_appointments_store_date ON appointments(store_id, appointment_date);
CREATE INDEX idx_appointments_status_time ON appointments(status, start_time);
```

### 2. 查询优化

- 使用预加载减少N+1查询
- 合理使用分页避免大结果集
- 利用覆盖索引提高查询效率

### 3. 连接池配置

```yaml
mysql:
  max_idle_conns: 10      # 最大空闲连接数
  max_open_conns: 100     # 最大打开连接数
  max_lifetime: 3600      # 连接最大生存时间
```

## 监控和运维

### 1. 关键指标

- **预约成功率**：预约创建成功的比例
- **经纪人利用率**：经纪人平均负载情况
- **排队时长**：用户平均排队等待时间
- **响应时间**：API接口响应时间

### 2. 日志记录

- **操作日志**：记录所有预约状态变更
- **性能日志**：记录慢查询和性能问题
- **错误日志**：记录系统异常和错误

### 3. 备份策略

```bash
# 每日全量备份
mysqldump -u username -p anjuke_appointment > backup_$(date +%Y%m%d).sql

# 增量备份（基于binlog）
mysqlbinlog --start-datetime="2025-01-30 00:00:00" mysql-bin.000001 > incremental_backup.sql
```

## 常见问题

### Q1: 如何处理高并发预约？

A: 使用以下策略：
- 数据库连接池优化
- 读写分离（主从复制）
- 缓存热点数据
- 异步处理非关键操作

### Q2: 如何保证预约不超售？

A: 通过以下机制：
- 数据库事务保证原子性
- 乐观锁防止并发冲突
- 实时库存检查
- 排队机制缓冲高峰

### Q3: 如何优化查询性能？

A: 采用以下方法：
- 合理设计索引
- 使用查询缓存
- 分页查询大结果集
- 定期分析慢查询

## 迁移检查清单

- [ ] 数据库环境准备完成
- [ ] 配置文件更新完成
- [ ] 数据库表结构创建完成
- [ ] 测试数据导入完成
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 性能测试通过
- [ ] 监控配置完成
- [ ] 备份策略制定完成
- [ ] 上线方案确认完成

## 总结

MySQL版本的预约系统相比MongoDB版本具有以下优势：

1. **更强的数据一致性**：ACID事务支持
2. **更好的查询能力**：复杂SQL查询支持
3. **更成熟的生态**：丰富的工具和经验
4. **更好的性能**：针对关系型数据优化

建议在生产环境中使用MySQL版本，以获得更好的稳定性和性能表现。