package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"zjMall/internal/product-service/model"

	"gorm.io/gorm"
)

const (
	SkuStatusOnShelf  = 1
	SkuStatusOffShelf = 2
	SkuStatusDisabled = 3
)

type SkuListFilter struct {
	Page      int32
	PageSize  int32
	ProductID string
	Status    int32
	Keyword   string
	MinPrice  float64
	MaxPrice  float64
	Offset    int
	Limit     int
}

type SkuRepository interface {
	CreateSku(ctx context.Context, sku *model.Sku) error
	CreateSkuWithAttributes(ctx context.Context, sku *model.Sku, attributeValueIDs []string) error
	GetSkuByID(ctx context.Context, id string) (*model.Sku, error)
	UpdateSku(ctx context.Context, sku *model.Sku) error
	DeleteSku(ctx context.Context, id string) error
	ListSkus(ctx context.Context, filter *SkuListFilter) ([]*model.Sku, int64, error)
	BatchCreateSkus(ctx context.Context, productID string, skus []*model.Sku) error

	// SKU 属性关联相关方法
	AddSkuAttribute(ctx context.Context, skuID, attributeValueID string) error
	RemoveSkuAttribute(ctx context.Context, skuID, attributeValueID string) error
	GetSkuAttributes(ctx context.Context, skuID string) ([]*model.AttributeValue, error)
	BatchSetSkuAttributes(ctx context.Context, skuID string, attributeValueIDs []string) error
	// GetMinPriceByProductIDs 批量获取商品最低SKU价格，返回 product_id -> min_price
	GetMinPriceByProductIDs(ctx context.Context, productIDs []string) (map[string]float64, error)
}

type skuRepository struct {
	db *gorm.DB
}

func NewSkuRepository(db *gorm.DB) SkuRepository {
	return &skuRepository{
		db: db,
	}
}

func (r *skuRepository) CreateSku(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

// CreateSkuWithAttributes 创建SKU并同时设置属性关联（事务）
func (r *skuRepository) CreateSkuWithAttributes(ctx context.Context, sku *model.Sku, attributeValueIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建SKU
		if err := tx.Create(sku).Error; err != nil {
			return err
		}

		// 2. 如果有属性值ID列表，批量设置属性关联
		if len(attributeValueIDs) > 0 {
			// 去重
			attributeValueIDMap := make(map[string]struct{}, len(attributeValueIDs))
			uniqueAttributeValueIDs := make([]string, 0, len(attributeValueIDs))
			for _, id := range attributeValueIDs {
				if id != "" {
					if _, exists := attributeValueIDMap[id]; !exists {
						attributeValueIDMap[id] = struct{}{}
						uniqueAttributeValueIDs = append(uniqueAttributeValueIDs, id)
					}
				}
			}

			if len(uniqueAttributeValueIDs) > 0 {
				// 检查所有属性值是否存在
				var count int64
				if err := tx.Model(&model.AttributeValue{}).
					Where("id IN ?", uniqueAttributeValueIDs).
					Count(&count).Error; err != nil {
					return err
				}
				if int(count) != len(uniqueAttributeValueIDs) {
					return fmt.Errorf("部分属性值不存在")
				}

				// 批量插入关联
				skuAttributes := make([]*model.SkuAttribute, 0, len(uniqueAttributeValueIDs))
				for _, attributeValueID := range uniqueAttributeValueIDs {
					skuAttributes = append(skuAttributes, &model.SkuAttribute{
						SkuID:            sku.ID,
						AttributeValueID: attributeValueID,
					})
				}
				if err := tx.CreateInBatches(skuAttributes, 50).Error; err != nil {
					if strings.Contains(err.Error(), "Duplicate entry") ||
						strings.Contains(err.Error(), "UNIQUE constraint") {
						return fmt.Errorf("部分SKU属性关联已存在")
					}
					return err
				}
			}
		}

		return nil
	})
}

func (r *skuRepository) GetSkuByID(ctx context.Context, id string) (*model.Sku, error) {
	var sku model.Sku
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sku, nil
}

func (r *skuRepository) UpdateSku(ctx context.Context, sku *model.Sku) error {
	return r.db.WithContext(ctx).
		Model(&model.Sku{}).
		Where("id = ?", sku.ID).
		Updates(sku).Error
}

func (r *skuRepository) DeleteSku(ctx context.Context, id string) error {
	// 直接软删除 SKU 记录（如有 SKU 属性关联，可在后续属性仓库中处理级联）
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Sku{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("SKU 不存在")
		}
		return err
	}
	return nil
}

