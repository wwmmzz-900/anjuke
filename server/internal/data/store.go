package data

import (
	"context"
	"fmt"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// storeDBRepo 数据库门店仓储实现
type storeDBRepo struct {
	data *DataDB
	log  *log.Helper
}

// NewStoreDBRepo 创建数据库门店仓储实例
func NewStoreDBRepo(data *DataDB, logger log.Logger) domain.StoreRepo {
	return &storeDBRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateStore 创建门店
func (r *storeDBRepo) CreateStore(ctx context.Context, store *domain.StoreInfo) (*domain.StoreInfo, error) {
	storeModel := &StoreModel{
		StoreName:     store.StoreName,
		CompanyID:     store.CompanyID,
		Address:       store.Address,
		Phone:         store.Phone,
		BusinessHours: store.BusinessHours,
		Rating:        store.Rating,
		ReviewCount:   store.ReviewCount,
		IsActive:      store.IsActive,
	}

	err := r.data.GetDB().WithContext(ctx).Create(storeModel).Error
	if err != nil {
		r.log.Errorf("创建门店失败: %v", err)
		return nil, fmt.Errorf("创建门店失败: %v", err)
	}

	result := storeModel.ToStoreInfo()
	r.log.Infof("门店创建成功: ID=%d, Name=%s", result.ID, result.StoreName)
	return result, nil
}

// GetStoreByID 根据ID获取门店信息
func (r *storeDBRepo) GetStoreByID(ctx context.Context, id uint64) (*domain.StoreInfo, error) {
	var store StoreModel
	err := r.data.GetDB().WithContext(ctx).First(&store, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warnf("门店不存在，ID: %d", id)
			return nil, fmt.Errorf("门店不存在")
		}
		r.log.Errorf("查询门店失败，ID: %d, 错误: %v", id, err)
		return nil, fmt.Errorf("查询门店失败: %v", err)
	}

	return store.ToStoreInfo(), nil
}

// UpdateStore 更新门店信息
func (r *storeDBRepo) UpdateStore(ctx context.Context, store *domain.StoreInfo) (*domain.StoreInfo, error) {
	updates := map[string]interface{}{
		"store_name":     store.StoreName,
		"company_id":     store.CompanyID,
		"address":        store.Address,
		"phone":          store.Phone,
		"business_hours": store.BusinessHours,
		"rating":         store.Rating,
		"review_count":   store.ReviewCount,
		"is_active":      store.IsActive,
	}

	err := r.data.GetDB().WithContext(ctx).Model(&StoreModel{}).Where("id = ?", store.ID).Updates(updates).Error
	if err != nil {
		r.log.Errorf("更新门店失败: %v", err)
		return nil, fmt.Errorf("更新门店失败: %v", err)
	}

	// 重新获取更新后的数据
	return r.GetStoreByID(ctx, store.ID)
}

// DeleteStore 删除门店
func (r *storeDBRepo) DeleteStore(ctx context.Context, id uint64) error {
	err := r.data.GetDB().WithContext(ctx).Delete(&StoreModel{}, id).Error
	if err != nil {
		r.log.Errorf("删除门店失败: %v", err)
		return fmt.Errorf("删除门店失败: %v", err)
	}

	r.log.Infof("门店删除成功: ID=%d", id)
	return nil
}

// ListStores 获取门店列表
func (r *storeDBRepo) ListStores(ctx context.Context, page, pageSize int32, keyword string, companyID uint64) ([]*domain.StoreInfo, int64, error) {
	var stores []StoreModel
	var total int64

	query := r.data.GetDB().WithContext(ctx).Model(&StoreModel{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("store_name LIKE ? OR address LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 公司筛选
	if companyID != 0 {
		query = query.Where("company_id = ?", companyID)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取门店总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Offset(int(offset)).Limit(int(pageSize)).Find(&stores).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取门店列表失败: %v", err)
	}

	// 转换为领域对象
	result := make([]*domain.StoreInfo, len(stores))
	for i, store := range stores {
		result[i] = store.ToStoreInfo()
	}

	return result, total, nil
}

// GetStoresByCompanyID 根据公司ID获取门店列表
func (r *storeDBRepo) GetStoresByCompanyID(ctx context.Context, companyID uint64) ([]*domain.StoreInfo, error) {
	var stores []StoreModel

	// 根据公司ID查询门店
	err := r.data.GetDB().WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Find(&stores).Error

	if err != nil {
		return nil, fmt.Errorf("查询公司门店失败: %v", err)
	}

	// 转换为领域对象
	result := make([]*domain.StoreInfo, len(stores))
	for i, store := range stores {
		result[i] = store.ToStoreInfo()
	}

	return result, nil
}
