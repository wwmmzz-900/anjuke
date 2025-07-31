package data

import (
	"context"
	"fmt"
	"strconv"

	"anjuke/server/internal/domain"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// realtorDBRepo 数据库经纪人仓储实现
type realtorDBRepo struct {
	data *DataDB
	log  *log.Helper
}

// NewRealtorDBRepo 创建数据库经纪人仓储实例
func NewRealtorDBRepo(data *DataDB, logger log.Logger) domain.RealtorRepo {
	return &realtorDBRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateRealtor 创建经纪人
func (r *realtorDBRepo) CreateRealtor(ctx context.Context, realtor *domain.RealtorInfo) (*domain.RealtorInfo, error) {
	// 需要将string类型的ServiceYears转换为int32
	serviceYears, _ := strconv.Atoi(realtor.ServiceYears)

	model := &RealtorModel{
		RealtorName:          realtor.RealtorName,
		BusinessArea:         realtor.BusinessArea,
		SecondHandScore:      int32(realtor.SecondHandScore),
		RentalScore:          int32(realtor.RentalScore),
		ServiceYears:         int32(serviceYears),
		ServicePeopleCount:   int32(realtor.ServicePeopleCount),
		MainBusinessArea:     realtor.MainBusinessArea,
		MainResidentialAreas: realtor.MainResidentialAreas,
		CompanyID:            realtor.CompanyID,
		StoreID:              realtor.StoreID,
		Phone:                realtor.Phone,
		Avatar:               realtor.Avatar,
		IsActive:             realtor.IsActive,
	}

	err := r.data.GetDB().WithContext(ctx).Create(model).Error
	if err != nil {
		r.log.Errorf("创建经纪人失败: %v", err)
		return nil, fmt.Errorf("创建经纪人失败: %v", err)
	}

	result := model.ToRealtorInfo()
	r.log.Infof("经纪人创建成功: ID=%d, Name=%s", result.ID, result.RealtorName)
	return result, nil
}

// GetRealtorByID 根据ID获取经纪人
func (r *realtorDBRepo) GetRealtorByID(ctx context.Context, id uint64) (*domain.RealtorInfo, error) {
	var realtor RealtorModel
	err := r.data.GetDB().WithContext(ctx).First(&realtor, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warnf("经纪人不存在，ID: %d", id)
			return nil, fmt.Errorf("经纪人不存在")
		}
		r.log.Errorf("查询经纪人失败，ID: %d, 错误: %v", id, err)
		return nil, fmt.Errorf("查询经纪人失败: %v", err)
	}

	return realtor.ToRealtorInfo(), nil
}

// UpdateRealtor 更新经纪人信息
func (r *realtorDBRepo) UpdateRealtor(ctx context.Context, realtor *domain.RealtorInfo) (*domain.RealtorInfo, error) {
	// 需要将string类型的ServiceYears转换为int32
	serviceYears, _ := strconv.Atoi(realtor.ServiceYears)

	updates := map[string]interface{}{
		"realtor_name":           realtor.RealtorName,
		"business_area":          realtor.BusinessArea,
		"second_hand_score":      int32(realtor.SecondHandScore),
		"rental_score":           int32(realtor.RentalScore),
		"service_years":          int32(serviceYears),
		"service_people_count":   int32(realtor.ServicePeopleCount),
		"main_business_area":     realtor.MainBusinessArea,
		"main_residential_areas": realtor.MainResidentialAreas,
		"company_id":             realtor.CompanyID,
		"store_id":               realtor.StoreID,
		"phone":                  realtor.Phone,
		"avatar":                 realtor.Avatar,
		"is_active":              realtor.IsActive,
	}

	err := r.data.GetDB().WithContext(ctx).Model(&RealtorModel{}).Where("id = ?", realtor.ID).Updates(updates).Error
	if err != nil {
		r.log.Errorf("更新经纪人失败: %v", err)
		return nil, fmt.Errorf("更新经纪人失败: %v", err)
	}

	// 重新获取更新后的数据
	return r.GetRealtorByID(ctx, realtor.ID)
}

// DeleteRealtor 删除经纪人
func (r *realtorDBRepo) DeleteRealtor(ctx context.Context, id uint64) error {
	err := r.data.GetDB().WithContext(ctx).Delete(&RealtorModel{}, id).Error
	if err != nil {
		r.log.Errorf("删除经纪人失败: %v", err)
		return fmt.Errorf("删除经纪人失败: %v", err)
	}

	r.log.Infof("经纪人删除成功: ID=%d", id)
	return nil
}

// GetRealtorsByStoreID 根据门店ID获取经纪人列表
func (r *realtorDBRepo) GetRealtorsByStoreID(ctx context.Context, storeID uint64) ([]*domain.RealtorInfo, error) {
	var realtors []RealtorModel

	// 根据门店ID查询经纪人
	err := r.data.GetDB().WithContext(ctx).
		Where("store_id = ? AND is_active = ?", storeID, true).
		Find(&realtors).Error

	if err != nil {
		return nil, fmt.Errorf("查询门店经纪人失败: %v", err)
	}

	// 转换为领域对象
	result := make([]*domain.RealtorInfo, len(realtors))
	for i, realtor := range realtors {
		result[i] = realtor.ToRealtorInfo()
	}

	return result, nil
}

// ListRealtors 获取经纪人列表
func (r *realtorDBRepo) ListRealtors(ctx context.Context, page, pageSize int32, keyword string, storeID uint64) ([]*domain.RealtorInfo, int64, error) {
	var realtors []RealtorModel
	var total int64

	query := r.data.GetDB().WithContext(ctx).Model(&RealtorModel{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("realtor_name LIKE ? OR business_area LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 门店筛选
	if storeID != 0 {
		query = query.Where("store_id = ?", storeID)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取经纪人总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Offset(int(offset)).Limit(int(pageSize)).Find(&realtors).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取经纪人列表失败: %v", err)
	}

	// 转换为领域对象
	result := make([]*domain.RealtorInfo, len(realtors))
	for i, realtor := range realtors {
		result[i] = realtor.ToRealtorInfo()
	}

	return result, total, nil
}
