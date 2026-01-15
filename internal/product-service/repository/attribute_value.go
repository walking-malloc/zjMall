package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"zjMall/internal/product-service/model"

	"gorm.io/gorm"
)

type AttributeValueListFilter struct {
	Page        int32
	PageSize    int32
	AttributeID string
	Keyword     string
	Offset      int
	Limit       int
}

type AttributeValueRepository interface {
	CreateAttributeValue(ctx context.Context, attributeValue *model.AttributeValue) error
	GetAttributeValueByID(ctx context.Context, id string) (*model.AttributeValue, error)
	UpdateAttributeValue(ctx context.Context, attributeValue *model.AttributeValue) error
	DeleteAttributeValue(ctx context.Context, id string) error
	ListAttributeValues(ctx context.Context, filter *AttributeValueListFilter) ([]*model.AttributeValue, int64, error)

	GetAttributeValueBySkuID(ctx context.Context, skuIDs []string) ([]*model.AttributeValue, error)
}

type attributeValueRepository struct {
	db *gorm.DB
}

func NewAttributeValueRepository(db *gorm.DB) AttributeValueRepository {
	return &attributeValueRepository{
		db: db,
	}
}

func (r *attributeValueRepository) CreateAttributeValue(ctx context.Context, attributeValue *model.AttributeValue) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先检查对应 attribute_id 是否存在
		var attribute model.Attribute
		err := tx.Model(&model.Attribute{}).
			Where("id = ?", attributeValue.AttributeID).
			First(&attribute).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("属性不存在")
			}
			return err
		}

		err = tx.Create(attributeValue).Error
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("该属性下已存在同名属性值")
			}
			return err
		}
		return nil
	})
	return err
}

func (r *attributeValueRepository) GetAttributeValueByID(ctx context.Context, id string) (*model.AttributeValue, error) {
	var attributeValue model.AttributeValue
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&attributeValue).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attributeValue, nil
}

func (r *attributeValueRepository) UpdateAttributeValue(ctx context.Context, attributeValue *model.AttributeValue) error {
	err := r.db.WithContext(ctx).
		Model(&model.AttributeValue{}).
		Where("id = ?", attributeValue.ID).
		Updates(attributeValue).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("该属性下已存在同名属性值")
		}
		return err
	}
	return nil
}

func (r *attributeValueRepository) DeleteAttributeValue(ctx context.Context, id string) error {
	// 检查是否有 SKU 属性关联此属性值
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exists := false
		err := tx.Raw("SELECT EXISTS(SELECT 1 FROM sku_attributes WHERE attribute_value_id = ? AND deleted_at IS NULL)", id).
			Scan(&exists).Error
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("该属性值已被SKU使用，无法删除")
		}

		// 软删除属性值
		err = tx.Where("id = ?", id).Delete(&model.AttributeValue{}).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (r *attributeValueRepository) ListAttributeValues(ctx context.Context, filter *AttributeValueListFilter) ([]*model.AttributeValue, int64, error) {
	// 构建查询
	query := r.db.WithContext(ctx).Model(&model.AttributeValue{})

	// 添加过滤条件
	if filter.AttributeID != "" {
		query = query.Where("attribute_id = ?", filter.AttributeID)
	}
	if filter.Keyword != "" {
		// 安全处理LIKE查询
		safeKeyword := strings.ReplaceAll(filter.Keyword, `\`, `\\`)
		safeKeyword = strings.ReplaceAll(safeKeyword, "%", "\\%")
		safeKeyword = strings.ReplaceAll(safeKeyword, "_", "\\_")
		query = query.Where("value LIKE ?", "%"+safeKeyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var attributeValues []*model.AttributeValue
	err := query.
		Order("sort_order DESC, created_at ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&attributeValues).Error

	if err != nil {
		return nil, 0, err
	}

	return attributeValues, total, nil
}

func (r *attributeValueRepository) GetAttributeValueBySkuID(ctx context.Context, skuIDs []string) ([]*model.AttributeValue, error) {
	var attributeValues []*model.AttributeValue
	//通过sku_attribute表查询attribute_value_id
	err := r.db.WithContext(ctx).Model(&model.AttributeValue{}).
		Joins("JOIN sku_attributes ON attribute_values.id = sku_attributes.attribute_value_id").
		Where("sku_attributes.sku_id IN (?)", skuIDs).
		Find(&attributeValues).Error
	if err != nil {
		return nil, err
	}
	return attributeValues, nil

}
