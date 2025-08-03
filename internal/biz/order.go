package biz

import (
	"anjuke/internal/model"
	"anjuke/internal/model/params"
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type OrderRepo interface {
	// 数据库查询
	GetRentalOrderByDB(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, error)
	// 缓存查询
	GetRentalOrderByCache(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, bool, error)
	// 缓存更新
	SetRentalOrderToCache(ctx context.Context, tenantId uint, page, pageSize int, orders *[]model.RentalOrder, total int64) error
	// 检查熔断器状态
	IsCircuitBreakerOpen() bool
	// 获取降级数据
	GetDegradedOrderList(ctx context.Context, tenantId uint) (*[]model.RentalOrder, int64, error)
	// 获取订单详情
	GetOrderDetail(ctx context.Context, id uint) (*model.RentalOrder, error)
	// 创建订单
	CreateOrder(ctx context.Context, order *model.RentalOrder) error
	// 检查房源是否可用
	CheckHouseAvailable(ctx context.Context, houseId uint, rentStart, rentEnd time.Time) (bool, error)
	// 取消订单
	CancelOrder(ctx context.Context, id uint, cancelReason string) error
}
type OrderUsecase struct {
	repo OrderRepo
	log  *log.Helper
}

func NewOrderUsecase(repo OrderRepo, logger log.Logger) *OrderUsecase {
	return &OrderUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (o *OrderUsecase) GetOrderDetail(ctx context.Context, id uint) (*model.RentalOrder, error) {

	return o.repo.GetOrderDetail(ctx, id)
}

func (o *OrderUsecase) GetTenantOrderList(ctx context.Context, tenantId uint, page, pageSize int) (*[]model.RentalOrder, int64, error) {
	// 1. 参数优化
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20 // 限制最大页大小为50
	}

	// 2. 检查缓存
	cacheOrders, cacheTotal, hit, err := o.repo.GetRentalOrderByCache(ctx, tenantId, page, pageSize)
	if err != nil {
		o.log.Errorf("缓存查询异常: %v", err)
	}
	if hit {
		return cacheOrders, cacheTotal, nil
	}

	// 3. 缓存未命中, 检查熔断器
	if o.repo.IsCircuitBreakerOpen() {
		o.log.Warnf("熔断器开启, 执行降级策略, tenantId: %d", tenantId)
		return o.repo.GetDegradedOrderList(ctx, tenantId)
	}

	// 4. 正常查询数据库
	return o.repo.GetRentalOrderByDB(ctx, tenantId, page, pageSize)
}

func (o *OrderUsecase) CreateOrder(ctx context.Context, params *params.CreateOrderParams) (string, error) {
	// 1. 参数验证
	if params.HouseId == 0 || params.TenantId == 0 || params.LandlordId == 0 {
		return "", fmt.Errorf("房源ID、租客ID、房东ID不能为空")
	}
	if params.TenantPhone == "" {
		return "", fmt.Errorf("租客手机号不能为空")
	}
	if params.RentAmount <= 0 || params.Deposit < 0 {
		return "", fmt.Errorf("租金必须大于0，押金不能小于0")
	}

	// 2. 解析时间
	rentStart, err := time.Parse("2006-01-02", params.RentStart)
	if err != nil {
		return "", fmt.Errorf("租期开始时间格式错误: %v", err)
	}
	rentEnd, err := time.Parse("2006-01-02", params.RentEnd)
	if err != nil {
		return "", fmt.Errorf("租期结束时间格式错误: %v", err)
	}
	if rentEnd.Before(rentStart) {
		return "", fmt.Errorf("租期结束时间不能早于开始时间")
	}

	// 3. 检查房源是否可用
	available, err := o.repo.CheckHouseAvailable(ctx, params.HouseId, rentStart, rentEnd)
	if err != nil {
		return "", fmt.Errorf("检查房源可用性失败: %v", err)
	}
	if !available {
		return "", fmt.Errorf("该房源在指定时间段内不可用")
	}

	// 4. 生成订单号
	orderNo := o.generateOrderNo()

	// 5. 创建订单对象
	order := &model.RentalOrder{
		OrderNo:     orderNo,
		HouseId:     &params.HouseId,
		TenantId:    &params.TenantId,
		LandlordId:  &params.LandlordId,
		TenantPhone: params.TenantPhone,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
		RentAmount:  &params.RentAmount,
		Deposit:     &params.Deposit,
		Status:      model.OrderStatusPending,
		SignedAt:    nil, // 创建时未签约
	}

	// 6. 保存订单
	if err := o.repo.CreateOrder(ctx, order); err != nil {
		return "", fmt.Errorf("创建订单失败: %v", err)
	}

	//o.log.Infof("订单创建成功, 订单号: %s, 租客ID: %d, 房源ID: %d", orderNo, params.TenantId, params.HouseId)
	return orderNo, nil
}

// generateOrderNo 生成订单号
func (o *OrderUsecase) generateOrderNo() string {
	return fmt.Sprintf("RO%d", time.Now().UnixNano()/1000000)
}

// CancelOrder 取消订单
func (o *OrderUsecase) CancelOrder(ctx context.Context, id uint, cancelReason string) error {
	// 1. 参数验证
	if id == 0 {
		return fmt.Errorf("订单ID不能为空")
	}
	if cancelReason == "" {
		return fmt.Errorf("取消原因不能为空")
	}

	// 2. 获取订单详情
	order, err := o.repo.GetOrderDetail(ctx, id)
	if err != nil {
		return fmt.Errorf("获取订单详情失败: %v", err)
	}

	// 3. 检查订单状态是否可以取消
	if order.Status == model.OrderStatusCancelled {
		return fmt.Errorf("订单已经取消，无法重复取消")
	}
	if order.Status == model.OrderStatusCompleted {
		return fmt.Errorf("订单已完成，无法取消")
	}

	// 4. 执行取消操作
	if err := o.repo.CancelOrder(ctx, id, cancelReason); err != nil {
		return fmt.Errorf("取消订单失败: %v", err)
	}

	//o.log.Infof("订单取消成功, 订单ID: %d, 订单号: %s, 取消原因: %s", id, order.OrderNo, cancelReason)
	return nil
}
