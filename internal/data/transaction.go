package data

import (
	"anjuke/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
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
