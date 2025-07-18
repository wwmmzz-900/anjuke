package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/wwmmzz-900/anjuke/internal/biz"
)

type TransactionRepo struct {
	data *Data
	log  *log.Helper
}

func NewTransactionRepo(data *Data, logger log.Logger) biz.TransactionRepo {
	return &TransactionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
