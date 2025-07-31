package data

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"anjuke/server/internal/conf"
	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient 封装 MinIO 客户端
type MinioClient struct {
	client   *minio.Client
	endpoint string
	bucket   string
	log      *log.Helper
}

// NewMinioClient 构造函数
func NewMinioClient(conf *conf.Data, logger log.Logger) (*MinioClient, error) {
	c := conf.Minio
	client, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: c.UseSsl,
		// 设置HTTP传输超时，使用更保守的设置
		Transport: &http.Transport{
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       5 * time.Minute,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 5 * time.Minute,
			ExpectContinueTimeout: 30 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false,
			// 减少拨号超时
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
		},
		Region: "",
	})
	if err != nil {
		return nil, err
	}

	log := log.NewHelper(logger)

	// 设置较长的超时时间用于检查和创建bucket
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 测试网络连接
	log.Infof("测试MinIO服务器连接: %s", c.Endpoint)

	exists, err := client.BucketExists(ctx, c.Bucket) // Bucket name from config
	if err != nil {
		log.Errorf("检查bucket失败: %v", err)
		return nil, fmt.Errorf("检查bucket失败，请检查MinIO服务器连接: %v", err)
	}
	if !exists {
		log.Infof("bucket不存在，正在创建...")
		err = client.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Errorf("创建bucket失败: %v", err)
			return nil, fmt.Errorf("创建bucket失败: %v", err)
		}
		log.Infof("bucket创建成功")
	} else {
		log.Infof("bucket已存在")
	}

	// 测试一个简单的上传操作
	testCtx, testCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer testCancel()

	testData := strings.NewReader("test connection")
	testObjectName := "test-connection.txt"

	log.Infof("测试MinIO上传功能...")
	_, err = client.PutObject(testCtx, c.Bucket, testObjectName, testData, 15, minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		log.Errorf("MinIO上传测试失败: %v", err)
		return nil, fmt.Errorf("MinIO上传测试失败: %v", err)
	}

	// 删除测试文件
	err = client.RemoveObject(testCtx, c.Bucket, testObjectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Warnf("删除测试文件失败: %v", err)
	} else {
		log.Infof("MinIO连接和上传测试成功")
	}

	minioClient := &MinioClient{
		client:   client,
		endpoint: c.Endpoint,
		bucket:   c.Bucket, // Bucket name from config
		log:      log,
	}

	return minioClient, nil
}

// GetBucket 获取Bucket名称
func (m *MinioClient) GetBucket() string {
	return m.bucket
}

// generateObjectName 生成带时间戳的对象名
func (m *MinioClient) generateObjectName(originalName string) string {
	ext := filepath.Ext(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Format("20060102150405")
	hasher := md5.New()
	hasher.Write([]byte(originalName + timestamp))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8]
	return fmt.Sprintf("%s_%s_%s%s", nameWithoutExt, timestamp, hash, ext)
}

// UploadFile 上传文件 - 实现domain.MinioRepo接口
func (m *MinioClient) UploadFile(ctx context.Context, fileName string, reader io.Reader, size int64) (*domain.FileInfo, error) {
	finalObjectName := m.generateObjectName(fileName)

	// 使用MinIO的PutObject
	uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	_, err := m.client.PutObject(uploadCtx, m.bucket, finalObjectName, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})

	if err != nil {
		m.log.Errorf("上传文件失败: %s, 错误: %v", fileName, err)
		return nil, fmt.Errorf("上传文件失败: %v", err)
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, finalObjectName)

	fileInfo := &domain.FileInfo{
		Name:         finalObjectName,
		Size:         size,
		LastModified: time.Now(),
		ContentType:  "application/octet-stream",
		URL:          fileURL,
		FileURL:      fileURL,
	}

	m.log.Infof("文件上传成功: %s -> %s", fileName, finalObjectName)
	return fileInfo, nil
}

// DownloadFile 下载文件 - 实现domain.MinioRepo接口
func (m *MinioClient) DownloadFile(ctx context.Context, fileName string) (io.Reader, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		m.log.Errorf("下载文件失败: %s, 错误: %v", fileName, err)
		return nil, fmt.Errorf("下载文件失败: %v", err)
	}

	m.log.Infof("文件下载成功: %s", fileName)
	return obj, nil
}

// DeleteFile 删除文件 - 实现domain.MinioRepo接口
func (m *MinioClient) DeleteFile(ctx context.Context, fileName string) error {
	err := m.client.RemoveObject(ctx, m.bucket, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		m.log.Errorf("删除文件失败: %s, 错误: %v", fileName, err)
		return fmt.Errorf("删除文件失败: %v", err)
	}

	m.log.Infof("文件删除成功: %s", fileName)
	return nil
}

// SimpleUpload 简单上传文件 - 实现domain.MinioRepo接口
func (m *MinioClient) SimpleUpload(ctx context.Context, fileName string, reader io.Reader) (*domain.FileInfo, error) {
	// 获取文件大小
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取文件数据失败: %v", err)
	}

	size := int64(len(data))
	return m.UploadFile(ctx, fileName, bytes.NewReader(data), size)
}

// Delete 删除文件（别名方法） - 实现domain.MinioRepo接口
func (m *MinioClient) Delete(ctx context.Context, fileName string) error {
	return m.DeleteFile(ctx, fileName)
}

