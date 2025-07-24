package biz

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wwmmzz-900/anjuke/api/house/v3"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

type House struct {
	HouseID     int64
	Title       string
	Description string
	Price       float64
	Area        float64
	Layout      string
	ImageURL    string
}

type HouseRepo interface {
	GetUserPricePreference(ctx context.Context, userID int64) (float64, float64, error)
	GetPersonalRecommendList(ctx context.Context, minPrice, maxPrice float64, page, pageSize int) ([]*House, int, error)
	GetRecommendList(ctx context.Context, page, pageSize int) ([]*House, int, error)
	CreateReservation(ctx context.Context, reservation *model.HouseReservation) error
	HasReservation(ctx context.Context, userID, houseID int64) (bool, error)
}

type HouseUsecase struct {
	repo HouseRepo
}

func NewHouseUsecase(repo HouseRepo) *HouseUsecase {
	return &HouseUsecase{repo: repo}
}

// 个性化推荐
func (uc *HouseUsecase) PersonalRecommendList(ctx context.Context, userID int64, page, pageSize int) ([]*House, int, error) {
	// 1. 获取用户偏好的价格区间
	minPrice, maxPrice, err := uc.repo.GetUserPricePreference(ctx, userID)
	if err != nil || minPrice == 0 && maxPrice == 0 {
		// 没有浏览记录时，给一个默认区间
		minPrice, maxPrice = 800, 1500
	}
	// 2. 查询推荐房源
	return uc.repo.GetPersonalRecommendList(ctx, minPrice, maxPrice, page, pageSize)
}

// 普通推荐列表
func (uc *HouseUsecase) RecommendList(ctx context.Context, page, pageSize int) ([]*House, int, error) {
	// 直接查询推荐房源
	return uc.repo.GetRecommendList(ctx, page, pageSize)
}

// 预约看房业务逻辑
func (uc *HouseUsecase) ReserveHouse(ctx context.Context, req *pb.ReserveHouseRequest) error {
	// 1. 校验是否已预约
	has, err := uc.repo.HasReservation(ctx, req.UserId, req.HouseId)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("您已预约过该房源")
	}

	// 2. 构造预约记录
	reservation := &model.HouseReservation{
		LandlordID:  req.LandlordId,
		UserID:      req.UserId,
		UserName:    req.UserName,
		HouseID:     req.HouseId,
		HouseTitle:  req.HouseTitle,
		ReserveTime: req.ReserveTime,
		CreatedAt:   time.Now().Unix(),
	}

	// 3. 保存预约
	return uc.repo.CreateReservation(ctx, reservation)
}
