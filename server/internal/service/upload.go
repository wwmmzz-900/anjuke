package service

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

type UploadService struct {
	log *log.Helper
}

func NewUploadService(logger log.Logger) *UploadService {
	return &UploadService{
		log: log.NewHelper(logger),
	}
}

// SimpleUpload 简化的上传接口，主要用于兼容性
func (s *UploadService) SimpleUpload(ctx context.Context, filename string, data []byte) (string, error) {
	s.log.Infof("接收文件: %s (%d KB)", filename, len(data)/1024)

	// 这里可以添加一些基本的验证逻辑
	if len(data) == 0 {
		return "", fmt.Errorf("文件数据为空")
	}

	// 返回一个占位符URL，实际的上传逻辑已经移到HTTP处理程序中
	return fmt.Sprintf("placeholder://%s", filename), nil
}
