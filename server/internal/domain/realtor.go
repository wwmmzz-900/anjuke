package domain

import "context"

// RealtorRepo 经纪人仓储接口
type RealtorRepo interface {
	CreateRealtor(ctx context.Context, realtor *RealtorInfo) (*RealtorInfo, error)
	GetRealtorByID(ctx context.Context, id uint64) (*RealtorInfo, error)
	UpdateRealtor(ctx context.Context, realtor *RealtorInfo) (*RealtorInfo, error)
	DeleteRealtor(ctx context.Context, id uint64) error
	ListRealtors(ctx context.Context, page, pageSize int32, keyword string, storeID uint64) ([]*RealtorInfo, int64, error)
	GetRealtorsByStoreID(ctx context.Context, storeID uint64) ([]*RealtorInfo, error)
}
