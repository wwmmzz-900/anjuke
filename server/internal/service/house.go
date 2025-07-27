package service

import (
	"context"

	commonv1 "anjuke/server/api/common/v1"
	pb "anjuke/server/api/house/v3"
	"anjuke/server/internal/biz"

	"google.golang.org/protobuf/types/known/anypb"
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

func (s *HouseService) CreateHouse(ctx context.Context, req *pb.CreateHouseRequest) (*commonv1.BaseResponse, error) {
	// TODO: 实现具体的业务逻辑
	// 这里只是示例实现

	// 构建响应数据
	data := &pb.CreateHouseData{
		HouseId:   "house_" + "123456", // 示例ID
		Title:     req.Title,
		Address:   req.Address,
		Price:     req.Price,
		Type:      req.Type,
		Area:      req.Area,
		CreatedAt: "2024-01-01 12:00:00",
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
		Msg:  "房屋创建成功",
		Data: anyData,
	}, nil
}
