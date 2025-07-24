package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketManager WebSocket连接管理器
type WebSocketManager struct {
	mutex           sync.RWMutex
	connections     map[int64]*websocket.Conn    // userID -> connection
	houseConnections map[int64]map[int64]*websocket.Conn // houseID -> userID -> connection
	sessionKeys     map[int64]string             // userID -> sessionKey
	upgrader        websocket.Upgrader
}

// NewWebSocketManager 创建新的WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections:      make(map[int64]*websocket.Conn),
		houseConnections: make(map[int64]map[int64]*websocket.Conn),
		sessionKeys:      make(map[int64]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// AddConnection 添加用户连接
func (wm *WebSocketManager) AddConnection(userID int64, conn *websocket.Conn) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	// 如果用户已有连接，先关闭旧连接
	if oldConn, exists := wm.connections[userID]; exists {
		oldConn.Close()
	}
	
	wm.connections[userID] = conn
	log.Printf("用户 %d 的WebSocket连接已添加", userID)
}

// AddHouseConnection 添加房源相关的用户连接
func (wm *WebSocketManager) AddHouseConnection(houseID, userID int64, conn *websocket.Conn) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	// 添加到全局连接池
	if oldConn, exists := wm.connections[userID]; exists {
		oldConn.Close()
	}
	wm.connections[userID] = conn
	
	// 添加到房源连接池
	if wm.houseConnections[houseID] == nil {
		wm.houseConnections[houseID] = make(map[int64]*websocket.Conn)
	}
	wm.houseConnections[houseID][userID] = conn
	
	log.Printf("用户 %d 的房源 %d WebSocket连接已添加", userID, houseID)
}

// RemoveConnection 移除用户连接
func (wm *WebSocketManager) RemoveConnection(userID int64) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	// 从全局连接池移除
	if conn, exists := wm.connections[userID]; exists {
		conn.Close()
		delete(wm.connections, userID)
	}
	
	// 从房源连接池移除
	for houseID, houseConns := range wm.houseConnections {
		if _, exists := houseConns[userID]; exists {
			delete(houseConns, userID)
			// 如果房源没有其他连接，删除房源条目
			if len(houseConns) == 0 {
				delete(wm.houseConnections, houseID)
			}
		}
	}
	
	// 移除会话密钥
	delete(wm.sessionKeys, userID)
	
	log.Printf("用户 %d 的WebSocket连接已移除", userID)
}

// SendMessageToUser 向指定用户发送消息
func (wm *WebSocketManager) SendMessageToUser(userID int64, message interface{}) error {
	wm.mutex.RLock()
	conn, exists := wm.connections[userID]
	wm.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("用户 %d 未连接", userID)
	}
	
	// 将消息转换为JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}
	
	// 发送消息
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		// 如果发送失败，移除连接
		wm.RemoveConnection(userID)
		return fmt.Errorf("发送消息失败: %v", err)
	}
	
	return nil
}

// SendMessageToHouse 向房源的所有用户发送消息
func (wm *WebSocketManager) SendMessageToHouse(houseID int64, message interface{}) error {
	wm.mutex.RLock()
	houseConns, exists := wm.houseConnections[houseID]
	if !exists {
		wm.mutex.RUnlock()
		return fmt.Errorf("房源 %d 没有连接的用户", houseID)
	}
	
	// 复制连接映射以避免在锁内进行网络操作
	connsCopy := make(map[int64]*websocket.Conn)
	for userID, conn := range houseConns {
		connsCopy[userID] = conn
	}
	wm.mutex.RUnlock()
	
	// 将消息转换为JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}
	
	// 向所有用户发送消息
	var failedUsers []int64
	for userID, conn := range connsCopy {
		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Printf("向用户 %d 发送消息失败: %v", userID, err)
			failedUsers = append(failedUsers, userID)
		}
	}
	
	// 移除失败的连接
	for _, userID := range failedUsers {
		wm.RemoveConnection(userID)
	}
	
	return nil
}

// GetHouseUsers 获取房源的所有在线用户
func (wm *WebSocketManager) GetHouseUsers(houseID int64) []int64 {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	houseConns, exists := wm.houseConnections[houseID]
	if !exists {
		return []int64{}
	}
	
	users := make([]int64, 0, len(houseConns))
	for userID := range houseConns {
		users = append(users, userID)
	}
	
	return users
}

// GetOnlineUsers 获取所有在线用户
func (wm *WebSocketManager) GetOnlineUsers() []int64 {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	users := make([]int64, 0, len(wm.connections))
	for userID := range wm.connections {
		users = append(users, userID)
	}
	
	return users
}

// GetOnlineUserCount 获取在线用户数量
func (wm *WebSocketManager) GetOnlineUserCount() int {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	return len(wm.connections)
}

// IsUserOnline 检查用户是否在线
func (wm *WebSocketManager) IsUserOnline(userID int64) bool {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	_, exists := wm.connections[userID]
	return exists
}

// SetSessionKey 设置用户会话密钥
func (wm *WebSocketManager) SetSessionKey(userID int64, sessionKey string) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	
	wm.sessionKeys[userID] = sessionKey
}

// GetSessionKey 获取用户会话密钥
func (wm *WebSocketManager) GetSessionKey(userID int64) (string, bool) {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	
	key, exists := wm.sessionKeys[userID]
	return key, exists
}

// BroadcastMessage 向所有在线用户广播消息
func (wm *WebSocketManager) BroadcastMessage(message interface{}) error {
	wm.mutex.RLock()
	connsCopy := make(map[int64]*websocket.Conn)
	for userID, conn := range wm.connections {
		connsCopy[userID] = conn
	}
	wm.mutex.RUnlock()
	
	// 将消息转换为JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}
	
	// 向所有用户发送消息
	var failedUsers []int64
	for userID, conn := range connsCopy {
		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Printf("向用户 %d 广播消息失败: %v", userID, err)
			failedUsers = append(failedUsers, userID)
		}
	}
	
	// 移除失败的连接
	for _, userID := range failedUsers {
		wm.RemoveConnection(userID)
	}
	
	return nil
}