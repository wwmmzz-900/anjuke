package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type HouseRepo interface {
	CreateHouse(ctx context.Context, house *House) (int64, error)
}
type HouseUsecase struct {
	repo HouseRepo
	log  *log.Helper
}
type House struct {
	HouseID                 int64   `gorm:"column:house_id;primaryKey;autoIncrement" json:"house_id"`
	Title                   string  `gorm:"column:title" json:"title"`
	Description             string  `gorm:"column:description" json:"description"`
	LandlordID              int64   `gorm:"column:landlord_id" json:"landlord_id"`
	Address                 string  `gorm:"column:address" json:"address"`
	RegionID                int64   `gorm:"column:region_id" json:"region_id"`
	CommunityID             int64   `gorm:"column:community_id" json:"community_id"`
	Price                   float64 `gorm:"column:price" json:"price"`
	Area                    float32 `gorm:"column:area" json:"area"`
	Layout                  string  `gorm:"column:layout" json:"layout"`
	Floor                   string  `gorm:"column:floor" json:"floor"`
	OwnershipCertificateUrl string  `gorm:"column:ownership_certificate_url" json:"ownership_certificate_url"`
	Orientation             string  `gorm:"column:orientation" json:"orientation"`
	Decoration              string  `gorm:"column:decoration" json:"decoration"`
	Facilities              string  `gorm:"column:facilities" json:"facilities"`
}

func (House) TableName() string {
	return "house"
}
func NewHouseUsecase(repo HouseRepo, logger log.Logger) *HouseUsecase {
	return &HouseUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *HouseUsecase) CreateHouse(ctx context.Context, house *House) (int64, error) {
	return uc.repo.CreateHouse(ctx, house)
}
