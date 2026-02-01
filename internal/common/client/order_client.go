package client

import (
	"context"
	"fmt"
	"log"
	"time"
	orderv1 "zjMall/gen/go/api/proto/order"
	"zjMall/internal/common/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// OrderClient 订单服务客户端接口
type OrderClient interface {
	// GetOrder 获取订单信息（不校验用户ID，用于支付回调等场景）
	GetOrderByNo(ctx context.Context, orderNo string) (*orderv1.Order, error)
	// MarkOrderPaid 标记订单已支付
	MarkOrderPaid(ctx context.Context, orderNo, payChannel, payTradeNo string) error
	// Close 关闭连接
	Close() error
}

type orderClient struct {
	conn   *grpc.ClientConn
	client orderv1.OrderServiceClient
}

// NewOrderClient 创建订单服务客户端
func NewOrderClient(addr string) (OrderClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: false,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接订单服务失败: %w", err)
	}

	client := orderv1.NewOrderServiceClient(conn)

	log.Printf("✅ 订单服务客户端连接成功: %s", addr)

	return &orderClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetOrderByNo 获取订单信息（不校验用户ID）
func (c *orderClient) GetOrderByNo(ctx context.Context, orderNo string) (*orderv1.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 从 context 中获取 userID（如果有的话，用于日志）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID != "" {
		md := metadata.New(map[string]string{
			string(middleware.UserIDKey): userID,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	log.Printf("🔍 [OrderClient] GetOrderByNo: orderNo=%s", orderNo)

	resp, err := c.client.GetOrder(ctx, &orderv1.GetOrderRequest{
		OrderNo: orderNo,
	})
	if err != nil {
		return nil, fmt.Errorf("调用订单服务失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("订单服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	if resp.Order == nil {
		return nil, fmt.Errorf("订单不存在")
	}
	return resp.Order, nil
}

// MarkOrderPaid 标记订单已支付
func (c *orderClient) MarkOrderPaid(ctx context.Context, orderNo, payChannel, payTradeNo string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	log.Printf("🔍 [OrderClient] MarkOrderPaid: orderNo=%s, payChannel=%s, payTradeNo=%s", orderNo, payChannel, payTradeNo)

	resp, err := c.client.MarkOrderPaid(ctx, &orderv1.MarkOrderPaidRequest{
		OrderNo:    orderNo,
		PayChannel: payChannel,
		PayTradeNo: payTradeNo,
	})
	if err != nil {
		return fmt.Errorf("调用订单服务失败: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("订单服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	return nil
}

// Close 关闭连接
func (c *orderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
