# MinIO 分片上传完整指南

## 功能概述

本项目实现了完整的MinIO分片上传功能，支持：

- ✅ **自动分片上传**：大文件自动分割为5MB分片
- ✅ **断点续传**：上传中断后可从断点继续
- ✅ **进度监控**：实时显示上传进度和速度
- ✅ **智能选择**：自动选择普通上传或分片上传
- ✅ **错误重试**：每个分片支持3次重试
- ✅ **资源清理**：自动清理超时的未完成上传

## 核心组件

### 1. MultipartUploader 分片上传器
```go
type MultipartUploader struct {
    client *MinioClient
    mutex  sync.RWMutex
}
```

### 2. MultipartUploadInfo 上传信息
```go
type MultipartUploadInfo struct {
    UploadID      string               // 上传ID
    ObjectName    string               // 对象名称
    Bucket        string               // 存储桶
    TotalSize     int64                // 总大小
    ChunkSize     int64                // 分片大小
    TotalChunks   int                  // 总分片数
    UploadedParts []minio.CompletePart // 已上传分片
    CreatedAt     time.Time            // 创建时间
    UpdatedAt     time.Time            // 更新时间
}
```

## 使用方法

### 1. 智能上传（推荐）

```go
// 自动选择普通上传或分片上传
url, err := minioClient.SmartUpload(ctx, "video.mp4", file, fileSize, "video/mp4")
if err != nil {
    log.Printf("上传失败: %v", err)
    return
}
fmt.Printf("上传成功: %s\n", url)
```

### 2. 带进度的智能上传

```go
// 进度回调函数
progressCallback := func(uploaded, total int64) {
    percentage := float64(uploaded) / float64(total) * 100
    fmt.Printf("上传进度: %.2f%% (%d/%d bytes)\n", percentage, uploaded, total)
}

url, err := minioClient.SmartUploadWithProgress(ctx, "large_file.zip", file, fileSize, "application/zip", progressCallback)
```

### 3. 手动分片上传

```go
// 1. 开始分片上传
uploadInfo, err := minioClient.StartMultipartUpload(ctx, "large_file.zip", "application/zip", fileSize)
if err != nil {
    return err
}

uploader := minioClient.GetUploader()

// 2. 上传分片（示例：上传第一个分片）
part, err := uploader.UploadPart(ctx, uploadInfo, 1, partReader, partSize)
if err != nil {
    uploader.AbortMultipartUpload(ctx, uploadInfo)
    return err
}
uploadInfo.UploadedParts = append(uploadInfo.UploadedParts, *part)

// 3. 完成上传
url, err := uploader.CompleteMultipartUpload(ctx, uploadInfo)
```

### 4. 断点续传

```go
// 恢复之前中断的上传
url, err := minioClient.ResumeMultipartUpload(ctx, uploadInfo, file)
if err != nil {
    log.Printf("断点续传失败: %v", err)
    return
}
fmt.Printf("断点续传成功: %s\n", url)
```

## 实际使用示例

### 视频文件上传

```go
func UploadVideo(ctx context.Context, videoPath string) error {
    // 打开视频文件
    file, err := os.Open(videoPath)
    if err != nil {
        return fmt.Errorf("打开视频文件失败: %v", err)
    }
    defer file.Close()

    // 获取文件信息
    fileInfo, err := file.Stat()
    if err != nil {
        return fmt.Errorf("获取文件信息失败: %v", err)
    }

    fmt.Printf("上传视频: %s (%.2f MB)\n", 
        fileInfo.Name(), float64(fileInfo.Size())/(1024*1024))

    // 进度回调
    startTime := time.Now()
    progressCallback := func(uploaded, total int64) {
        percentage := float64(uploaded) / float64(total) * 100
        elapsed := time.Since(startTime)
        speed := float64(uploaded) / elapsed.Seconds() / 1024 / 1024 // MB/s
        
        fmt.Printf("进度: %.1f%% - 速度: %.2f MB/s\n", percentage, speed)
    }

    // 上传视频
    url, err := minioClient.SmartUploadWithProgress(
        ctx, fileInfo.Name(), file, fileInfo.Size(), 
        "video/mp4", progressCallback)
    
    if err != nil {
        return fmt.Errorf("视频上传失败: %v", err)
    }

    elapsed := time.Since(startTime)
    avgSpeed := float64(fileInfo.Size()) / elapsed.Seconds() / 1024 / 1024
    fmt.Printf("上传完成! 耗时: %v, 平均速度: %.2f MB/s\n", elapsed, avgSpeed)
    fmt.Printf("访问地址: %s\n", url)
    
    return nil
}
```

### 批量文件上传

