package middleware

import (
	"context"
	"net/http"
	"strings"

	"anjuke/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

// JWTKey 用于在context中存储JWT claims的key
type JWTKey struct{}

// JWTMiddleware JWT中间件
func JWTMiddleware(jwtIns *jwt.JWT) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 从HTTP请求中获取token
			if tr, ok := req.(interface{ Header() http.Header }); ok {
				authHeader := tr.Header().Get("Authorization")
				if authHeader == "" {
					return nil, errors.Unauthorized("UNAUTHORIZED", "missing authorization header")
				}

				// 检查Authorization header格式
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid authorization header format")
				}

				tokenString := parts[1]
				if tokenString == "" {
					return nil, errors.Unauthorized("UNAUTHORIZED", "missing token")
				}

				// 解析JWT token
				claims, err := jwtIns.ParseToken(tokenString)
				if err != nil {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid token")
				}

				// 将claims存储到context中
				ctx = context.WithValue(ctx, JWTKey{}, claims)
			}

			return handler(ctx, req)
		}
	}
}

// GetClaimsFromContext 从context中获取JWT claims
func GetClaimsFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(JWTKey{}).(*jwt.Claims)
	return claims, ok
}

// GetUserIDFromContext 从context中获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return 0, false
	}
	return claims.UserID, true
}

// GetUsernameFromContext 从context中获取用户名
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.Username, true
}
