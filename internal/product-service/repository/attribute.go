package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"zjMall/internal/product-service/model"

	"gorm.io/gorm"
)

type AttributeListFilter struct {
	Page       int32
	PageSize   int32
	CategoryID string
	Type       int32
	IsRequired int32
	Keyword    string
	Offset     int
	Limit      int
}

type AttributeRepository interface {
	CreateAttribute(ctx context.Context, attribute *model.Attribute) error
	GetAttributeByID(ctx context.Context, id string) (*model.Attribute, error)
	UpdateAttribute(ctx context.Context, attribute *model.Attribute) error
	DeleteAttribute(ctx context.Context, id string) error
	ListAttributes(ctx context.Context, filter *AttributeListFilter) ([]*model.Attribute, int64, error)
}

type attributeRepository struct {
	db *gorm.DB
}

func NewAttributeRepository(db *gorm.DB) AttributeRepository {
	return &attributeRepository{
		db: db,
	}
}

func (r *attributeRepository) CreateAttribute(ctx context.Context, attribute *model.Attribute) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//先检查对应category_id是否存在
		var category model.Category
		err := tx.Model(&model.Category{}).
			Where("id = ?", attribute.CategoryID).
			First(&category).Error
		if err != nil {
			return err
		}

		err = tx.Create(attribute).Error
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("该类目下已存在同名属性")
			}
			return err
		}
		return nil
	})
	return err
}

func (r *attributeRepository) GetAttributeByID(ctx context.Context, id string) (*model.Attribute, error) {
	var attribute model.Attribute
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&attribute).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attribute, nil
}

func (r *attributeRepository) UpdateAttribute(ctx context.Context, attribute *model.Attribute) error {
	err := r.db.WithContext(ctx).
		Model(&model.Attribute{}).
		Where("id = ?", attribute.ID).
		Updates(attribute).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("该类目下已存在同名属性")
		}
		return err
	}
	return nil
}

func (r *attributeRepository) DeleteAttribute(ctx context.Context, id string) error {
	// 检查是否有属性值关联此属性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exists := false
		err := tx.Raw("SELECT EXISTS(SELECT 1 FROM attribute_values WHERE attribute_id = ? AND deleted_at IS NULL) as exists", id).
			Scan(&exists).Error
		if exists {
			return fmt.Errorf("该属性下存在属性值，无法删除")
		}
		// 删除属性
		err = tx.Where("id = ?", id).Delete(&model.Attribute{}).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (r *attributeRepository) ListAttributes(ctx context.Context, filter *AttributeListFilter) ([]*model.Attribute, int64, error) {
	// 构建查询
	query := r.db.WithContext(ctx).Model(&model.Attribute{})

	// 添加过滤条件
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.Type > 0 {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.IsRequired >= 0 {
		query = query.Where("is_required = ?", filter.IsRequired)
	}
	if filter.Keyword != "" {
		// 安全处理LIKE查询
		safeKeyword := strings.ReplaceAll(filter.Keyword, `\`, `\\`)
		safeKeyword = strings.ReplaceAll(safeKeyword, "%", "\\%")
		safeKeyword = strings.ReplaceAll(safeKeyword, "_", "\\_")
		query = query.Where("name LIKE ?", "%"+safeKeyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var attributes []*model.Attribute
	err := query.
		Order("sort_order DESC, created_at ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&attributes).Error

	if err != nil {
		return nil, 0, err
	}

	return attributes, total, nil
}
