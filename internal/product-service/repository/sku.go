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
	GetSkuByID(ctx context.Context, id string) (*model.Sku, error)
	UpdateSku(ctx context.Context, sku *model.Sku) error
	DeleteSku(ctx context.Context, id string) error
	ListSkus(ctx context.Context, filter *SkuListFilter) ([]*model.Sku, int64, error)
	BatchCreateSkus(ctx context.Context, productID string, skus []*model.Sku) error
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
