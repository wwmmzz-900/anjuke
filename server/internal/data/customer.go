package data

import (
	"github.com/go-kratos/kratos/v2/log"
)

type CustomerRepo struct {
	data *Data
	log  *log.Helper
}

func NewCustomerRepo(data *Data, logger log.Logger) *CustomerRepo {
	return &CustomerRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
