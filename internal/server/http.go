package server

import (
	"net/http" // 标准库

	v6 "github.com/wwmmzz-900/anjuke/api/customer/v6"
	v1 "github.com/wwmmzz-900/anjuke/api/helloworld/v1"
	v3 "github.com/wwmmzz-900/anjuke/api/house/v3"
	v5 "github.com/wwmmzz-900/anjuke/api/points/v5"
	v4 "github.com/wwmmzz-900/anjuke/api/transaction/v4"
	v2 "github.com/wwmmzz-900/anjuke/api/user/v2"
	"github.com/wwmmzz-900/anjuke/internal/conf"
	"github.com/wwmmzz-900/anjuke/internal/service"

	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

var wsHub = service.WsHub // 全局Hub

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

	// 在 HTTP Server 启动前注册 WebSocket 路由
	// 注册 WebSocket 路由（用 net/http）
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil || userID <= 0 {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		wsHub.HandleWS(w, r, userID)
	})

	return srv
}
