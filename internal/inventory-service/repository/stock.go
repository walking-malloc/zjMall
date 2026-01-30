package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"zjMall/internal/inventory-service/model"

	"gorm.io/gorm"
)

// DeductItem 表示单个 SKU 的扣减请求
type DeductItem struct {
	SKUID    string
	Quantity int64
}

// StockRepository 库存仓储接口
type StockRepository interface {
	// GetBySKUID 根据 skuID 查询库存
	GetBySKUID(ctx context.Context, skuID string) (*model.Stock, error)
	// BatchGetBySKUID 批量查询库存
	BatchGetBySKUID(ctx context.Context, skuIDs []string) (map[string]*model.Stock, error)
	// TryDeductStocks 批量尝试扣减库存（使用乐观锁，防止超卖，支持幂等性检查）
	// orderNo: 订单号，用于日志记录和幂等性检查
	// items: 需要扣减的SKU列表
	TryDeductStocks(ctx context.Context, orderNo string, items []DeductItem) error
	// RollbackStocks 批量回滚库存（加回）
	// orderNo: 订单号，用于日志记录
	// items: 需要回滚的SKU列表
	RollbackStocks(ctx context.Context, orderNo string, items []DeductItem) error
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

// TryDeductStocks 批量尝试扣减库存（使用乐观锁，防止超卖，支持幂等性检查）
func (r *stockRepository) TryDeductStocks(ctx context.Context, orderNo string, items []DeductItem) error {
	if orderNo == "" {
		return fmt.Errorf("订单号不能为空")
	}
	if len(items) == 0 {
		return fmt.Errorf("扣减项不能为空")
	}

	// 参数校验
	for _, item := range items {
		if item.SKUID == "" || item.Quantity <= 0 {
			return fmt.Errorf("非法库存扣减请求: sku_id=%s, quantity=%d", item.SKUID, item.Quantity)
		}
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 收集所有需要扣减的 SKU ID
	skuIDs := make([]string, 0, len(items))
	for _, item := range items {
		skuIDs = append(skuIDs, item.SKUID)
	}

	// 批量查询库存记录（包含版本号）
	var stocks []model.Stock
	if err := tx.Where("sku_id IN ?", skuIDs).Find(&stocks).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询库存失败: %w", err)
	}

	// 构建 SKU ID 到库存记录的映射
	stockMap := make(map[string]*model.Stock)
	for i := range stocks {
		stockMap[stocks[i].SKUID] = &stocks[i]
	}

	// 准备需要扣减的项，检查库存是否充足
	var toDeduct []DeductItem
	for _, item := range items {
		// 检查库存记录是否存在
		stock, exists := stockMap[item.SKUID]
		if !exists {
			tx.Rollback()
			return fmt.Errorf("SKU %s 的库存记录不存在", item.SKUID)
		}

		// 检查库存是否充足
		if stock.AvailableStock < item.Quantity {
			tx.Rollback()
			return fmt.Errorf("SKU %s 库存不足: 当前库存=%d, 需要扣减=%d", item.SKUID, stock.AvailableStock, item.Quantity)
		}

		toDeduct = append(toDeduct, item)
	}

	// 如果没有需要扣减的项，直接提交事务
	if len(toDeduct) == 0 {
		tx.Commit()
		return nil
	}

	// 批量更新库存（使用乐观锁）
	// 先尝试插入 log（幂等性检查），如果已存在则跳过整个扣减流程
	// UPDATE inventory_stocks
	// SET available_stock = available_stock - ?, version = version + 1
	// WHERE sku_id = ? AND available_stock >= ? AND version = ?
	for _, item := range toDeduct {
		stock := stockMap[item.SKUID]

		// 先尝试插入 log（幂等性检查：如果已存在，说明已经扣减过）
		logEntry := &model.StockLog{
			SKUID:        item.SKUID,
			ChangeAmount: -item.Quantity,
			Reason:       "deduct",
			RefID:        orderNo,
		}
		if err := tx.Create(logEntry).Error; err != nil {
			// 检查是否是唯一索引冲突（幂等性：同一个订单号重复扣减）
			if errors.Is(err, gorm.ErrDuplicatedKey) ||
				strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") ||
				strings.Contains(err.Error(), "duplicate key") {
				// 幂等：已经扣减过，跳过
				log.Printf("ℹ️ [StockRepository] TryDeductStocks: 订单 %s 已扣减过 SKU %s 的库存，幂等跳过", orderNo, item.SKUID)
				continue
			}
			tx.Rollback()
			return fmt.Errorf("写入库存日志失败 sku_id=%s: %w", item.SKUID, err)
		}

		// log 插入成功，执行库存扣减
		res := tx.Model(&model.Stock{}).
			Where("sku_id = ? AND available_stock >= ? AND version = ?", item.SKUID, item.Quantity, stock.Version).
			Updates(map[string]interface{}{
				"available_stock": gorm.Expr("available_stock - ?", item.Quantity),
				"version":         gorm.Expr("version + 1"),
			})

		if res.Error != nil {
			tx.Rollback()
			return fmt.Errorf("扣减库存失败 sku_id=%s: %w", item.SKUID, res.Error)
		}
		if res.RowsAffected == 0 {
			tx.Rollback()
			// RowsAffected=0 可能的原因：
			// 1. version 不匹配（乐观锁冲突，被其他请求修改）
			// 2. available_stock < quantity（库存不足）
			// 3. sku_id 不存在（但前面已经检查过，理论上不会发生）
			return fmt.Errorf("SKU %s 库存扣减失败: 可能被其他请求并发修改（乐观锁冲突）或库存不足（当前库存可能已不足 %d）", item.SKUID, item.Quantity)
		}
	}

	return tx.Commit().Error
}

// RollbackStocks 批量回滚库存（加回）
func (r *stockRepository) RollbackStocks(ctx context.Context, orderNo string, items []DeductItem) error {
	if orderNo == "" {
		return fmt.Errorf("订单号不能为空")
	}
	if len(items) == 0 {
		return nil // 空列表直接返回成功
	}

	// 参数校验
	for _, item := range items {
		if item.SKUID == "" || item.Quantity <= 0 {
			return fmt.Errorf("非法库存回滚请求: sku_id=%s, quantity=%d", item.SKUID, item.Quantity)
		}
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 收集所有需要回滚的 SKU ID
	skuIDs := make([]string, 0, len(items))
	for _, item := range items {
		skuIDs = append(skuIDs, item.SKUID)
	}

	// 批量查询库存记录（包含版本号）
	var stocks []model.Stock
	if err := tx.Where("sku_id IN ?", skuIDs).Find(&stocks).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询库存失败: %w", err)
	}

	// 构建 SKU ID 到库存记录的映射
	stockMap := make(map[string]*model.Stock)
	for i := range stocks {
		stockMap[stocks[i].SKUID] = &stocks[i]
	}

	// 批量更新库存（使用乐观锁）
	for _, item := range items {
		stock, exists := stockMap[item.SKUID]
		if !exists {
			log.Printf("⚠️ RollbackStocks: 未找到库存记录 sku_id=%s，跳过回滚", item.SKUID)
			continue
		}

		res := tx.Model(&model.Stock{}).
			Where("sku_id = ? AND version = ?", item.SKUID, stock.Version).
			Updates(map[string]interface{}{
				"available_stock": gorm.Expr("available_stock + ?", item.Quantity),
				"version":         gorm.Expr("version + 1"),
			})

		if res.Error != nil {
			tx.Rollback()
			return fmt.Errorf("回滚库存失败 sku_id=%s: %w", item.SKUID, res.Error)
		}
		if res.RowsAffected == 0 {
			log.Printf("⚠️ RollbackStocks: SKU %s 回滚失败: 可能被其他请求并发修改（乐观锁冲突），跳过", item.SKUID)
			continue
		}

		// 先尝试插入 log（幂等性检查：如果已存在，说明已经回滚过）
		logEntry := &model.StockLog{
			SKUID:        item.SKUID,
			ChangeAmount: +item.Quantity,
			Reason:       "rollback",
			RefID:        orderNo,
		}
		if err := tx.Create(logEntry).Error; err != nil {
			// 检查是否是唯一索引冲突（幂等性：同一个订单号重复回滚）
			if errors.Is(err, gorm.ErrDuplicatedKey) ||
				strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") ||
				strings.Contains(err.Error(), "duplicate key") {
				// 幂等：已经回滚过，跳过
				log.Printf("ℹ️ [StockRepository] RollbackStocks: 订单 %s 已回滚过 SKU %s 的库存，幂等跳过", orderNo, item.SKUID)
				continue
			}
			tx.Rollback()
			return fmt.Errorf("写入库存日志失败 sku_id=%s: %w", item.SKUID, err)
		}
	}

	return tx.Commit().Error
}
