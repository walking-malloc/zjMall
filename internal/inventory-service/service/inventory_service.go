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

// TryDeductStocks 尝试为订单扣减多个 SKU 的库存（全部成功才算成功）
// 这里采用简化实现：串行依次扣减，一旦某个失败，调用方负责回滚。
func (s *InventoryService) TryDeductStocks(ctx context.Context, items []ItemQuantity) error {
	for _, item := range items {
		if item.SKUID == "" || item.Quantity <= 0 {
			return fmt.Errorf("非法库存扣减请求: sku_id=%s, quantity=%d", item.SKUID, item.Quantity)
		}
		if err := s.stockRepo.TryDeductStock(ctx, item.SKUID, item.Quantity); err != nil {
			return fmt.Errorf("扣减库存失败 sku_id=%s: %w", item.SKUID, err)
		}
	}
	return nil
}

// RollbackStocks 为多个 SKU 回滚库存
func (s *InventoryService) RollbackStocks(ctx context.Context, items []ItemQuantity) error {
	for _, item := range items {
		if item.SKUID == "" || item.Quantity <= 0 {
			continue
		}
		if err := s.stockRepo.RollbackStock(ctx, item.SKUID, item.Quantity); err != nil {
			return fmt.Errorf("回滚库存失败 sku_id=%s: %w", item.SKUID, err)
		}
	}
	return nil
}
