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
	BrandDetailCacheKey               = "product:brand:detail:%s"
	BrandListCacheKey                 = "product:brand:list"              // 全量列表缓存
	BrandGroupByLetterCacheKey        = "product:brand:groupbyletter:all" // 所有首字母分组缓存
	BrandNullCacheKey                 = "product:brand:null:%s"
	SingleFlightBrandDetailKey        = "product:brand:sf:detail:%s"
	SingleFlightBrandListKey          = "product:brand:sf:list"              // 固定key，因为缓存全量数据
	SingleFlightBrandGroupByLetterKey = "product:brand:sf:groupbyletter:all" // 固定key，因为缓存所有分组
)

type BrandListFliter struct {
	Offset      int
	Limit       int
	Status      int
	Keyword     string
	FirstLetter string
	Country     string
}

// BrandGroupByLetter 按首字母分组的品牌结构
type BrandGroupByLetter struct {
	FirstLetter string         // 首字母（如：A、B、C）
	Brands      []*model.Brand // 该首字母下的品牌列表
}
type BrandRepository interface {
	CreateBrand(ctx context.Context, brand *model.Brand) error
	GetBrandByID(ctx context.Context, id string) (*model.Brand, error)
	UpdateBrand(ctx context.Context, brand *model.Brand) error
	DeleteBrand(ctx context.Context, id string) error
	ListBrands(ctx context.Context, filter *BrandListFliter) ([]*model.Brand, error)
	GetBrandsByFirstLetter(ctx context.Context, status int32) ([]*BrandGroupByLetter, error)

	// 品牌类目关联方法
	AddBrandCategory(ctx context.Context, brandID, categoryID string) error
	RemoveBrandCategory(ctx context.Context, brandID, categoryID string) error
	GetBrandCategories(ctx context.Context, brandID string) ([]*model.Category, error)
	BatchSetBrandCategories(ctx context.Context, brandID string, categoryIDs []string) error
}

type brandRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
	sf        singleflight.Group
}

