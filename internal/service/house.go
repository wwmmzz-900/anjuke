package service

import (
	"context"
	"github.com/wwmmzz-900/anjuke/internal/biz"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"
)

type HouseService struct {
	pb.UnimplementedHouseServer
	v3uc *biz.HouseUsecase
}

func NewHouseService(v3uc *biz.HouseUsecase) *HouseService {
	return &HouseService{
		v3uc: v3uc,
	}
}

func (s *HouseService) CreateHouse(ctx context.Context, req *pb.CreateHouseRequest) (*pb.CreateHouseReply, error) {
	return &pb.CreateHouseReply{}, nil
}
