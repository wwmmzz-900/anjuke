package domain

import "context"

// StoreRepo 门店仓储接口
type StoreRepo interface {
	CreateStore(ctx context.Context, store *StoreInfo) (*StoreInfo, error)
	GetStoreByID(ctx context.Context, id uint64) (*StoreInfo, error)
	UpdateStore(ctx context.Context, store *StoreInfo) (*StoreInfo, error)
	DeleteStore(ctx context.Context, id uint64) error
	ListStores(ctx context.Context, page, pageSize int32, keyword string, companyID uint64) ([]*StoreInfo, int64, error)
	GetStoresByCompanyID(ctx context.Context, companyID uint64) ([]*StoreInfo, error)
}
