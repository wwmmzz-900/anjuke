package params

// CreateOrderParams 创建订单参数
type CreateOrderParams struct {
	HouseId     uint
	TenantId    uint
	LandlordId  uint
	TenantPhone string
	RentStart   string
	RentEnd     string
	RentAmount  float64
	Deposit     float64
}
