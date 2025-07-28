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
	return &pb.CreateHouseReply{}, nil
}

// 点赞房源
func (s *HouseService) LikeProperty(ctx context.Context, req *pb.LikePropertyRequest) (*pb.LikePropertyReply, error) {
	if req.PropertyId <= 0 || req.UserId <= 0 {
		return &pb.LikePropertyReply{Success: false, Message: "无效的参数"}, nil
	}

	err := s.v3uc.LikeProperty(ctx, req.PropertyId, req.UserId)
	if err != nil {
		return &pb.LikePropertyReply{Success: false, Message: err.Error()}, nil
	}

	count, _ := s.v3uc.GetPropertyLikeCount(ctx, req.PropertyId)
	return &pb.LikePropertyReply{
		Success:   true,
		Message:   "点赞成功",
		LikeCount: count,
	}, nil
}

// 取消点赞房源
func (s *HouseService) UnlikeProperty(ctx context.Context, req *pb.UnlikePropertyRequest) (*pb.UnlikePropertyReply, error) {
	if req.PropertyId <= 0 || req.UserId <= 0 {
		return &pb.UnlikePropertyReply{Success: false, Message: "无效的参数"}, nil
	}

	err := s.v3uc.UnlikeProperty(ctx, req.PropertyId, req.UserId)
	if err != nil {
		return &pb.UnlikePropertyReply{Success: false, Message: err.Error()}, nil
	}

	count, _ := s.v3uc.GetPropertyLikeCount(ctx, req.PropertyId)
	return &pb.UnlikePropertyReply{
		Success:   true,
		Message:   "取消点赞成功",
		LikeCount: count,
	}, nil
}

// 检查用户是否已点赞
func (s *HouseService) IsPropertyLiked(ctx context.Context, req *pb.IsPropertyLikedRequest) (*pb.IsPropertyLikedReply, error) {
	if req.PropertyId <= 0 || req.UserId <= 0 {
		return &pb.IsPropertyLikedReply{Liked: false, LikeCount: 0}, nil
	}

	liked, err := s.v3uc.IsPropertyLiked(ctx, req.PropertyId, req.UserId)
	if err != nil {
		return &pb.IsPropertyLikedReply{Liked: false, LikeCount: 0}, err
	}

	count, _ := s.v3uc.GetPropertyLikeCount(ctx, req.PropertyId)
	return &pb.IsPropertyLikedReply{
		Liked:     liked,
		LikeCount: count,
	}, nil
}

// 获取房源点赞数
func (s *HouseService) GetPropertyLikeCount(ctx context.Context, req *pb.GetPropertyLikeCountRequest) (*pb.GetPropertyLikeCountReply, error) {
	if req.PropertyId <= 0 {
		return &pb.GetPropertyLikeCountReply{Count: 0}, nil
	}

	count, err := s.v3uc.GetPropertyLikeCount(ctx, req.PropertyId)
	if err != nil {
		return &pb.GetPropertyLikeCountReply{Count: 0}, err
	}

	return &pb.GetPropertyLikeCountReply{Count: count}, nil
}

// 获取用户点赞列表
func (s *HouseService) GetUserLikeList(ctx context.Context, req *pb.GetUserLikeListRequest) (*pb.GetUserLikeListReply, error) {
	if req.UserId <= 0 {
		return &pb.GetUserLikeListReply{}, nil
	}

	// 处理分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	house, err := s.v3uc.GetHouseDetail(ctx, req.UserId)
	if err != nil {
		return &pb.GetUserLikeListReply{}, err
	}

	// 调用业务层方法
	likes, totalCount, err := s.v3uc.GetUserLikeList(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return &pb.GetUserLikeListReply{}, err
	}

	// 转换为响应格式
	var properties []*pb.PropertyLike
	for _, like := range likes {
		properties = append(properties, &pb.PropertyLike{
			Id:            int64(like.ID),
			PropertyId:    like.PropertyId,
			UserId:        like.UserId,
			CreatedAt:     like.CreatedAt.Unix(),
			PropertyTitle: house.Title,
			PropertyImage: house.OwnershipCertificateUrl,
			PropertyPrice: house.Price,
		})
	}

	return &pb.GetUserLikeListReply{
		Properties: properties,
		TotalCount: totalCount,
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}
