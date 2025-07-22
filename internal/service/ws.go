package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 房源时段预约池
var houseTimeSlotPool = struct {
	sync.RWMutex
	data map[int64]map[string]int64 // houseID -> timeSlot(如"2024-05-11T14:00") -> userID
}{data: make(map[int64]map[string]int64)}

// 房源WebSocket连接池
var houseWSConnPool = struct {
	sync.RWMutex
	data map[int64]map[int64]*websocket.Conn // houseID -> userID -> conn
	keys map[int64]string                    // userID -> 加密密钥
}{
	data: make(map[int64]map[int64]*websocket.Conn),
	keys: make(map[int64]string),
}

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 注册WebSocket连接（用户进入房源详情页时调用）
func HandleHouseWS(w http.ResponseWriter, r *http.Request) {
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	fmt.Printf("WebSocket connected successfully: houseID=%d, userID=%d, remoteAddr=%s\n",
		houseID, userID, r.RemoteAddr)

	// 注册连接
	houseWSConnPool.Lock()
	if houseWSConnPool.data[houseID] == nil {
		houseWSConnPool.data[houseID] = make(map[int64]*websocket.Conn)
	}
	houseWSConnPool.data[houseID][userID] = conn
	houseWSConnPool.Unlock()

	fmt.Printf("新连接: houseID=%d, userID=%d, 当前连接池: %+v\n", houseID, userID, houseWSConnPool.data)

	fmt.Printf("当前连接池: %+v\n", houseWSConnPool.data)

	// 生成会话密钥
	sessionKey := GenerateSessionKey()
	
	// 存储用户的会话密钥
	houseWSConnPool.Lock()
	houseWSConnPool.keys[userID] = sessionKey
	houseWSConnPool.Unlock()
	
	// 发送连接成功消息，包含会话密钥
	welcomeMsg := map[string]interface{}{
		"type":        "connection",
		"message":     "WebSocket 连接成功",
		"house_id":    houseID,
		"user_id":     userID,
		"session_key": sessionKey,
		"sequence":    GlobalSequenceManager.GetNextSequence(userID),
		"timestamp":   time.Now().Unix(),
	}
	conn.WriteJSON(welcomeMsg)

	// 监听连接关闭和消息
	go func() {
		defer func() {
			fmt.Printf("WebSocket disconnected: houseID=%d, userID=%d\n", houseID, userID)
			houseWSConnPool.Lock()
			if houseWSConnPool.data[houseID] != nil {
				delete(houseWSConnPool.data[houseID], userID)
				// 如果该房源没有其他连接，删除房源条目
				if len(houseWSConnPool.data[houseID]) == 0 {
					delete(houseWSConnPool.data, houseID)
				}
			}
			houseWSConnPool.Unlock()
			conn.Close()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket error: %v\n", err)
				}
				break
			}

			fmt.Printf("收到原始消息: %s\n", string(message))
			
			// 尝试解析JSON消息
			var msgData map[string]interface{}
			if err := json.Unmarshal(message, &msgData); err == nil {
				// 成功解析为JSON
				fmt.Printf("解析JSON成功: %+v\n", msgData)
				
				// 处理不同格式的消息
				if to, hasTo := msgData["to"].(float64); hasTo {
					// ApiPost格式的消息: {"from": 2, "to": 1, "message": "你好"}
					from, _ := msgData["from"].(float64)
					content, _ := msgData["message"].(string)
					encrypted, _ := msgData["encrypted"].(bool)
					sequence, hasSequence := msgData["sequence"].(float64)
					
					targetID := int64(to)
					fromID := int64(from)
					if fromID == 0 {
						fromID = userID // 如果没有指定发送者，使用当前用户ID
					}
					
					// 获取或生成序列号
					var seq int64
					if hasSequence {
						seq = int64(sequence)
					} else {
						seq = GlobalSequenceManager.GetNextSequence(fromID)
					}
					
					fmt.Printf("ApiPost格式: 用户 %d 向用户 %d 发送消息: %s (序列号: %d)\n", 
						fromID, targetID, content, seq)
					
					// 如果消息已加密，则不需要再次加密
					if !encrypted {
						// 获取发送者的密钥
						houseWSConnPool.RLock()
						senderKey, hasSenderKey := houseWSConnPool.keys[fromID]
						houseWSConnPool.RUnlock()
						
						// 如果有密钥，则加密消息
						if hasSenderKey {
							key := GenerateKey(senderKey)
							encryptedContent, err := EncryptMessage([]byte(content), key)
							if err == nil {
								content = encryptedContent
								encrypted = true
							} else {
								fmt.Printf("加密消息失败: %v\n", err)
							}
						}
					}
					
					// 发送确认消息给发送者
					conn.WriteJSON(map[string]interface{}{
						"type": "system",
						"message": "消息已发送",
						"sequence": seq,
						"timestamp": time.Now().Unix(),
					})
					
					// 转发消息给接收者
					pushToUser(houseID, targetID, map[string]interface{}{
						"type": "chat",
						"from": fromID,
						"message": content,
						"to": targetID,
						"encrypted": encrypted,
						"sequence": seq,
						"timestamp": time.Now().Unix(),
					})
				} else if action, ok := msgData["action"].(string); ok {
					// 处理特定动作
					switch action {
					case "login":
						// 处理登录消息
						fmt.Printf("用户 %d 登录\n", userID)
						conn.WriteJSON(map[string]interface{}{
							"type": "system",
							"message": "登录成功",
							"user_id": userID,
							"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
						})
						
					case "message":
						// 处理聊天消息
						content, _ := msgData["content"].(string)
						toUserID, _ := msgData["to"].(float64)
						encrypted, _ := msgData["encrypted"].(bool)
						sequence, hasSequence := msgData["sequence"].(float64)
						
						if toUserID > 0 {
							targetID := int64(toUserID)
							
							// 获取或生成序列号
							var seq int64
							if hasSequence {
								seq = int64(sequence)
							} else {
								seq = GlobalSequenceManager.GetNextSequence(userID)
							}
							
							fmt.Printf("用户 %d 向用户 %d 发送消息: %s (序列号: %d)\n", 
								userID, targetID, content, seq)
							
							// 如果消息未加密，尝试加密
							if !encrypted {
								// 获取发送者的密钥
								houseWSConnPool.RLock()
								senderKey, hasSenderKey := houseWSConnPool.keys[userID]
								houseWSConnPool.RUnlock()
								
								// 如果有密钥，则加密消息
								if hasSenderKey {
									key := GenerateKey(senderKey)
									encryptedContent, err := EncryptMessage([]byte(content), key)
									if err == nil {
										content = encryptedContent
										encrypted = true
									} else {
										fmt.Printf("加密消息失败: %v\n", err)
									}
								}
							}
							
							// 发送确认消息给发送者
							conn.WriteJSON(map[string]interface{}{
								"type": "system",
								"message": "消息已发送",
								"sequence": seq,
								"timestamp": time.Now().Unix(),
							})
							
							// 转发消息给接收者
							pushToUser(houseID, targetID, map[string]interface{}{
								"type": "chat",
								"from": userID,
								"content": content,
								"encrypted": encrypted,
								"sequence": seq,
								"timestamp": time.Now().Unix(),
							})
						} else {
							conn.WriteJSON(map[string]interface{}{
								"type": "error",
								"message": "缺少接收者ID",
							})
						}
						
					default:
						// 未知动作
						fmt.Printf("未知动作: %s\n", action)
						conn.WriteJSON(map[string]interface{}{
							"type": "error",
							"message": "未知动作",
						})
					}
				} else {
					// 没有action字段，简单回显消息
					fmt.Println("消息没有action字段，回显")
					conn.WriteJSON(map[string]interface{}{
						"type": "echo",
						"data": msgData,
						"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
					})
				}
			} else {
				// JSON解析失败
				fmt.Printf("JSON解析失败: %v\n", err)
				conn.WriteJSON(map[string]interface{}{
					"type": "error",
					"message": "消息格式错误，请发送有效的JSON",
					"original": string(message),
				})
			}
		}
	}()
}