```go
func BatchUpload(ctx context.Context, filePaths []string) error {
    for i, filePath := range filePaths {
        fmt.Printf("上传文件 %d/%d: %s\n", i+1, len(filePaths), filePath)
        
        file, err := os.Open(filePath)
        if err != nil {
            fmt.Printf("跳过文件 %s: %v\n", filePath, err)
            continue
        }
        
        fileInfo, _ := file.Stat()
        
        url, err := minioClient.SmartUpload(ctx, fileInfo.Name(), file, fileInfo.Size(), "application/octet-stream")
        file.Close()
        
        if err != nil {
            fmt.Printf("文件 %s 上传失败: %v\n", filePath, err)
            continue
        }
        
        fmt.Printf("文件 %s 上传成功: %s\n", filePath, url)
    }
    
    return nil
}
```

## 配置参数

### 分片上传常量

```go
const (
    DefaultChunkSize = 5 * 1024 * 1024    // 5MB per part
    LargeFileThreshold = 10 * 1024 * 1024 // 10MB threshold
    MaxChunkSize = 100 * 1024 * 1024      // 100MB max per part
    MinChunkSize = 1024 * 1024            // 1MB min per part
    MaxPartsLimit = 10000                 // Maximum parts limit
)
```

### 上传策略

| 文件大小 | 上传策略 | 分片大小 | 说明 |
|---------|---------|---------|------|
| < 10MB | 普通上传 | - | 直接上传，速度快 |
| 10MB - 100MB | 分片上传 | 5MB | 平衡速度和可靠性 |
| 100MB - 1GB | 分片上传 | 5MB | 支持断点续传 |
| > 1GB | 分片上传 | 5MB | 大文件必须分片 |

## 错误处理

### 常见错误及解决方案

1. **网络中断**
   ```go
   // 保存上传信息，稍后断点续传
   if err != nil {
       saveUploadInfo(uploadInfo) // 自定义保存逻辑
       return fmt.Errorf("上传中断，可稍后续传: %v", err)
   }
   ```

2. **分片上传失败**
   ```go
   // 自动重试3次，失败后取消上传
   part, err := uploader.UploadPart(ctx, uploadInfo, partNumber, reader, partSize)
   if err != nil {
       uploader.AbortMultipartUpload(ctx, uploadInfo)
       return fmt.Errorf("分片上传失败: %v", err)
   }
   ```

3. **存储空间不足**
   ```go
   // 清理旧的未完成上传
   err := uploader.CleanupIncompleteUploads(ctx, "", 24*time.Hour)
   if err != nil {
       log.Printf("清理失败: %v", err)
   }
   ```

## 性能优化

### 1. 并发上传（未来版本）
```go
// 可以实现并发上传多个分片
// 注意：需要控制并发数，避免过多连接
```

### 2. 内存优化
```go
// 使用固定大小的缓冲区，避免内存泄漏
buffer := make([]byte, DefaultChunkSize)
// 重复使用buffer
```

### 3. 网络优化
```go
// 在MinIO客户端配置中设置合适的超时和连接池
Transport: &http.Transport{
    MaxIdleConns:          100,
    MaxIdleConnsPerHost:   10,
    IdleConnTimeout:       90 * time.Second,
    TLSHandshakeTimeout:   10 * time.Second,
    ResponseHeaderTimeout: 30 * time.Second,
}
```

## 监控和维护

### 1. 清理未完成上传
```go
// 定期清理超过24小时的未完成上传
func CleanupRoutine() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        err := uploader.CleanupIncompleteUploads(context.Background(), "", 24*time.Hour)
        if err != nil {
            log.Printf("清理失败: %v", err)
        }
    }
}
```

### 2. 监控上传状态
```go
// 列出所有未完成的上传
uploads, err := uploader.ListIncompleteUploads(ctx, "")
if err != nil {
    log.Printf("获取未完成上传失败: %v", err)
    return
}

for _, upload := range uploads {
    fmt.Printf("未完成上传: %s (开始时间: %v)\n", upload.Key, upload.Initiated)
}
```

## 最佳实践

1. **文件大小判断**：根据文件大小自动选择上传策略
2. **进度显示**：为用户提供清晰的上传进度反馈
3. **错误处理**：妥善处理网络中断等异常情况
4. **资源清理**：定期清理未完成的上传，释放存储空间
5. **断点续传**：保存上传状态，支持中断后继续
6. **并发控制**：避免同时上传过多文件导致资源耗尽

## 总结

这个MinIO分片上传实现提供了：

- 🚀 **高性能**：5MB分片，支持GB级大文件
- 🔄 **可靠性**：断点续传，网络中断不怕
- 📊 **可观测**：实时进度，详细日志
- 🛡️ **健壮性**：自动重试，错误处理
- 🔧 **易用性**：智能选择，简单API

适用于视频上传、文档管理、数据备份等各种大文件处理场景。