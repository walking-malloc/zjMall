package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"zjMall/pkg"

	"google.golang.org/grpc/metadata"
)

// ContextKey 用于从 context 中获取用户ID
type ContextKey string

const UserIDKey ContextKey = "user_id"

// GetUserIDFromContext 从 context 中获取用户ID
// 优先从 HTTP context 中获取，如果没有则从 gRPC metadata 中获取
func GetUserIDFromContext(ctx context.Context) string {
	// 1. 优先从 HTTP context 中获取（由 HTTP 中间件设置）
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		log.Printf("从 HTTP context 获取 user_id: %s", userID)
		return userID
	}

	// 2. 从 gRPC metadata 中获取（由 gRPC Gateway 传递）
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("从 gRPC metadata 获取 user_id: %v", md)
		userIDs := md.Get(string(UserIDKey))
		if len(userIDs) > 0 && userIDs[0] != "" {
			log.Printf("从 gRPC metadata 获取 user_id: %s", userIDs[0])
			return userIDs[0]
		}
	}

	return ""
}

// 白名单路径（不需要认证的接口）
var publicPaths = []string{
	"/api/v1/users/register",     // 注册
	"/api/v1/users/login",        // 登录
	"/api/v1/users/login-by-sms", // 短信登录
	"/api/v1/users/sms-code",     // 获取短信验证码
	"/healthz",                   // 健康检查
	"/swagger/",                  // Swagger 文档
}

// isPublicPath 检查路径是否在白名单中
func isPublicPath(path string) bool {
	for _, publicPath := range publicPaths {
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// Auth 认证中间件
func Auth() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// OPTIONS 预检请求直接放行（由 CORS 中间件处理）
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// 如果是公开路径，直接放行
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 从请求头获取 Token
			// 支持两种格式：
			// 1. Authorization: Bearer <token>
			// 2. Authorization: <token>

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"code": 401, "message": "未登录，请先登录"}`, http.StatusUnauthorized)
				return
			}

			// 提取 Token（去除 "Bearer " 前缀）
			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			if token == "" {
				http.Error(w, `{"code": 401, "message": "Token 格式错误"}`, http.StatusUnauthorized)
				return
			}

			// 验证 Token 并获取 Claims（包含用户ID和角色）
			claims, err := pkg.VerifyJWTWithClaims(token)
			if err != nil {
				log.Println("VerifyJWT error:", err)
				http.Error(w, `{"code": 401, "message": "Token 无效或已过期"}`, http.StatusUnauthorized)
				return
			}
			log.Printf("[Auth] claims.Roles: %v", claims.Roles)
			// 将用户ID和角色放入 Context，后续 handler 可以从 context 中获取
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			if len(claims.Roles) > 0 {
				ctx = context.WithValue(ctx, RolesKey, claims.Roles)
			}
			r = r.WithContext(ctx)

			// 继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}
