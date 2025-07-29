package biz

import (
	"context"
)

type AlipayBizRepo interface {
	AlipayPay(ctx context.Context, orderId string, amount string) (string, error)
	AlipayQuery(ctx context.Context, orderId string) (status, tradeNo, payTime string, err error)
}

type PaymentUsecase struct {
	repo AlipayBizRepo
}

func NewPaymentUsecase(repo AlipayBizRepo) *PaymentUsecase {
	return &PaymentUsecase{repo: repo}
}

func (uc *PaymentUsecase) AlipayPay(ctx context.Context, orderId string, amount string) (string, error) {
	return uc.repo.AlipayPay(ctx, orderId, amount)
}

func (uc *PaymentUsecase) AlipayQuery(ctx context.Context, orderId string) (string, string, string, error) {
	return uc.repo.AlipayQuery(ctx, orderId)
}
