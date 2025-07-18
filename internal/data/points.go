package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/wwmmzz-900/anjuke/internal/biz"
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
