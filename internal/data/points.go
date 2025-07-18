package data

import (
	"anjuke/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type PointsRepo struct {
	data *Data
	log  *log.Helper
}

func NewPointsRepo(data *Data, logger log.Logger) biz.PointsRepo {
	return &PointsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