func NewBrandRepository(db *gorm.DB, cacheRepo cache.CacheRepository, sf singleflight.Group) BrandRepository {
	return &brandRepository{
		db:        db,
		cacheRepo: cacheRepo,
		sf:        sf,
	}
}
func (r *brandRepository) CreateBrand(ctx context.Context, brand *model.Brand) error {
	err := r.db.WithContext(ctx).Create(brand).Error
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("品牌名称已存在")
		}
		return err
	}
	newCtx := context.Background()
	// 删除详情缓存和空值缓存
	r.cacheRepo.Delete(newCtx, fmt.Sprintf(BrandDetailCacheKey, brand.ID))
	r.cacheRepo.Delete(newCtx, fmt.Sprintf(BrandNullCacheKey, brand.ID))
	// 删除列表缓存（因为新增了品牌）
	r.cacheRepo.Delete(newCtx, BrandListCacheKey)
	// 删除按首字母分组缓存（因为新增了品牌，所有分组都需要更新）
	r.cacheRepo.Delete(newCtx, BrandGroupByLetterCacheKey)

	return nil
}
func (r *brandRepository) GetBrandByID(ctx context.Context, id string) (*model.Brand, error) {

	//缓存中没有，则从数据库中获取
	sfKey := fmt.Sprintf(SingleFlightBrandDetailKey, id)
	nullKey := fmt.Sprintf(BrandNullCacheKey, id)
	sfResult, err, _ := r.sf.Do(sfKey, func() (interface{}, error) { //防击穿
		//先从缓存中获取
		result, err := r.cacheRepo.Get(ctx, fmt.Sprintf(BrandDetailCacheKey, id))
		if err == nil && result != "" {
			var brand model.Brand
			err = json.Unmarshal([]byte(result), &brand)
			return &brand, nil
		}
		//检查空值缓存
		nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
		if nullResult == "1" {
			log.Printf("[BrandRepository] GetBrandByID null-cache hit inside singleflight, id=%s", id)
			return nil, nil
		}
		//没有缓存，则从数据库中获取
		var brand model.Brand
		err = r.db.WithContext(ctx).Where("id = ?", id).First(&brand).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[BrandRepository] GetBrandByID record not found in DB, id=%s", id)
				r.cacheRepo.Set(context.Background(), nullKey, "1", 5*time.Minute) //如果数据库中没有，则设置空值缓存
				return nil, nil
			}
			return nil, err
		}
		//设置缓存
		data, _ := json.Marshal(brand)
		if cacheErr := r.cacheRepo.Set(context.Background(), fmt.Sprintf(BrandDetailCacheKey, id), string(data), time.Minute*10); cacheErr != nil {
			log.Printf("[BrandRepository] GetBrandByID set cache error: %v", cacheErr)
		}

		return &brand, nil
	})
	if err != nil {
		return nil, err
	}
	if sfResult == nil {
		return nil, nil
	}
	brands, ok := sfResult.(*model.Brand)
	if !ok {
		return nil, fmt.Errorf("type assertion to *model.Brand failed or brand is nil")
	}
	return brands, nil
}
func (r *brandRepository) UpdateBrand(ctx context.Context, brand *model.Brand) error {
	//读取当前记录版本号
	var current model.Brand
	err := r.db.WithContext(ctx).Where("id = ?", brand.ID).First(&current).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("brand not found: %s", brand.ID)
		}
		return err
	}
	brand.Version = current.Version + 1
	//更新时检查版本号
	result := r.db.WithContext(ctx).Model(&model.Brand{}).Where("id = ? and version = ?", brand.ID, current.Version).Updates(brand)
	if result.Error != nil {
		return result.Error
	}
	//检查是否更新成功
	if result.RowsAffected == 0 {
		return fmt.Errorf("version mismatch, current version: %d, new version: %d", current.Version, brand.Version)
	}
	newCtx := context.Background()
	// 删除详情缓存和空值缓存
	r.cacheRepo.Delete(newCtx, fmt.Sprintf(BrandDetailCacheKey, brand.ID))
	r.cacheRepo.Delete(newCtx, fmt.Sprintf(BrandNullCacheKey, brand.ID))
	// 删除列表缓存（品牌信息可能影响排序或筛选）
	r.cacheRepo.Delete(newCtx, BrandListCacheKey)
	// 删除按首字母分组缓存（如果first_letter或status变了，所有分组都需要更新）
	r.cacheRepo.Delete(newCtx, BrandGroupByLetterCacheKey)

	return nil
}
func (r *brandRepository) DeleteBrand(ctx context.Context, id string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询品牌信息，用于删除相关缓存
		var brand model.Brand
		err := tx.WithContext(ctx).Where("id = ?", id).First(&brand).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("brand not found: %s", id)
			}
			return err
		}

		// 执行软删除（注意：Where要在Delete之前）
		err = tx.WithContext(ctx).Where("id = ?", id).Delete(&model.Brand{}).Error
		if err != nil {
			return err
		}

		newCtx := context.Background()
		// 删除详情缓存和空值缓存
		r.cacheRepo.Delete(newCtx, fmt.Sprintf(BrandDetailCacheKey, id))
		r.cacheRepo.Delete(newCtx, fmt.Sprintf(BrandNullCacheKey, id))
		// 删除列表缓存（因为删除了品牌）
		r.cacheRepo.Delete(newCtx, BrandListCacheKey)
		// 删除按首字母分组缓存（因为删除了品牌，所有分组都需要更新）
		r.cacheRepo.Delete(newCtx, BrandGroupByLetterCacheKey)

		return nil
	})
	return err
}
func (r *brandRepository) ListBrands(ctx context.Context, filter *BrandListFliter) ([]*model.Brand, error) {

	// 使用singleflight防击穿（key固定，因为缓存的是全量数据）
	sfResult, err, _ := r.sf.Do(SingleFlightBrandListKey, func() (interface{}, error) {
		// 先从缓存中获取全量列表（类似category的做法）
		result, err := r.cacheRepo.Get(ctx, BrandListCacheKey)
		if err == nil && result != "" {
			var brands []*model.Brand
			err = json.Unmarshal([]byte(result), &brands)
			if err != nil {
				return nil, err
			}
			// 在内存中过滤
			filteredBrands := filterBrands(brands, filter)
			return filteredBrands, nil
		}
		// 缓存中没有，则从数据库中获取全量数据
		var brands []*model.Brand
		err = r.db.WithContext(ctx).Order("sort_order DESC").Find(&brands).Error
		if err != nil {
			return nil, err
		}

		// 将全量数据缓存到redis（即使为空数组也缓存）

		data, _ := json.Marshal(brands)
		if cacheErr := r.cacheRepo.Set(context.Background(), BrandListCacheKey, string(data), time.Hour*12); cacheErr != nil {
			log.Printf("[BrandRepository] ListBrands set cache error: %v", cacheErr)
		}
		return brands, nil
	})
	if err != nil {
		return nil, err
	}
	if sfResult == nil {
		return []*model.Brand{}, nil
	}
	brands, ok := sfResult.([]*model.Brand)
	if !ok {
		return nil, fmt.Errorf("type assertion to []*model.Brand failed")
	}

	// 在内存中过滤
	filteredBrands := filterBrands(brands, filter)
	return filteredBrands, nil
}

