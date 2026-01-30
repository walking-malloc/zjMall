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
	// DeductStock 批量扣减库存（使用乐观锁，防止超卖，支持幂等性检查）
	// orderID: 订单号，作为幂等键
	// items: 需要扣减的 SKU 列表，批量操作在一个事务中完成
	DeductStock(ctx context.Context, orderID string, items []*inventoryv1.SkuQuantity) error
	// RollbackStock 批量回滚库存（订单取消/关闭时调用，使用乐观锁）
	// orderID: 订单号，用于日志记录
	// items: 需要回滚的 SKU 列表，批量操作在一个事务中完成
	RollbackStock(ctx context.Context, orderID string, items []*inventoryv1.SkuQuantity) error
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
			Time:                30 * time.Second, // 每30秒发送一次ping（降低频率）
			Timeout:             5 * time.Second,  // ping超时时间
			PermitWithoutStream: false,            // 只在有活跃流时发送ping
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

// DeductStock 批量扣减库存（使用乐观锁，防止超卖，支持幂等性检查）
// 批量操作在一个事务中完成，全部成功或全部失败
func (c *inventoryClient) DeductStock(ctx context.Context, orderID string, items []*inventoryv1.SkuQuantity) error {
	if len(items) == 0 {
		return fmt.Errorf("扣减项不能为空")
	}
	if orderID == "" {
		return fmt.Errorf("订单号不能为空")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // 批量操作可能需要更长时间
	defer cancel()

	resp, err := c.client.DeductStock(ctx, &inventoryv1.DeductStockRequest{
		OrderId: orderID,
		Items:   items,
	})
	if err != nil {
		return fmt.Errorf("调用库存服务扣减失败: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("库存服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	return nil
}

// RollbackStock 批量回滚库存（订单取消/关闭时调用，使用乐观锁）
// 批量操作在一个事务中完成，部分失败会继续处理其他项
func (c *inventoryClient) RollbackStock(ctx context.Context, orderID string, items []*inventoryv1.SkuQuantity) error {
	if len(items) == 0 {
		return nil // 空列表直接返回成功
	}
	if orderID == "" {
		return fmt.Errorf("订单号不能为空")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // 批量操作可能需要更长时间
	defer cancel()

	resp, err := c.client.RollbackStock(ctx, &inventoryv1.RollbackStockRequest{
		OrderId: orderID,
		Items:   items,
	})
	if err != nil {
		return fmt.Errorf("调用库存服务回滚失败: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("库存服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}
	return nil
}

// Close 关闭连接
func (c *inventoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
