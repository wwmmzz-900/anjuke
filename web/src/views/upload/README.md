# 文件上传功能说明

## 功能概述

本系统提供了完整的文件上传功能，支持多种上传方式和文件类型，集成了 MinIO 对象存储服务。

## 主要特性

### 🚀 上传方式
- **单文件上传** - 支持单个文件上传
- **多文件上传** - 支持批量文件上传
- **拖拽上传** - 支持拖拽文件到指定区域上传
- **手动上传** - 选择文件后手动触发上传
- **自定义上传** - 完全自定义的上传逻辑

### 📁 文件类型支持
- **图片文件** - JPG, PNG, GIF, WebP
- **文档文件** - PDF, DOC, DOCX, TXT
- **压缩文件** - ZIP, RAR, 7Z
- **其他文件** - 支持自定义文件类型

### 🔧 高级功能
- **上传进度显示** - 实时显示上传进度
- **文件预览** - 支持图片和PDF在线预览
- **文件管理** - 上传后的文件列表管理
- **文件删除** - 支持删除已上传的文件
- **断点续传** - 大文件支持断点续传（后端支持）
- **分片上传** - 大文件自动分片上传

## 后端接口

### 文件上传接口

#### 单文件上传
```
POST /user/uploadFile
Content-Type: multipart/form-data

参数:
- file: 文件对象

响应:
{
  "code": 0,
  "msg": "上传成功",
  "data": {
    "url": "文件访问URL",
    "objectName": "对象存储中的文件名"
  }
}
```

#### 多文件上传
```
POST /user/uploadFiles
Content-Type: multipart/form-data

参数:
- files: 文件数组

响应:
{
  "code": 0,
  "msg": "上传成功",
  "data": {
    "urls": ["文件URL数组"],
    "objectNames": ["对象名数组"]
  }
}
```

#### 文件删除
```
DELETE /user/deleteFile?objectName=文件对象名

响应:
{
  "code": 0,
  "msg": "删除成功"
}
```

#### 清理未完成上传
```
POST /user/cleanupUploads
Content-Type: application/json

参数:
{
  "prefix": "文件前缀",
  "olderThan": "24h"
}
```

## 前端组件使用

### FileUpload 组件

```vue
<template>
  <FileUpload
    v-model="fileList"
    :multiple="true"
    :drag="true"
    :max-size="100"
    accept="image/*,.pdf,.doc,.docx"
    tip="支持图片、PDF、Word文件，最大100MB"
    @success="handleUploadSuccess"
    @error="handleUploadError"
  />
</template>

<script>
import FileUpload from '@/components/FileUpload.vue'

export default {
  components: {
    FileUpload
  },
  data() {
    return {
      fileList: []
    }
  },
  methods: {
    handleUploadSuccess(response, file, fileList) {
      console.log('上传成功:', response)
    },
    handleUploadError(error, file, fileList) {
      console.error('上传失败:', error)
    }
  }
}
</script>
```

### 组件属性

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| action | String | '/api/user/uploadFile' | 上传地址 |
| multiple | Boolean | false | 是否支持多选 |
| accept | String | '' | 接受的文件类型 |
| limit | Number | 10 | 最大上传数量 |
| maxSize | Number | 100 | 文件大小限制(MB) |
| autoUpload | Boolean | true | 是否自动上传 |
| showFileList | Boolean | true | 是否显示文件列表 |
| drag | Boolean | false | 是否启用拖拽上传 |
| disabled | Boolean | false | 是否禁用 |
| buttonText | String | '选择文件' | 按钮文字 |
| tip | String | '' | 提示文字 |

### 组件事件

| 事件名 | 参数 | 说明 |
|--------|------|------|
| success | (response, file, fileList) | 上传成功 |
| error | (error, file, fileList) | 上传失败 |
| progress | (event, file, fileList) | 上传进度 |
| change | (file, fileList) | 文件状态改变 |
| remove | (file, fileList) | 文件移除 |

## 使用示例

### 1. 头像上传
```vue
<el-upload
  class="avatar-uploader"
  action="/api/user/uploadFile"
  :show-file-list="false"
  :on-success="handleAvatarSuccess"
  :before-upload="beforeAvatarUpload"
>
  <img v-if="avatarUrl" :src="avatarUrl" class="avatar" />
  <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
</el-upload>
```

### 2. 图片列表上传
```vue
<el-upload
  action="/api/user/uploadFiles"
  list-type="picture-card"
  :on-preview="handlePictureCardPreview"
  multiple
  accept="image/*"
>
  <el-icon><Plus /></el-icon>
</el-upload>
```

