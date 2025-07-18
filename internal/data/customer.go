package data

import (
	"anjuke/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
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
