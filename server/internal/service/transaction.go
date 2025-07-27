package service

import (
	"context"

	commonv1 "anjuke/server/api/common/v1"
	pb "anjuke/server/api/transaction/v4"
	"anjuke/server/internal/biz"

	"google.golang.org/protobuf/types/known/anypb"
)

type TransactionService struct {
	pb.UnimplementedTransactionServer
	v4uc *biz.TransactionUsecase
}

func NewTransactionService(v4uc *biz.TransactionUsecase) *TransactionService {
	return &TransactionService{
		v4uc: v4uc,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*commonv1.BaseResponse, error) {
	// TODO: 实现具体的业务逻辑
	// 这里只是示例实现

	// 构建响应数据
	data := &pb.CreateTransactionData{
		TransactionId: "txn_" + "123456", // 示例ID
		UserId:        req.UserId,
		Amount:        req.Amount,
		Type:          req.Type,
		Status:        "created",
		CreatedAt:     "2024-01-01 12:00:00",
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
		Msg:  "交易创建成功",
		Data: anyData,
	}, nil
}
