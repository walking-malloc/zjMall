package client

import (
	"context"
	"fmt"
	"log"
	"time"
	productv1 "zjMall/gen/go/api/proto/product"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ProductClient 商品服务客户端接口
type ProductClient interface {
	// GetProduct 获取商品详情（包含 SKU 列表）
	// 返回商品信息和 SKU 列表
	GetProduct(ctx context.Context, productID string) (*productv1.ProductInfo, []*productv1.SkuInfo, error)
	// Close 关闭连接
	Close() error
}

type productClient struct {
	conn   *grpc.ClientConn
	client productv1.ProductServiceClient
}

// NewProductClient 创建商品服务客户端
// addr: 商品服务 gRPC 地址，例如 "localhost:50053"
func NewProductClient(addr string) (ProductClient, error) {
	// 配置连接参数
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// 建立连接（使用 context 控制超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接商品服务失败: %w", err)
	}

	client := productv1.NewProductServiceClient(conn)

	log.Printf("✅ 商品服务客户端连接成功: %s", addr)

	return &productClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetProduct 获取商品详情（包含 SKU 列表）
func (c *productClient) GetProduct(ctx context.Context, productID string) (*productv1.ProductInfo, []*productv1.SkuInfo, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.GetProduct(ctx, &productv1.GetProductRequest{
		ProductId:   productID,
		IncludeSkus: true, // 包含 SKU 列表
	})

	if err != nil {
		return nil, nil, fmt.Errorf("调用商品服务失败: %w", err)
	}

	if resp.Code != 0 {
		return nil, nil, fmt.Errorf("商品服务返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}

	if resp.Product == nil {
		return nil, nil, fmt.Errorf("商品不存在: product_id=%s", productID)
	}

	return resp.Product, resp.Skus, nil
}

// Close 关闭连接
func (c *productClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
