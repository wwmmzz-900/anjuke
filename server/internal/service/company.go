package service

import (
	commonv1 "anjuke/server/api/common/v1"
	v1 "anjuke/server/api/company/v1"
	"anjuke/server/internal/biz"
	"anjuke/server/internal/domain"
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/anypb"
)

// CompanyService 公司服务
type CompanyService struct {
	v1.UnimplementedCompanyServer
	companyUc *biz.CompanyUsecase
	log       *log.Helper
}

// NewCompanyService 创建公司服务实例
func NewCompanyService(companyUc *biz.CompanyUsecase, logger log.Logger) *CompanyService {
	return &CompanyService{
		companyUc: companyUc,
		log:       log.NewHelper(logger),
	}
}

// 辅助函数：将domain.CompanyInfo转换为proto.CompanyInfo
func domainCompanyToProto(company *domain.CompanyInfo) *v1.CompanyInfo {
	return &v1.CompanyInfo{
		Id:            company.ID,
		FullName:      company.FullName,
		ShortName:     company.ShortName,
		BusinessScope: company.BusinessScope,
		Address:       company.Address,
		Phone:         company.Phone,
		CompanyLogo:   company.CompanyLogo,
	}
}

// CreateCompany 创建公司
func (s *CompanyService) CreateCompany(ctx context.Context, req *v1.CreateCompanyRequest) (*commonv1.BaseResponse, error) {
	company := &domain.CompanyInfo{
		FullName:      req.FullName,
		ShortName:     req.ShortName,
		BusinessScope: req.BusinessScope,
		Address:       req.Address,
		Phone:         req.Phone,
		CompanyLogo:   req.CompanyLogo,
	}

	result, err := s.companyUc.CreateCompany(ctx, company)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoCompany := domainCompanyToProto(result)
	data, _ := anypb.New(protoCompany)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "创建成功",
		Data: data,
	}, nil
}

