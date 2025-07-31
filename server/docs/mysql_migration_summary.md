# Company和Appointment模块MySQL迁移总结

## 修改概述

本次修改将company模块和appointment模块从MongoDB存储修正为MySQL存储，主要涉及以下文件的修改：

## 1. Domain层修改

### server/internal/domain/company.go
- 将MongoDB的`primitive.ObjectID`类型改为MySQL的`uint64`自增主键
- 将MongoDB的`primitive.DateTime`改为Go标准的`time.Time`
- 移除MongoDB特有的bson标签，使用标准的json标签
- 将数组类型字段改为JSON字符串存储（如BusinessArea、MainBusinessArea等）

## 2. Data层修改

### server/internal/data/company.go
- 完全重写为MySQL实现
- 创建了`CompanyMySQLRepo`结构体替代原来的MongoDB实现
- 添加了MySQL表模型：
  - `CompanyModel` - 公司表
  - `CompanyStoreModel` - 门店表  
  - `CompanyRealtorModel` - 经纪人表
- 实现了所有CRUD操作的MySQL版本
- 添加了模型转换方法

### server/internal/data/appointment.go
- 添加了`RealtorWorkingHoursModel`表模型
- 完善了工作时间管理相关方法：
  - `GetRealtorWorkingHours`
  - `CreateRealtorWorkingHours`
  - `UpdateRealtorWorkingHours`
  - `DeleteRealtorWorkingHours`

### server/internal/data/data_mysql.go
- 修复了GORM配置中的日志问题
- 添加了自动迁移功能，包含所有表模型
- 配置了数据库连接池参数

## 3. 表结构设计

### 公司相关表
```sql
-- companies 公司表
CREATE TABLE companies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    company_logo VARCHAR(255),
    full_name VARCHAR(100) NOT NULL,
    short_name VARCHAR(50) NOT NULL,
    business_scope TEXT,
    address VARCHAR(200) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_full_name (full_name),
    INDEX idx_short_name (short_name)
);

-- stores 门店表
CREATE TABLE stores (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    store_name VARCHAR(100) NOT NULL,
    company_id BIGINT NOT NULL,
    address VARCHAR(200) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    business_hours VARCHAR(50),
    rating DECIMAL(3,2) DEFAULT 0,
    review_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_company_id (company_id),
    FOREIGN KEY (company_id) REFERENCES companies(id)
);

-- realtors 经纪人表
CREATE TABLE realtors (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    realtor_name VARCHAR(50) NOT NULL,
    business_area JSON,
    second_hand_score INT DEFAULT 0,
    rental_score INT DEFAULT 0,
    service_years VARCHAR(20),
    service_people_count INT DEFAULT 0,
    main_business_area JSON,
    main_residential_areas JSON,
    company_id BIGINT NOT NULL,
    store_id BIGINT NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_company_id (company_id),
    INDEX idx_store_id (store_id),
    FOREIGN KEY (company_id) REFERENCES companies(id),
    FOREIGN KEY (store_id) REFERENCES stores(id)
);
```

### 预约相关表
- `appointments` - 预约主表
- `appointment_logs` - 预约日志表
- `store_working_hours` - 门店工作时间表
- `realtor_working_hours` - 经纪人工作时间表
- `realtor_status` - 经纪人状态表

## 4. 主要改进

1. **性能优化**：使用MySQL的自增主键替代MongoDB的ObjectID
2. **关系完整性**：添加了外键约束确保数据一致性
3. **索引优化**：为常用查询字段添加了索引
4. **类型安全**：使用强类型的uint64替代字符串ID
5. **事务支持**：利用MySQL的事务特性确保数据一致性

## 5. 注意事项

1. 数组字段使用JSON格式存储，需要在应用层进行序列化/反序列化
2. 所有ID字段都使用uint64类型，需要注意类型转换
3. 时间字段统一使用time.Time类型
4. 需要确保MySQL版本支持JSON字段类型（MySQL 5.7+）

## 6. 后续工作

1. 更新相关的服务层和业务层代码
2. 编写数据迁移脚本（从MongoDB迁移到MySQL）
3. 更新API文档和测试用例
4. 性能测试和优化