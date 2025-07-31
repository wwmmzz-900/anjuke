# Domain层价值分析和优化建议

## 当前Domain层的问题

### 1. **混合了不同层次的概念**
```go
// ❌ 这些应该在Service层，不是Domain层
type CreateAppointmentRequest struct {
    UserID          int64  `json:"user_id" validate:"required"`
    StoreID         string `json:"store_id" validate:"required"`
    // ...
}

type CreateAppointmentResponse struct {
    Appointment *AppointmentInfo `json:"appointment"`
    NeedQueue   bool             `json:"need_queue"`
    // ...
}
```

### 2. **接口定义过于具体**
```go
// ❌ 这些接口太具体，应该更抽象
type AppointmentRepo interface {
    GetAppointmentsByUser(ctx context.Context, userID int64, page, pageSize int32) ([]*AppointmentInfo, int64, error)
    GetAppointmentsByRealtor(ctx context.Context, realtorID uint64, date time.Time) ([]*AppointmentInfo, error)
    // 太多具体的查询方法
}
```

### 3. **数据模型和业务模型混合**
```go
// ❌ 这更像数据模型，不是领域模型
type CompanyInfo struct {
    ID            uint64    `json:"id"`                     // MySQL 自增主键
    CompanyLogo   string    `json:"company_logo,omitempty"` // 公司Logo URL
    CreatedAt     time.Time `json:"created_at"`             // 创建时间
    UpdatedAt     time.Time `json:"updated_at"`             // 更新时间
}
```

## 优化建议

### 方案1: 保留Domain层但重构 (推荐)

#### 1.1 纯粹的领域模型
```go
// ✅ 纯粹的业务实体
type Company struct {
    id           CompanyID
    name         CompanyName
    businessScope BusinessScope
    contactInfo  ContactInfo
}

// ✅ 值对象
type CompanyID struct {
    value uint64
}

type CompanyName struct {
    fullName  string
    shortName string
}
```

#### 1.2 简化的仓储接口
```go
// ✅ 更抽象的接口
type CompanyRepository interface {
    Save(ctx context.Context, company *Company) error
    FindByID(ctx context.Context, id CompanyID) (*Company, error)
    FindAll(ctx context.Context, criteria SearchCriteria) ([]*Company, error)
    Delete(ctx context.Context, id CompanyID) error
}
```

#### 1.3 移除请求/响应对象
```go
// ❌ 移除这些，放到Service层
// type CreateCompanyRequest struct { ... }
// type CreateCompanyResponse struct { ... }
```

### 方案2: 简化Domain层 (当前可行)

保持当前结构，但做以下调整：

#### 2.1 分离关注点
```go
// domain/entity.go - 纯业务实体
type CompanyInfo struct {
    ID            uint64
    FullName      string
    ShortName     string
    BusinessScope string
    Address       string
    Phone         string
}

// domain/repository.go - 仓储接口
type CompanyRepo interface {
    Create(ctx context.Context, company *CompanyInfo) (*CompanyInfo, error)
    GetByID(ctx context.Context, id string) (*CompanyInfo, error)
    Update(ctx context.Context, company *CompanyInfo) (*CompanyInfo, error)
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, criteria ListCriteria) ([]*CompanyInfo, int64, error)
}

// service/dto.go - 请求响应对象移到这里
type CreateCompanyRequest struct { ... }
type CreateCompanyResponse struct { ... }
```

### 方案3: 移除Domain层 (不推荐)

如果选择移除Domain层：

#### 3.1 后果
- 违反DDD原则
- 业务逻辑分散
- 难以测试
- 不符合Kratos框架设计

#### 3.2 替代方案
```go
// internal/model/company.go
type Company struct { ... }

// internal/repository/company.go  
type CompanyRepository interface { ... }

// internal/service/company.go
type CompanyService struct { ... }
```

## 结论

### ✅ Domain层应该保留，因为：
1. **符合DDD架构**：这是Kratos推荐的架构模式
2. **依赖倒置**：上层不依赖具体实现
3. **业务聚合**：业务规则集中管理
4. **可测试性**：便于单元测试
5. **可维护性**：清晰的层次结构

### 🔧 但需要优化：
1. **移除请求/响应对象**到Service层
2. **简化仓储接口**，更抽象
3. **分离数据字段**（如CreatedAt）和业务字段
4. **添加业务方法**到实体中

### 📝 推荐的重构步骤：
1. 先保持当前结构运行
2. 逐步移除Request/Response对象到Service层
3. 简化Repository接口
4. 添加业务方法到实体
5. 考虑引入值对象

你的Domain层是有价值的，只需要适当优化即可。