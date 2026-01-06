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
	CategoryAllListCacheKey = "product:category:list:all"  //所有类目扁平列表
	CategoryTreeCacheKey    = "product:category:tree"      //完整树形结构
	CategoryDetailCacheKey  = "product:category:detail:%s" //单个类目详情
	CategoryNullCacheKey    = "product:category:null:%s"   //空值缓存

	//singleflight(防止缓存击穿, 5分钟过期)
	SingleFlightDetailKey = "product:category:sf:detail:%s" //单个类目详情, 5分钟过期
	SingleFlightTreeKey   = "product:category:sf:tree"      //完整树形结构, 5分钟过期
	SingleFlightListKey   = "product:category:sf:list"      //所有类目扁平列表, 5分钟过期
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
}

type CategoryListFliter struct {
	ParentID  string
	Level     int32
	Status    int32
	IsVisible *bool
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
		log.Printf("[CategoryRepository] CreateCategory DB error: %v", err)
		return err
	}
	log.Printf("[CategoryRepository] CreateCategory success, id=%s, name=%s, parent_id=%s", category.ID, category.Name, category.ParentID)
	go func() {
		// 删除所有相关缓存
		log.Printf("[CategoryRepository] CreateCategory invalidate caches for id=%s", category.ID)
		r.cacheRepo.Delete(context.Background(), CategoryAllListCacheKey)
		r.cacheRepo.Delete(context.Background(), CategoryTreeCacheKey)
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(CategoryDetailCacheKey, category.ID))
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(CategoryNullCacheKey, category.ID))
	}()
	return nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	// 先查缓存
	key := fmt.Sprintf(CategoryDetailCacheKey, id)
	result, err := r.getFromCache(ctx, key)
	if err == nil && result != nil {
		log.Printf("[CategoryRepository] GetCategoryByID cache hit, id=%s", id)
		return result, nil
	} //如果缓存查到可以直接返回

	//检查空值缓存
	nullKey := fmt.Sprintf(CategoryNullCacheKey, id)
	nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
	if nullResult == "1" {
		log.Printf("[CategoryRepository] GetCategoryByID null-cache hit, id=%s", id)
		return nil, nil
	}

	// 使用singleflight防止缓存击穿
	sfKey := fmt.Sprintf(SingleFlightDetailKey, id)
	log.Printf("[CategoryRepository] GetCategoryByID cache miss, enter singleflight, id=%s", id)
	sfResult, err, _ := r.sf.Do(sfKey, func() (interface{}, error) {
		//检查空值缓存
		nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
		if nullResult == "1" {
			log.Printf("[CategoryRepository] GetCategoryByID null-cache hit inside singleflight, id=%s", id)
			return nil, nil
		}
		//如果没有空值，则从数据库查询
		var category model.Category
		err = r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) { //如果记录不存在，设置缓存为null防止缓存穿透
				log.Printf("[CategoryRepository] GetCategoryByID record not found in DB, id=%s", id)
				go func() {
					r.cacheRepo.Set(context.Background(), nullKey, "1", 5*time.Minute)
				}()
				return nil, nil
			}
			log.Printf("[CategoryRepository] GetCategoryByID DB error, id=%s, err=%v", id, err)
			return nil, err
		}

		//异步写缓存
		go func() {
			if err := r.setToCache(context.Background(), key, &category, 5*time.Minute); err != nil {
				log.Printf("[CategoryRepository] GetCategoryByID setToCache failed, id=%s, err=%v", id, err)
			}
		}()
		return &category, nil
	})

	if err != nil {
		log.Printf("[CategoryRepository] GetCategoryByID singleflight error, id=%s, err=%v", id, err)
		return nil, err
	}
	if sfResult == nil {
		log.Printf("[CategoryRepository] GetCategoryByID singleflight result is nil, id=%s", id)
		return nil, nil
	}
	category, ok := sfResult.(*model.Category)
	if !ok {
		log.Printf("[CategoryRepository] GetCategoryByID type assertion failed, id=%s", id)
		return nil, fmt.Errorf("type assertion to *model.Category failed or category is nil")
	}
	log.Printf("[CategoryRepository] GetCategoryByID success from DB, id=%s", id)
	return category, nil
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, category *model.Category) error {
	err := r.db.WithContext(ctx).Where("id = ?", category.ID).Updates(category).Error
	if err != nil {
		log.Printf("[CategoryRepository] UpdateCategory DB error, id=%s, err=%v", category.ID, err)
		return err
	}
	log.Printf("[CategoryRepository] UpdateCategory success, id=%s", category.ID)
	go func() {
		log.Printf("[CategoryRepository] UpdateCategory invalidate caches for id=%s", category.ID)
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryDetailCacheKey, category.ID)) //删除单个类目详情缓存
		r.cacheRepo.Delete(ctx, CategoryAllListCacheKey)                          //删除所有类目扁平列表缓存
		r.cacheRepo.Delete(ctx, CategoryTreeCacheKey)                             //删除完整树形结构缓存
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryNullCacheKey, category.ID))   //删除空值缓存
	}()
	return nil
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id string) error {
	var category model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		log.Printf("[CategoryRepository] DeleteCategory find DB error, id=%s, err=%v", id, err)
		return err
	}
	err = r.db.WithContext(ctx).Delete(&category).Error //软删除
	if err != nil {
		log.Printf("[CategoryRepository] DeleteCategory DB delete error, id=%s, err=%v", id, err)
		return err
	}
	log.Printf("[CategoryRepository] DeleteCategory success (soft delete), id=%s", id)
	go func() {
		log.Printf("[CategoryRepository] DeleteCategory invalidate caches for id=%s", category.ID)
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryDetailCacheKey, category.ID)) //删除单个类目详情缓存
		r.cacheRepo.Delete(ctx, CategoryAllListCacheKey)                          //删除所有类目扁平列表缓存
		r.cacheRepo.Delete(ctx, CategoryTreeCacheKey)                             //删除完整树形结构缓存
		r.cacheRepo.Delete(ctx, fmt.Sprintf(CategoryNullCacheKey, category.ID))   //删除空值缓存
	}()
	return err
}

