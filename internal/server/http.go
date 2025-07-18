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
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, house *service.HouseService, transaction *service.TransactionService, points *service.PointsService, customer *service.CustomerService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	v2.RegisterUserHTTPServer(srv, user)
	v3.RegisterHouseHTTPServer(srv, house)
	v4.RegisterTransactionHTTPServer(srv, transaction)
	v5.RegisterPointsHTTPServer(srv, points)
	v6.RegisterCustomerHTTPServer(srv, customer)
	return srv
}