func (r *skuRepository) ListSkus(ctx context.Context, filter *SkuListFilter) ([]*model.Sku, int64, error) {
	// 合法化分页参数
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	query := r.db.WithContext(ctx).Model(&model.Sku{})

	// 过滤条件
	if filter.ProductID != "" {
		query = query.Where("product_id = ?", filter.ProductID)
	}
	if filter.Status > 0 {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		safeKeyword := strings.ReplaceAll(filter.Keyword, `\`, `\\`)
		safeKeyword = strings.ReplaceAll(safeKeyword, "%", "\\%")
		safeKeyword = strings.ReplaceAll(safeKeyword, "_", "\\_")
		query = query.Where("(sku_code LIKE ? OR barcode LIKE ? OR name LIKE ?)",
			"%"+safeKeyword+"%",
			"%"+safeKeyword+"%",
			"%"+safeKeyword+"%")
	}
	if filter.MinPrice > 0 {
		query = query.Where("price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("price <= ?", filter.MaxPrice)
	}

	// 总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := int((filter.Page - 1) * filter.PageSize)
	var skus []*model.Sku
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(int(filter.PageSize)).
		Find(&skus).Error; err != nil {
		return nil, 0, err
	}

	return skus, total, nil
}

func (r *skuRepository) BatchCreateSkus(ctx context.Context, productID string, skus []*model.Sku) error {
	if len(skus) == 0 {
		return nil
	}

	// 设置 product_id，防止调用方漏填
	for _, sku := range skus {
		sku.ProductID = productID
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(skus, 50).Error; err != nil {
			// 处理唯一键冲突（如 sku_code）
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("部分 SKU 编码已存在")
			}
			return err
		}
		return nil
	})
}

// AddSkuAttribute 添加 SKU 属性关联
func (r *skuRepository) AddSkuAttribute(ctx context.Context, skuID, attributeValueID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查 SKU 是否存在
		var sku model.Sku
		if err := tx.Where("id = ?", skuID).First(&sku).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("SKU 不存在: %s", skuID)
			}
			return err
		}

		// 检查属性值是否存在
		var attrValue model.AttributeValue
		if err := tx.Where("id = ?", attributeValueID).First(&attrValue).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("属性值不存在: %s", attributeValueID)
			}
			return err
		}

		// 创建关联（唯一索引会自动防止重复）
		skuAttr := &model.SkuAttribute{
			SkuID:            skuID,
			AttributeValueID: attributeValueID,
		}
		if err := tx.Create(skuAttr).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("该 SKU 已关联该属性值")
			}
			return err
		}

		return nil
	})
}

// RemoveSkuAttribute 删除 SKU 属性关联
func (r *skuRepository) RemoveSkuAttribute(ctx context.Context, skuID, attributeValueID string) error {
	result := r.db.WithContext(ctx).
		Where("sku_id = ? AND attribute_value_id = ?", skuID, attributeValueID).
		Delete(&model.SkuAttribute{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("SKU 属性关联不存在")
	}
	return nil
}

// GetSkuAttributes 查询 SKU 的属性值列表
func (r *skuRepository) GetSkuAttributes(ctx context.Context, skuID string) ([]*model.AttributeValue, error) {
	var attrs []*model.AttributeValue

	err := r.db.WithContext(ctx).
		Table("attribute_values").
		Joins("INNER JOIN sku_attributes ON sku_attributes.attribute_value_id = attribute_values.id").
		Where("sku_attributes.sku_id = ?", skuID).
		Order("attribute_values.sort_order ASC, attribute_values.created_at ASC").
		Find(&attrs).Error
	if err != nil {
		return nil, err
	}

	return attrs, nil
}

// BatchSetSkuAttributes 批量设置 SKU 属性关联（先删除旧的，再插入新的）
func (r *skuRepository) BatchSetSkuAttributes(ctx context.Context, skuID string, attributeValueIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查 SKU 是否存在
		var sku model.Sku
		if err := tx.Where("id = ?", skuID).First(&sku).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("SKU 不存在: %s", skuID)
			}
			return err
		}

		// 去重属性值 ID
		if len(attributeValueIDs) > 0 {
			seen := make(map[string]struct{}, len(attributeValueIDs))
			unique := make([]string, 0, len(attributeValueIDs))
			for _, id := range attributeValueIDs {
				if id == "" {
					continue
				}
				if _, ok := seen[id]; !ok {
					seen[id] = struct{}{}
					unique = append(unique, id)
				}
			}
			attributeValueIDs = unique
		}

		// 检查属性值是否全部存在
		if len(attributeValueIDs) > 0 {
			var count int64
			if err := tx.Model(&model.AttributeValue{}).
				Where("id IN ?", attributeValueIDs).
				Count(&count).Error; err != nil {
				return err
			}
			if int(count) != len(attributeValueIDs) {
				return fmt.Errorf("部分属性值不存在")
			}
		}

		// 删除旧关联
		if err := tx.Where("sku_id = ?", skuID).Delete(&model.SkuAttribute{}).Error; err != nil {
			return err
		}

		// 插入新关联
		if len(attributeValueIDs) > 0 {
			records := make([]*model.SkuAttribute, 0, len(attributeValueIDs))
			for _, id := range attributeValueIDs {
				records = append(records, &model.SkuAttribute{
					SkuID:            skuID,
					AttributeValueID: id,
				})
			}
			if err := tx.CreateInBatches(records, 50).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetMinPriceByProductIDs 批量获取商品最低SKU价格（仅上架且未删除）
func (r *skuRepository) GetMinPriceByProductIDs(ctx context.Context, productIDs []string) (map[string]float64, error) {
	if len(productIDs) == 0 {
		return map[string]float64{}, nil
	}
	type minPriceRow struct {
		ProductID string  `gorm:"column:product_id"`
		MinPrice  float64 `gorm:"column:min_price"`
	}
	var rows []minPriceRow
	err := r.db.WithContext(ctx).Model(&model.Sku{}).
		Select("product_id, MIN(price) as min_price").
		Where("product_id IN ? AND status = ?", productIDs, SkuStatusOnShelf).
		Where("deleted_at IS NULL").
		Group("product_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.ProductID] = row.MinPrice
	}
	return result, nil
}
