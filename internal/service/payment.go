package service

import (
	"context"

	pb "anjuke/api/payment/v1"
	"anjuke/internal/biz"
)

type PaymentService struct {
	pb.UnimplementedPaymentServer
	uc *biz.PaymentUsecase
}

func NewPaymentService(uc *biz.PaymentUsecase) *PaymentService {
	return &PaymentService{uc: uc}
}

func (s *PaymentService) AlipayPay(ctx context.Context, req *pb.AlipayPayRequest) (*pb.AlipayPayReply, error) {
	payUrl, err := s.uc.AlipayPay(ctx, req.OrderId, req.Amount)
	if err != nil {
		return nil, err
	}
	return &pb.AlipayPayReply{
		PayUrl: payUrl,
	}, nil
}

func (s *PaymentService) AlipayQuery(ctx context.Context, req *pb.AlipayQueryRequest) (*pb.AlipayQueryReply, error) {
	status, tradeNo, payTime, err := s.uc.AlipayQuery(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &pb.AlipayQueryReply{
		OrderId: req.OrderId,
		Status:  status,
		TradeNo: tradeNo,
		PayTime: payTime,
	}, nil
}
