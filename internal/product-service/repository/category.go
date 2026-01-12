package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
	"zjMall/internal/common/cache"
	"zjMall/internal/product-service/model"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	CategoryNullCacheKey = "product:category:null:%s" //空值缓存

)

// 树形结构（嵌套对象）
type CategoryTreeNode struct {
	ID        string
	Name      string
	Level     int8
	IsLeaf    bool
	IsVisible bool
	SortOrder int32
	Icon      string
	Status    int8
	Children  []*CategoryTreeNode // 子类目
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CategoryListFliter struct {
	ParentID  string
	Level     int32
	Status    int32
	IsVisible bool
	Keyword   string
	Offset    int32
	Limit     int32
}
type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *model.Category) error
	GetCategoryByID(ctx context.Context, id string) (*model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) error
	DeleteCategory(ctx context.Context, id string) error
	ListCategories(ctx context.Context, filter *CategoryListFliter) ([]*model.Category, error)
	GetCategoryTree(ctx context.Context, maxlevel int32, isvisible bool, status int32) ([]*CategoryTreeNode, error)

	getFromCache(ctx context.Context, key string) (*model.Category, error)
	setToCache(ctx context.Context, key string, category *model.Category, expiration time.Duration) error
}

type categoryRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
	sf        singleflight.Group //防止缓存击穿
}

func NewCategoryRepository(db *gorm.DB, cacheRepo cache.CacheRepository, sf singleflight.Group) CategoryRepository {
	return &categoryRepository{
		db:        db,
		cacheRepo: cacheRepo,
		sf:        sf,
	}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category *model.Category) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		// 检查是否是唯一性约束错误
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("类目名称已存在")
		}
		return err
	}
	log.Printf("[CategoryRepository] CreateCategory success, id=%s, name=%s, parent_id=%s", category.ID, category.Name, category.ParentID)

	return nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {

	nullKey := fmt.Sprintf(CategoryNullCacheKey, id)

	//检查空值缓存
	nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
	if nullResult == "1" {
		log.Printf("[CategoryRepository] GetCategoryByID null-cache hit inside singleflight, id=%s", id)
		return nil, nil
	}

	log.Printf("[CategoryRepository] GetCategoryByID cache miss, enter singleflight, id=%s", id)

	var category model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { //如果记录不存在，设置缓存为null防止缓存穿透
			log.Printf("[CategoryRepository] GetCategoryByID record not found in DB, id=%s", id)
			r.cacheRepo.Set(context.Background(), nullKey, "1", 5*time.Minute)
			return nil, nil
		}
		log.Printf("[CategoryRepository] GetCategoryByID DB error, id=%s, err=%v", id, err)
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, category *model.Category) error {
	//读取当前记录版本号
	var current model.Category
	err := r.db.WithContext(ctx).Where("id = ?", category.ID).First(&current).Error
	if err != nil {
		return err
	}
	category.Version = current.Version + 1
	updates := map[string]interface{}{
		"name":       category.Name,
		"is_leaf":    category.IsLeaf,    // false 也会更新
		"is_visible": category.IsVisible, // false 也会更新
		"sort_order": category.SortOrder, // 0 也会更新
		"icon":       category.Icon,
		"status":     category.Status,
		"version":    category.Version,
	}
	result := r.db.WithContext(ctx).Model(&model.Category{}).Where("id = ? and version = ?", category.ID, current.Version).Updates(updates)
	if result.Error != nil {
		log.Printf("[CategoryRepository] UpdateCategory DB error, id=%s, err=%v", category.ID, err)
		return err
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("version mismatch, current version: %d, new version: %d", current.Version, category.Version)
	}

	return nil
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id string) error {
	var category model.Category
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//先锁定该类目，避免子类目会创建
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", id).
			First(&category).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("category not found: %s", id)
			}
			return err
		}
		//先检查该类目下是否有子类目
		tx.Exec("SELECT 1 FROM categories WHERE parent_id = ? For Update", id)
		//再检查该类目下是否有商品和品牌，有就无法删除
		tx.Exec("SELECT 1 FROM products WHERE category_id = ? For update", id)
		tx.Exec("SELECT 1 FROM brands WHERE category_id = ? For update", id)

		//没有子类目，则删除该类目
		err = tx.Delete(&category).Error //软删除
		if err != nil {
			log.Printf("[CategoryRepository] DeleteCategory DB delete error, id=%s, err=%v", id, err)
			return err
		}
		//删除类目品牌关联
		err = tx.Where("category_id = ?", id).Delete(&model.BrandCategory{}).Error
		if err != nil {
			return err
		}
		log.Printf("[CategoryRepository] DeleteCategory success (soft delete), id=%s", id)
		return nil
	})

	return err
}