// GetCompany 获取公司信息
func (s *CompanyService) GetCompany(ctx context.Context, req *v1.GetCompanyRequest) (*commonv1.BaseResponse, error) {
	result, err := s.companyUc.GetCompany(ctx, strconv.FormatUint(req.Id, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoCompany := domainCompanyToProto(result)
	data, _ := anypb.New(protoCompany)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// UpdateCompany 更新公司信息
func (s *CompanyService) UpdateCompany(ctx context.Context, req *v1.UpdateCompanyRequest) (*commonv1.BaseResponse, error) {
	company := &domain.CompanyInfo{
		ID:            req.Id,
		FullName:      req.FullName,
		ShortName:     req.ShortName,
		BusinessScope: req.BusinessScope,
		Address:       req.Address,
		Phone:         req.Phone,
		CompanyLogo:   req.CompanyLogo,
	}

	result, err := s.companyUc.UpdateCompany(ctx, company)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoCompany := domainCompanyToProto(result)
	data, _ := anypb.New(protoCompany)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "更新成功",
		Data: data,
	}, nil
}

// DeleteCompany 删除公司
func (s *CompanyService) DeleteCompany(ctx context.Context, req *v1.DeleteCompanyRequest) (*commonv1.BaseResponse, error) {
	err := s.companyUc.DeleteCompany(ctx, strconv.FormatUint(req.Id, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "删除成功",
	}, nil
}

// ListCompanies 查询公司列表
func (s *CompanyService) ListCompanies(ctx context.Context, req *v1.ListCompaniesRequest) (*commonv1.BaseResponse, error) {
	companies, total, err := s.companyUc.ListCompanies(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	var protoCompanies []*v1.CompanyInfo
	for _, company := range companies {
		protoCompanies = append(protoCompanies, domainCompanyToProto(company))
	}

	response := &v1.ListCompaniesResponse{
		Companies: protoCompanies,
		Total:     int32(total),
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	data, _ := anypb.New(response)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// GetCompanyStores 获取公司下的所有门店
func (s *CompanyService) GetCompanyStores(ctx context.Context, req *v1.GetCompanyStoresRequest) (*commonv1.BaseResponse, error) {
	stores, err := s.companyUc.GetCompanyStores(ctx, strconv.FormatUint(req.CompanyId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	var protoStores []*v1.StoreInfo
	for _, store := range stores {
		protoStores = append(protoStores, &v1.StoreInfo{
			Id:        store.ID,
			StoreName: store.StoreName,
			Address:   store.Address,
			Phone:     store.Phone,
			CompanyId: store.CompanyID,
		})
	}

	data, _ := anypb.New(&v1.ListStoresResponse{
		Stores: protoStores,
	})

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// StoreService 门店服务
type StoreService struct {
	v1.UnimplementedStoreServer
	storeUc   *biz.StoreUsecase
	companyUc *biz.CompanyUsecase
	log       *log.Helper
}

// NewStoreService 创建门店服务实例
func NewStoreService(storeUc *biz.StoreUsecase, companyUc *biz.CompanyUsecase, logger log.Logger) *StoreService {
	return &StoreService{
		storeUc:   storeUc,
		companyUc: companyUc,
		log:       log.NewHelper(logger),
	}
}

// 辅助函数：将domain.StoreInfo转换为proto.StoreInfo
func domainStoreToProto(store *domain.StoreInfo) *v1.StoreInfo {
	return &v1.StoreInfo{
		Id:        store.ID,
		StoreName: store.StoreName,
		Address:   store.Address,
		Phone:     store.Phone,
		CompanyId: store.CompanyID,
	}
}

// CreateStore 创建门店
func (s *StoreService) CreateStore(ctx context.Context, req *v1.CreateStoreRequest) (*commonv1.BaseResponse, error) {
	// 首先验证公司是否存在
	_, err := s.companyUc.GetCompany(ctx, strconv.FormatUint(req.CompanyId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 400,
			Msg:  "指定的公司不存在，请先创建公司或检查公司ID是否正确",
		}, nil
	}

	store := &domain.StoreInfo{
		StoreName:     req.StoreName,
		Address:       req.Address,
		Phone:         req.Phone,
		CompanyID:     req.CompanyId,
		BusinessHours: req.BusinessHours,
		Rating:        req.Rating,
		ReviewCount:   req.ReviewCount,
		IsActive:      req.IsActive,
	}

	result, err := s.storeUc.CreateStore(ctx, store)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  "创建门店失败: " + err.Error(),
		}, nil
	}

	protoStore := domainStoreToProto(result)
	data, _ := anypb.New(protoStore)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "创建成功",
		Data: data,
	}, nil
}

// GetStore 获取门店信息
func (s *StoreService) GetStore(ctx context.Context, req *v1.GetStoreRequest) (*commonv1.BaseResponse, error) {
	result, err := s.storeUc.GetStore(ctx, strconv.FormatUint(req.Id, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoStore := domainStoreToProto(result)
	data, _ := anypb.New(protoStore)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// UpdateStore 更新门店信息
func (s *StoreService) UpdateStore(ctx context.Context, req *v1.UpdateStoreRequest) (*commonv1.BaseResponse, error) {
	// 如果要更新公司ID，需要验证新的公司是否存在
	if req.CompanyId > 0 {
		_, err := s.companyUc.GetCompany(ctx, strconv.FormatUint(req.CompanyId, 10))
		if err != nil {
			return &commonv1.BaseResponse{
				Code: 400,
				Msg:  "指定的公司不存在，请先创建公司或检查公司ID是否正确",
			}, nil
		}
	}

	store := &domain.StoreInfo{
		ID:        req.Id,
		StoreName: req.StoreName,
		Address:   req.Address,
		Phone:     req.Phone,
		CompanyID: req.CompanyId,
	}

	result, err := s.storeUc.UpdateStore(ctx, store)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoStore := domainStoreToProto(result)
	data, _ := anypb.New(protoStore)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "更新成功",
		Data: data,
	}, nil
}

// DeleteStore 删除门店
func (s *StoreService) DeleteStore(ctx context.Context, req *v1.DeleteStoreRequest) (*commonv1.BaseResponse, error) {
	err := s.storeUc.DeleteStore(ctx, strconv.FormatUint(req.Id, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "删除成功",
	}, nil
}

// ListStores 查询门店列表
func (s *StoreService) ListStores(ctx context.Context, req *v1.ListStoresRequest) (*commonv1.BaseResponse, error) {
	stores, total, err := s.storeUc.ListStores(ctx, req.Page, req.PageSize, req.Keyword, strconv.FormatUint(req.CompanyId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	var protoStores []*v1.StoreInfo
	for _, store := range stores {
		protoStores = append(protoStores, domainStoreToProto(store))
	}

	response := &v1.ListStoresResponse{
		Stores:   protoStores,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	data, _ := anypb.New(response)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// GetStoreRealtors 获取门店下的经纪人列表
func (s *StoreService) GetStoreRealtors(ctx context.Context, req *v1.GetStoreRealtorsRequest) (*commonv1.BaseResponse, error) {
	realtors, err := s.storeUc.GetStoreRealtors(ctx, strconv.FormatUint(req.StoreId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	var protoRealtors []*v1.RealtorInfo
	for _, realtor := range realtors {
		// 解析JSON字符串为数组
		var businessArea []string
		var mainBusinessArea []string
		var mainResidentialAreas []string

		json.Unmarshal([]byte(realtor.BusinessArea), &businessArea)
		json.Unmarshal([]byte(realtor.MainBusinessArea), &mainBusinessArea)
		json.Unmarshal([]byte(realtor.MainResidentialAreas), &mainResidentialAreas)

		protoRealtors = append(protoRealtors, &v1.RealtorInfo{
			Id:                   realtor.ID,
			RealtorName:          realtor.RealtorName,
			BusinessArea:         businessArea,
			SecondHandScore:      int32(realtor.SecondHandScore),
			RentalScore:          int32(realtor.RentalScore),
			ServiceYears:         realtor.ServiceYears,
			ServicePeopleCount:   int32(realtor.ServicePeopleCount),
			MainBusinessArea:     mainBusinessArea,
			MainResidentialAreas: mainResidentialAreas,
			CompanyId:            realtor.CompanyID,
			StoreId:              realtor.StoreID,
		})
	}

	data, _ := anypb.New(&v1.ListRealtorsResponse{
		Realtors: protoRealtors,
	})

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// RealtorService 经纪人服务
type RealtorService struct {
	v1.UnimplementedRealtorServer
	realtorUc *biz.RealtorUsecase
	storeUc   *biz.StoreUsecase
	companyUc *biz.CompanyUsecase
	log       *log.Helper
}

// NewRealtorService 创建经纪人服务实例
func NewRealtorService(realtorUc *biz.RealtorUsecase, storeUc *biz.StoreUsecase, companyUc *biz.CompanyUsecase, logger log.Logger) *RealtorService {
	return &RealtorService{
		realtorUc: realtorUc,
		storeUc:   storeUc,
		companyUc: companyUc,
		log:       log.NewHelper(logger),
	}
}

// 辅助函数：将domain.RealtorInfo转换为proto.RealtorInfo
func domainRealtorToProto(realtor *domain.RealtorInfo) *v1.RealtorInfo {
	// 解析JSON字符串为数组
	var businessArea []string
	var mainBusinessArea []string
	var mainResidentialAreas []string

	json.Unmarshal([]byte(realtor.BusinessArea), &businessArea)
	json.Unmarshal([]byte(realtor.MainBusinessArea), &mainBusinessArea)
	json.Unmarshal([]byte(realtor.MainResidentialAreas), &mainResidentialAreas)

	return &v1.RealtorInfo{
		Id:                   realtor.ID,
		RealtorName:          realtor.RealtorName,
		BusinessArea:         businessArea,
		SecondHandScore:      int32(realtor.SecondHandScore),
		RentalScore:          int32(realtor.RentalScore),
		ServiceYears:         realtor.ServiceYears,
		ServicePeopleCount:   int32(realtor.ServicePeopleCount),
		MainBusinessArea:     mainBusinessArea,
		MainResidentialAreas: mainResidentialAreas,
		CompanyId:            realtor.CompanyID,
		StoreId:              realtor.StoreID,
	}
}

// CreateRealtor 创建经纪人
func (s *RealtorService) CreateRealtor(ctx context.Context, req *v1.CreateRealtorRequest) (*commonv1.BaseResponse, error) {
	// 验证公司是否存在
	_, err := s.companyUc.GetCompany(ctx, strconv.FormatUint(req.CompanyId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 400,
			Msg:  "指定的公司不存在，请先创建公司或检查公司ID是否正确",
		}, nil
	}

	// 验证门店是否存在
	store, err := s.storeUc.GetStore(ctx, strconv.FormatUint(req.StoreId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 400,
			Msg:  "指定的门店不存在，请先创建门店或检查门店ID是否正确",
		}, nil
	}

	// 验证门店是否属于指定的公司
	if store.CompanyID != req.CompanyId {
		return &commonv1.BaseResponse{
			Code: 400,
			Msg:  "指定的门店不属于该公司，请检查公司ID和门店ID的对应关系",
		}, nil
	}

	// 将数组序列化为JSON字符串
	businessAreaJSON, _ := json.Marshal(req.BusinessArea)
	mainBusinessAreaJSON, _ := json.Marshal(req.MainBusinessArea)
	mainResidentialAreasJSON, _ := json.Marshal(req.MainResidentialAreas)

	realtor := &domain.RealtorInfo{
		RealtorName:          req.RealtorName,
		BusinessArea:         string(businessAreaJSON),
		SecondHandScore:      int(req.SecondHandScore),
		RentalScore:          int(req.RentalScore),
		ServiceYears:         req.ServiceYears,
		ServicePeopleCount:   int(req.ServicePeopleCount),
		MainBusinessArea:     string(mainBusinessAreaJSON),
		MainResidentialAreas: string(mainResidentialAreasJSON),
		CompanyID:            req.CompanyId,
		StoreID:              req.StoreId,
	}

	result, err := s.realtorUc.CreateRealtor(ctx, realtor)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoRealtor := domainRealtorToProto(result)
	data, _ := anypb.New(protoRealtor)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "创建成功",
		Data: data,
	}, nil
}

// GetRealtor 获取经纪人信息
func (s *RealtorService) GetRealtor(ctx context.Context, req *v1.GetRealtorRequest) (*commonv1.BaseResponse, error) {
	result, err := s.realtorUc.GetRealtor(ctx, strconv.FormatUint(req.Id, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoRealtor := domainRealtorToProto(result)
	data, _ := anypb.New(protoRealtor)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}

// UpdateRealtor 更新经纪人信息
func (s *RealtorService) UpdateRealtor(ctx context.Context, req *v1.UpdateRealtorRequest) (*commonv1.BaseResponse, error) {
	// 如果要更新公司ID，需要验证新的公司是否存在
	if req.CompanyId > 0 {
		_, err := s.companyUc.GetCompany(ctx, strconv.FormatUint(req.CompanyId, 10))
		if err != nil {
			return &commonv1.BaseResponse{
				Code: 400,
				Msg:  "指定的公司不存在，请先创建公司或检查公司ID是否正确",
			}, nil
		}
	}

	// 如果要更新门店ID，需要验证新的门店是否存在
	var store *domain.StoreInfo
	if req.StoreId > 0 {
		var err error
		store, err = s.storeUc.GetStore(ctx, strconv.FormatUint(req.StoreId, 10))
		if err != nil {
			return &commonv1.BaseResponse{
				Code: 400,
				Msg:  "指定的门店不存在，请先创建门店或检查门店ID是否正确",
			}, nil
		}
	}

	// 如果同时指定了公司ID和门店ID，需要验证门店是否属于指定的公司
	if req.CompanyId > 0 && req.StoreId > 0 && store != nil {
		if store.CompanyID != req.CompanyId {
			return &commonv1.BaseResponse{
				Code: 400,
				Msg:  "指定的门店不属于该公司，请检查公司ID和门店ID的对应关系",
			}, nil
		}
	}

	// 将数组序列化为JSON字符串
	businessAreaJSON, _ := json.Marshal(req.BusinessArea)
	mainBusinessAreaJSON, _ := json.Marshal(req.MainBusinessArea)
	mainResidentialAreasJSON, _ := json.Marshal(req.MainResidentialAreas)

	realtor := &domain.RealtorInfo{
		ID:                   req.Id,
		RealtorName:          req.RealtorName,
		BusinessArea:         string(businessAreaJSON),
		SecondHandScore:      int(req.SecondHandScore),
		RentalScore:          int(req.RentalScore),
		ServiceYears:         req.ServiceYears,
		ServicePeopleCount:   int(req.ServicePeopleCount),
		MainBusinessArea:     string(mainBusinessAreaJSON),
		MainResidentialAreas: string(mainResidentialAreasJSON),
		CompanyID:            req.CompanyId,
		StoreID:              req.StoreId,
	}

	result, err := s.realtorUc.UpdateRealtor(ctx, realtor)
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	protoRealtor := domainRealtorToProto(result)
	data, _ := anypb.New(protoRealtor)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "更新成功",
		Data: data,
	}, nil
}

// DeleteRealtor 删除经纪人
func (s *RealtorService) DeleteRealtor(ctx context.Context, req *v1.DeleteRealtorRequest) (*commonv1.BaseResponse, error) {
	err := s.realtorUc.DeleteRealtor(ctx, strconv.FormatUint(req.Id, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "删除成功",
	}, nil
}

// ListRealtors 查询经纪人列表
func (s *RealtorService) ListRealtors(ctx context.Context, req *v1.ListRealtorsRequest) (*commonv1.BaseResponse, error) {
	realtors, total, err := s.realtorUc.ListRealtors(ctx, req.Page, req.PageSize, req.Keyword, strconv.FormatUint(req.StoreId, 10))
	if err != nil {
		return &commonv1.BaseResponse{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	var protoRealtors []*v1.RealtorInfo
	for _, realtor := range realtors {
		protoRealtors = append(protoRealtors, domainRealtorToProto(realtor))
	}

	response := &v1.ListRealtorsResponse{
		Realtors: protoRealtors,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	data, _ := anypb.New(response)

	return &commonv1.BaseResponse{
		Code: 0,
		Msg:  "查询成功",
		Data: data,
	}, nil
}
