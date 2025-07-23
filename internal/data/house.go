package data

import (
	"anjuke/internal/biz"
	"context"
	"encoding/json"
	"fmt"
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
	// 1. 写入数据库
	result := r.data.db.Create(house)
	if result.Error != nil {
		return 0, result.Error
	}

	// 2. 更新 Redis 缓存
	// 更新详情缓存
	detailKey := fmt.Sprintf("house:detail:%d", house.HouseID)
	// 使用 JSON 格式序列化房屋信息curl -sSf https://tinygo.org/install.sh | sh
	houseJSON, err := json.Marshal(house)
	if err != nil {
		r.log.Errorf("Failed to marshal house data: %v", err)
	} else {
		err = r.data.rdb.Set(ctx, detailKey, houseJSON, 0).Err()
		if err != nil {
			r.log.Errorf("Failed to set house detail cache: %v", err)
		}
	}

	// 更新列表缓存（如有，key 设计可根据实际情况调整）

	listKey := "house:list:all"
	err = r.data.rdb.LPush(ctx, listKey, house.HouseID).Err()
	if err != nil {
		r.log.Errorf("Failed to update house list cache: %v", err)
	}

	return house.HouseID, nil
	/*//1 写入数据库
	// 插入数据到数据库
	result := r.data.db.Create(house)
	if result.Error != nil {
		return 0, result.Error
	}

	// 2. 清理相关 Redis 缓存
	// 清理详情缓存
	detailKey := fmt.Sprintf("house:detail:%d", house.HouseID)
	r.data.rdb.Del(ctx, detailKey)

	// 清理列表缓存（如有，key设计可根据实际情况调整）
	// 例如：r.data.rdb.Del(ctx, "house:list:all")
	// 如果有分页、区域等条件，可以用通配符批量删除
	// keys, _ := r.data.rdb.Keys(ctx, "house:list:*").Result()
	// for _, key := range keys {
	//     r.data.rdb.Del(ctx, key)
	// }
	return house.HouseID, nil*/
}
