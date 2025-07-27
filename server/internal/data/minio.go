package data

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
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
	uploader *MultipartUploader // 分片上传器
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

	// 初始化分片上传器，传入自身指针
	minioClient.uploader = NewMultipartUploader(minioClient, client, c.Bucket, c.Endpoint, logger)

	return minioClient, nil
}

// GetBucket 获取Bucket名称
func (m *MinioClient) GetBucket() string {
	return m.bucket
}

// 前后端协作分片上传相关方法

// SimpleUpload 简单上传（小文件）
func (m *MinioClient) SimpleUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// 简化日志输出
	m.log.Infof("开始上传: %s", objectName)

	finalObjectName := m.uploader.generateObjectName(objectName)

	_, err := m.client.PutObject(ctx, m.bucket, finalObjectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		m.log.Errorf("上传失败: %v", err)
		return "", fmt.Errorf("上传失败: %v", err)
	}

	url := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, finalObjectName)
	m.log.Infof("上传完成: %s", url)

	return url, nil
}

// SmartUpload 统一上传接口，供外部调用
func (m *MinioClient) SmartUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
	return m.uploader.SmartUpload(ctx, objectName, reader, size, contentType, progressCb)
}

// Download 下载文件
func (m *MinioClient) Download(ctx context.Context, objectName string) (io.Reader, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Delete 删除文件
func (m *MinioClient) Delete(ctx context.Context, objectName string) error {
	// 删除前先检查文件是否存在
	_, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		respErr := minio.ToErrorResponse(err)
		if respErr.Code == "NoSuchKey" || respErr.Code == "NotFound" {
			return fmt.Errorf("文件不存在或已被删除")
		}
		return fmt.Errorf("检查文件状态失败: %v", err)
	}

	err = m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	// 删除后再次检查，确保文件已被删除
	_, err = m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err == nil {
		return fmt.Errorf("文件删除后依然存在")
	}
	respErr := minio.ToErrorResponse(err)
	if respErr.Code != "NoSuchKey" && respErr.Code != "NotFound" {
		return fmt.Errorf("删除后检查文件状态失败: %v", err)
	}

	return nil
}

// CleanupIncompleteUploads 清理未完成的分片上传
func (m *MinioClient) CleanupIncompleteUploads(ctx context.Context, prefix string, olderThan time.Duration) error {
	m.log.Infof("开始清理未完成的上传: prefix=%s, olderThan=%v", prefix, olderThan)

	// 获取未完成的上传
	uploadsCh := m.client.ListIncompleteUploads(ctx, m.bucket, prefix, true)

	var count int
	for upload := range uploadsCh {
		if upload.Err != nil {
			m.log.Errorf("列出未完成上传失败: %v", upload.Err)
			continue
		}

		// 检查上传时间
		if time.Since(upload.Initiated) > olderThan {
			// minio-go v7 无法直接中止分片上传，这里仅做日志记录
			m.log.Warnf("检测到未完成上传但无法自动中止: %s (uploadId=%s)", upload.Key, upload.UploadID)
			count++
		}
	}

	m.log.Infof("清理完成，共清理 %d 个未完成上传", count)
	return nil
}

// 确保MinioClient实现了domain.MinioRepo接口
var _ domain.MinioRepo = (*MinioClient)(nil)

// 实现domain.MinioRepo接口
func NewMinioRepo(mc *MinioClient) domain.MinioRepo {
	return mc
}

// 自动补全 MinioRepo 接口方法
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
		})
	}

	return files, nil
}

func (m *MinioClient) GetFileInfo(ctx context.Context, objectName string) (*domain.FileInfo, error) {
	// TODO: 实现或调用实际逻辑
	return nil, nil
}

func (m *MinioClient) SearchFiles(ctx context.Context, keyword string, maxKeys int) ([]domain.FileInfo, error) {
	var files []domain.FileInfo

	opts := minio.ListObjectsOptions{
		Recursive: true,
		MaxKeys:   maxKeys,
	}

	for object := range m.client.ListObjects(ctx, m.bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("搜索文件失败: %v", object.Err)
		}

		if keyword == "" || (len(object.Key) > 0 && containsIgnoreCase(object.Key, keyword)) {
			fileURL := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, object.Key)
			files = append(files, domain.FileInfo{
				Name:         object.Key,
				Size:         object.Size,
				LastModified: object.LastModified,
				ContentType:  "",
				ETag:         object.ETag,
				URL:          fileURL,
			})
		}
	}

	return files, nil
}

func (m *MinioClient) GetFileStats(ctx context.Context, prefix string) (map[string]any, error) {
	var totalFiles int32
	var totalSize int64
	var todayFiles int32

	today := time.Now().Format("2006-01-02")

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
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

	stats := map[string]any{
		"totalUploads":   totalFiles,
		"successUploads": totalFiles, // 假设所有文件都上传成功
		"totalSize":      totalSize,
		"todayUploads":   todayFiles,
	}

	m.log.Infof("文件统计: %+v", stats)
	return stats, nil
}

func containsIgnoreCase(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (stringContainsFold(s, substr)))
}

func stringContainsFold(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
