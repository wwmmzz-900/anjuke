package data

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/minio/minio-go/v7"
)

const MultipartThreshold = 5 * 1024 * 1024 // 5MB

// MultipartUploader 分片上传管理器
type MultipartUploader struct {
	minioClient *MinioClient
	client      *minio.Client
	bucket      string
	endpoint    string
	log         *log.Helper
}

func NewMultipartUploader(minioClient *MinioClient, client *minio.Client, bucket, endpoint string, logger log.Logger) *MultipartUploader {
	return &MultipartUploader{
		minioClient: minioClient,
		client:      client,
		bucket:      bucket,
		endpoint:    endpoint,
		log:         log.NewHelper(logger),
	}
}

// 生成带时间戳的对象名
func (m *MultipartUploader) generateObjectName(originalName string) string {
	ext := filepath.Ext(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Format("20060102150405")
	hasher := md5.New()
	hasher.Write([]byte(originalName + timestamp))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8]
	return fmt.Sprintf("%s_%s_%s%s", nameWithoutExt, timestamp, hash, ext)
}

// SmartUpload 自动分片/普通上传，支持进度回调
func (m *MultipartUploader) SmartUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
	// 简化日志输出
	m.log.Infof("处理上传: %s", objectName)

	// 如果文件小于5MB，直接用SimpleUpload
	if size < MultipartThreshold {
		return m.minioClient.SimpleUpload(ctx, objectName, reader, size, contentType)
	}

	finalObjectName := m.generateObjectName(objectName)
	pr := reader
	if progressCb != nil {
		pr = &progressReader{r: reader, total: size, cb: progressCb}
	}

	// 对于大文件采用分片上传，提高可靠性
	if size > 50*1024*1024 { // 50MB以上使用分片上传
		m.log.Infof("开始分片上传: %s", finalObjectName)

		// 使用官方推荐的方式创建分片上传
		multipartOpts := minio.PutObjectOptions{ContentType: contentType}

		// 添加重试逻辑
		maxRetries := 3
		var lastErr error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			if attempt > 1 {
				m.log.Infof("重试上传 (第 %d 次)", attempt)
			}

			// 重置reader位置（如果可能的话）
			if attempt > 1 {
				if seeker, ok := pr.(io.Seeker); ok {
					if _, err := seeker.Seek(0, io.SeekStart); err != nil {
						m.log.Errorf("无法重置文件读取位置: %v", err)
						return "", fmt.Errorf("无法重置文件读取位置: %v", err)
					}
				} else {
					m.log.Errorf("无法重试上传：reader不支持seek操作")
					break
				}
			}

			// 使用标准上传但增加超时监控
			uploadCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			defer cancel()

			info, err := m.client.PutObject(uploadCtx, m.bucket, finalObjectName, pr, size, multipartOpts)
			if err == nil {
				m.log.Infof("上传成功: %s", info.Key)
				lastErr = nil
				break
			}

			m.log.Errorf("上传尝试 %d 失败: %v", attempt, err)
			lastErr = err

			// 最后一次尝试失败，返回错误
			if attempt == maxRetries {
				return "", fmt.Errorf("在 %d 次尝试后上传失败: %v", maxRetries, err)
			}

			// 重试延迟，随着尝试次数增加
			retryDelay := time.Duration(attempt) * 3 * time.Second
			m.log.Infof("等待 %v 后重试...", retryDelay)
			time.Sleep(retryDelay)
		}

		// 如果有错误，返回
		if lastErr != nil {
			return "", fmt.Errorf("上传失败: %v", lastErr)
		}
	} else {
		// 对于较小的文件，继续使用PutObject
		m.log.Infof("使用普通上传方式: %s, size: %d", finalObjectName, size)
		_, err := m.client.PutObject(ctx, m.bucket, finalObjectName, pr, size, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			m.log.Errorf("MinIO PutObject失败: %v", err)
			return "", fmt.Errorf("MinIO上传失败: %v", err)
		}
	}

	url := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucket, finalObjectName)
	m.log.Infof("文件已上传: %s", finalObjectName)
	return url, nil
}

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
