package data

import (
	"github.com/go-kratos/kratos/v2/log"
)

type HouseRepo struct {
	data *Data
	log  *log.Helper
}

func NewHouseRepo(data *Data, logger log.Logger) *HouseRepo {
	return &HouseRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
