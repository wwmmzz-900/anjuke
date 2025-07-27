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

func TestUploadService_EdgeCasesExtended(t *testing.T) {
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

func TestUploadService_ConcurrentUploadsExtended(t *testing.T) {
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

func TestUploadService_ContextCancellationExtended(t *testing.T) {
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
