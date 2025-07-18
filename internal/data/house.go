package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/wwmmzz-900/anjuke/internal/biz"
)

type HouseRepo struct {
	data *Data
	log  *log.Helper
}

func NewHouseRepo(data *Data, logger log.Logger) biz.HouseRepo {
	return &HouseRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
