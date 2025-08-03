package data

import (
	"github.com/afex/hystrix-go/hystrix"
	"time"
)

// CircuitBreaker 熔断器接口（与OrderCacheRepo中定义一致）
type CircuitBreaker interface {
	IsOpen() bool // 检查熔断器是否开启
}

// HystrixCircuit 基于Hystrix的熔断器实现
type HystrixCircuit struct {
	command string // 熔断器命令名（唯一标识，如"order_query"）
}

func (h *HystrixCircuit) Get(key string) (interface{}, bool) {
	//TODO implement me
	panic("implement me")
}

func (h *HystrixCircuit) Set(key string, value interface{}, ttl time.Duration) {
	//TODO implement me
	panic("implement me")
}

// NewHystrixCircuit 创建熔断器实例
func NewHystrixCircuit(command string) *HystrixCircuit {
	// 配置熔断器参数（根据业务调整）
	hystrix.ConfigureCommand(command, hystrix.CommandConfig{
		Timeout:               1000, // 超时时间（毫秒）
		MaxConcurrentRequests: 100,  // 最大并发请求数
		ErrorPercentThreshold: 50,   // 错误率阈值（超过则熔断）
		SleepWindow:           5000, // 熔断后休眠时间（毫秒）
	})
	return &HystrixCircuit{command: command}
}

// IsOpen 检查熔断器是否开启（新版本兼容写法）
func (h *HystrixCircuit) IsOpen() bool {
	// 获取熔断器实例
	circuit, _, err := hystrix.GetCircuit(h.command)
	if err != nil {
		// 若熔断器未初始化，默认返回"未开启"
		return false
	}
	// 检查熔断器状态（Open表示开启）
	return circuit.IsOpen()
}
