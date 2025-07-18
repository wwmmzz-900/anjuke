package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/wwmmzz-900/anjuke/internal/biz"
)

type CustomerRepo struct {
	data *Data
	log  *log.Helper
}

func NewCustomerRepo(data *Data, logger log.Logger) biz.CustomerRepo {
	return &CustomerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
