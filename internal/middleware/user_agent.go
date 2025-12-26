// Package middleware provides HTTP middleware components for the application.
package middleware

import (
	"context"
	"net/http"
)

// 定义一个自定义的上下文 key 类型（非导出，避免外部冲突）
type userAgentContextKey struct{}

// GetUserAgent 从给定的上下文中检索存储的 User-Agent 信息。如果不存在，则返回空字符串。
func GetUserAgent(ctx context.Context) string {
	if val, ok := ctx.Value(userAgentContextKey{}).(string); ok {
		return val
	}
	return ""
}

// UserAgentMiddleware 是一个中间件，它从传入的请求中提取 User-Agent 头部信息，并将其存储在请求上下文中以供后续检索。
type UserAgentMiddleware struct{}

func NewUserAgentMiddleware() *UserAgentMiddleware {
	return &UserAgentMiddleware{}
}

func (m *UserAgentMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val := r.Header.Get("User-Agent")
		reqCtx := r.Context()
		ctx := context.WithValue(reqCtx, userAgentContextKey{}, val)
		newReq := r.WithContext(ctx)

		// Passthrough to next handler
		next(w, newReq)
	}
}
