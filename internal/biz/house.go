package biz

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type PropertyLike struct {
	gorm.Model
	PropertyId int64 `json:"property_id"` // 房源ID
	UserId     int64 `json:"user_id"`     // 用户ID
}

func (*PropertyLike) TableName() string {
	return "property_likes"
}

type House struct {
	HouseId                 int64     `json:"house_id"`                  // 房源ID
	Title                   string    `json:"title"`                     // 房源标题
	Description             string    `json:"description"`               // 房源描述
	LandlordId              int64     `json:"landlord_id"`               // 发布人ID
	Address                 string    `json:"address"`                   // 详细地址
	RegionId                int64     `json:"region_id"`                 // 区域/小区ID
	CommunityId             int64     `json:"community_id"`              // 小区ID
	Price                   float64   `json:"price"`                     // 价格
	Area                    float32   `json:"area"`                      // 面积
	Layout                  string    `json:"layout"`                    // 户型
	Floor                   string    `json:"floor"`                     // 楼层
	OwnershipCertificateUrl string    `json:"ownership_certificate_url"` // 产权证明图片
	Orientation             string    `json:"orientation"`               // 朝向
	Decoration              string    `json:"decoration"`                // 装修
	Facilities              string    `json:"facilities"`                // 配套设施（逗号分隔）
	Status                  string    `json:"status"`                    // 状态（
	CreatedAt               time.Time `json:"created_at"`                // 发布时间
	UpdatedAt               time.Time `json:"updated_at"`                // 更新时间
	DeletedAt               time.Time `json:"deleted_at"`                // 删除时间
}

func (*House) TableName() string {
	return "house"
}

type HouseRepo interface {
	// 点赞房源
	LikeProperty(ctx context.Context, propertyId, userId int64) error
	// 取消点赞房源
	UnlikeProperty(ctx context.Context, propertyId, userId int64) error
	// 检查用户是否已点赞
	IsPropertyLiked(ctx context.Context, propertyId, userId int64) (bool, error)
	// 获取房源点赞数
	GetPropertyLikeCount(ctx context.Context, propertyId int64) (int64, error)
	// 获取用户点赞列表
	GetUserLikeList(ctx context.Context, userId int64, page, pageSize int) ([]*PropertyLike, int64, error)
	// 获取房源详情
	GetHouseDetail(ctx context.Context, houseId int64) (*House, error)
}

type HouseUsecase struct {
	repo HouseRepo
	log  *log.Helper
}

func NewHouseUsecase(repo HouseRepo, logger log.Logger) *HouseUsecase {
	return &HouseUsecase{repo: repo, log: log.NewHelper(logger)}
}

// LikeProperty 点赞房源
func (uc *HouseUsecase) LikeProperty(ctx context.Context, propertyId, userId int64) error {
	if propertyId <= 0 || userId <= 0 {
		uc.log.Errorf("invalid parameters: propertyId=%d, userId=%d", propertyId, userId)
		return errors.New("invalid parameters")
	}

	err := uc.repo.LikeProperty(ctx, propertyId, userId)
	if err != nil {
		uc.log.Errorf("failed to like property: %v", err)
		return err
	}
	return nil
}

// UnlikeProperty 取消点赞房源
func (uc *HouseUsecase) UnlikeProperty(ctx context.Context, propertyId, userId int64) error {
	if propertyId <= 0 || userId <= 0 {
		uc.log.Errorf("invalid parameters: propertyId=%d, userId=%d", propertyId, userId)
		return errors.New("invalid parameters")
	}

	err := uc.repo.UnlikeProperty(ctx, propertyId, userId)
	if err != nil {
		uc.log.Errorf("failed to unlike property: %v", err)
		return err
	}
	return nil
}

// IsPropertyLiked 检查用户是否已点赞房源
func (uc *HouseUsecase) IsPropertyLiked(ctx context.Context, propertyId, userId int64) (bool, error) {
	if propertyId <= 0 || userId <= 0 {
		uc.log.Errorf("invalid parameters: propertyId=%d, userId=%d", propertyId, userId)
		return false, errors.New("invalid parameters")
	}

	liked, err := uc.repo.IsPropertyLiked(ctx, propertyId, userId)
	if err != nil {
		uc.log.Errorf("failed to check like status: %v", err)
		return false, err
	}
	return liked, nil
}

// GetPropertyLikeCount 获取房源点赞数
func (uc *HouseUsecase) GetPropertyLikeCount(ctx context.Context, propertyId int64) (int64, error) {
	if propertyId <= 0 {
		uc.log.Errorf("invalid parameter: propertyId=%d", propertyId)
		return 0, errors.New("invalid parameter")
	}

	count, err := uc.repo.GetPropertyLikeCount(ctx, propertyId)
	if err != nil {
		uc.log.Errorf("failed to get like count: %v", err)
		return 0, err
	}
	return count, nil
}

// GetUserLikeList 获取用户点赞列表
func (uc *HouseUsecase) GetUserLikeList(ctx context.Context, userId int64, page, pageSize int) ([]*PropertyLike, int64, error) {
	if userId <= 0 {
		uc.log.Errorf("invalid parameter: userId=%d", userId)
		return nil, 0, errors.New("invalid parameter")
	}

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	likeList, total, err := uc.repo.GetUserLikeList(ctx, userId, page, pageSize)
	if err != nil {
		uc.log.Errorf("failed to get user like list: %v", err)
		return nil, 0, err
	}

	return likeList, total, nil
}

// GetHouseDetail 获取房源详情
func (uc *HouseUsecase) GetHouseDetail(ctx context.Context, houseId int64) (*House, error) {
	if houseId <= 0 {
		uc.log.Errorf("invalid parameter: houseId=%d", houseId)
		return nil, errors.New("invalid parameter")
	}

	house, err := uc.repo.GetHouseDetail(ctx, houseId)
	if err != nil {
		uc.log.Errorf("failed to get house detail: %v", err)
		return nil, err
	}

	return house, nil
}
