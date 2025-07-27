package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	uploadv1 "anjuke/server/api/upload/v1"
	"anjuke/server/internal/mocks"
	"anjuke/server/internal/testutil"
)

func TestUploadService_SimpleUpload(t *testing.T) {
	tests := []struct {
		name          string
		req           *uploadv1.SimpleUploadRequest
		mockMinioRepo func() *mocks.MockMinioRepo
		expectError   bool
		errorContains string
	}{
		{
			name: "简单上传成功",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "test.txt",
				FileData:    []byte("Hello, World!"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "http://localhost:9000/test/test.txt", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "大文件使用智能上传",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "large-file.txt",
				FileData:    make([]byte, 6*1024*1024), // 6MB 文件
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "http://localhost:9000/test/large-file.txt", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "文件名为空",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "",
				FileData:    []byte("Hello, World!"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "http://localhost:9000/test/", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "文件内容为空",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "empty.txt",
				FileData:    []byte{},
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "http://localhost:9000/test/empty.txt", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "上传失败",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "fail.txt",
				FileData:    []byte("test content"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "", errors.New("storage error")
					},
				}
			},
			expectError:   true,
			errorContains: "文件上传失败",
		},
		{
			name: "大文件上传失败",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "large-fail.txt",
				FileData:    make([]byte, 6*1024*1024), // 6MB 文件
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "", errors.New("smart upload failed")
					},
				}
			},
			expectError:   true,
			errorContains: "文件上传失败",
		},
		{
			name: "特殊字符文件名",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "测试文件-2024.txt",
				FileData:    []byte("中文内容测试"),
				ContentType: "text/plain; charset=utf-8",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "http://localhost:9000/test/测试文件-2024.txt", nil
					},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUploadService(tt.mockMinioRepo(), testutil.MockLogger())

			resp, err := service.SimpleUpload(context.Background(), tt.req)

			testutil.AssertError(t, err, tt.expectError, "SimpleUpload error check")

			if tt.expectError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				testutil.AssertNotNil(t, resp, "SimpleUpload response")
				if resp.Url == "" {
					t.Error("SimpleUpload() URL should not be empty")
				}
				if tt.req.Filename != "" && !strings.Contains(resp.Url, tt.req.Filename) {
					t.Errorf("SimpleUpload() URL should contain filename: %s", resp.Url)
				}
			}
		})
	}
}

func TestUploadService_SmartUpload(t *testing.T) {
	tests := []struct {
		name          string
		req           *SmartUploadRequest
		mockMinioRepo func() *mocks.MockMinioRepo
		expectError   bool
		errorContains string
	}{
		{
			name: "智能上传成功",
			req: &SmartUploadRequest{
				Filename:    "test.txt",
				FileData:    []byte("Hello, World!"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						if progressCb != nil {
							progressCb(size, size)
						}
						return "http://localhost:9000/test/test.txt", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "大文件智能上传",
			req: &SmartUploadRequest{
				Filename:    "large-file.zip",
				FileData:    make([]byte, 10*1024*1024), // 10MB 文件
				ContentType: "application/zip",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						// 模拟进度回调
						if progressCb != nil {
							progressCb(size/2, size) // 50% 进度
							progressCb(size, size)   // 100% 完成
						}
						return "http://localhost:9000/test/large-file.zip", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "文件名为空",
			req: &SmartUploadRequest{
				Filename:    "",
				FileData:    []byte("Hello, World!"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "http://localhost:9000/test/", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "上传失败",
			req: &SmartUploadRequest{
				Filename:    "fail.txt",
				FileData:    []byte("test content"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "", errors.New("network timeout")
					},
				}
			},
			expectError:   true,
			errorContains: "network timeout",
		},
		{
			name: "二进制文件上传",
			req: &SmartUploadRequest{
				Filename:    "image.jpg",
				FileData:    []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}, // JPEG 文件头
				ContentType: "image/jpeg",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "http://localhost:9000/test/image.jpg", nil
					},
				}
			},
			expectError: false,
		},
		{
			name: "空文件上传",
			req: &SmartUploadRequest{
				Filename:    "empty.txt",
				FileData:    []byte{},
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						if progressCb != nil {
							progressCb(0, 0)
						}
						return "http://localhost:9000/test/empty.txt", nil
					},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUploadService(tt.mockMinioRepo(), testutil.MockLogger())

			resp, err := service.SmartUpload(context.Background(), tt.req)

			testutil.AssertError(t, err, tt.expectError, "SmartUpload error check")

			if tt.expectError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				testutil.AssertNotNil(t, resp, "SmartUpload response")
				if resp.Url == "" {
					t.Error("SmartUpload() URL should not be empty")
				}
				if tt.req.Filename != "" && !strings.Contains(resp.Url, tt.req.Filename) {
					t.Errorf("SmartUpload() URL should contain filename: %s", resp.Url)
				}
			}
		})
	}
}

