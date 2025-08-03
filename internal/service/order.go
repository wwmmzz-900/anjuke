package service

import (
	pb "anjuke/api/order/v1"
	"anjuke/internal/biz"
	"anjuke/internal/model/params"
	"context"
	"fmt"
)

type OrderService struct {
	pb.UnimplementedOrderServer
	uc *biz.OrderUsecase
}

func NewOrderService(uc *biz.OrderUsecase) *OrderService {
	return &OrderService{uc: uc}
}

func (s *OrderService) GetTenantOrderList(ctx context.Context, req *pb.GetTenantOrderListRequest) (*pb.GetTenantOrderListReply, error) {
	tenantId := req.GetTenantId()
	orderList, count, err := s.uc.GetTenantOrderList(ctx, uint(tenantId), int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var result []*pb.OrderInfo
	for _, order := range *orderList {
		orderInfo := &pb.OrderInfo{
			OrderNo:      order.OrderNo,
			HouseId:      uint32(*order.HouseId),
			TenantId:     uint32(*order.TenantId),
			LandlordId:   uint32(*order.LandlordId),
			TenantPhone:  order.TenantPhone,
			RentStart:    order.RentStart.Format("2006-01-02"),
			RentEnd:      order.RentEnd.Format("2006-01-02"),
			RentAmount:   *order.RentAmount,
			Deposit:      *order.Deposit,
			Status:       string(order.Status),
			CreatedAt:    order.CreatedAt.Format("2006-01-02"),
			CancelReason: order.CancelReason,
		}

		if order.SignedAt != nil {
			orderInfo.SignedAt = order.SignedAt.Format("2006-01-02")
		}
		if order.CancelledAt != nil {
			orderInfo.CancelledAt = order.CancelledAt.Format("2006-01-02 15:04:05")
		}

		result = append(result, orderInfo)
	}

	return &pb.GetTenantOrderListReply{
		List:     result,
		Total:    count,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	}, nil
}
func (s *OrderService) GetOrderDetail(ctx context.Context, req *pb.GetOrderDetailRequest) (*pb.GetOrderDetailReply, error) {
	order, err := s.uc.GetOrderDetail(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	orderInfo := &pb.OrderInfo{
		OrderNo:      order.OrderNo,
		HouseId:      uint32(*order.HouseId),
		TenantId:     uint32(*order.TenantId),
		LandlordId:   uint32(*order.LandlordId),
		TenantPhone:  order.TenantPhone,
		RentStart:    order.RentStart.Format("2006-01-02 15:04:05"),
		RentEnd:      order.RentEnd.Format("2006-01-02 15:04:05"),
		RentAmount:   *order.RentAmount,
		Deposit:      *order.Deposit,
		Status:       string(order.Status),
		CreatedAt:    order.CreatedAt.Format("2006-01-02 15:04:05"),
		CancelReason: order.CancelReason,
	}

	if order.SignedAt != nil {
		orderInfo.SignedAt = order.SignedAt.Format("2006-01-02 15:04:05")
	}
	if order.CancelledAt != nil {
		orderInfo.CancelledAt = order.CancelledAt.Format("2006-01-02 15:04:05")
	} // 注意：这里需要处理可能的空指针异常

	return &pb.GetOrderDetailReply{
		Order: orderInfo,
	}, nil
}
func (s *OrderService) GetOrderList(ctx context.Context, req *pb.GetOrderListRequest) (*pb.GetOrderListReply, error) {
	return &pb.GetOrderListReply{}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderReply, error) {
	orderNo, err := s.uc.CreateOrder(ctx, &params.CreateOrderParams{
		HouseId:     uint(req.GetHouseId()),
		TenantId:    uint(req.GetTenantId()),
		LandlordId:  uint(req.GetLandlordId()),
		TenantPhone: req.GetTenantPhone(),
		RentStart:   req.GetRentStart(),
		RentEnd:     req.GetRentEnd(),
		RentAmount:  req.GetRentAmount(),
		Deposit:     req.GetDeposit(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderReply{
		OrderNo: orderNo,
		Message: "订单创建成功",
	}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderReply, error) {
	err := s.uc.CancelOrder(ctx, uint(req.GetId()), req.GetCancelReason())
	if err != nil {
		return &pb.CancelOrderReply{
			Message: fmt.Sprintf("取消订单失败: %v", err),
			Success: false,
		}, nil
	}

	return &pb.CancelOrderReply{
		Message: "订单取消成功",
		Success: true,
	}, nil
}
