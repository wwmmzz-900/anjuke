package service

import (
	"anjuke/internal/biz"
	"context"

	pb "anjuke/api/transaction/v4"
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

func (s *TransactionService) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionReply, error) {
	return &pb.CreateTransactionReply{}, nil
}
