package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// TraceIDKey Context 中存储 Trace ID 的 key
const TraceIDKey = "traceID"

// TraceID Trace ID 中间件
// 用于分布式追踪，每个请求分配一个唯一的 Trace ID
func TraceID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 尝试从请求头获取 Trace ID
			traceID := r.Header.Get("X-Trace-ID")

			// 如果请求头中没有，生成一个新的
			if traceID == "" {
				traceID = generateTraceID()
			}

			// 将 Trace ID 添加到响应头，方便客户端追踪
			w.Header().Set("X-Trace-ID", traceID)

			// 将 Trace ID 添加到 Context 中，供后续使用
			ctx := context.WithValue(r.Context(), TraceIDKey, traceID)

			// 使用新的 Context 继续处理
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTraceID 从 Context 中获取 Trace ID
// 在业务代码中使用：traceID := middleware.GetTraceID(ctx)
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// generateTraceID 生成唯一的 Trace ID（32位十六进制字符串）
func generateTraceID() string {
	bytes := make([]byte, 16) // 16字节 = 128位
	rand.Read(bytes)
	return hex.EncodeToString(bytes) // 转换为32位十六进制字符串
}
