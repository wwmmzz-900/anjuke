package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket连接管理器
type WebSocketManager struct {
	// 存储用户ID到WebSocket连接的映射
	connections map[int64]*websocket.Conn
	// 读写锁保护连接映射
	mutex sync.RWMutex
	// WebSocket升级器
	upgrader websocket.Upgrader
}

// 创建新的WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[int64]*websocket.Conn),
		upgrader: websocket.Upgrader{
			// 允许跨域请求
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			// 设置读写缓冲区大小
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// 发送消息给指定用户
func (wm *WebSocketManager) SendMessageToUser(userID int64, message interface{}) error {
	wm.mutex.RLock()
	conn, exists := wm.connections[userID]
	wm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("用户 %d 未连接", userID)
	}

	// 将消息转换为JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}

	// 发送消息
	if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
		// 发送失败，移除连接
		wm.removeConnection(userID)
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("向用户 %d 发送消息: %s", userID, string(messageJSON))
	return nil
}

// 添加连接
func (wm *WebSocketManager) addConnection(userID int64, conn *websocket.Conn) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	// 如果用户已有连接，先关闭旧连接
	if oldConn, exists := wm.connections[userID]; exists {
		oldConn.Close()
	}
	
	wm.connections[userID] = conn
}

// 移除连接
func (wm *WebSocketManager) removeConnection(userID int64) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	delete(wm.connections, userID)
}

// 获取在线用户数量
func (wm *WebSocketManager) GetOnlineUserCount() int {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	return len(wm.connections)
}

// 获取在线用户列表
func (wm *WebSocketManager) GetOnlineUsers() []int64 {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	users := make([]int64, 0, len(wm.connections))
	for userID := range wm.connections {
		users = append(users, userID)
	}
	return users
}