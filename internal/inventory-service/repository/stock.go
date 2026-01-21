package repository

import (
	"context"
	"fmt"
	"log"

	"zjMall/internal/inventory-service/model"

	"gorm.io/gorm"
)

// StockRepository 库存仓储接口
type StockRepository interface {
	// GetBySKUID 根据 skuID 查询库存
	GetBySKUID(ctx context.Context, skuID string) (*model.Stock, error)
	// BatchGetBySKUID 批量查询库存
	BatchGetBySKUID(ctx context.Context, skuIDs []string) (map[string]*model.Stock, error)
	// TryDeductStock 尝试扣减库存（并发安全，防止超卖）
	TryDeductStock(ctx context.Context, skuID string, quantity int64) error
	// RollbackStock 回滚库存（加回）
	RollbackStock(ctx context.Context, skuID string, quantity int64) error
}

type stockRepository struct {
	db *gorm.DB
}

// NewStockRepository 创建库存仓储
func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

// GetBySKUID 根据 skuID 查询库存
func (r *stockRepository) GetBySKUID(ctx context.Context, skuID string) (*model.Stock, error) {
	var stock model.Stock
	if err := r.db.WithContext(ctx).
		Where("sku_id = ?", skuID).
		First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询库存失败: %w", err)
	}
	return &stock, nil
}

// BatchGetBySKUID 批量查询库存
func (r *stockRepository) BatchGetBySKUID(ctx context.Context, skuIDs []string) (map[string]*model.Stock, error) {
	if len(skuIDs) == 0 {
		return map[string]*model.Stock{}, nil
	}

	var list []model.Stock
	if err := r.db.WithContext(ctx).
		Where("sku_id IN ?", skuIDs).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("批量查询库存失败: %w", err)
	}

	result := make(map[string]*model.Stock, len(list))
	for i := range list {
		s := list[i]
		result[s.SKUID] = &s
	}
	return result, nil
}

// TryDeductStock 尝试扣减库存（防止超卖）
func (r *stockRepository) TryDeductStock(ctx context.Context, skuID string, quantity int64) error {
	if quantity <= 0 {
		return fmt.Errorf("扣减数量必须大于0")
	}

	// 简化实现：使用一条条件更新 SQL 防止超卖
	// UPDATE inventory_stocks
	// SET available_stock = available_stock - ?
	// WHERE sku_id = ? AND available_stock >= ?
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	res := tx.
		Model(&model.Stock{}).
		Where("sku_id = ? AND available_stock >= ?", skuID, quantity).
		Update("available_stock", gorm.Expr("available_stock - ?", quantity))

	if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("扣减库存失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("库存不足或记录不存在")
	}

	// 写入库存变动日志（负数表示扣减）
	logEntry := &model.StockLog{
		SKUID:        skuID,
		ChangeAmount: -quantity,
		Reason:       "deduct", // 后续可细分为 order_created 等
	}
	if err := tx.Create(logEntry).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("写入库存日志失败: %w", err)
	}

	return tx.Commit().Error
}

// RollbackStock 回滚库存（加回）
func (r *stockRepository) RollbackStock(ctx context.Context, skuID string, quantity int64) error {
	if quantity <= 0 {
		return fmt.Errorf("回滚数量必须大于0")
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	res := tx.
		Model(&model.Stock{}).
		Where("sku_id = ?", skuID).
		Update("available_stock", gorm.Expr("available_stock + ?", quantity))

	if res.Error != nil {
		tx.Rollback()
		return fmt.Errorf("回滚库存失败: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		log.Printf("⚠️ RollbackStock: 未找到库存记录 sku_id=%s，忽略回滚", skuID)
		tx.Rollback()
		return nil
	}

	logEntry := &model.StockLog{
		SKUID:        skuID,
		ChangeAmount: +quantity,
		Reason:       "rollback",
	}
	if err := tx.Create(logEntry).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("写入库存日志失败: %w", err)
	}

	return tx.Commit().Error
}
