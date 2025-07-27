package service

import (
	uploadv1 "anjuke/server/api/upload/v1"
	"anjuke/server/internal/domain"
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type UploadService struct {
	minioRepo domain.MinioRepo
	log       *log.Helper
}

func NewUploadService(minioRepo domain.MinioRepo, logger log.Logger) *UploadService {
	return &UploadService{
		minioRepo: minioRepo,
		log:       log.NewHelper(logger),
	}
}

// MinioRepo 返回底层的MinioRepo实例，用于直接访问
func (s *UploadService) MinioRepo() domain.MinioRepo {
	return s.minioRepo
}

type SmartUploadRequest struct {
	Filename    string
	ContentType string
	FileData    []byte
}

type SmartUploadReply struct {
	Url string
}

// SmartUpload 统一上传接口
func (s *UploadService) SmartUpload(ctx context.Context, req *SmartUploadRequest) (*SmartUploadReply, error) {
	reader := bytes.NewReader(req.FileData)
	size := int64(len(req.FileData))
	url, err := s.minioRepo.SmartUpload(ctx, req.Filename, reader, size, req.ContentType, nil)
	if err != nil {
		return nil, err
	}
	return &SmartUploadReply{Url: url}, nil
}

// 实现uploadv1.UploadServiceHTTPServer接口
func (s *UploadService) SimpleUpload(ctx context.Context, req *uploadv1.SimpleUploadRequest) (*uploadv1.SimpleUploadReply, error) {
	// 简化日志输出
	s.log.Infof("接收文件: %s (%d KB)", req.Filename, len(req.FileData)/1024)

	reader := bytes.NewReader(req.FileData)
	size := int64(len(req.FileData))

	// 设置上传超时
	uploadCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	const MultipartThreshold = 5 * 1024 * 1024 // 5MB
	var url string
	var err error
	if size >= MultipartThreshold {
		url, err = s.minioRepo.SmartUpload(uploadCtx, req.Filename, reader, size, req.ContentType, nil)
	} else {
		url, err = s.minioRepo.SimpleUpload(uploadCtx, req.Filename, reader, size, req.ContentType)
	}
	if err != nil {
		s.log.Errorf("上传失败: %v", err)
		return nil, fmt.Errorf("文件上传失败: %v", err)
	}

	s.log.Infof("处理完成: %s", req.Filename)
	return &uploadv1.SimpleUploadReply{Url: url}, nil
}
