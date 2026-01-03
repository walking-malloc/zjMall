package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"zjMall/internal/common/cache"
	"zjMall/internal/product-service/model"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	CategoryCachePrefix     = "product:category:%s"
	CategoryListCachePrefix = "product:category:list:%s"
	SingleFlightGroupKey    = "product:category:singleflight:%s"
	CategoryNullCacheKey    = "product:category:null:%s"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *model.Category) error
	GetCategoryByID(ctx context.Context, id string) (*model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) error
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context, parentID string) ([]*model.Category, error)

	getFromCache(ctx context.Context, key string) (*model.Category, error)
	setToCache(ctx context.Context, key string, category *model.Category, expiration time.Duration) error
}

type categoryRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
	sf        singleflight.Group //防止缓存击穿
}

func NewCategoryRepository(db *gorm.DB, cacheRepo cache.CacheRepository) CategoryRepository {
	return &categoryRepository{
		db:        db,
		cacheRepo: cacheRepo,
	}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category *model.Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return err
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryCachePrefix, category.ID))
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryListCachePrefix, category.ParentID))
	}()
	return nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	// 先查缓存
	key := fmt.Sprintf(CategoryCachePrefix, id)
	result, err := r.getFromCache(ctx, key)
	if err == nil && result != nil {
		return result, nil
	}

	//检查空值缓存
	nullKey := fmt.Sprintf(CategoryNullCacheKey, id)
	nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
	if nullResult == "1" {
		return nil, nil
	}

	// 使用singleflight防止缓存击穿
	sfKey := fmt.Sprintf(SingleFlightGroupKey, id)
	sfResult, err, _ := r.sf.Do(sfKey, func() (interface{}, error) {
		//检查空值缓存
		nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
		if nullResult == "1" {
			return nil, nil
		}

		var category model.Category
		err = r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&category).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { //如果记录不存在，设置缓存为null防止缓存穿透
				go func() {
					r.cacheRepo.Set(context.Background(), nullKey, "1", 5*time.Minute)
				}()
				return nil, nil
			}
			return nil, err
		}

		//异步写缓存
		go func() {
			if err := r.setToCache(context.Background(), key, &category, time.Hour*24); err != nil {
				log.Printf("set to cache failed: %v", err)
			}
		}()
		return &category, nil
	})

	if err != nil {
		return nil, err
	}
	if sfResult == nil {
		return nil, nil
	}
	category, ok := sfResult.(*model.Category)
	if !ok {
		return nil, fmt.Errorf("type assertion to *model.Category failed or category is nil")
	}
	return category, nil
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, category *model.Category) error {
	err := r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", category.ID).Updates(category).Error
	if err != nil {
		return err
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryCachePrefix, category.ID))
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryListCachePrefix, category.ParentID))
	}()
	return nil
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id string) error {
	var category *model.Category
	err := r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).First(category).Error
	if err != nil {
		return err
	}

	if category == nil {
		return fmt.Errorf("category not found")
	}

	category.DeletedAt = time.Now()
	err = r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).Updates(category).Error
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryCachePrefix, id))
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryListCachePrefix, category.ParentID))
	}()
	return err
}

func (r *categoryRepository) ListCategories(ctx context.Context, parentID string) ([]*model.Category, error) {
	var categories []*model.Category
	//先查缓存
	key := fmt.Sprintf(CategoryListCachePrefix, parentID)
	result, err := r.cacheRepo.Get(ctx, key)
	if err == nil && result != "" {
		err = json.Unmarshal([]byte(result), &categories)
		if err == nil {
			return categories, nil
		}
	} //反序列化失败，从数据库查询

	//从数据库查询
	if parentID == "" {
		err = r.db.WithContext(ctx).Model(&model.Category{}).Where("parent_id IS NULL").Order("sort_order ASC").Find(&categories).Error
	} else {
		err = r.db.WithContext(ctx).Model(&model.Category{}).Where("parent_id = ?", parentID).Order("sort_order ASC").Find(&categories).Error
	}

	if err != nil {
		return nil, err
	}

	go func() {
		data, _ := json.Marshal(categories)
		r.cacheRepo.Set(ctx, fmt.Sprintf(CategoryListCachePrefix, parentID), string(data), time.Hour*24)
	}()
	return categories, nil
}

func (r *categoryRepository) getFromCache(ctx context.Context, key string) (*model.Category, error) {
	result, err := r.cacheRepo.Get(ctx, key)
	if err == nil && result != "" {
		var category model.Category
		err = json.Unmarshal([]byte(result), &category)
		if err == nil {
			return &category, nil
		}
	}
	return nil, err
}
func (r *categoryRepository) setToCache(ctx context.Context, key string, category *model.Category, expiration time.Duration) error {
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}
	err = r.cacheRepo.Set(ctx, key, string(data), expiration)
	if err != nil {
		return err
	}

	return nil
}

func getCategoryListCacheKey(parentID string) string {
	if parentID == "" {
		return fmt.Sprintf(CategoryListCachePrefix, "root")
	}
	return fmt.Sprintf(CategoryListCachePrefix, parentID)
}
