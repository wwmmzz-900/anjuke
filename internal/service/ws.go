package service

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 生产环境请加强校验
}

// WebSocketHub 管理所有用户的WebSocket连接
type WebSocketHub struct {
	clients map[int64]*websocket.Conn // 用户ID到连接的映射
	mu      sync.Mutex
}

// 新建Hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients: make(map[int64]*websocket.Conn),
	}
}

// 处理WebSocket连接
func (hub *WebSocketHub) HandleWS(w http.ResponseWriter, r *http.Request, userID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	hub.mu.Lock()
	hub.clients[userID] = conn
	hub.mu.Unlock()

	// 监听关闭
	go func() {
		defer func() {
			hub.mu.Lock()
			delete(hub.clients, userID)
			hub.mu.Unlock()
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// 推送消息给指定用户
func (hub *WebSocketHub) SendToUser(userID int64, message string) error {
	hub.mu.Lock()
	conn, ok := hub.clients[userID]
	hub.mu.Unlock()
	if !ok {
		return nil // 用户不在线可忽略或做离线处理
	}
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}
