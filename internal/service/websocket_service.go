package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wwmmzz-900/anjuke/internal/model"
)

// WebSocketManager WebSocket连接管理器
type WebSocketManager struct {
	mutex            sync.RWMutex
	connections      map[int64]*websocket.Conn           // userID -> connection
	houseConnections map[int64]map[int64]*websocket.Conn // houseID -> userID -> connection
	sessionKeys      map[int64]string                    // userID -> sessionKey
	upgrader         websocket.Upgrader
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

// GetConnectionStats 获取连接统计信息
func (wm *WebSocketManager) GetConnectionStats() map[string]interface{} {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	stats := make(map[string]interface{})

	// 获取在线用户数量
	stats["total_connections"] = len(wm.connections)

	// 获取在线用户列表
	onlineUsers := make([]int64, 0, len(wm.connections))
	for userID := range wm.connections {
		onlineUsers = append(onlineUsers, userID)
	}
	stats["online_users"] = onlineUsers

	// 获取房源连接信息
	houseStats := make(map[string]interface{})
	for houseID, connections := range wm.houseConnections {
		houseStats[fmt.Sprintf("house_%d", houseID)] = len(connections)
	}
	stats["total_houses"] = len(wm.houseConnections)
	stats["houses"] = houseStats

	return stats
}

// =============================================================================
// WebSocket HTTP 处理器和业务逻辑
// =============================================================================

// WebSocketService WebSocket业务服务
type WebSocketService struct {
	manager *WebSocketManager
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService(manager *WebSocketManager) *WebSocketService {
	return &WebSocketService{
		manager: manager,
	}
}

// HandleHouseWS 处理房源WebSocket连接
func (ws *WebSocketService) HandleHouseWS(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	houseIDStr := r.URL.Query().Get("house_id")
	userIDStr := r.URL.Query().Get("user_id")

	if houseIDStr == "" || userIDStr == "" {
		http.Error(w, "缺少必要参数 house_id 或 user_id", http.StatusBadRequest)
		return
	}

	houseID, err := strconv.ParseInt(houseIDStr, 10, 64)
	if err != nil {
		http.Error(w, "无效的 house_id 参数", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "无效的 user_id 参数", http.StatusBadRequest)
		return
	}

	// 升级为 WebSocket 连接
	conn, err := ws.manager.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	log.Printf("WebSocket connected successfully: houseID=%d, userID=%d, remoteAddr=%s\n",
		houseID, userID, r.RemoteAddr)

	// 注册连接
	ws.manager.AddHouseConnection(houseID, userID, conn)

	// 生成会话密钥
	sessionKey := GenerateSessionKey()
	ws.manager.SetSessionKey(userID, sessionKey)

	// 发送连接成功消息
	welcomeMsg := map[string]interface{}{
		"type":        string(model.WSMessageTypeConnection),
		"message":     "WebSocket 连接成功",
		"house_id":    houseID,
		"user_id":     userID,
		"session_key": sessionKey,
		"sequence":    GlobalSequenceManager.GetNextSequence(userID),
		"timestamp":   time.Now().Unix(),
	}
	conn.WriteJSON(welcomeMsg)

	// 启动消息处理协程
	go ws.handleConnection(conn, houseID, userID)
}

// handleConnection 处理WebSocket连接的消息
func (ws *WebSocketService) handleConnection(conn *websocket.Conn, houseID, userID int64) {
	defer func() {
		log.Printf("WebSocket disconnected: houseID=%d, userID=%d\n", houseID, userID)
		ws.manager.RemoveConnection(userID)
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		log.Printf("收到原始消息: %s\n", string(message))

		// 解析并处理消息
		if err := ws.processMessage(conn, houseID, userID, message); err != nil {
			log.Printf("处理消息失败: %v\n", err)
			conn.WriteJSON(map[string]interface{}{
				"type":    "error",
				"message": "消息处理失败",
			})
		}
	}
}

// processMessage 处理接收到的消息
func (ws *WebSocketService) processMessage(conn *websocket.Conn, houseID, userID int64, message []byte) error {
	var msgData map[string]interface{}
	if err := json.Unmarshal(message, &msgData); err != nil {
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	log.Printf("解析JSON成功: %+v\n", msgData)

	// 处理不同类型的消息
	if to, hasTo := msgData["to"].(float64); hasTo {
		return ws.handleDirectMessage(conn, houseID, userID, msgData, int64(to))
	} else if action, ok := msgData["action"].(string); ok {
		return ws.handleActionMessage(conn, houseID, userID, msgData, action)
	} else {
		// 回显消息
		return ws.handleEchoMessage(conn, msgData)
	}
}

// handleDirectMessage 处理直接消息
func (ws *WebSocketService) handleDirectMessage(conn *websocket.Conn, houseID, userID int64, msgData map[string]interface{}, targetID int64) error {
	from, _ := msgData["from"].(float64)
	content, _ := msgData["message"].(string)
	encrypted, _ := msgData["encrypted"].(bool)
	sequence, hasSequence := msgData["sequence"].(float64)

	fromID := int64(from)
	if fromID == 0 {
		fromID = userID
	}

	// 获取或生成序列号
	var seq int64
	if hasSequence {
		seq = int64(sequence)
	} else {
		seq = GlobalSequenceManager.GetNextSequence(fromID)
	}

	log.Printf("用户 %d 向用户 %d 发送消息: %s (序列号: %d)\n", fromID, targetID, content, seq)

	// 处理消息加密
	messageContent := content
	if !encrypted {
		if sessionKey, exists := ws.manager.GetSessionKey(fromID); exists {
			key := GenerateKey(sessionKey)
			if encryptedContent, err := EncryptMessage([]byte(content), key); err == nil {
				messageContent = encryptedContent
				encrypted = true
			}
		}
	}

	// 发送确认消息给发送者
	conn.WriteJSON(map[string]interface{}{
		"type":      "system",
		"message":   "消息已发送",
		"sequence":  seq,
		"timestamp": time.Now().Unix(),
	})

	// 解密消息给接收者
	messageForReceiver := messageContent
	if encrypted {
		if sessionKey, exists := ws.manager.GetSessionKey(fromID); exists {
			key := GenerateKey(sessionKey)
			if decryptedContent, err := DecryptMessage(messageContent, key); err == nil {
				messageForReceiver = string(decryptedContent)
			} else {
				messageForReceiver = "[消息解密失败]"
			}
		} else {
			messageForReceiver = "[无法解密消息]"
		}
	}

	// 转发消息给接收者
	return ws.manager.SendMessageToUser(targetID, map[string]interface{}{
		"type":      "chat",
		"from":      fromID,
		"message":   messageForReceiver,
		"content":   messageForReceiver,
		"to":        targetID,
		"encrypted": false,
		"sequence":  seq,
		"timestamp": time.Now().Unix(),
	})
}

// handleActionMessage 处理动作消息
func (ws *WebSocketService) handleActionMessage(conn *websocket.Conn, houseID, userID int64, msgData map[string]interface{}, action string) error {
	switch action {
	case "login":
		return ws.handleLogin(conn, userID)
	case "message":
		return ws.handleChatMessage(conn, houseID, userID, msgData)
	default:
		return fmt.Errorf("未知动作: %s", action)
	}
}

// handleLogin 处理登录消息
func (ws *WebSocketService) handleLogin(conn *websocket.Conn, userID int64) error {
	log.Printf("用户 %d 登录\n", userID)
	return conn.WriteJSON(map[string]interface{}{
		"type":      "system",
		"message":   "登录成功",
		"user_id":   userID,
		"timestamp": time.Now().Unix(),
	})
}

// handleChatMessage 处理聊天消息
func (ws *WebSocketService) handleChatMessage(conn *websocket.Conn, houseID, userID int64, msgData map[string]interface{}) error {
	content, _ := msgData["content"].(string)
	toUserID, _ := msgData["to"].(float64)
	encrypted, _ := msgData["encrypted"].(bool)
	sequence, hasSequence := msgData["sequence"].(float64)

	if toUserID <= 0 {
		return conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "缺少接收者ID",
		})
	}

	targetID := int64(toUserID)

	// 获取或生成序列号
	var seq int64
	if hasSequence {
		seq = int64(sequence)
	} else {
		seq = GlobalSequenceManager.GetNextSequence(userID)
	}

	log.Printf("用户 %d 向用户 %d 发送消息: %s (序列号: %d)\n", userID, targetID, content, seq)

	// 处理消息加密/解密逻辑（与handleDirectMessage类似）
	messageContent := content
	if !encrypted {
		if sessionKey, exists := ws.manager.GetSessionKey(userID); exists {
			key := GenerateKey(sessionKey)
			if encryptedContent, err := EncryptMessage([]byte(content), key); err == nil {
				messageContent = encryptedContent
				encrypted = true
			}
		}
	}

	// 发送确认消息
	conn.WriteJSON(map[string]interface{}{
		"type":      "system",
		"message":   "消息已发送",
		"sequence":  seq,
		"timestamp": time.Now().Unix(),
	})

	// 解密并转发消息
	messageForReceiver := messageContent
	if encrypted {
		if sessionKey, exists := ws.manager.GetSessionKey(userID); exists {
			key := GenerateKey(sessionKey)
			if decryptedContent, err := DecryptMessage(messageContent, key); err == nil {
				messageForReceiver = string(decryptedContent)
			} else {
				messageForReceiver = "[消息解密失败]"
			}
		}
	}

	return ws.manager.SendMessageToUser(targetID, map[string]interface{}{
		"type":      "chat",
		"from":      userID,
		"message":   messageForReceiver,
		"content":   messageForReceiver,
		"encrypted": false,
		"sequence":  seq,
		"timestamp": time.Now().Unix(),
	})
}

// handleEchoMessage 处理回显消息
func (ws *WebSocketService) handleEchoMessage(conn *websocket.Conn, msgData map[string]interface{}) error {
	log.Println("消息没有action字段，回显")
	return conn.WriteJSON(map[string]interface{}{
		"type":      "echo",
		"data":      msgData,
		"timestamp": time.Now().Unix(),
	})
}

// =============================================================================
// HTTP 处理器函数（保持向后兼容）
// =============================================================================

// 全局WebSocket服务实例
var globalWebSocketService *WebSocketService

// 初始化全局WebSocket服务
func init() {
	if GlobalWebSocketManager != nil {
		globalWebSocketService = NewWebSocketService(GlobalWebSocketManager)
	}
}

// HandleHouseWS 全局WebSocket处理器（向后兼容）
func HandleHouseWS(w http.ResponseWriter, r *http.Request) {
	if globalWebSocketService == nil {
		globalWebSocketService = NewWebSocketService(GlobalWebSocketManager)
	}
	globalWebSocketService.HandleHouseWS(w, r)
}

// HandleWSStats HTTP处理器：获取WebSocket连接统计信息
func HandleWSStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := GlobalWebSocketManager.GetConnectionStats()

	// 添加序列号管理器统计信息
	sequenceStats := make(map[string]interface{})
	GlobalSequenceManager.mutex.RLock()
	sequenceStats["total_users"] = len(GlobalSequenceManager.sequences)
	GlobalSequenceManager.mutex.RUnlock()
	stats["sequences"] = sequenceStats

	jsonData, err := json.Marshal(map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": stats,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"msg":"内部服务器错误","data":null}`))
		return
	}

	w.Write(jsonData)
}

