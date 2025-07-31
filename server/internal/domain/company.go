package domain

import (
	"context"
	"time"
)

// CompanyInfo 公司业务实体
type CompanyInfo struct {
	ID            uint64    `json:"id"`
	CompanyLogo   string    `json:"company_logo,omitempty"`
	FullName      string    `json:"full_name"`
	ShortName     string    `json:"short_name"`
	BusinessScope string    `json:"business_scope"`
	Address       string    `json:"address"`
	Phone         string    `json:"phone"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IsValid 验证公司信息是否有效
func (c *CompanyInfo) IsValid() bool {
	return c.FullName != "" && c.Phone != ""
}

// StoreInfo 门店业务实体
type StoreInfo struct {
	ID                   uint64    `json:"id"`
	StoreName            string    `json:"store_name"`
	CompanyID            uint64    `json:"company_id"`
	Address              string    `json:"address"`
	Phone                string    `json:"phone"`
	BusinessHours        string    `json:"business_hours"`
	Rating               float64   `json:"rating"`
	ReviewCount          int32     `json:"review_count"`
	IsActive             bool      `json:"is_active"`
	MainBusinessArea     string    `json:"main_business_area"`
	MainResidentialAreas string    `json:"main_residential_areas"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// IsValid 验证门店信息是否有效
func (s *StoreInfo) IsValid() bool {
	return s.StoreName != "" && s.CompanyID != 0
}

// RealtorInfo 经纪人业务实体
type RealtorInfo struct {
	ID                   uint64    `json:"id"`
	RealtorName          string    `json:"realtor_name"`
	BusinessArea         string    `json:"business_area"`
	SecondHandScore      int       `json:"second_hand_score"`
	RentalScore          int       `json:"rental_score"`
	ServiceYears         string    `json:"service_years"`
	ServicePeopleCount   int       `json:"service_people_count"`
	MainBusinessArea     string    `json:"main_business_area"`
	MainResidentialAreas string    `json:"main_residential_areas"`
	CompanyID            uint64    `json:"company_id"`
	StoreID              uint64    `json:"store_id"`
	Phone                string    `json:"phone,omitempty"`
	Avatar               string    `json:"avatar,omitempty"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// IsValid 验证经纪人信息是否有效
func (r *RealtorInfo) IsValid() bool {
	return r.RealtorName != "" && r.CompanyID != 0 && r.StoreID != 0
}

// CompanyRepo 公司仓储接口
type CompanyRepo interface {
	CreateCompany(ctx context.Context, company *CompanyInfo) (*CompanyInfo, error)
	GetCompanyByID(ctx context.Context, id uint64) (*CompanyInfo, error)
	UpdateCompany(ctx context.Context, company *CompanyInfo) (*CompanyInfo, error)
	DeleteCompany(ctx context.Context, id uint64) error
	ListCompanies(ctx context.Context, page, pageSize int32, keyword string) ([]*CompanyInfo, int64, error)
}
