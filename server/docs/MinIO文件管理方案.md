# MinIO 直接文件管理方案

## 概述

本项目采用直接从 MinIO 获取文件信息的方案，而不是创建额外的数据库表。这种设计简化了系统架构，保证了数据一致性。

## 架构设计

```
前端 Vue.js
    ↓ HTTP API
后端 Go (Kratos)
    ↓ MinIO SDK
MinIO 对象存储
```

## 核心接口

### 1. 文件上传
- `POST /user/uploadFile` - 单文件上传
- `POST /user/uploadFiles` - 多文件上传

### 2. 文件管理
- `GET /user/fileList` - 获取文件列表
- `GET /user/uploadStats` - 获取统计信息
- `GET /user/fileInfo` - 获取文件详情
- `DELETE /user/deleteFile` - 删除文件

### 3. 维护接口
- `POST /user/cleanupUploads` - 清理未完成上传

## 实现特点

### 优势
1. **数据一致性**: MinIO 是唯一数据源
2. **系统简化**: 无需维护数据库表
3. **实时性**: 文件信息始终最新
4. **维护简单**: 减少数据同步逻辑

### 功能支持
- ✅ 文件列表查询
- ✅ 分页支持
- ✅ 文件名搜索
- ✅ 统计信息
- ✅ 文件详情查询

## 使用示例

### 获取文件列表
```bash
GET /user/fileList?page=1&pageSize=10&keyword=test
```

### 获取统计信息
```bash
GET /user/uploadStats
```

### 上传文件
```bash
POST /user/uploadFile
Content-Type: multipart/form-data

file: [文件数据]
```

## 性能考虑

- MinIO ListObjects API 性能良好
- 支持前缀过滤和客户端分页
- 适合中小规模文件管理场景
- 可通过前端缓存优化用户体验

## 扩展性

如需更复杂功能，可考虑：
1. 添加 Redis 缓存层
2. 使用 Elasticsearch 建立索引
3. 混合数据库存储业务元数据

## 配置要求

确保 MinIO 服务正常运行：
- 端点: 14.103.235.216:9000
- 存储桶: mybucket
- 访问密钥已正确配置