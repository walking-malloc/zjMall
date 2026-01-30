package client

import (
	"context"
	"fmt"
	"log"
	"time"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/common/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// UserClient 用户服务客户端接口
type UserClient interface {
	// GetUserAddress 获取用户地址
	// addressID 为空时返回默认地址，否则返回指定地址
	GetUserAddress(ctx context.Context, addressID string) (*userv1.Address, error)
	// Close 关闭连接
	Close() error
}

type userClient struct {
	conn   *grpc.ClientConn
	client userv1.UserServiceClient
}

// NewUserClient 创建用户服务客户端
// addr: 用户服务 gRPC 地址，例如 "localhost:50052"
func NewUserClient(addr string) (UserClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second, // 每30秒发送一次ping（降低频率）
			Timeout:             5 * time.Second,  // ping超时时间
			PermitWithoutStream: false,            // 只在有活跃流时发送ping
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接用户服务失败: %w", err)
	}

	client := userv1.NewUserServiceClient(conn)

	log.Printf("✅ 用户服务客户端连接成功: %s", addr)

	return &userClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetUserAddress 获取用户地址
// addressID 为空时返回默认地址，否则返回指定地址
func (c *userClient) GetUserAddress(ctx context.Context, addressID string) (*userv1.Address, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 从 context 中获取 userID（由订单服务的认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("无法获取用户ID，请确保已登录")
	}

	// 将 userID 放入 gRPC metadata，传递给用户服务
	md := metadata.New(map[string]string{
		string(middleware.UserIDKey): userID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	log.Printf("🔍 [UserClient] GetUserAddress: userID=%s, addressID=%s", userID, addressID)

	resp, err := c.client.GetUserAddress(ctx, &userv1.GetUserAddressRequest{
		AddressId: addressID,
	})
	if err != nil {
		return nil, fmt.Errorf("调用用户服务失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("用户服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("地址不存在")
	}
	return resp.Data, nil
}

// Close 关闭连接
func (c *userClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