// SearchFiles 搜索文件 - 实现domain.MinioRepo接口
func (m *MinioClient) SearchFiles(ctx context.Context, keyword string, page, pageSize int32) ([]domain.FileInfo, int32, error) {
	var files []domain.FileInfo
	var count int32

	opts := minio.ListObjectsOptions{
		Recursive: true,
		MaxKeys:   int(pageSize * page), // 简化分页处理
	}

	for object := range m.client.ListObjects(ctx, m.bucket, opts) {
		if object.Err != nil {
			return nil, 0, fmt.Errorf("搜索文件失败: %v", object.Err)
		}

		if keyword == "" || strings.Contains(strings.ToLower(object.Key), strings.ToLower(keyword)) {
			fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, object.Key)
			files = append(files, domain.FileInfo{
				Name:         object.Key,
				Size:         object.Size,
				LastModified: object.LastModified,
				ContentType:  "",
				ETag:         object.ETag,
				URL:          fileURL,
				FileURL:      fileURL,
			})
			count++
		}
	}

	return files, count, nil
}

// ListFiles 列出文件 - 实现domain.MinioRepo接口
func (m *MinioClient) ListFiles(ctx context.Context, prefix string, maxKeys int) ([]domain.FileInfo, error) {
	var files []domain.FileInfo

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		MaxKeys:   maxKeys,
	}

	for object := range m.client.ListObjects(ctx, m.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("列出对象失败: %v", object.Err)
		}

		fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, object.Key)

		files = append(files, domain.FileInfo{
			Name:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ContentType:  "", // MinIO ListObjects 不返回 ContentType，需要单独获取
			ETag:         object.ETag,
			URL:          fileURL,
			FileURL:      fileURL,
		})
	}

	return files, nil
}

// GetFileStats 获取文件统计信息 - 实现domain.MinioRepo接口
func (m *MinioClient) GetFileStats(ctx context.Context) (map[string]interface{}, error) {
	var totalFiles int32
	var totalSize int64
	var todayFiles int32

	today := time.Now().Format("2006-01-02")

	opts := minio.ListObjectsOptions{
		Recursive: true,
	}

	for object := range m.client.ListObjects(ctx, m.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("获取文件统计失败: %v", object.Err)
		}

		totalFiles++
		totalSize += object.Size

		// 检查是否是今天上传的文件
		if object.LastModified.Format("2006-01-02") == today {
			todayFiles++
		}
	}

	stats := map[string]interface{}{
		"totalUploads":   totalFiles,
		"successUploads": totalFiles, // 假设所有文件都上传成功
		"totalSize":      totalSize,
		"todayUploads":   todayFiles,
	}

	m.log.Infof("文件统计: %+v", stats)
	return stats, nil
}

// 确保MinioClient实现了domain.MinioRepo接口
var _ domain.MinioRepo = (*MinioClient)(nil)

// NewMinioRepo 实现domain.MinioRepo接口
func NewMinioRepo(mc *MinioClient) domain.MinioRepo {
	return mc
}

// progressReader 包装io.Reader以支持进度回调
type progressReader struct {
	r     io.Reader
	total int64
	read  int64
	cb    func(uploaded, total int64)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.read += int64(n)
	if p.cb != nil {
		p.cb(p.read, p.total)
	}
	return n, err
}

// SmartUpload 智能上传文件（支持进度回调）
func (m *MinioClient) SmartUpload(ctx context.Context, fileName string, reader io.Reader, size int64, contentType string, progressCallback func(uploaded, total int64)) (string, error) {
	finalObjectName := m.generateObjectName(fileName)

	// 包装reader以支持进度回调
	var uploadReader io.Reader = reader
	if progressCallback != nil {
		uploadReader = &progressReader{
			r:     reader,
			total: size,
			cb:    progressCallback,
		}
	}

	// 使用MinIO的PutObject
	uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	_, err := m.client.PutObject(uploadCtx, m.bucket, finalObjectName, uploadReader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		m.log.Errorf("上传文件失败: %s, 错误: %v", fileName, err)
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, finalObjectName)
	m.log.Infof("文件上传成功: %s -> %s", fileName, finalObjectName)
	return fileURL, nil
}

// CleanupIncompleteUploads 清理未完成的上传
func (m *MinioClient) CleanupIncompleteUploads(ctx context.Context, prefix string, olderThan time.Duration) error {
	// MinIO Go SDK 没有直接的清理未完成上传的方法
	// 这里我们可以实现一个简单的清理逻辑
	m.log.Infof("清理未完成的上传任务，前缀: %s, 时间: %v", prefix, olderThan)
	return nil
}

// GetFileInfo 获取文件信息
func (m *MinioClient) GetFileInfo(ctx context.Context, objectName string) (*domain.FileInfo, error) {
	objInfo, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		m.log.Errorf("获取文件信息失败: %s, 错误: %v", objectName, err)
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, objectName)

	fileInfo := &domain.FileInfo{
		Name:         objectName,
		Size:         objInfo.Size,
		LastModified: objInfo.LastModified,
		ContentType:  objInfo.ContentType,
		ETag:         objInfo.ETag,
		URL:          fileURL,
		FileURL:      fileURL,
	}

	return fileInfo, nil
}