func (r *brandRepository) GetBrandsByFirstLetter(ctx context.Context, status int32) ([]*BrandGroupByLetter, error) {
	// 先从缓存中获取所有首字母分组
	result, err := r.cacheRepo.Get(ctx, BrandGroupByLetterCacheKey)
	if err == nil && result != "" {
		var groups []*BrandGroupByLetter
		err = json.Unmarshal([]byte(result), &groups)
		if err != nil {
			return nil, err
		}
		// 如果传入了status参数，需要过滤
		if status > 0 {
			groups = filterBrandGroupsByStatus(groups, status)
		}
		return groups, nil
	}

	// 使用singleflight防击穿（固定key，因为缓存所有分组）
	sfResult, err, _ := r.sf.Do(SingleFlightBrandGroupByLetterKey, func() (interface{}, error) {
		// 缓存中没有，则从数据库中获取所有品牌
		var brands []*model.Brand
		query := r.db.WithContext(ctx)
		if status > 0 {
			query = query.Where("status = ?", status)
		}
		err = query.Order("first_letter ASC, sort_order DESC").Find(&brands).Error
		if err != nil {
			return nil, err
		}

		// 按首字母分组
		groups := groupBrandsByLetter(brands)

		// 将分组结果缓存到redis（即使为空数组也缓存）

		data, _ := json.Marshal(groups)
		if cacheErr := r.cacheRepo.Set(context.Background(), BrandGroupByLetterCacheKey, string(data), time.Hour*1); cacheErr != nil {
			log.Printf("[BrandRepository] GetBrandsByFirstLetter set cache error: %v", cacheErr)
		}
		return groups, nil
	})
	if err != nil {
		return nil, err
	}
	if sfResult == nil {
		return []*BrandGroupByLetter{}, nil
	}
	groups, ok := sfResult.([]*BrandGroupByLetter)
	if !ok {
		return nil, fmt.Errorf("type assertion to []*BrandGroupByLetter failed")
	}
	return groups, nil
}

func filterBrands(brands []*model.Brand, filter *BrandListFliter) []*model.Brand {
	if filter == nil {
		return brands
	}

	// 先过滤条件（注意：brands已经过滤了软删除，所以这里不需要再检查）
	var filteredBrands []*model.Brand
	for _, brand := range brands {
		// 状态过滤
		if filter.Status != 0 && int(brand.Status) != filter.Status {
			continue
		}
		// 关键词过滤（大小写不敏感）
		if filter.Keyword != "" && !strings.Contains(strings.ToLower(brand.Name), strings.ToLower(filter.Keyword)) {
			continue
		}
		// 首字母过滤
		if filter.FirstLetter != "" && brand.FirstLetter != filter.FirstLetter {
			continue
		}
		// 国家过滤
		if filter.Country != "" && brand.Country != filter.Country {
			continue
		}
		filteredBrands = append(filteredBrands, brand)
	}

	// 再分页（先过滤再分页，顺序不能错）
	if filter.Offset > 0 && filter.Offset < len(filteredBrands) {
		filteredBrands = filteredBrands[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(filteredBrands) {
		filteredBrands = filteredBrands[:filter.Limit]
	}
	return filteredBrands
}

// groupBrandsByLetter 按首字母分组品牌
func groupBrandsByLetter(brands []*model.Brand) []*BrandGroupByLetter {
	// 使用map按首字母分组
	groupMap := make(map[string][]*model.Brand)
	for _, brand := range brands {
		firstLetter := brand.FirstLetter
		if firstLetter == "" {
			firstLetter = "#" // 空首字母归为"#"
		}
		groupMap[firstLetter] = append(groupMap[firstLetter], brand)
	}

	// 转换为切片并排序
	groups := make([]*BrandGroupByLetter, 0, len(groupMap))
	for firstLetter, brands := range groupMap {
		groups = append(groups, &BrandGroupByLetter{
			FirstLetter: firstLetter,
			Brands:      brands,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].FirstLetter < groups[j].FirstLetter
	})

	return groups
}

// filterBrandGroupsByStatus 按状态过滤品牌分组
func filterBrandGroupsByStatus(groups []*BrandGroupByLetter, status int32) []*BrandGroupByLetter {
	filteredGroups := make([]*BrandGroupByLetter, 0)
	for _, group := range groups {
		var filteredBrands []*model.Brand
		for _, brand := range group.Brands {
			if int32(brand.Status) == status {
				filteredBrands = append(filteredBrands, brand)
			}
		}
		// 只保留有品牌的组
		if len(filteredBrands) > 0 {
			filteredGroups = append(filteredGroups, &BrandGroupByLetter{
				FirstLetter: group.FirstLetter,
				Brands:      filteredBrands,
			})
		}
	}
	return filteredGroups
}

// ============================================
// 品牌类目关联方法
// ============================================

// AddBrandCategory 添加品牌类目关联
// 注意：不需要加锁，因为：
// 1. 数据库唯一索引 uk_brand_category 已经提供了并发保护
// 2. 即使两个请求同时插入相同的关联，唯一索引会保证只有一个成功
// 3. 错误处理已经捕获了唯一约束冲突，返回友好的错误信息
func (r *brandRepository) AddBrandCategory(ctx context.Context, brandID, categoryID string) error {
	// 使用事务 + FOR UPDATE 锁定记录，防止在检查后被删除
	// 这样可以避免"检查时存在，插入时已被删除"的并发问题
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查品牌是否存在，并锁定记录（FOR UPDATE）
		var brand model.Brand
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", brandID).
			First(&brand).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("品牌不存在: %s", brandID)
			}
			return err
		}

		// 检查类目是否存在，并锁定记录（FOR UPDATE）
		var category model.Category
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", categoryID).
			First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("类目不存在: %s", categoryID)
			}
			return err
		}

		// 创建关联（唯一索引会自动防止重复）
		brandCategory := &model.BrandCategory{
			BrandID:    brandID,
			CategoryID: categoryID,
		}
		if err := tx.Create(brandCategory).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("该品牌已关联该类目")
			}
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("[BrandRepository] AddBrandCategory success, brand_id=%s, category_id=%s", brandID, categoryID)
	return nil
}

