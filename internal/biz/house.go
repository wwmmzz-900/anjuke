package biz

import "github.com/go-kratos/kratos/v2/log"

type HouseRepo interface {
}
type HouseUsecase struct {
	repo HouseRepo
	log  *log.Helper
}

func NewHouseUsecase(repo HouseRepo, logger log.Logger) *HouseUsecase {
	return &HouseUsecase{repo: repo, log: log.NewHelper(logger)}
}
