package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"zjMall/internal/common/metrics"
)

// PrometheusMetrics 是一个中间件，用于记录 Prometheus 指标
func PrometheusMetrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 复用 logging.go 中的 responseWriter 来捕获状态码
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 执行下一个处理器
			next.ServeHTTP(rw, r)

			// 计算耗时
			duration := time.Since(start).Seconds()

			// 规范化路径（避免高基数标签）
			path := normalizePath(r.URL.Path)

			// 记录指标
			statusCode := strconv.Itoa(rw.statusCode)
			metrics.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				path,
				statusCode,
			).Inc()

			metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				path,
			).Observe(duration)
		})
	}
}

// normalizePath 规范化路径，避免高基数标签
// 例如：/api/v1/products/123 -> /api/v1/products/:id
func normalizePath(path string) string {
	// 移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// 规范化常见路径参数
	normalizations := map[string]string{
		"/api/v1/products/":  "/api/v1/products/:id",
		"/api/v1/users/":     "/api/v1/users/:id",
		"/api/v1/orders/":    "/api/v1/orders/:id",
		"/api/v1/carts/":     "/api/v1/carts/:id",
		"/api/v1/payments/":  "/api/v1/payments/:id",
		"/api/v1/inventory/": "/api/v1/inventory/:id",
	}

	for prefix, replacement := range normalizations {
		if strings.HasPrefix(path, prefix) {
			// 检查是否是带ID的路径
			parts := strings.Split(path[len(prefix):], "/")
			if len(parts) > 0 && parts[0] != "" {
				return replacement
			}
		}
	}

	return path
}