func (r *categoryRepository) ListCategories(ctx context.Context, filter *CategoryListFliter) ([]*model.Category, error) {

	var categories []*model.Category
	err := r.db.WithContext(ctx).Order("sort_order Desc").Find(&categories).Error
	if err != nil {
		log.Printf("[CategoryRepository] ListCategories DB error: %v", err)
		return nil, err
	}

	filteredCategories := filterCategories(categories, filter)
	log.Printf("[CategoryRepository] ListCategories DB query done, total=%d, filtered=%d", len(categories), len(filteredCategories))
	return filteredCategories, nil

}

func (r *categoryRepository) GetCategoryTree(ctx context.Context, maxlevel int32, isvisible bool, status int32) ([]*CategoryTreeNode, error) {

	allcategories, err := r.ListCategories(ctx, &CategoryListFliter{
		Level:     maxlevel,
		IsVisible: isvisible,
		Status:    status,
	})
	if err != nil {
		log.Printf("[CategoryRepository] GetCategoryTree ListCategories error: %v", err)
		return nil, err
	}
	tree := buildTreeFromFlatList(allcategories)

	return tree, nil
}
func filterCategories(allcategories []*model.Category, filter *CategoryListFliter) []*model.Category {
	if len(allcategories) == 0 || filter == nil {
		return allcategories
	}
	var result []*model.Category = make([]*model.Category, 0, len(allcategories))
	for _, category := range allcategories {
		if filter.ParentID != "" && category.ParentID != filter.ParentID {
			continue
		}
		if filter.Level > 0 && category.Level > (int8)(filter.Level) {
			continue
		}
		if filter.Status > 0 && category.Status != (int8)(filter.Status) {
			continue
		}
		if category.IsVisible != filter.IsVisible {
			continue
		}
		if filter.Keyword != "" && !strings.Contains(category.Name, filter.Keyword) {
			continue
		}
		result = append(result, category)
	}

	if filter.Offset > 0 && filter.Offset < int32(len(result)) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < int32(len(result)) {
		result = result[:filter.Limit]
	}
	return result
}
func buildTreeFromFlatList(flatList []*model.Category) []*CategoryTreeNode {
	if len(flatList) == 0 {
		return []*CategoryTreeNode{}
	}

	nodeMap := make(map[string]*CategoryTreeNode) //建立id与节点的映射
	var rootNodes []*CategoryTreeNode

	//第一遍遍历,构建节点映射
	for _, category := range flatList {
		node := &CategoryTreeNode{
			ID:        category.ID,
			Name:      category.Name,
			Level:     category.Level,
			IsLeaf:    category.IsLeaf,
			IsVisible: category.IsVisible,
			SortOrder: category.SortOrder,
			Icon:      category.Icon,
			Status:    category.Status,
			Children:  []*CategoryTreeNode{},
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		}
		nodeMap[category.ID] = node
	}

	//第二遍遍历
	for _, category := range flatList {
		node := nodeMap[category.ID]

		if category.ParentID == "" {
			rootNodes = append(rootNodes, node)
		} else {
			parent, ok := nodeMap[category.ParentID]
			if ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	sortTreeNodes(rootNodes)
	return rootNodes
}
func sortTreeNodes(nodes []*CategoryTreeNode) {
	//先按sort_order降序排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].SortOrder > nodes[j].SortOrder
	})
	for _, node := range nodes {
		if len(node.Children) > 0 {
			sortTreeNodes(node.Children)
		}
	}
}
func (r *categoryRepository) getFromCache(ctx context.Context, key string) (*model.Category, error) {
	result, err := r.cacheRepo.Get(ctx, key)
	if err == nil && result != "" {
		log.Printf("[CategoryRepository] getFromCache hit, key=%s", key)
		var category model.Category
		err = json.Unmarshal([]byte(result), &category)
		if err == nil {
			return &category, nil
		}
	}
	if err != nil {
		log.Printf("[CategoryRepository] getFromCache Get error, key=%s, err=%v", key, err)
	}
	return nil, err
}
func (r *categoryRepository) setToCache(ctx context.Context, key string, category *model.Category, expiration time.Duration) error {
	data, err := json.Marshal(category)
	if err != nil {
		log.Printf("[CategoryRepository] setToCache marshal error, key=%s, err=%v", key, err)
		return err
	}
	err = r.cacheRepo.Set(ctx, key, string(data), expiration)
	if err != nil {
		log.Printf("[CategoryRepository] setToCache Set error, key=%s, err=%v", key, err)
		return err
	}
	log.Printf("[CategoryRepository] setToCache success, key=%s, expiration=%s", key, expiration.String())
	return nil
}
