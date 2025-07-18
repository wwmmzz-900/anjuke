package biz

import "github.com/go-kratos/kratos/v2/log"

type PointsRepo interface {
}
type PointsUsecase struct {
	repo PointsRepo
	log  *log.Helper
}

func NewPointsUsecase(repo PointsRepo, logger log.Logger) *PointsUsecase {
	return &PointsUsecase{repo: repo, log: log.NewHelper(logger)}
}
