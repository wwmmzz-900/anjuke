package server

import (
	v6 "anjuke/api/customer/v6"
	v1 "anjuke/api/helloworld/v1"
	v3 "anjuke/api/house/v3"
	v5 "anjuke/api/points/v5"
	v4 "anjuke/api/transaction/v4"
	v2 "anjuke/api/user/v2"
	"anjuke/internal/conf"
	"anjuke/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, house *service.HouseService, transaction *service.TransactionService, points *service.PointsService, customer *service.CustomerService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterGreeterServer(srv, greeter)
	v2.RegisterUserServer(srv, user)
	v3.RegisterHouseServer(srv, house)
	v4.RegisterTransactionServer(srv, transaction)
	v5.RegisterPointsServer(srv, points)
	v6.RegisterCustomerServer(srv, customer)
	return srv
}
