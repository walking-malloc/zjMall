package client

import (
	"context"
	"fmt"
	"log"
	"time"

	inventoryv1 "zjMall/gen/go/api/proto/inventory"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// InventoryClient 库存服务客户端接口
type InventoryClient interface {
	// GetStock 获取单个 SKU 的可用库存
	GetStock(ctx context.Context, skuID string) (int64, error)
	// BatchGetStock 批量获取多个 SKU 的库存信息，返回 map[sku_id]Stock
	BatchGetStock(ctx context.Context, skuIDs []string) (map[string]*inventoryv1.Stock, error)
	// Close 关闭连接
	Close() error
}

type inventoryClient struct {
	conn   *grpc.ClientConn
	client inventoryv1.InventoryServiceClient
}

// NewInventoryClient 创建库存服务客户端
// addr: 库存服务 gRPC 地址，例如 "localhost:50054"
func NewInventoryClient(addr string) (InventoryClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接库存服务失败: %w", err)
	}

	client := inventoryv1.NewInventoryServiceClient(conn)

	log.Printf("✅ 库存服务客户端连接成功: %s", addr)

	return &inventoryClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetStock 获取单个 SKU 的可用库存
func (c *inventoryClient) GetStock(ctx context.Context, skuID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.GetStock(ctx, &inventoryv1.GetStockRequest{
		SkuId: skuID,
	})
	if err != nil {
		return 0, fmt.Errorf("调用库存服务失败: %w", err)
	}
	if resp.Code != 0 {
		return 0, fmt.Errorf("库存服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	if resp.Data == nil {
		return 0, nil
	}
	return resp.Data.AvailableStock, nil
}

// BatchGetStock 批量获取库存
func (c *inventoryClient) BatchGetStock(ctx context.Context, skuIDs []string) (map[string]*inventoryv1.Stock, error) {
	if len(skuIDs) == 0 {
		return map[string]*inventoryv1.Stock{}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.BatchGetStock(ctx, &inventoryv1.BatchGetStockRequest{
		SkuIds: skuIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("调用库存服务失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("库存服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	if resp.Data == nil {
		return map[string]*inventoryv1.Stock{}, nil
	}
	return resp.Data, nil
}

// Close 关闭连接
func (c *inventoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
