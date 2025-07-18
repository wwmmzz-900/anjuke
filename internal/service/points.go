package service

import (
	"context"
	"github.com/wwmmzz-900/anjuke/internal/biz"

	pb "github.com/wwmmzz-900/anjuke/api/points/v5"
)

type PointsService struct {
	pb.UnimplementedPointsServer
	v5uc *biz.PointsUsecase
}

func NewPointsService(v5uc *biz.PointsUsecase) *PointsService {
	return &PointsService{
		v5uc: v5uc,
	}
}

func (s *PointsService) CreatePoints(ctx context.Context, req *pb.CreatePointsRequest) (*pb.CreatePointsReply, error) {
	return &pb.CreatePointsReply{}, nil
}