// 推送消息给指定用户
func pushToUser(houseID, userID int64, msg interface{}) {
	houseWSConnPool.RLock()
	defer houseWSConnPool.RUnlock()
	
	// 首先尝试在同一个房源中查找用户
	if connections, exists := houseWSConnPool.data[houseID]; exists && connections[userID] != nil {
		conn := connections[userID]
		err := conn.WriteJSON(msg)
		fmt.Printf("在同一房源中推送给用户%d: %+v, err=%v\n", userID, msg, err)
		return
	}
	
	// 如果在同一房源中找不到，则在所有房源中查找
	for currentHouseID, connections := range houseWSConnPool.data {
		if conn, exists := connections[userID]; exists && conn != nil {
			err := conn.WriteJSON(msg)
			fmt.Printf("在房源%d中找到用户%d并推送消息: %+v, err=%v\n", 
				currentHouseID, userID, msg, err)
			return
		}
	}
	
	fmt.Printf("用户%d未连接，无法推送\n", userID)
}

// 推送消息给所有正在查看该房源的用户
func pushToAll(houseID int64, msg interface{}) {
	houseWSConnPool.RLock()
	defer houseWSConnPool.RUnlock()

	if connections, exists := houseWSConnPool.data[houseID]; exists {
		for userID, conn := range connections {
			if err := conn.WriteJSON(msg); err != nil {
				fmt.Printf("Failed to send message to user %d: %v\n", userID, err)
			}
		}
	}
}

// 获取连接状态信息（用于调试）
func GetConnectionStats() map[string]interface{} {
	houseWSConnPool.RLock()
	defer houseWSConnPool.RUnlock()

	stats := make(map[string]interface{})
	totalConnections := 0

	for houseID, connections := range houseWSConnPool.data {
		stats[fmt.Sprintf("house_%d", houseID)] = len(connections)
		totalConnections += len(connections)
	}

	stats["total_connections"] = totalConnections
	stats["total_houses"] = len(houseWSConnPool.data)

	return stats
}

// HTTP 处理器：获取 WebSocket 连接统计信息
func HandleWSStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := GetConnectionStats()

	// 简单的 JSON 编码
	fmt.Fprintf(w, `{"code":0,"msg":"success","data":%v}`, stats)
}

// HTTP 处理器：提供 WebSocket 测试页面
func HandleWSTestPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/websocket_test.html")
}

// HTTP 处理器：提供安全聊天页面
func HandleSecureChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/secure_chat.html")
}
