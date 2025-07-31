package server

import (
	companyv1 "anjuke/server/api/company/v1"
	v1 "anjuke/server/api/helloworld/v1"
	v5 "anjuke/server/api/points/v5"
	v2 "anjuke/server/api/user/v2"
	"anjuke/server/internal/conf"
	"anjuke/server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, points *service.PointsService, company *service.CompanyService, store *service.StoreService, realtor *service.RealtorService, logger log.Logger) *grpc.Server {
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
	v5.RegisterPointsServer(srv, points)
	companyv1.RegisterCompanyServer(srv, company)
	companyv1.RegisterStoreServer(srv, store)
	companyv1.RegisterRealtorServer(srv, realtor)
	return srv
}
