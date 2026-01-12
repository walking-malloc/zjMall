package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"zjMall/internal/common/cache"
	"zjMall/internal/product-service/model"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type TagListFilter struct {
	Page     int32
	PageSize int32
	Type     int32
	Status   int32
	Keyword  string
	Offset   int
	Limit    int
}

type TagRepository interface {
	CreateTag(ctx context.Context, tag *model.Tag) error
	GetTagByID(ctx context.Context, id string) (*model.Tag, error)
	UpdateTag(ctx context.Context, tag *model.Tag) error
	DeleteTag(ctx context.Context, id string) error
	ListTags(ctx context.Context, filter *TagListFilter) ([]*model.Tag, int64, error)
}

type tagRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
	sf        singleflight.Group
}

func NewTagRepository(db *gorm.DB, cacheRepo cache.CacheRepository, sf singleflight.Group) TagRepository {
	return &tagRepository{
		db:        db,
		cacheRepo: cacheRepo,
		sf:        sf,
	}
}

func (r *tagRepository) CreateTag(ctx context.Context, tag *model.Tag) error {

	err := r.db.WithContext(ctx).Create(tag).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("标签名称已存在")
		}
		return err
	}
	return nil
}

func (r *tagRepository) GetTagByID(ctx context.Context, id string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) UpdateTag(ctx context.Context, tag *model.Tag) error {
	err := r.db.WithContext(ctx).
		Where("id = ?", tag.ID).
		Updates(tag).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("标签名称已存在")
		}
		return err
	}
	return nil
}

func (r *tagRepository) DeleteTag(ctx context.Context, id string) error {
	// 检查是否有商品关联此标签
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除tag和商品关联的数据
		err := tx.Where("tag_id = ?", id).Delete(&model.ProductTag{}).Error
		if err != nil {
			return err
		}
		// 删除tag
		err = r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tag{}).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("标签不存在")
			}
			return err
		}
		return nil
	})
	return err
}

func (r *tagRepository) ListTags(ctx context.Context, filter *TagListFilter) ([]*model.Tag, int64, error) {
	// 构建查询
	query := r.db.WithContext(ctx).Model(&model.Tag{})

	// 添加过滤条件
	if filter.Type > 0 {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status > 0 {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		// 安全处理LIKE查询
		safeKeyword := strings.ReplaceAll(filter.Keyword, `\`, `\\`)
		safeKeyword = strings.ReplaceAll(filter.Keyword, "%", "\\%")
		safeKeyword = strings.ReplaceAll(safeKeyword, "_", "\\_")
		query = query.Where("name LIKE ?", "%"+safeKeyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var tags []*model.Tag
	err := query.
		Order("sort_order DESC, created_at ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Find(&tags).Error

	if err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}
