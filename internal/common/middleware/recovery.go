package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	TraceID string `json:"trace_id,omitempty"`
}

// Recovery 错误恢复中间件
// 捕获 panic，防止程序崩溃，返回 JSON 格式的错误响应
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 记录错误（包含 Trace ID）
					traceID := GetTraceID(r.Context())
					log.Printf("Panic recovered: %v\nTraceID: %s\n%s",
						err, traceID, debug.Stack())

					// 设置响应头
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					// 返回 JSON 格式的错误响应
					response := ErrorResponse{
						Error:   "Internal Server Error",
						TraceID: traceID,
					}
					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
