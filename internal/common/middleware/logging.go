package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// LogEntry 结构化日志条目
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Duration   string `json:"duration_ms"`
	RemoteAddr string `json:"remote_addr"`
	UserAgent  string `json:"user_agent"`
	TraceID    string `json:"trace_id"`
}

// responseWriter 包装 http.ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logging 日志中间件（企业级：结构化 JSON 日志）
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// 结构化日志（JSON 格式，便于日志收集系统解析）
			entry := LogEntry{
				Timestamp:  time.Now().Format(time.RFC3339),
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rw.statusCode,
				Duration:   formatDuration(duration),
				RemoteAddr: r.RemoteAddr,
				UserAgent:  r.UserAgent(),
				TraceID:    GetTraceID(r.Context()), // 从 Context 获取 Trace ID
			}

			// JSON 格式输出（便于 ELK、Loki 等日志收集系统解析）
			logData, err := json.Marshal(entry)
			if err != nil {
				// 如果 JSON 序列化失败，使用文本格式
				log.Printf("[%s] %s %s - %d - %v - TraceID: %s",
					r.Method, r.URL.Path, r.RemoteAddr, rw.statusCode, duration, entry.TraceID)
			} else {
				log.Println(string(logData))
			}
		})
	}
}

// formatDuration 格式化耗时（毫秒）
func formatDuration(d time.Duration) string {
	return d.String()
}
