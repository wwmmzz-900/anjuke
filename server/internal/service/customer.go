package service

import (
	"context"

	commonv1 "anjuke/server/api/common/v1"
	pb "anjuke/server/api/customer/v6"
	"anjuke/server/internal/biz"

	"google.golang.org/protobuf/types/known/anypb"
)

type CustomerService struct {
	pb.UnimplementedCustomerServer
	v6uc *biz.CustomerUsecase
}

func NewCustomerService(v6uc *biz.CustomerUsecase) *CustomerService {
	return &CustomerService{
		v6uc: v6uc,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*commonv1.BaseResponse, error) {
	// TODO: 实现具体的业务逻辑
	// 这里只是示例实现

	// 构建响应数据
	data := &pb.CreateCustomerData{
		CustomerId: "cust_" + "123456", // 示例ID
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		Address:    req.Address,
		CreatedAt:  "2024-01-01 12:00:00",
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
		Msg:  "客户创建成功",
		Data: anyData,
	}, nil
}
