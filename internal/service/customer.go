package service

import (
	"anjuke/internal/biz"
	"context"

	pb "anjuke/api/customer/v6"
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

func (s *CustomerService) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerReply, error) {
	return &pb.CreateCustomerReply{}, nil
}
