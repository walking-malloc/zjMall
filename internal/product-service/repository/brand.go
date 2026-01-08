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
		return err
	}
	go func() {
		// 删除详情缓存和空值缓存
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(BrandDetailCacheKey, brand.ID))
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(BrandNullCacheKey, brand.ID))
		// 删除列表缓存（因为新增了品牌）
		r.cacheRepo.Delete(context.Background(), BrandListCacheKey)
		// 删除按首字母分组缓存（因为新增了品牌，所有分组都需要更新）
		r.cacheRepo.Delete(context.Background(), BrandGroupByLetterCacheKey)
	}()
	return nil
}
func (r *brandRepository) GetBrandByID(ctx context.Context, id string) (*model.Brand, error) {
	//先从缓存中获取
	result, err := r.cacheRepo.Get(ctx, fmt.Sprintf(BrandDetailCacheKey, id))
	if err == nil && result != "" {
		var brand model.Brand
		err = json.Unmarshal([]byte(result), &brand)
		return &brand, nil
	}

	//检查空值缓存
	nullKey := fmt.Sprintf(BrandNullCacheKey, id)
	nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
	if nullResult == "1" {
		log.Printf("[BrandRepository] GetBrandByID null-cache hit, id=%s", id)
		return nil, nil
	}

	//缓存中没有，则从数据库中获取
	sfKey := fmt.Sprintf(SingleFlightBrandDetailKey, id)
	sfResult, err, _ := r.sf.Do(sfKey, func() (interface{}, error) { //防击穿
		//检查空值缓存
		nullResult, _ := r.cacheRepo.Get(ctx, nullKey)
		if nullResult == "1" {
			log.Printf("[BrandRepository] GetBrandByID null-cache hit inside singleflight, id=%s", id)
			return nil, nil
		}
		var brand model.Brand
		err := r.db.WithContext(ctx).Where("id = ?", id).First(&brand).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[BrandRepository] GetBrandByID record not found in DB, id=%s", id)
				go func() {
					r.cacheRepo.Set(context.Background(), nullKey, "1", 5*time.Minute)
				}()
				return nil, nil
			}
			return nil, err
		}
		go func() {
			data, _ := json.Marshal(brand)
			if cacheErr := r.cacheRepo.Set(context.Background(), fmt.Sprintf(BrandDetailCacheKey, id), string(data), time.Minute*10); cacheErr != nil {
				log.Printf("[BrandRepository] GetBrandByID set cache error: %v", cacheErr)
			}
		}()
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
	err := r.db.WithContext(ctx).Model(&model.Brand{}).Where("id = ?", brand.ID).Updates(brand).Error
	if err != nil {
		return err
	}
	go func() {
		// 删除详情缓存和空值缓存
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(BrandDetailCacheKey, brand.ID))
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(BrandNullCacheKey, brand.ID))
		// 删除列表缓存（品牌信息可能影响排序或筛选）
		r.cacheRepo.Delete(context.Background(), BrandListCacheKey)
		// 删除按首字母分组缓存（如果first_letter或status变了，所有分组都需要更新）
		r.cacheRepo.Delete(context.Background(), BrandGroupByLetterCacheKey)
	}()
	return nil
}
func (r *brandRepository) DeleteBrand(ctx context.Context, id string) error {
	// 先查询品牌信息，用于删除相关缓存
	var brand model.Brand
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("brand not found: %s", id)
		}
		return err
	}

	// 执行软删除（注意：Where要在Delete之前）
	err = r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Brand{}).Error
	if err != nil {
		return err
	}

	go func() {
		// 删除详情缓存和空值缓存
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(BrandDetailCacheKey, id))
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(BrandNullCacheKey, id))
		// 删除列表缓存（因为删除了品牌）
		r.cacheRepo.Delete(context.Background(), BrandListCacheKey)
		// 删除按首字母分组缓存（因为删除了品牌，所有分组都需要更新）
		r.cacheRepo.Delete(context.Background(), BrandGroupByLetterCacheKey)
	}()
	return nil
}
func (r *brandRepository) ListBrands(ctx context.Context, filter *BrandListFliter) ([]*model.Brand, error) {
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

	// 使用singleflight防击穿（key固定，因为缓存的是全量数据）
	sfResult, err, _ := r.sf.Do(SingleFlightBrandListKey, func() (interface{}, error) {
		// 缓存中没有，则从数据库中获取全量数据
		var brands []*model.Brand
		err = r.db.WithContext(ctx).Order("sort_order DESC").Find(&brands).Error
		if err != nil {
			return nil, err
		}

		// 将全量数据缓存到redis（即使为空数组也缓存）
		go func() {
			data, _ := json.Marshal(brands)
			if cacheErr := r.cacheRepo.Set(context.Background(), BrandListCacheKey, string(data), time.Hour*1); cacheErr != nil {
				log.Printf("[BrandRepository] ListBrands set cache error: %v", cacheErr)
			}
		}()
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
		go func() {
			data, _ := json.Marshal(groups)
			if cacheErr := r.cacheRepo.Set(context.Background(), BrandGroupByLetterCacheKey, string(data), time.Hour*1); cacheErr != nil {
				log.Printf("[BrandRepository] GetBrandsByFirstLetter set cache error: %v", cacheErr)
			}
		}()
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
