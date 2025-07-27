package mocks

import (
	"context"
	"io"
	"strings"
	"time"

	"anjuke/server/internal/domain"
)

// MockUserRepo 模拟用户仓储
type MockUserRepo struct {
	SmsRiskControlFunc            func(ctx context.Context, phone, deviceID, ip string) error
	RealNameFunc                  func(ctx context.Context, user *domain.RealName) (*domain.RealName, error)
	UpdateUserStatusFunc          func(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error)
	CheckPhoneExistsFunc          func(ctx context.Context, phone string) (bool, error)
	CreateUserFunc                func(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error)
	GetUserByPhoneAndPasswordFunc func(ctx context.Context, phone, password string) (*domain.UserBase, error)
	GetUserByPhoneFunc            func(ctx context.Context, phone string) (*domain.UserBase, error)
}

func (m *MockUserRepo) SmsRiskControl(ctx context.Context, phone, deviceID, ip string) error {
	if m.SmsRiskControlFunc != nil {
		return m.SmsRiskControlFunc(ctx, phone, deviceID, ip)
	}
	return nil
}

func (m *MockUserRepo) RealName(ctx context.Context, user *domain.RealName) (*domain.RealName, error) {
	if m.RealNameFunc != nil {
		return m.RealNameFunc(ctx, user)
	}
	return user, nil
}

func (m *MockUserRepo) UpdateUserStatus(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	if m.UpdateUserStatusFunc != nil {
		return m.UpdateUserStatusFunc(ctx, user)
	}
	return user, nil
}

func (m *MockUserRepo) CheckPhoneExists(ctx context.Context, phone string) (bool, error) {
	if m.CheckPhoneExistsFunc != nil {
		return m.CheckPhoneExistsFunc(ctx, phone)
	}
	return false, nil
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *domain.UserBase) (*domain.UserBase, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return user, nil
}

func (m *MockUserRepo) GetUserByPhoneAndPassword(ctx context.Context, phone, password string) (*domain.UserBase, error) {
	if m.GetUserByPhoneAndPasswordFunc != nil {
		return m.GetUserByPhoneAndPasswordFunc(ctx, phone, password)
	}
	return &domain.UserBase{Phone: phone}, nil
}

func (m *MockUserRepo) GetUserByPhone(ctx context.Context, phone string) (*domain.UserBase, error) {
	if m.GetUserByPhoneFunc != nil {
		return m.GetUserByPhoneFunc(ctx, phone)
	}
	return &domain.UserBase{Phone: phone}, nil
}

// MockSmsRepo 模拟短信仓储
type MockSmsRepo struct {
	SendSmsFunc   func(ctx context.Context, phone, scene string) (string, error)
	VerifySmsFunc func(ctx context.Context, phone, code, scene string) (bool, error)
}

func (m *MockSmsRepo) SendSms(ctx context.Context, phone, scene string) (string, error) {
	if m.SendSmsFunc != nil {
		return m.SendSmsFunc(ctx, phone, scene)
	}
	return "短信发送成功", nil
}

func (m *MockSmsRepo) VerifySms(ctx context.Context, phone, code, scene string) (bool, error) {
	if m.VerifySmsFunc != nil {
		return m.VerifySmsFunc(ctx, phone, code, scene)
	}
	return code == "123456", nil
}

// MockMinioRepo 模拟MinIO仓储
type MockMinioRepo struct {
	SimpleUploadFunc             func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	SmartUploadFunc              func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error)
	DownloadFunc                 func(ctx context.Context, objectName string) (io.Reader, error)
	DeleteFunc                   func(ctx context.Context, objectName string) error
	ListFilesFunc                func(ctx context.Context, prefix string, maxKeys int) ([]domain.FileInfo, error)
	SearchFilesFunc              func(ctx context.Context, keyword string, maxKeys int) ([]domain.FileInfo, error)
	GetFileStatsFunc             func(ctx context.Context, prefix string) (map[string]any, error)
	GetFileInfoFunc              func(ctx context.Context, objectName string) (*domain.FileInfo, error)
	CleanupIncompleteUploadsFunc func(ctx context.Context, prefix string, olderThan time.Duration) error
	GetBucketFunc                func() string
}

func (m *MockMinioRepo) SimpleUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	if m.SimpleUploadFunc != nil {
		return m.SimpleUploadFunc(ctx, objectName, reader, size, contentType)
	}
	return "http://localhost:9000/test/" + objectName, nil
}

func (m *MockMinioRepo) SmartUpload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
	if m.SmartUploadFunc != nil {
		return m.SmartUploadFunc(ctx, objectName, reader, size, contentType, progressCb)
	}
	if progressCb != nil {
		progressCb(size, size)
	}
	return "http://localhost:9000/test/" + objectName, nil
}

func (m *MockMinioRepo) Delete(ctx context.Context, objectName string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, objectName)
	}
	return nil
}

func (m *MockMinioRepo) ListFiles(ctx context.Context, prefix string, maxKeys int) ([]domain.FileInfo, error) {
	if m.ListFilesFunc != nil {
		return m.ListFilesFunc(ctx, prefix, maxKeys)
	}
	return []domain.FileInfo{
		{
			Name:         "test.txt",
			Size:         1024,
			LastModified: time.Now(),
			ContentType:  "text/plain",
			ETag:         "test-etag",
			URL:          "http://localhost:9000/test/test.txt",
		},
	}, nil
}

func (m *MockMinioRepo) SearchFiles(ctx context.Context, keyword string, maxKeys int) ([]domain.FileInfo, error) {
	if m.SearchFilesFunc != nil {
		return m.SearchFilesFunc(ctx, keyword, maxKeys)
	}
	return []domain.FileInfo{}, nil
}

func (m *MockMinioRepo) Download(ctx context.Context, objectName string) (io.Reader, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, objectName)
	}
	return strings.NewReader("mock file content"), nil
}

func (m *MockMinioRepo) GetFileStats(ctx context.Context, prefix string) (map[string]any, error) {
	if m.GetFileStatsFunc != nil {
		return m.GetFileStatsFunc(ctx, prefix)
	}
	return map[string]any{
		"totalUploads":   int32(10),
		"successUploads": int32(10),
		"totalSize":      int64(10240),
		"todayUploads":   int32(5),
	}, nil
}

func (m *MockMinioRepo) GetBucket() string {
	if m.GetBucketFunc != nil {
		return m.GetBucketFunc()
	}
	return "test-bucket"
}

func (m *MockMinioRepo) GetFileInfo(ctx context.Context, objectName string) (*domain.FileInfo, error) {
	if m.GetFileInfoFunc != nil {
		return m.GetFileInfoFunc(ctx, objectName)
	}
	return &domain.FileInfo{
		Name: objectName,
		Size: 1024,
		URL:  "http://localhost:9000/test/" + objectName,
	}, nil
}

func (m *MockMinioRepo) CleanupIncompleteUploads(ctx context.Context, prefix string, olderThan time.Duration) error {
	if m.CleanupIncompleteUploadsFunc != nil {
		return m.CleanupIncompleteUploadsFunc(ctx, prefix, olderThan)
	}
	return nil
}
