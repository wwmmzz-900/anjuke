package domain

// SearchCriteria 通用搜索条件
type SearchCriteria struct {
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
	Keyword  string `json:"keyword"`
}

// CompanyID 公司ID值对象
type CompanyID struct {
	Value uint64
}

// StoreID 门店ID值对象
type StoreID struct {
	Value uint64
}

// RealtorID 经纪人ID值对象
type RealtorID struct {
	Value uint64
}
