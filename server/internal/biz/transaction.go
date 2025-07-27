package biz

import "github.com/go-kratos/kratos/v2/log"

type TransactionRepo interface {
}
type TransactionUsecase struct {
	repo TransactionRepo
	log  *log.Helper
}

func NewTransactionUsecase(repo TransactionRepo, logger log.Logger) *TransactionUsecase {
	return &TransactionUsecase{repo: repo, log: log.NewHelper(logger)}
}
