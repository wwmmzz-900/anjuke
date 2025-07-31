package biz

import (
	"anjuke/server/internal/domain"
	"context"
	"fmt"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
)

// CompanyUsecase 公司业务逻辑
type CompanyUsecase struct {
	companyRepo domain.CompanyRepo
	storeRepo   domain.StoreRepo
	log         *log.Helper
}

// NewCompanyUsecase 创建公司业务逻辑实例
func NewCompanyUsecase(companyRepo domain.CompanyRepo, storeRepo domain.StoreRepo, logger log.Logger) *CompanyUsecase {
	return &CompanyUsecase{
		companyRepo: companyRepo,
		storeRepo:   storeRepo,
		log:         log.NewHelper(logger),
	}
}

// CreateCompany 创建公司
func (uc *CompanyUsecase) CreateCompany(ctx context.Context, company *domain.CompanyInfo) (*domain.CompanyInfo, error) {
	// 业务验证
	if company.FullName == "" {
		return nil, fmt.Errorf("公司全称不能为空")
	}
	if company.Phone == "" {
		return nil, fmt.Errorf("联系电话不能为空")
	}

	return uc.companyRepo.CreateCompany(ctx, company)
}

// GetCompany 获取公司信息
func (uc *CompanyUsecase) GetCompany(ctx context.Context, id string) (*domain.CompanyInfo, error) {
	// 将string ID转换为uint64
	companyID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的公司ID格式: %s", id)
	}
	return uc.companyRepo.GetCompanyByID(ctx, companyID)
}

// UpdateCompany 更新公司信息
func (uc *CompanyUsecase) UpdateCompany(ctx context.Context, company *domain.CompanyInfo) (*domain.CompanyInfo, error) {
	// 业务验证
	if company.FullName == "" {
		return nil, fmt.Errorf("公司全称不能为空")
	}
	if company.Phone == "" {
		return nil, fmt.Errorf("联系电话不能为空")
	}

	return uc.companyRepo.UpdateCompany(ctx, company)
}

// DeleteCompany 删除公司
func (uc *CompanyUsecase) DeleteCompany(ctx context.Context, id string) error {
	// 将string ID转换为uint64
	companyID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的公司ID格式: %s", id)
	}
	return uc.companyRepo.DeleteCompany(ctx, companyID)
}

// ListCompanies 分页查询公司列表
func (uc *CompanyUsecase) ListCompanies(ctx context.Context, page, pageSize int32, keyword string) ([]*domain.CompanyInfo, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	return uc.companyRepo.ListCompanies(ctx, page, pageSize, keyword)
}

// GetCompanyStores 获取公司下的所有门店
func (uc *CompanyUsecase) GetCompanyStores(ctx context.Context, companyID string) ([]*domain.StoreInfo, error) {
	// 将string ID转换为uint64
	id, err := strconv.ParseUint(companyID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的公司ID格式: %s", companyID)
	}

	// 先验证公司是否存在
	_, err = uc.companyRepo.GetCompanyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.storeRepo.GetStoresByCompanyID(ctx, id)
}

// StoreUsecase 门店业务逻辑
type StoreUsecase struct {
	storeRepo   domain.StoreRepo
	companyRepo domain.CompanyRepo
	realtorRepo domain.RealtorRepo
	log         *log.Helper
}

// NewStoreUsecase 创建门店业务逻辑实例
func NewStoreUsecase(storeRepo domain.StoreRepo, companyRepo domain.CompanyRepo, realtorRepo domain.RealtorRepo, logger log.Logger) *StoreUsecase {
	return &StoreUsecase{
		storeRepo:   storeRepo,
		companyRepo: companyRepo,
		realtorRepo: realtorRepo,
		log:         log.NewHelper(logger),
	}
}

// CreateStore 创建门店
func (uc *StoreUsecase) CreateStore(ctx context.Context, store *domain.StoreInfo) (*domain.StoreInfo, error) {
	// 业务验证
	if store.StoreName == "" {
		return nil, fmt.Errorf("门店名称不能为空")
	}
	if store.CompanyID == 0 {
		return nil, fmt.Errorf("必须指定所属公司")
	}

	return uc.storeRepo.CreateStore(ctx, store)
}

// GetStore 获取门店信息
func (uc *StoreUsecase) GetStore(ctx context.Context, id string) (*domain.StoreInfo, error) {
	// 将string ID转换为uint64
	storeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的门店ID格式: %s", id)
	}
	return uc.storeRepo.GetStoreByID(ctx, storeID)
}

