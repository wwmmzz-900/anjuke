package data

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"anjuke/server/internal/domain"
)

// CompanyModel 公司表模型
type CompanyModel struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement;comment:公司ID，主键，自增" json:"id"`
	CompanyLogo   string    `gorm:"type:varchar(255);comment:公司logo图片地址" json:"company_logo"`
	FullName      string    `gorm:"type:varchar(100);not null;comment:公司全称，不能为空" json:"full_name"`
	ShortName     string    `gorm:"type:varchar(50);not null;comment:公司简称，不能为空" json:"short_name"`
	BusinessScope string    `gorm:"type:text;comment:公司经营范围" json:"business_scope"`
	Address       string    `gorm:"type:varchar(200);comment:公司地址" json:"address"`
	Phone         string    `gorm:"type:varchar(20);comment:公司联系电话" json:"phone"`
	CreatedAt     time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;comment:更新时间，自动更新" json:"updated_at"`

	// 关联关系
	Stores   []StoreModel   `gorm:"foreignKey:CompanyID" json:"stores"`
	Realtors []RealtorModel `gorm:"foreignKey:CompanyID" json:"realtors"`
}

func (CompanyModel) TableName() string {
	return "companies" // 公司表
}

// ToCompanyInfo 转换为领域对象
func (c *CompanyModel) ToCompanyInfo() *domain.CompanyInfo {
	return &domain.CompanyInfo{
		ID:            c.ID,
		CompanyLogo:   c.CompanyLogo,
		FullName:      c.FullName,
		ShortName:     c.ShortName,
		BusinessScope: c.BusinessScope,
		Address:       c.Address,
		Phone:         c.Phone,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// StoreModel 门店表模型
type StoreModel struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement;comment:门店ID，主键，自增" json:"id"`
	StoreName            string    `gorm:"type:varchar(100);not null;comment:门店名称，不能为空" json:"store_name"`
	CompanyID            uint64    `gorm:"not null;index;comment:所属公司ID，外键，不能为空" json:"company_id"`
	Address              string    `gorm:"type:varchar(200);not null;comment:门店地址，不能为空" json:"address"`
	Phone                string    `gorm:"type:varchar(20);not null;comment:门店联系电话，不能为空" json:"phone"`
	BusinessHours        string    `gorm:"type:varchar(50);comment:营业时间" json:"business_hours"`
	Rating               float64   `gorm:"type:decimal(3,2);default:0;comment:门店评分，默认为0" json:"rating"`
	ReviewCount          int32     `gorm:"default:0;comment:评价数量，默认为0" json:"review_count"`
	IsActive             bool      `gorm:"default:true;comment:是否激活，默认为true" json:"is_active"`
	MainBusinessArea     string    `gorm:"type:varchar(100);comment:主营业区域" json:"main_business_area"`
	MainResidentialAreas string    `gorm:"type:text;comment:主要覆盖小区，文本格式" json:"main_residential_areas"`
	CreatedAt            time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime;comment:更新时间，自动更新" json:"updated_at"`

	// 关联关系
	Company  CompanyModel   `gorm:"foreignKey:CompanyID" json:"company"`
	Realtors []RealtorModel `gorm:"foreignKey:StoreID" json:"realtors"`
}

func (StoreModel) TableName() string {
	return "stores" // 门店表
}

// ToStoreInfo 转换为领域对象
func (s *StoreModel) ToStoreInfo() *domain.StoreInfo {
	return &domain.StoreInfo{
		ID:            s.ID,
		StoreName:     s.StoreName,
		CompanyID:     s.CompanyID,
		Address:       s.Address,
		Phone:         s.Phone,
		BusinessHours: s.BusinessHours,
		Rating:        s.Rating,
		ReviewCount:   s.ReviewCount,
		IsActive:      s.IsActive,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

// RealtorModel 经纪人表模型
type RealtorModel struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement;comment:经纪人ID，主键，自增" json:"id"`
	RealtorName          string    `gorm:"type:varchar(50);not null;comment:经纪人姓名，不能为空" json:"realtor_name"`
	BusinessArea         string    `gorm:"type:varchar(100);comment:业务区域" json:"business_area"`
	SecondHandScore      int32     `gorm:"type:tinyint;default:0;comment:二手房评分（0-100分），默认为0" json:"second_hand_score"`
	RentalScore          int32     `gorm:"type:tinyint;default:0;comment:租房评分（0-100分），默认为0" json:"rental_score"`
	ServiceYears         int32     `gorm:"default:0;comment:服务年限，默认为0" json:"service_years"`
	ServicePeopleCount   int32     `gorm:"default:0;comment:服务客户数量，默认为0" json:"service_people_count"`
	MainBusinessArea     string    `gorm:"type:varchar(100);comment:主营业区域" json:"main_business_area"`
	MainResidentialAreas string    `gorm:"type:text;comment:主要覆盖小区，文本格式" json:"main_residential_areas"`
	CompanyID            uint64    `gorm:"not null;index;comment:所属公司ID，外键，不能为空" json:"company_id"`
	StoreID              uint64    `gorm:"not null;index;comment:所属门店ID，外键，不能为空" json:"store_id"`
	Phone                string    `gorm:"type:varchar(20);not null;comment:联系电话，不能为空" json:"phone"`
	Avatar               string    `gorm:"type:varchar(255);comment:头像图片地址" json:"avatar"`
	IsActive             bool      `gorm:"default:true;comment:是否激活，默认为true" json:"is_active"`
	CreatedAt            time.Time `gorm:"autoCreateTime;comment:创建时间，自动生成" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime;comment:更新时间，自动更新" json:"updated_at"`

	// 关联关系
	Company CompanyModel `gorm:"foreignKey:CompanyID" json:"company"`
	Store   StoreModel   `gorm:"foreignKey:StoreID" json:"store"`
}

func (RealtorModel) TableName() string {
	return "realtors" // 经纪人表
}

// ToRealtorInfo 转换为领域对象
func (r *RealtorModel) ToRealtorInfo() *domain.RealtorInfo {
	return &domain.RealtorInfo{
		ID:                   r.ID,
		RealtorName:          r.RealtorName,
		BusinessArea:         r.BusinessArea,
		SecondHandScore:      int(r.SecondHandScore),
		RentalScore:          int(r.RentalScore),
		ServiceYears:         fmt.Sprintf("%d", r.ServiceYears),
		ServicePeopleCount:   int(r.ServicePeopleCount),
		MainBusinessArea:     r.MainBusinessArea,
		MainResidentialAreas: r.MainResidentialAreas,
		CompanyID:            r.CompanyID,
		StoreID:              r.StoreID,
		Phone:                r.Phone,
		Avatar:               r.Avatar,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}

// CompanyDBRepo 公司数据库仓储实现
type CompanyDBRepo struct {
	data *Data
}

// NewCompanyDBRepo 创建公司数据库仓储
func NewCompanyDBRepo(data *Data) domain.CompanyRepo {
	return &CompanyDBRepo{
		data: data,
	}
}

// CreateCompany 创建公司
func (r *CompanyDBRepo) CreateCompany(ctx context.Context, company *domain.CompanyInfo) (*domain.CompanyInfo, error) {
	model := &CompanyModel{
		CompanyLogo:   company.CompanyLogo,
		FullName:      company.FullName,
		ShortName:     company.ShortName,
		BusinessScope: company.BusinessScope,
		Address:       company.Address,
		Phone:         company.Phone,
	}

	if err := r.data.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return model.ToCompanyInfo(), nil
}

// GetCompanyByID 根据ID获取公司
func (r *CompanyDBRepo) GetCompanyByID(ctx context.Context, id uint64) (*domain.CompanyInfo, error) {
	var model CompanyModel
	if err := r.data.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return model.ToCompanyInfo(), nil
}

// UpdateCompany 更新公司
func (r *CompanyDBRepo) UpdateCompany(ctx context.Context, company *domain.CompanyInfo) (*domain.CompanyInfo, error) {
	updates := map[string]interface{}{
		"company_logo":   company.CompanyLogo,
		"full_name":      company.FullName,
		"short_name":     company.ShortName,
		"business_scope": company.BusinessScope,
		"address":        company.Address,
		"phone":          company.Phone,
		"updated_at":     time.Now(),
	}

	if err := r.data.db.WithContext(ctx).Model(&CompanyModel{}).
		Where("id = ?", company.ID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	// 返回更新后的公司信息
	return r.GetCompanyByID(ctx, company.ID)
}

// DeleteCompany 删除公司
func (r *CompanyDBRepo) DeleteCompany(ctx context.Context, id uint64) error {
	return r.data.db.WithContext(ctx).Delete(&CompanyModel{}, id).Error
}

// ListCompanies 获取公司列表
func (r *CompanyDBRepo) ListCompanies(ctx context.Context, page, pageSize int32, keyword string) ([]*domain.CompanyInfo, int64, error) {
	var models []CompanyModel
	var total int64

	query := r.data.db.WithContext(ctx).Model(&CompanyModel{})

	// 如果有关键词，添加搜索条件
	if keyword != "" {
		query = query.Where("full_name LIKE ? OR short_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(int(offset)).Limit(int(pageSize)).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换为领域对象
	companies := make([]*domain.CompanyInfo, len(models))
	for i, model := range models {
		companies[i] = model.ToCompanyInfo()
	}

	return companies, total, nil
}

// GetCompanyStores 获取公司的门店列表
func (r *CompanyDBRepo) GetCompanyStores(ctx context.Context, companyID uint64) ([]*domain.StoreInfo, error) {
	var models []StoreModel
	if err := r.data.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	stores := make([]*domain.StoreInfo, len(models))
	for i, model := range models {
		stores[i] = model.ToStoreInfo()
	}

	return stores, nil
}

// GetCompanyRealtors 获取公司的经纪人列表
func (r *CompanyDBRepo) GetCompanyRealtors(ctx context.Context, companyID uint64) ([]*domain.RealtorInfo, error) {
	var models []RealtorModel
	if err := r.data.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	realtors := make([]*domain.RealtorInfo, len(models))
	for i, model := range models {
		realtors[i] = model.ToRealtorInfo()
	}

	return realtors, nil
}

// CreateStore 创建门店
func (r *CompanyDBRepo) CreateStore(ctx context.Context, store *domain.StoreInfo) (*domain.StoreInfo, error) {
	model := &StoreModel{
		StoreName:            store.StoreName,
		CompanyID:            store.CompanyID,
		Address:              store.Address,
		Phone:                store.Phone,
		BusinessHours:        store.BusinessHours,
		Rating:               store.Rating,
		ReviewCount:          store.ReviewCount,
		IsActive:             store.IsActive,
		MainBusinessArea:     store.MainBusinessArea,
		MainResidentialAreas: store.MainResidentialAreas,
	}

	if err := r.data.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return model.ToStoreInfo(), nil
}

// GetStoreByID 根据ID获取门店
func (r *CompanyDBRepo) GetStoreByID(ctx context.Context, id uint64) (*domain.StoreInfo, error) {
	var model StoreModel
	if err := r.data.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return model.ToStoreInfo(), nil
}

// UpdateStore 更新门店
func (r *CompanyDBRepo) UpdateStore(ctx context.Context, store *domain.StoreInfo) error {
	updates := map[string]interface{}{
		"store_name":             store.StoreName,
		"address":                store.Address,
		"phone":                  store.Phone,
		"business_hours":         store.BusinessHours,
		"rating":                 store.Rating,
		"review_count":           store.ReviewCount,
		"is_active":              store.IsActive,
		"main_business_area":     store.MainBusinessArea,
		"main_residential_areas": store.MainResidentialAreas,
		"updated_at":             time.Now(),
	}

	return r.data.db.WithContext(ctx).Model(&StoreModel{}).
		Where("id = ?", store.ID).
		Updates(updates).Error
}

// DeleteStore 删除门店
func (r *CompanyDBRepo) DeleteStore(ctx context.Context, id uint64) error {
	return r.data.db.WithContext(ctx).Delete(&StoreModel{}, id).Error
}

// CreateRealtor 创建经纪人
func (r *CompanyDBRepo) CreateRealtor(ctx context.Context, realtor *domain.RealtorInfo) (*domain.RealtorInfo, error) {
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

	if err := r.data.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return model.ToRealtorInfo(), nil
}

// GetRealtorByID 根据ID获取经纪人
func (r *CompanyDBRepo) GetRealtorByID(ctx context.Context, id uint64) (*domain.RealtorInfo, error) {
	var model RealtorModel
	if err := r.data.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return model.ToRealtorInfo(), nil
}

// UpdateRealtor 更新经纪人
func (r *CompanyDBRepo) UpdateRealtor(ctx context.Context, realtor *domain.RealtorInfo) error {
	updates := map[string]interface{}{
		"realtor_name":           realtor.RealtorName,
		"business_area":          realtor.BusinessArea,
		"second_hand_score":      realtor.SecondHandScore,
		"rental_score":           realtor.RentalScore,
		"service_years":          realtor.ServiceYears,
		"service_people_count":   realtor.ServicePeopleCount,
		"main_business_area":     realtor.MainBusinessArea,
		"main_residential_areas": realtor.MainResidentialAreas,
		"store_id":               realtor.StoreID,
		"phone":                  realtor.Phone,
		"avatar":                 realtor.Avatar,
		"is_active":              realtor.IsActive,
		"updated_at":             time.Now(),
	}

	return r.data.db.WithContext(ctx).Model(&RealtorModel{}).
		Where("id = ?", realtor.ID).
		Updates(updates).Error
}

// DeleteRealtor 删除经纪人
func (r *CompanyDBRepo) DeleteRealtor(ctx context.Context, id uint64) error {
	return r.data.db.WithContext(ctx).Delete(&RealtorModel{}, id).Error
}

// GetStoreRealtors 获取门店的经纪人列表
func (r *CompanyDBRepo) GetStoreRealtors(ctx context.Context, storeID uint64) ([]*domain.RealtorInfo, error) {
	var models []RealtorModel
	if err := r.data.db.WithContext(ctx).
		Where("store_id = ? AND is_active = ?", storeID, true).
		Find(&models).Error; err != nil {
		return nil, err
	}

	realtors := make([]*domain.RealtorInfo, len(models))
	for i, model := range models {
		realtors[i] = model.ToRealtorInfo()
	}

	return realtors, nil
}
