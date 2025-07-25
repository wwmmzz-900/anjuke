package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	v6 "github.com/wwmmzz-900/anjuke/api/customer/v6"
	v1 "github.com/wwmmzz-900/anjuke/api/helloworld/v1"
	v3 "github.com/wwmmzz-900/anjuke/api/house/v3"
	v5 "github.com/wwmmzz-900/anjuke/api/points/v5"
	v4 "github.com/wwmmzz-900/anjuke/api/transaction/v4"
	v2 "github.com/wwmmzz-900/anjuke/api/user/v2"
	"github.com/wwmmzz-900/anjuke/internal/conf"
	"github.com/wwmmzz-900/anjuke/internal/model"
	"github.com/wwmmzz-900/anjuke/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServerNoAuth creates a new HTTP server without authentication middleware.
func NewHTTPServerNoAuth(c *conf.Server, greeter *service.GreeterService, user *service.UserService, house *service.HouseService, transaction *service.TransactionService, points *service.PointsService, customer *service.CustomerService, logger log.Logger) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(
			recovery.Recovery(),
		),
		// 添加请求解码器选项，支持更多Content-Type
		kratoshttp.RequestDecoder(func(r *kratoshttp.Request, v interface{}) error {
			// 设置默认Content-Type为application/json
			if r.Header.Get("Content-Type") == "" {
				r.Header.Set("Content-Type", "application/json")
			}
			return kratoshttp.DefaultRequestDecoder(r, v)
		}),
		kratoshttp.ResponseEncoder(func(w kratoshttp.ResponseWriter, r *kratoshttp.Request, v interface{}) error {
			// 设置响应Content-Type
			w.Header().Set("Content-Type", "application/json")
			return kratoshttp.DefaultResponseEncoder(w, r, v)
		}),
	}

	// 设置网络和地址
	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		// 使用不同的端口，避免与原服务冲突
		opts = append(opts, kratoshttp.Address("0.0.0.0:8001"))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := kratoshttp.NewServer(opts...)

	// 注册所有服务
	v1.RegisterGreeterHTTPServer(srv, greeter)
	v2.RegisterUserHTTPServer(srv, user)
	v3.RegisterHouseHTTPServer(srv, house)
	v4.RegisterTransactionHTTPServer(srv, transaction)
	v5.RegisterPointsHTTPServer(srv, points)
	v6.RegisterCustomerHTTPServer(srv, customer)

	// 注册 WebSocket 路由
	srv.HandleFunc("/ws/house", service.HandleHouseWS)
	srv.HandleFunc("/api/websocket/stats", service.HandleWSStats)
	srv.HandleFunc("/websocket-test", service.HandleWSTestPage)
	srv.HandleFunc("/secure-chat", service.HandleSecureChatPage)

	// 添加聊天相关API路由
	srv.HandleFunc("/api/chat/messages", handleChatMessages)
	srv.HandleFunc("/api/chat/mark-read", handleMarkMessagesAsRead)
	srv.HandleFunc("/api/chat/unread-count", handleUnreadMessageCount)

	// 添加一个健康检查路由
	srv.HandleFunc("/health", func(w kratoshttp.ResponseWriter, r *kratoshttp.Request) {
		w.Write([]byte("OK"))
	})

	log.NewHelper(logger).Info("HTTP server without authentication started on port 8001")

	return srv
}

// 处理获取聊天消息列表请求
func handleChatMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取参数
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"msg":"缺少chat_id参数","data":null}`))
		return
	}

	// 获取分页参数
	page := model.DefaultPage
	pageSize := model.DefaultPageSize

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 使用分页参数（避免未使用变量警告）
	_ = page
	_ = pageSize

	// 调用聊天服务获取消息列表
	// 注意：这里需要获取ChatService实例，实际项目中应该通过依赖注入获取
	// 这里简化处理，直接返回示例数据

	// 构造响应
	response := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": map[string]interface{}{
			"total": 10,
			"list": []map[string]interface{}{
				{
					"id":            1,
					"chat_id":       chatID,
					"sender_id":     1001,
					"sender_name":   "张三",
					"receiver_id":   2001,
					"receiver_name": "李四",
					"type":          0,
					"content":       "您好，我想预约看房",
					"read":          true,
					"created_at":    time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
				},
				{
					"id":            2,
					"chat_id":       chatID,
					"sender_id":     2001,
					"sender_name":   "李四",
					"receiver_id":   1001,
					"receiver_name": "张三",
					"type":          0,
					"content":       "好的，请问您方便什么时间？",
					"read":          true,
					"created_at":    time.Now().Add(-50 * time.Minute).Format("2006-01-02 15:04:05"),
				},
			},
		},
	}

	// 返回JSON响应
	jsonData, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"msg":"内部服务器错误","data":null}`))
		return
	}

	w.Write(jsonData)
}

// 处理标记消息为已读请求
func handleMarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 只接受POST请求
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"code":405,"msg":"方法不允许","data":null}`))
		return
	}

	// 解析请求体
	var req struct {
		ChatID     string `json:"chat_id"`
		ReceiverID int64  `json:"receiver_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"msg":"无效的请求体","data":null}`))
		return
	}

	// 参数验证
	if req.ChatID == "" || req.ReceiverID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"msg":"缺少必要参数","data":null}`))
		return
	}

	// 调用聊天服务标记消息为已读
	// 注意：这里需要获取ChatService实例，实际项目中应该通过依赖注入获取
	// 这里简化处理，直接返回成功

	// 构造响应
	response := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": map[string]interface{}{
			"success": true,
		},
	}

	// 返回JSON响应
	jsonData, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"msg":"内部服务器错误","data":null}`))
		return
	}

	w.Write(jsonData)
}

// 处理获取未读消息数量请求
func handleUnreadMessageCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取参数
	receiverIDStr := r.URL.Query().Get("receiver_id")
	if receiverIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"msg":"缺少receiver_id参数","data":null}`))
		return
	}

	receiverID, err := strconv.ParseInt(receiverIDStr, 10, 64)
	if err != nil || receiverID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"msg":"无效的receiver_id参数","data":null}`))
		return
	}

	// 调用聊天服务获取未读消息数量
	// 注意：这里需要获取ChatService实例，实际项目中应该通过依赖注入获取
	// 这里简化处理，直接返回示例数据

	// 构造响应
	response := map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": map[string]interface{}{
			"count": 5,
		},
	}

	// 返回JSON响应
	jsonData, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"msg":"内部服务器错误","data":null}`))
		return
	}

	w.Write(jsonData)
}
