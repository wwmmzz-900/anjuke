package data

import (
	"anjuke/internal/biz"
	"context"
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
func (r *HouseRepo) CreateHouse(ctx context.Context, house *biz.House) (int64, error) {
	// 假设你用 GORM
	result := r.data.db.Create(house)
	if result.Error != nil {
		return 0, result.Error
	}
	return house.HouseID, nil
}