// HandleWSTestPage HTTP处理器：提供WebSocket测试页面
func HandleWSTestPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/websocket_test.html")
}

// HandleSecureChatPage HTTP处理器：提供安全聊天页面
func HandleSecureChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/secure_chat.html")
}

// =============================================================================
// 辅助函数（向后兼容）
// =============================================================================

// pushToUser 推送消息给指定用户（向后兼容）
func pushToUser(houseID, userID int64, msg interface{}) {
	err := GlobalWebSocketManager.SendMessageToUser(userID, msg)
	if err != nil {
		log.Printf("推送消息给用户%d失败: %v\n", userID, err)
	} else {
		log.Printf("成功推送消息给用户%d: %+v\n", userID, msg)
	}
}

// pushToAll 推送消息给所有正在查看该房源的用户（向后兼容）
func pushToAll(houseID int64, msg interface{}) {
	users := GlobalWebSocketManager.GetHouseUsers(houseID)

	for _, userID := range users {
		err := GlobalWebSocketManager.SendMessageToUser(userID, msg)
		if err != nil {
			log.Printf("推送消息给房源%d的用户%d失败: %v\n", houseID, userID, err)
		}
	}

	log.Printf("向房源%d的%d个用户推送了消息\n", houseID, len(users))
}

// GetConnectionStats 获取连接状态信息（向后兼容）
func GetConnectionStats() map[string]interface{} {
	return GlobalWebSocketManager.GetConnectionStats()
}