package service

import (
	"context"

	commonv1 "anjuke/server/api/common/v1"
	v1 "anjuke/server/api/helloworld/v1"
	"anjuke/server/internal/biz"
	"anjuke/server/internal/domain"

	"google.golang.org/protobuf/types/known/anypb"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer
	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*commonv1.BaseResponse, error) {
	g, err := s.uc.CreateGreeter(ctx, &domain.Greeter{Hello: in.Name})
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &v1.HelloData{
		Message: "Hello " + g.Hello,
		Name:    in.Name,
	}

	anyData, err := anypb.New(data)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "数据序列化失败",
			Data: nil,
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "问候成功",
		Data: anyData,
	}, nil
}
