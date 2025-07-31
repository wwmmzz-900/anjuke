// Package service 提供了积分模块的对外暴露服务
package service

import (
	"context"
	"fmt"
	"log"

	commonv1 "anjuke/server/api/common/v1"
	pb "anjuke/server/api/points/v5"
	"anjuke/server/internal/biz"
	"anjuke/server/internal/domain"

	"google.golang.org/protobuf/types/known/anypb"
)

// PointsUsecaseInterface 定义积分用例接口，用于测试时的mock
type PointsUsecaseInterface interface {
	GetUserPoints(ctx context.Context, userID uint64) (*domain.UserPoints, error)
	GetPointsHistory(ctx context.Context, userID uint64, page, pageSize int32, pointsType string) ([]*domain.PointsRecord, int32, error)
	CheckIn(ctx context.Context, userID uint64) (*domain.CheckInResult, error)
	EarnPointsByConsume(ctx context.Context, userID uint64, orderID string, amount int64) (*domain.EarnResult, error)
	UsePoints(ctx context.Context, userID uint64, points int64, orderID, description string) (*domain.UseResult, error)
}

// PointsService 实现了积分模块的 gRPC 和 HTTP 服务
type PointsService struct {
	pb.UnimplementedPointsServer
	uc PointsUsecaseInterface
}

// NewPointsService 创建积分服务
func NewPointsService(uc *biz.PointsUsecase) *PointsService {
	return &PointsService{
		uc: uc,
	}
}

// NewPointsServiceWithInterface 用于测试时创建服务
func NewPointsServiceWithInterface(uc PointsUsecaseInterface) *PointsService {
	return &PointsService{
		uc: uc,
	}
}

// GetUserPoints 查询用户积分余额
func (s *PointsService) GetUserPoints(ctx context.Context, req *pb.GetUserPointsRequest) (*commonv1.BaseResponse, error) {
	log.Printf("GetUserPoints参数: UserID=%d", req.UserId)

	if req.UserId == 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: user_id",
			Data: nil,
		}, nil
	}

	userPoints, err := s.uc.GetUserPoints(ctx, req.UserId)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &pb.GetUserPointsData{
		UserId:      userPoints.UserID,
		TotalPoints: userPoints.TotalPoints,
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
		Msg:  "查询成功",
		Data: anyData,
	}, nil
}

// GetPointsHistory 查询积分明细记录
func (s *PointsService) GetPointsHistory(ctx context.Context, req *pb.GetPointsHistoryRequest) (*commonv1.BaseResponse, error) {
	log.Printf("GetPointsHistory参数: UserID=%d, Page=%d, PageSize=%d, Type=%s", req.UserId, req.Page, req.PageSize, req.Type)

	if req.UserId == 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: user_id",
			Data: nil,
		}, nil
	}

	records, total, err := s.uc.GetPointsHistory(ctx, req.UserId, req.Page, req.PageSize, req.Type)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 转换为 protobuf 格式
	pbRecords := make([]*pb.PointsRecord, len(records))
	for i, record := range records {
		pbRecords[i] = &pb.PointsRecord{
			Id:          record.ID,
			UserId:      record.UserID,
			Type:        record.Type,
			Points:      record.Points,
			Description: record.Description,
			OrderId:     record.OrderID,
			Amount:      record.Amount,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// 计算分页信息
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20 // 默认每页20条
	}
	totalPages := int32((total + pageSize - 1) / pageSize)

	// 构建响应数据
	data := &pb.GetPointsHistoryData{
		Records: pbRecords,
		PageInfo: &commonv1.PageInfo{
			Page:       req.Page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
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
		Msg:  "查询成功",
		Data: anyData,
	}, nil
}

// CheckIn 签到获取积分
func (s *PointsService) CheckIn(ctx context.Context, req *pb.CheckInRequest) (*commonv1.BaseResponse, error) {
	log.Printf("CheckIn参数: UserID=%d", req.UserId)

	if req.UserId == 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: user_id",
			Data: nil,
		}, nil
	}

	result, err := s.uc.CheckIn(ctx, req.UserId)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &pb.CheckInData{
		PointsEarned:    result.PointsEarned,
		TotalPoints:     result.TotalPoints,
		ConsecutiveDays: result.ConsecutiveDays,
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
		Msg:  "签到成功",
		Data: anyData,
	}, nil
}

// EarnPointsByConsume 消费获取积分
func (s *PointsService) EarnPointsByConsume(ctx context.Context, req *pb.EarnPointsByConsumeRequest) (*commonv1.BaseResponse, error) {
	log.Printf("EarnPointsByConsume参数: UserID=%d, OrderID=%s, Amount=%d", req.UserId, req.OrderId, req.Amount)

	if req.UserId == 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: user_id",
			Data: nil,
		}, nil
	}
	if req.OrderId == "" {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: order_id",
			Data: nil,
		}, nil
	}
	if req.Amount <= 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "消费金额必须大于0",
			Data: nil,
		}, nil
	}

	result, err := s.uc.EarnPointsByConsume(ctx, req.UserId, req.OrderId, req.Amount)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &pb.EarnPointsByConsumeData{
		PointsEarned: result.PointsEarned,
		TotalPoints:  result.TotalPoints,
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
		Msg:  fmt.Sprintf("消费获得积分成功，获得%d积分", result.PointsEarned),
		Data: anyData,
	}, nil
}

// UsePoints 使用积分抵扣
func (s *PointsService) UsePoints(ctx context.Context, req *pb.UsePointsRequest) (*commonv1.BaseResponse, error) {
	log.Printf("UsePoints参数: UserID=%d, Points=%d, OrderID=%s, Description=%s", req.UserId, req.Points, req.OrderId, req.Description)

	if req.UserId == 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "缺少参数: user_id",
			Data: nil,
		}, nil
	}
	if req.Points <= 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  "使用积分数量必须大于0",
			Data: nil,
		}, nil
	}
	if req.Points%domain.UsePointsRate != 0 {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  fmt.Sprintf("积分数量必须是%d的倍数", domain.UsePointsRate),
			Data: nil,
		}, nil
	}

	result, err := s.uc.UsePoints(ctx, req.UserId, req.Points, req.OrderId, req.Description)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 1,
			Msg:  err.Error(),
			Data: nil,
		}, nil
	}

	// 构建响应数据
	data := &pb.UsePointsData{
		PointsUsed:     result.PointsUsed,
		AmountDeducted: result.AmountDeducted,
		TotalPoints:    result.TotalPoints,
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
		Msg:  fmt.Sprintf("积分使用成功，抵扣%.2f元", float64(result.AmountDeducted)/100),
		Data: anyData,
	}, nil
}