func TestUploadService_MinioRepo(t *testing.T) {
	mockRepo := &mocks.MockMinioRepo{}
	service := NewUploadService(mockRepo, testutil.MockLogger())

	repo := service.MinioRepo()
	testutil.AssertNotNil(t, repo, "MinioRepo should not be nil")
}

func TestUploadService_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *mocks.MockMinioRepo
		testFunc    func(*UploadService) error
		expectError bool
	}{
		{
			name: "超大文件名",
			setupMock: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "http://localhost:9000/test/" + objectName, nil
					},
				}
			},
			testFunc: func(service *UploadService) error {
				longFilename := strings.Repeat("a", 255) + ".txt"
				req := &uploadv1.SimpleUploadRequest{
					Filename:    longFilename,
					FileData:    []byte("test"),
					ContentType: "text/plain",
				}
				_, err := service.SimpleUpload(context.Background(), req)
				return err
			},
			expectError: false,
		},
		{
			name: "特殊字符文件名",
			setupMock: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "http://localhost:9000/test/" + objectName, nil
					},
				}
			},
			testFunc: func(service *UploadService) error {
				req := &SmartUploadRequest{
					Filename:    "file with spaces & symbols!@#$%^&*().txt",
					FileData:    []byte("test content"),
					ContentType: "text/plain",
				}
				_, err := service.SmartUpload(context.Background(), req)
				return err
			},
			expectError: false,
		},
		{
			name: "极小文件",
			setupMock: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						if size != 1 {
							return "", errors.New("unexpected file size")
						}
						return "http://localhost:9000/test/" + objectName, nil
					},
				}
			},
			testFunc: func(service *UploadService) error {
				req := &uploadv1.SimpleUploadRequest{
					Filename:    "tiny.txt",
					FileData:    []byte("a"), // 1 byte
					ContentType: "text/plain",
				}
				_, err := service.SimpleUpload(context.Background(), req)
				return err
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUploadService(tt.setupMock(), testutil.MockLogger())
			err := tt.testFunc(service)
			testutil.AssertError(t, err, tt.expectError, tt.name)
		})
	}
}

func TestUploadService_ConcurrentUploads(t *testing.T) {
	mockRepo := &mocks.MockMinioRepo{
		SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
			// 模拟一些处理时间
			return "http://localhost:9000/test/" + objectName, nil
		},
	}

	service := NewUploadService(mockRepo, testutil.MockLogger())

	// 并发上传测试
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := &uploadv1.SimpleUploadRequest{
				Filename:    fmt.Sprintf("concurrent-file-%d.txt", id),
				FileData:    []byte(fmt.Sprintf("content-%d", id)),
				ContentType: "text/plain",
			}
			_, err := service.SimpleUpload(context.Background(), req)
			results <- err
		}(i)
	}

	// 检查所有结果
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent upload %d failed: %v", i, err)
		}
	}
}

func TestUploadService_ContextCancellation(t *testing.T) {
	mockRepo := &mocks.MockMinioRepo{
		SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
			// 检查上下文是否被取消
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "http://localhost:9000/test/" + objectName, nil
			}
		},
	}

	service := NewUploadService(mockRepo, testutil.MockLogger())

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	req := &uploadv1.SimpleUploadRequest{
		Filename:    "cancelled.txt",
		FileData:    []byte("test"),
		ContentType: "text/plain",
	}

	_, err := service.SimpleUpload(ctx, req)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

func TestUploadService_SimpleUpload_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		req           *uploadv1.SimpleUploadRequest
		mockMinioRepo func() *mocks.MockMinioRepo
		expectError   bool
		errorContains string
	}{
		{
			name: "MinIO上传失败",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "test.txt",
				FileData:    []byte("Hello, World!"),
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SimpleUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
						return "", errors.New("MinIO connection failed")
					},
				}
			},
			expectError:   true,
			errorContains: "MinIO connection failed",
		},
		{
			name: "大文件使用智能上传失败",
			req: &uploadv1.SimpleUploadRequest{
				Filename:    "large.txt",
				FileData:    make([]byte, 6*1024*1024), // 6MB，超过5MB阈值
				ContentType: "text/plain",
			},
			mockMinioRepo: func() *mocks.MockMinioRepo {
				return &mocks.MockMinioRepo{
					SmartUploadFunc: func(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string, progressCb func(uploaded, total int64)) (string, error) {
						return "", errors.New("smart upload failed")
					},
				}
			},
			expectError:   true,
			errorContains: "smart upload failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUploadService(tt.mockMinioRepo(), testutil.MockLogger())

			resp, err := service.SimpleUpload(context.Background(), tt.req)

			testutil.AssertError(t, err, tt.expectError, "SimpleUpload error check")

			if tt.expectError {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				testutil.AssertNotNil(t, resp, "SimpleUpload response")
				if resp.Url == "" {
					t.Error("SimpleUpload() URL should not be empty")
				}
				if tt.req.Filename != "" && !strings.Contains(resp.Url, tt.req.Filename) {
					t.Errorf("SimpleUpload() URL should contain filename: %s", resp.Url)
				}
			}
		})
	}
}
