package middleware

import (
	"context"
	"log"
	"strings"

	"zjMall/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryAuthInterceptor 所有 gRPC Unary 的认证拦截器
// 支持两种方式：
// 1. 客户端调用：从 authorization header 获取 JWT token，验证后提取 userID
// 2. 服务间调用：直接从 user_id metadata 获取 userID（信任内部服务）
func UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 从 metadata 里取数据
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req) // 没有 metadata 就先放行，看业务自己怎么处理
	}

	var userID string

	// 优先检查是否有 user_id metadata（服务间调用）
	userIDVals := md.Get(string(UserIDKey))
	if len(userIDVals) > 0 && userIDVals[0] != "" {
		// 服务间调用，直接使用 user_id
		userID = userIDVals[0]
		log.Printf("🔍 [gRPC Auth] 从 user_id metadata 获取: %s", userID)
	} else {
		// 客户端调用，从 authorization header 获取 JWT token
		authVals := md.Get("authorization")
		if len(authVals) == 0 || authVals[0] == "" {
			// 没有 token，可以直接放行（由业务判断），也可以直接返回 Unauthenticated
			// 这里建议：放行，后面业务用 GetUserIDFromContext 判空，返回"未登录"
			return handler(ctx, req)
		}

		authHeader := authVals[0]
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			return handler(ctx, req)
		}

		// 验证 JWT，拿到 userId
		var err error
		userID, err = pkg.VerifyJWT(token)
		if err != nil {
			// 这里可以直接返回 401，也可以继续放行
			return nil, status.Error(codes.Unauthenticated, "Token 无效或已过期")
		}
		log.Printf("🔍 [gRPC Auth] 从 JWT token 验证获取: %s", userID)
	}

	// 将 userID 写入到 context，后续 handler 可以用 GetUserIDFromContext(ctx) 获取
	if userID != "" {
		ctx = context.WithValue(ctx, UserIDKey, userID)
	}

	// 继续后面的 handler
	return handler(ctx, req)
}
