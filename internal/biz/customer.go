package biz

import "github.com/go-kratos/kratos/v2/log"

type CustomerRepo interface {
}
type CustomerUsecase struct {
	repo CustomerRepo
	log  *log.Helper
}

func NewCustomerUsecase(repo CustomerRepo, logger log.Logger) *CustomerUsecase {
	return &CustomerUsecase{repo: repo, log: log.NewHelper(logger)}
}
