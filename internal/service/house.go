package service

import (
	"anjuke/internal/biz"
	"context"

	pb "anjuke/api/house/v3"
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
	// 1. 构造业务对象
	house := &biz.House{
		Title:                   req.Title,
		Description:             req.Description,
		LandlordID:              req.LandlordId,
		Address:                 req.Address,
		RegionID:                req.RegionId,
		CommunityID:             req.CommunityId,
		Price:                   req.Price,
		Area:                    req.Area,
		Layout:                  req.Layout,
		Floor:                   req.Floor,
		OwnershipCertificateUrl: req.OwnershipCertificateUrl,
		Orientation:             req.Orientation,
		Decoration:              req.Decoration,
		Facilities:              req.Facilities,
	}
	// 2. 调用 usecase 层
	houseID, err := s.v3uc.CreateHouse(ctx, house)
	if err != nil {
		return nil, err
	}
	// 3. 返回
	return &pb.CreateHouseReply{
		HouseId: houseID,
		Message: "发布成功",
	}, nil
}
