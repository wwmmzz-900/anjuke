package data

import (
	"anjuke/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
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
