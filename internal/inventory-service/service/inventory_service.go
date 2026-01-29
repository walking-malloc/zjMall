package service

import (
	"context"
	"fmt"

	"zjMall/internal/inventory-service/model"
	"zjMall/internal/inventory-service/repository"
)

// ItemQuantity 表示单个 SKU 的数量请求
type ItemQuantity struct {
	SKUID    string
	Quantity int64
}

// InventoryService 库存领域服务
type InventoryService struct {
	stockRepo repository.StockRepository
}

// NewInventoryService 创建库存服务
func NewInventoryService(stockRepo repository.StockRepository) *InventoryService {
	return &InventoryService{
		stockRepo: stockRepo,
	}
}

// GetStock 查询单个 SKU 库存
func (s *InventoryService) GetStock(ctx context.Context, skuID string) (*model.Stock, error) {
	if skuID == "" {
		return nil, fmt.Errorf("skuID 不能为空")
	}
	return s.stockRepo.GetBySKUID(ctx, skuID)
}

// BatchGetStock 批量查询库存
func (s *InventoryService) BatchGetStock(ctx context.Context, skuIDs []string) (map[string]*model.Stock, error) {
	if len(skuIDs) == 0 {
		return nil, nil
	}
	return s.stockRepo.BatchGetBySKUID(ctx, skuIDs)
}

// TryDeductStocks 尝试为订单扣减多个 SKU 的库存（批量操作，使用乐观锁）
// orderNo: 订单号，用于日志记录和幂等性检查
func (s *InventoryService) TryDeductStocks(ctx context.Context, orderNo string, items []ItemQuantity) error {
	if orderNo == "" {
		return fmt.Errorf("订单号不能为空")
	}
	if len(items) == 0 {
		return fmt.Errorf("扣减项不能为空")
	}

	// 转换为 repository 层的 DeductItem
	deductItems := make([]repository.DeductItem, 0, len(items))
	for _, item := range items {
		if item.SKUID == "" || item.Quantity <= 0 {
			return fmt.Errorf("非法库存扣减请求: sku_id=%s, quantity=%d", item.SKUID, item.Quantity)
		}
		deductItems = append(deductItems, repository.DeductItem{
			SKUID:    item.SKUID,
			Quantity: item.Quantity,
		})
	}

	return s.stockRepo.TryDeductStocks(ctx, orderNo, deductItems)
}

// RollbackStocks 为多个 SKU 回滚库存（批量操作，使用乐观锁）
// orderNo: 订单号，用于日志记录
func (s *InventoryService) RollbackStocks(ctx context.Context, orderNo string, items []ItemQuantity) error {
	if orderNo == "" {
		return fmt.Errorf("订单号不能为空")
	}
	if len(items) == 0 {
		return nil // 空列表直接返回成功
	}

	// 转换为 repository 层的 DeductItem
	deductItems := make([]repository.DeductItem, 0, len(items))
	for _, item := range items {
		if item.SKUID == "" || item.Quantity <= 0 {
			continue // 跳过无效项
		}
		deductItems = append(deductItems, repository.DeductItem{
			SKUID:    item.SKUID,
			Quantity: item.Quantity,
		})
	}

	if len(deductItems) == 0 {
		return nil
	}

	return s.stockRepo.RollbackStocks(ctx, orderNo, deductItems)
}