// UpdateStore 更新门店信息
func (uc *StoreUsecase) UpdateStore(ctx context.Context, store *domain.StoreInfo) (*domain.StoreInfo, error) {
	// 业务验证
	if store.StoreName == "" {
		return nil, fmt.Errorf("门店名称不能为空")
	}
	if store.CompanyID == 0 {
		return nil, fmt.Errorf("必须指定所属公司")
	}

	return uc.storeRepo.UpdateStore(ctx, store)
}

// DeleteStore 删除门店
func (uc *StoreUsecase) DeleteStore(ctx context.Context, id string) error {
	// 将string ID转换为uint64
	storeID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的门店ID格式: %s", id)
	}
	return uc.storeRepo.DeleteStore(ctx, storeID)
}

// ListStores 分页查询门店列表
func (uc *StoreUsecase) ListStores(ctx context.Context, page, pageSize int32, keyword string, companyID string) ([]*domain.StoreInfo, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 将string ID转换为uint64
	var id uint64 = 0
	if companyID != "" {
		var err error
		id, err = strconv.ParseUint(companyID, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("无效的公司ID格式: %s", companyID)
		}
	}

	return uc.storeRepo.ListStores(ctx, page, pageSize, keyword, id)
}

// GetStoreRealtors 获取门店下的所有经纪人
func (uc *StoreUsecase) GetStoreRealtors(ctx context.Context, storeID string) ([]*domain.RealtorInfo, error) {
	// 将string ID转换为uint64
	id, err := strconv.ParseUint(storeID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的门店ID格式: %s", storeID)
	}

	// 先验证门店是否存在
	_, err = uc.storeRepo.GetStoreByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.realtorRepo.GetRealtorsByStoreID(ctx, id)
}

// RealtorUsecase 经纪人业务逻辑
type RealtorUsecase struct {
	realtorRepo domain.RealtorRepo
	storeRepo   domain.StoreRepo
	companyRepo domain.CompanyRepo
	log         *log.Helper
}

// NewRealtorUsecase 创建经纪人业务逻辑实例
func NewRealtorUsecase(realtorRepo domain.RealtorRepo, storeRepo domain.StoreRepo, companyRepo domain.CompanyRepo, logger log.Logger) *RealtorUsecase {
	return &RealtorUsecase{
		realtorRepo: realtorRepo,
		storeRepo:   storeRepo,
		companyRepo: companyRepo,
		log:         log.NewHelper(logger),
	}
}

// CreateRealtor 创建经纪人
func (uc *RealtorUsecase) CreateRealtor(ctx context.Context, realtor *domain.RealtorInfo) (*domain.RealtorInfo, error) {
	// 业务验证
	if realtor.RealtorName == "" {
		return nil, fmt.Errorf("经纪人姓名不能为空")
	}
	if realtor.CompanyID == 0 {
		return nil, fmt.Errorf("必须指定所属公司")
	}
	if realtor.StoreID == 0 {
		return nil, fmt.Errorf("必须指定所属门店")
	}

	return uc.realtorRepo.CreateRealtor(ctx, realtor)
}

// GetRealtor 获取经纪人信息
func (uc *RealtorUsecase) GetRealtor(ctx context.Context, id string) (*domain.RealtorInfo, error) {
	// 将string ID转换为uint64
	realtorID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的经纪人ID格式: %s", id)
	}
	return uc.realtorRepo.GetRealtorByID(ctx, realtorID)
}

// UpdateRealtor 更新经纪人信息
func (uc *RealtorUsecase) UpdateRealtor(ctx context.Context, realtor *domain.RealtorInfo) (*domain.RealtorInfo, error) {
	// 业务验证
	if realtor.RealtorName == "" {
		return nil, fmt.Errorf("经纪人姓名不能为空")
	}
	if realtor.CompanyID == 0 {
		return nil, fmt.Errorf("必须指定所属公司")
	}
	if realtor.StoreID == 0 {
		return nil, fmt.Errorf("必须指定所属门店")
	}

	return uc.realtorRepo.UpdateRealtor(ctx, realtor)
}

// DeleteRealtor 删除经纪人
func (uc *RealtorUsecase) DeleteRealtor(ctx context.Context, id string) error {
	// 将string ID转换为uint64
	realtorID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的经纪人ID格式: %s", id)
	}
	return uc.realtorRepo.DeleteRealtor(ctx, realtorID)
}

// ListRealtors 分页查询经纪人列表
func (uc *RealtorUsecase) ListRealtors(ctx context.Context, page, pageSize int32, keyword string, storeID string) ([]*domain.RealtorInfo, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 将string ID转换为uint64
	var id uint64 = 0
	if storeID != "" {
		var err error
		id, err = strconv.ParseUint(storeID, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("无效的门店ID格式: %s", storeID)
		}
	}

	return uc.realtorRepo.ListRealtors(ctx, page, pageSize, keyword, id)
}
