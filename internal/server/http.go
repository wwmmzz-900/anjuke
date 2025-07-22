package server

import (
	v6 "github.com/wwmmzz-900/anjuke/api/customer/v6"
	v1 "github.com/wwmmzz-900/anjuke/api/helloworld/v1"
	v3 "github.com/wwmmzz-900/anjuke/api/house/v3"
	v5 "github.com/wwmmzz-900/anjuke/api/points/v5"
	v4 "github.com/wwmmzz-900/anjuke/api/transaction/v4"
	v2 "github.com/wwmmzz-900/anjuke/api/user/v2"
	"github.com/wwmmzz-900/anjuke/internal/conf"
	"github.com/wwmmzz-900/anjuke/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, house *service.HouseService, transaction *service.TransactionService, points *service.PointsService, customer *service.CustomerService, logger log.Logger) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratoshttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratoshttp.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	v2.RegisterUserHTTPServer(srv, user)
	v3.RegisterHouseHTTPServer(srv, house)
	v4.RegisterTransactionHTTPServer(srv, transaction)
	v5.RegisterPointsHTTPServer(srv, points)
	v6.RegisterCustomerHTTPServer(srv, customer)

	// 注册 WebSocket 路由到 Kratos HTTP 服务器
	srv.HandleFunc("/ws/house", service.HandleHouseWS)
	// 注册 WebSocket 统计信息路由
	srv.HandleFunc("/api/websocket/stats", service.HandleWSStats)
	// 注册 WebSocket 测试页面路由
	srv.HandleFunc("/websocket-test", service.HandleWSTestPage)
	// 注册安全聊天页面路由
	srv.HandleFunc("/secure-chat", service.HandleSecureChatPage)

	return srv
}

// 注意：WebSocket 处理逻辑已移至 service 包
