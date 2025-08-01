package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport"
)

// MiddlewareBuilder 中间件构建器
type MiddlewareBuilder struct {
	logger log.Logger
}

// NewMiddlewareBuilder 创建中间件构建器
func NewMiddlewareBuilder(logger log.Logger) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logger: logger,
	}
}

// Build 构建中间件链
func (b *MiddlewareBuilder) Build() middleware.Middleware {
	return middleware.Chain(
		recovery.Recovery(),
		tracing.Server(),
		metrics.Server(),
		b.logging(),
	)
}

// logging 日志中间件
func (b *MiddlewareBuilder) logging() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			t := time.Now()
			info, ok := transport.FromServerContext(ctx)
			if ok {
				b.logger.Log(log.LevelInfo, "start handling request", "method", info.Operation(), "time", t)
			}
			resp, err := handler(ctx, req)
			if ok {
				b.logger.Log(
					log.LevelInfo, "finish handling request",
					"method", info.Operation(),
					"time", t,
					"duration", time.Since(t),
					"error", err,
				)
			}
			return resp, err
		}
	}
}