### 3. 拖拽上传
```vue
<el-upload
  class="upload-demo"
  drag
  action="/api/user/uploadFile"
  multiple
>
  <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
  <div class="el-upload__text">
    将文件拖到此处，或<em>点击上传</em>
  </div>
</el-upload>
```

## 配置说明

### 文件类型限制
```javascript
// 图片文件
accept="image/*"

// 特定图片格式
accept=".jpg,.jpeg,.png,.gif"

// 文档文件
accept=".pdf,.doc,.docx,.txt"

// 多种类型
accept="image/*,.pdf,.doc,.docx"
```

### 文件大小限制
```javascript
// 在 beforeUpload 钩子中检查
beforeUpload(file) {
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isLt2M) {
    this.$message.error('文件大小不能超过 2MB!')
    return false
  }
  return true
}
```

## 注意事项

1. **文件大小限制**: 默认单个文件最大100MB，可根据需要调整
2. **文件类型检查**: 前端和后端都应该进行文件类型验证
3. **安全性**: 上传的文件应该进行安全扫描
4. **存储空间**: 注意监控存储空间使用情况
5. **网络超时**: 大文件上传时注意设置合适的超时时间

## 故障排除

### 常见问题

1. **上传失败**: 检查网络连接和后端服务状态
2. **文件过大**: 调整文件大小限制或使用分片上传
3. **类型不支持**: 检查 accept 属性和后端文件类型验证
4. **权限问题**: 确认用户有上传权限
5. **存储空间不足**: 检查 MinIO 存储空间

### 调试方法

1. 打开浏览器开发者工具查看网络请求
2. 检查控制台错误信息
3. 查看后端日志
4. 测试 MinIO 连接状态

## 真实数据集成

### 数据流程

1. **文件上传**: 前端调用 `/user/uploadFile` 或 `/user/uploadFiles` 接口
2. **后端处理**: 后端将文件上传到 MinIO 对象存储
3. **返回结果**: 后端返回文件访问URL和对象名
4. **前端展示**: 前端更新文件列表和统计信息

### API 响应格式

#### 上传成功响应
```json
{
  "code": 0,
  "msg": "上传成功",
  "data": {
    "url": "http://14.103.235.216:9000/mybucket/1640995200_example.jpg",
    "objectName": "1640995200_example.jpg"
  }
}
```

#### 文件列表响应 (需要后端实现)
```json
{
  "code": 0,
  "msg": "获取成功",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "example.jpg",
        "size": 1024000,
        "type": "image/jpeg",
        "url": "http://14.103.235.216:9000/mybucket/1640995200_example.jpg",
        "objectName": "1640995200_example.jpg",
        "uploadTime": "2024-01-15 10:30:00",
        "status": "success"
      }
    ],
    "total": 1
  }
}
```

#### 统计数据响应 (需要后端实现)
```json
{
  "code": 0,
  "msg": "获取成功",
  "data": {
    "totalUploads": 156,
    "successUploads": 148,
    "totalSize": 2684354560,
    "todayUploads": 23
  }
}
```

### 前端数据处理

#### 文件列表加载
```javascript
const loadFileList = async () => {
  try {
    const response = await uploadApi.getFileList({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      keyword: searchKeyword.value
    })
    
    fileList.value = response.list.map(file => ({
      id: file.id,
      name: file.name,
      size: file.size,
      type: file.type,
      url: file.url,
      uploadTime: file.uploadTime,
      status: file.status,
      objectName: file.objectName
    }))
    
    pagination.value.total = response.total
  } catch (error) {
    console.error('加载文件列表失败:', error)
  }
}
```

#### 上传成功处理
```javascript
const handleUploadSuccess = (response, file) => {
  // 添加到本地列表
  const newFile = {
    id: Date.now(),
    name: file.name,
    size: file.size,
    type: file.type,
    url: response.data?.url || response.url,
    uploadTime: new Date().toLocaleString(),
    status: 'success',
    objectName: response.data?.objectName || response.objectName
  }
  
  // 重新加载列表和统计
  loadFileList()
  loadUploadStats()
}
```

### 测试页面

访问 `/upload/test` 页面可以进行以下测试：

1. **简单上传测试**: 测试基本的文件上传功能
2. **拖拽上传测试**: 测试拖拽多文件上传
3. **API 直接测试**: 直接调用上传API进行测试
4. **接口功能测试**: 测试清理和统计接口

### 后端接口待实现

以下接口需要后端实现以支持完整功能：

1. `GET /user/fileList` - 获取文件列表
2. `GET /user/uploadStats` - 获取上传统计
3. `POST /user/cleanupUploads` - 清理未完成上传

如果这些接口暂未实现，前端会优雅降级，不会影响基本的上传功能。