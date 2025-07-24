package service

import (
	pb "anjuke/api/blacklist/v1"
	"anjuke/internal/biz"
	"context"

	"github.com/go-kratos/kratos/v2/errors"
)

type BlacklistService struct {
	pb.UnimplementedBlacklistServer
	uc *biz.BlacklistUsecase
}

// NewBlacklistService 创建黑名单服务
func NewBlacklistService(uc *biz.BlacklistUsecase) *BlacklistService {
	return &BlacklistService{uc: uc}
}

// AddToBlacklist 添加用户到黑名单
func (s *BlacklistService) AddToBlacklist(ctx context.Context, req *pb.AddToBlacklistRequest) (*pb.AddToBlacklistReply, error) {
	if req.UserId <= 0 {
		return nil, errors.BadRequest("INVALID_PARAMETER", "用户ID不能为空")
	}

	blacklist, err := s.uc.AddToBlacklist(ctx, req.UserId, req.Reason)
	if err != nil {
		if err == biz.ErrUserNotFound {
			return nil, errors.NotFound("USER_NOT_FOUND", "用户不存在")
		}
		if err == biz.ErrUserAlreadyBlacklisted {
			return nil, errors.BadRequest("USER_ALREADY_BLACKLISTED", "用户已在黑名单中")
		}
		return nil, errors.InternalServer("INTERNAL_ERROR", err.Error())
	}

	return &pb.AddToBlacklistReply{
		Success:     true,
		Message:     "添加黑名单成功",
		BlacklistId: blacklist.ID,
	}, nil
}

// RemoveFromBlacklist 从黑名单移除用户
func (s *BlacklistService) RemoveFromBlacklist(ctx context.Context, req *pb.RemoveFromBlacklistRequest) (*pb.RemoveFromBlacklistReply, error) {
	if req.UserId <= 0 {
		return nil, errors.BadRequest("INVALID_PARAMETER", "用户ID不能为空")
	}

	err := s.uc.RemoveFromBlacklist(ctx, req.UserId)
	if err != nil {
		if err == biz.ErrUserNotInBlacklist {
			return nil, errors.NotFound("USER_NOT_IN_BLACKLIST", "用户不在黑名单中")
		}
		return nil, errors.InternalServer("INTERNAL_ERROR", err.Error())
	}

	return &pb.RemoveFromBlacklistReply{
		Success: true,
		Message: "移除黑名单成功",
	}, nil
}

// CheckBlacklist 检查用户是否在黑名单中
func (s *BlacklistService) CheckBlacklist(ctx context.Context, req *pb.CheckBlacklistRequest) (*pb.CheckBlacklistReply, error) {
	if req.UserId <= 0 {
		return nil, errors.BadRequest("INVALID_PARAMETER", "用户ID不能为空")
	}

	blacklist, err := s.uc.CheckBlacklist(ctx, req.UserId)
	if err != nil {
		return nil, errors.InternalServer("INTERNAL_ERROR", err.Error())
	}

	if blacklist == nil {
		return &pb.CheckBlacklistReply{
			IsBlacklisted: false,
		}, nil
	}

	return &pb.CheckBlacklistReply{
		IsBlacklisted: true,
		Reason:        blacklist.Reason,
		CreatedAt:     blacklist.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetBlacklistList 获取黑名单列表
func (s *BlacklistService) GetBlacklistList(ctx context.Context, req *pb.GetBlacklistListRequest) (*pb.GetBlacklistListReply, error) {
	items, total, err := s.uc.GetBlacklistList(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, errors.InternalServer("INTERNAL_ERROR", err.Error())
	}

	var pbItems []*pb.BlacklistItem
	for _, item := range items {
		pbItems = append(pbItems, &pb.BlacklistItem{
			Id:        item.ID,
			UserId:    item.UserID,
			Reason:    item.Reason,
			CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
			UserName:  item.UserName,
			Phone:     item.Phone,
		})
	}

	return &pb.GetBlacklistListReply{
		Items:    pbItems,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
