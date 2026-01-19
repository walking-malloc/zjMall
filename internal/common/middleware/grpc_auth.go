package middleware

import (
	"context"
	"strings"

	"zjMall/pkg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryAuthInterceptor 所有 gRPC Unary 的认证拦截器
func UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 从 metadata 里取 Authorization
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req) // 没有 metadata 就先放行，看业务自己怎么处理
	}

	authVals := md.Get("authorization")
	if len(authVals) == 0 || authVals[0] == "" {
		// 没有 token，可以直接放行（由业务判断），也可以直接返回 Unauthenticated
		// 这里建议：放行，后面业务用 GetUserIDFromContext 判空，返回“未登录”
		return handler(ctx, req)
	}

	authHeader := authVals[0]
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return handler(ctx, req)
	}

	// 验证 JWT，拿到 userId
	userID, err := pkg.VerifyJWT(token)
	if err != nil {
		// 这里可以直接返回 401，也可以继续放行
		return nil, status.Error(codes.Unauthenticated, "Token 无效或已过期")
	}

	// 写入到 context，key 用你现在的 UserIDKey
	ctx = context.WithValue(ctx, UserIDKey, userID)

	// 继续后面的 handler，后面可以用 GetUserIDFromContext(ctx) 取
	return handler(ctx, req)
}