func (r *categoryRepository) ListCategories(ctx context.Context, filter *CategoryListFliter) ([]*model.Category, error) {
	log.Printf("[CategoryRepository] ListCategories called, filter=%+v", filter)
	//先获取所有类目扁平列表
	result, err := r.cacheRepo.Get(ctx, CategoryAllListCacheKey)
	if err == nil && result != "" {
		log.Printf("[CategoryRepository] ListCategories cache hit for all list")
		var categories []*model.Category
		err = json.Unmarshal([]byte(result), &categories)
		if err != nil {
			log.Printf("[CategoryRepository] ListCategories unmarshal cache error: %v", err)
			return nil, err
		}
		filteredCategories := filterByParentID(categories, filter)
		log.Printf("[CategoryRepository] ListCategories return from cache, total=%d, filtered=%d", len(categories), len(filteredCategories))
		return filteredCategories, nil
	}
	if err != nil {
		log.Printf("[CategoryRepository] ListCategories cache Get error (will fallback to DB): %v", err)
	} else {
		log.Printf("[CategoryRepository] ListCategories cache miss for all list")
	}
	//如果缓存没有，则从数据库查询
	log.Printf("[CategoryRepository] ListCategories enter singleflight for all list")
	sfResult, err, _ := r.sf.Do(SingleFlightListKey, func() (interface{}, error) {
		var categories []*model.Category
		err = r.db.WithContext(ctx).Order("sort_order Desc").Find(&categories).Error
		if err != nil {
			log.Printf("[CategoryRepository] ListCategories DB error: %v", err)
			return nil, err
		}
		go func() {
			data, _ := json.Marshal(categories)
			if cacheErr := r.cacheRepo.Set(ctx, CategoryAllListCacheKey, string(data), time.Hour*24); cacheErr != nil {
				log.Printf("[CategoryRepository] ListCategories set all-list cache error: %v", cacheErr)
			}
		}()

		filteredCategories := filterByParentID(categories, filter)
		log.Printf("[CategoryRepository] ListCategories DB query done, total=%d, filtered=%d", len(categories), len(filteredCategories))
		return filteredCategories, nil
	})
	if err != nil {
		log.Printf("[CategoryRepository] ListCategories singleflight error: %v", err)
		return nil, err
	}
	if sfResult == nil {
		log.Printf("[CategoryRepository] ListCategories singleflight result nil")
		return nil, nil
	}
	categories := sfResult.([]*model.Category)
	log.Printf("[CategoryRepository] ListCategories return from DB/singleflight, count=%d", len(categories))
	return categories, nil
}

func (r *categoryRepository) GetCategoryTree(ctx context.Context, maxlevel int32, isvisible bool, status int32) ([]*CategoryTreeNode, error) {
	log.Printf("[CategoryRepository] GetCategoryTree called, maxlevel=%d, isvisible=%v, status=%d", maxlevel, isvisible, status)
	result, err := r.cacheRepo.Get(ctx, CategoryTreeCacheKey)
	if err != nil {
		log.Printf("[CategoryRepository] GetCategoryTree cache Get error: %v", err)
		return nil, err
	}
	if result != "" {
		log.Printf("[CategoryRepository] GetCategoryTree cache hit")
		var tree []*CategoryTreeNode
		err = json.Unmarshal([]byte(result), &tree)
		return tree, nil
	}
	//如果缓存没有，则从redis查询扁平列表并构建树形结构
	allcategories, err := r.ListCategories(ctx, &CategoryListFliter{
		Level:     maxlevel,
		IsVisible: &isvisible,
		Status:    status,
	})
	if err != nil {
		log.Printf("[CategoryRepository] GetCategoryTree ListCategories error: %v", err)
		return nil, err
	}
	log.Printf("[CategoryRepository] GetCategoryTree fetched flat list, count=%d", len(allcategories))
	tree := buildTreeFromFlatList(allcategories)
	go func() {
		data, _ := json.Marshal(tree)
		if cacheErr := r.cacheRepo.Set(ctx, CategoryTreeCacheKey, string(data), time.Hour*24); cacheErr != nil {
			log.Printf("[CategoryRepository] GetCategoryTree set tree cache error: %v", cacheErr)
		}
	}()
	log.Printf("[CategoryRepository] GetCategoryTree built tree, root_count=%d", len(tree))
	return tree, nil
}
func filterByParentID(allcategories []*model.Category, filter *CategoryListFliter) []*model.Category {
	if len(allcategories) == 0 || filter == nil {
		return allcategories
	}
	var result []*model.Category = make([]*model.Category, 0, len(allcategories))
	for _, category := range allcategories {
		if filter.ParentID != "" && category.ParentID != filter.ParentID {
			continue
		}
		if filter.Level > 0 && category.Level != (int8)(filter.Level) {
			continue
		}
		if filter.Status > 0 && category.Status != (int8)(filter.Status) {
			continue
		}
		if filter.IsVisible != nil && category.IsVisible != *filter.IsVisible {
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
