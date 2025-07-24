package service

// 全局管理器实例
var (
	GlobalWebSocketManager *WebSocketManager
)

// InitGlobalManagers 初始化全局管理器
func InitGlobalManagers() {
	GlobalWebSocketManager = NewWebSocketManager()
	// GlobalSequenceManager 已在 crypto.go 中定义和初始化
}

// 确保在包加载时初始化全局管理器
func init() {
	InitGlobalManagers()
}