// RemoveBrandCategory 删除品牌类目关联
func (r *brandRepository) RemoveBrandCategory(ctx context.Context, brandID, categoryID string) error {
	result := r.db.WithContext(ctx).Where("brand_id = ? AND category_id = ?", brandID, categoryID).
		Delete(&model.BrandCategory{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("品牌类目关联不存在")
	}

	log.Printf("[BrandRepository] RemoveBrandCategory success, brand_id=%s, category_id=%s", brandID, categoryID)
	return nil
}

// GetBrandCategories 查询品牌的类目列表
func (r *brandRepository) GetBrandCategories(ctx context.Context, brandID string) ([]*model.Category, error) {
	var categories []*model.Category

	// 通过关联表查询类目
	err := r.db.WithContext(ctx).
		Table("categories").
		Joins("INNER JOIN brand_categories ON categories.id = brand_categories.category_id").
		Where("brand_categories.brand_id = ? AND brand_categories.deleted_at IS NULL", brandID).
		Where("categories.deleted_at IS NULL").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	log.Printf("[BrandRepository] GetBrandCategories success, brand_id=%s, count=%d", brandID, len(categories))
	return categories, nil
}

// BatchSetBrandCategories 批量设置品牌类目关联（先删除旧的，再插入新的）
func (r *brandRepository) BatchSetBrandCategories(ctx context.Context, brandID string, categoryIDs []string) error {
	// 检查品牌是否存在
	var brand model.Brand
	if err := r.db.WithContext(ctx).Where("id = ?", brandID).First(&brand).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("品牌不存在: %s", brandID)
		}
		return err
	}

	// 检查所有类目是否存在
	if len(categoryIDs) > 0 {
		var count int64
		err := r.db.WithContext(ctx).Model(&model.Category{}).
			Where("id IN ?", categoryIDs).
			Count(&count).Error
		if err != nil {
			return err
		}
		if int(count) != len(categoryIDs) {
			return fmt.Errorf("部分类目不存在")
		}
	}

	// 使用事务批量替换
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除该品牌的所有旧关联
		if err := tx.Where("brand_id = ?", brandID).Delete(&model.BrandCategory{}).Error; err != nil {
			return err
		}

		// 2. 如果有关联要添加，批量插入新关联
		if len(categoryIDs) > 0 {
			brandCategories := make([]*model.BrandCategory, 0, len(categoryIDs))
			for _, categoryID := range categoryIDs {
				brandCategories = append(brandCategories, &model.BrandCategory{
					BrandID:    brandID,
					CategoryID: categoryID,
				})
			}
			if err := tx.CreateInBatches(brandCategories, 50).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("[BrandRepository] BatchSetBrandCategories success, brand_id=%s, category_count=%d", brandID, len(categoryIDs))
	return nil
}
