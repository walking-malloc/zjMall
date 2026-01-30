package client

import (
	"context"
	"fmt"
	"log"
	"time"

	cartv1 "zjMall/gen/go/api/proto/cart"

	"zjMall/internal/common/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// CartClient 购物车服务客户端接口
type CartClient interface {
	// RemoveItems 批量删除购物车商品
	RemoveItems(ctx context.Context, itemIDs []string) error
	// Close 关闭连接
	Close() error
}

type cartClient struct {
	conn   *grpc.ClientConn
	client cartv1.CartServiceClient
}

// NewCartClient 创建购物车服务客户端
// addr: 购物车服务 gRPC 地址，例如 "localhost:50054"
func NewCartClient(addr string) (CartClient, error) {
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
		return nil, fmt.Errorf("连接购物车服务失败: %w", err)
	}

	client := cartv1.NewCartServiceClient(conn)

	log.Printf("✅ 购物车服务客户端连接成功: %s", addr)

	return &cartClient{
		conn:   conn,
		client: client,
	}, nil
}

// RemoveItems 批量删除购物车商品
func (c *cartClient) RemoveItems(ctx context.Context, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil // 空列表直接返回成功
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 从 context 中获取 userID（由订单服务的认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return fmt.Errorf("无法获取用户ID，请确保已登录")
	}

	// 将 userID 放入 gRPC metadata，传递给购物车服务
	md := metadata.New(map[string]string{
		string(middleware.UserIDKey): userID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	log.Printf("🔍 [CartClient] RemoveItems: userID=%s, itemIDs=%v", userID, itemIDs)

	resp, err := c.client.RemoveItems(ctx, &cartv1.RemoveItemsRequest{
		ItemIds: itemIDs,
	})
	if err != nil {
		return fmt.Errorf("调用购物车服务失败: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("购物车服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}

	log.Printf("✅ [CartClient] RemoveItems: 成功删除 %d 个购物车项", len(itemIDs))
	return nil
}

// Close 关闭连接
func (c *cartClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
