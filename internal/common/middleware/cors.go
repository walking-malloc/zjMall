package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins   []string // 允许的源
	AllowedMethods   []string // 允许的 HTTP 方法
	AllowedHeaders   []string // 允许的请求头
	ExposedHeaders   []string // 暴露给客户端的响应头
	AllowCredentials bool     // 是否允许携带凭证
	MaxAge           int      // 预检请求缓存时间（秒）
}

// DefaultCORSConfig 返回默认的 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"}, // 允许所有源
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "X-Trace-ID"},
		ExposedHeaders:   []string{"X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           3600, // 1小时
	}
}

// CORS CORS 中间件（企业级：支持配置化）
func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查 Origin 是否在允许列表中
			if isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// 设置允许的方法
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))

			// 设置允许的请求头
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))

			// 设置暴露的响应头
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))

			// 设置是否允许凭证
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// 设置预检请求缓存时间
			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// 处理预检请求（OPTIONS）
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 继续处理实际请求
			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed 检查 Origin 是否在允许列表中
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		// 支持通配符 "*"
		if allowed == "*" {
			return true
		}
		// 精确匹配
		if allowed == origin {
			return true
		}
	}

	return false
}
