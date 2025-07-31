package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
)

// ProgressHub 管理所有WebSocket连接和上传进度
type ProgressHub struct {
	// 已注册的连接
	connections map[string][]*websocket.Conn
	// 每个上传ID的进度
	progress map[string]float64
	// 互斥锁
	mu sync.Mutex
	// 日志
	log *log.Helper
}

// 全局进度管理器
var GlobalProgressHub *ProgressHub

// 初始化全局进度管理器
func InitProgressHub(logger log.Logger) {
	GlobalProgressHub = &ProgressHub{
		connections: make(map[string][]*websocket.Conn),
		progress:    make(map[string]float64),
		log:         log.NewHelper(logger),
	}

	// 启动清理协程
	go GlobalProgressHub.cleanupRoutine()
}

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler 处理WebSocket连接请求
func (s *Server) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// 获取上传ID
	uploadID := r.URL.Query().Get("uploadID")
	if uploadID == "" {
		http.Error(w, "Missing uploadID", http.StatusBadRequest)
		return
	}

	s.log.Infof("WebSocket连接请求: uploadID=%s", uploadID)

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// 注册连接
	GlobalProgressHub.mu.Lock()
	if _, exists := GlobalProgressHub.connections[uploadID]; !exists {
		GlobalProgressHub.connections[uploadID] = make([]*websocket.Conn, 0)
		GlobalProgressHub.progress[uploadID] = 0
	}
	GlobalProgressHub.connections[uploadID] = append(GlobalProgressHub.connections[uploadID], conn)
	currentProgress := GlobalProgressHub.progress[uploadID]
	GlobalProgressHub.mu.Unlock()

	s.log.Infof("WebSocket连接建立: uploadID=%s", uploadID)

	// 发送当前进度
	sendProgressUpdate(conn, uploadID, currentProgress, "连接已建立")

	// 监听关闭
	go func() {
		defer conn.Close()
		for {
			// 保持连接并监听客户端消息
			_, _, err := conn.ReadMessage()
			if err != nil {
				// 连接已关闭，移除连接
				GlobalProgressHub.mu.Lock()
				for i, c := range GlobalProgressHub.connections[uploadID] {
					if c == conn {
						GlobalProgressHub.connections[uploadID] = append(
							GlobalProgressHub.connections[uploadID][:i],
							GlobalProgressHub.connections[uploadID][i+1:]...,
						)
						break
					}
				}
				// 如果没有更多连接，清理数据
				if len(GlobalProgressHub.connections[uploadID]) == 0 {
					delete(GlobalProgressHub.connections, uploadID)
					delete(GlobalProgressHub.progress, uploadID)
				}
				GlobalProgressHub.mu.Unlock()
				break
			}
		}
	}()
}

// UpdateProgress 更新上传进度并通知所有连接的客户端
func (h *ProgressHub) UpdateProgress(uploadID string, progress float64, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.progress[uploadID] = progress

	// 只在关键节点记录日志
	if progress == 0 || progress == 100 || status == "上传失败" || status == "处理完成" {
		h.log.Infof("上传进度: uploadID=%s, progress=%.0f%%, status=%s", uploadID, progress, status)
	}

	if conns, exists := h.connections[uploadID]; exists && len(conns) > 0 {
		activeConnections := 0
		for _, conn := range conns {
			if conn != nil {
				err := sendProgressUpdate(conn, uploadID, progress, status)
				if err == nil {
					activeConnections++
				} else {
					h.log.Errorf("发送进度更新失败: uploadID=%s, error=%v", uploadID, err)
				}
			}
		}
		// 只在失败时记录详细信息
		if activeConnections == 0 {
			h.log.Warnf("无法通知客户端: uploadID=%s", uploadID)
		}
	} else {
		// 只在最终状态时警告连接已断开
		if status == "处理完成" {
			h.log.Warnf("连接已断开: uploadID=%s", uploadID)
		}
	}
}

// 发送进度更新到客户端
func sendProgressUpdate(conn *websocket.Conn, uploadID string, progress float64, status string) error {
	// 确保状态不为空
	if status == "" {
		status = "处理中"
	}

	message := map[string]interface{}{
		"uploadID": uploadID,
		"progress": int(progress), // 转换为整数
		"status":   status,
	}

	err := conn.WriteJSON(message)
	if err != nil {
		return err
	}
	return nil
}

// CleanupUpload 清理特定上传ID的连接
func (h *ProgressHub) CleanupUpload(uploadID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, exists := h.connections[uploadID]; exists {
		// 优雅关闭所有连接
		for _, conn := range conns {
			if conn != nil {
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "上传完成"))
				conn.Close()
			}
		}

		// 清理数据
		delete(h.connections, uploadID)
		delete(h.progress, uploadID)
		h.log.Infof("上传完成，连接已清理: uploadID=%s", uploadID)
	}
}

// 定期清理未使用的上传ID
func (h *ProgressHub) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		for uploadID, conns := range h.connections {
			if len(conns) == 0 {
				delete(h.connections, uploadID)
				delete(h.progress, uploadID)
			}
		}
		h.mu.Unlock()
	}
}
