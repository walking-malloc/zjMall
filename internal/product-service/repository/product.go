package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"zjMall/internal/common/cache"
	"zjMall/internal/product-service/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ProductDetailCachedKey = "product:detail:%s"
	ProductNullCachedKey   = "product:null:%s" //空值缓存，防止缓存击穿
	ProductLockKey         = "product:lock:%s" //分布式锁

	ProductStatusDraft     = 1
	ProductStatusToAudit   = 2
	ProductStatusAuditPass = 3
	ProductStatusOnShelf   = 4
	ProductStatusOffShelf  = 5
	ProductStatusDeleted   = 6
)

type ProductListFliter struct {
	Page       int32
	PageSize   int32
	CategoryId string
	BrandId    string
	Status     int32
	Keyword    string
	StartTime  *time.Time
	EndTime    *time.Time
	SortBy     string
	SortOrder  string
}
type ProductListResult struct {
	Products []*model.Product
	Total    int64
}

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	GetProduct(ctx context.Context, id string) (*model.Product, error)
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, filter *ProductListFliter) (*ProductListResult, error)
	OnShelfProduct(ctx context.Context, id string) error
	OffShelfProduct(ctx context.Context, id string) error
	SubmitProductAudit(ctx context.Context, id string) error
	AuditProduct(ctx context.Context, id string, result int32) error
}

type productRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository
}

func NewProductRepository(db *gorm.DB, cache cache.CacheRepository) ProductRepository {
	return &productRepository{db: db, cacheRepo: cache}
}
func (r *productRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		//先检查对应的brandId和categoryId是否存在
		//先检查categoryId是否存在
		var category model.Category
		err := tx.Model(&model.Category{}).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", product.CategoryID).First(&category).Error
		if err != nil {
			return err
		}
		if category.ID == "" {
			return errors.New("category not found")
		}
		//再检查brandId是否存在
		var brand model.Brand
		err = tx.Model(&model.Brand{}).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", product.BrandID).First(&brand).Error
		if err != nil {
			return err
		}
		if brand.ID == "" {
			return errors.New("brand not found")
		}

		//创建商品
		err = tx.Model(&model.Product{}).Create(product).Error
		if err != nil {
			return err
		}

		return nil

	})

	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, product.ID)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, product.ID))   //删除空值缓存
	}()
	return err
}

func (r *productRepository) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	//先从缓存中获取
	result, err := r.cacheRepo.Get(ctx, fmt.Sprintf(ProductDetailCachedKey, id))
	if err == nil && result != "" {
		var product model.Product
		err = json.Unmarshal([]byte(result), &product)
		if err != nil {
			return nil, err
		}
		return &product, nil
	}

	//如果没有缓存，构建空值缓存
	nullKey := fmt.Sprintf(ProductNullCachedKey, id)
	r.cacheRepo.Set(ctx, nullKey, "1", 5*time.Minute+time.Duration(rand.Intn(60))*time.Second) //设置5分钟+随机60秒防止缓存雪崩
	//然后从数据库中获取
	var product model.Product
	err = r.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	if product.ID == "" {
		return nil, errors.New("product not found")
	}

	go func() {
		data, err := json.Marshal(product)
		if err != nil {
			return
		}
		ctx2 := context.Background()
		r.cacheRepo.Set(ctx2, fmt.Sprintf(ProductDetailCachedKey, id), string(data), 5*time.Minute+time.Duration(rand.Intn(60))*time.Second)
	}()
	return &product, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *model.Product) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		//先检查对应的brandId和categoryId是否存在
		//先检查categoryId是否存在
		var category model.Category
		err := tx.Model(&model.Category{}).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", product.CategoryID).First(&category).Error
		if err != nil {
			return err
		}
		if category.ID == "" {
			return errors.New("category not found")
		}
		//再检查brandId是否存在
		var brand model.Brand
		err = tx.Model(&model.Brand{}).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", product.BrandID).First(&brand).Error
		if err != nil {
			return err
		}
		if brand.ID == "" {
			return errors.New("brand not found")
		}
		//更新商品
		err = tx.Model(&model.Product{}).Where("id = ?", product.ID).Updates(product).Error
		if err != nil {
			return err
		}
		return nil
	})
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, product.ID)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, product.ID))   //删除空值缓存
	}()
	return err
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	err := r.db.Where("id = ?", id).Delete(&model.Product{}).Error
	if err != nil {
		return err
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) ListProducts(ctx context.Context, filter *ProductListFliter) (*ProductListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 { // 防DoS
		filter.PageSize = 100
	}
	// 2. 构建查询（动态条件）
	query := r.db.WithContext(ctx).Model(&model.Product{})

	// 只添加非空的过滤条件
	if filter.CategoryId != "" {
		query = query.Where("category_id = ?", filter.CategoryId)
	}
	if filter.BrandId != "" {
		query = query.Where("brand_id = ?", filter.BrandId)
	}
	if filter.Status > 0 { // 0表示不过滤状态
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		// 安全处理LIKE查询
		safeKeyword := strings.ReplaceAll(filter.Keyword, "%", "\\%")
		safeKeyword = strings.ReplaceAll(safeKeyword, "_", "\\_")
		query = query.Where("(title LIKE ? OR subtitle LIKE ?)",
			"%"+safeKeyword+"%", "%"+safeKeyword+"%")
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	// 3. 获取总数（分页必须）
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 4. 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	var products []*model.Product

	// 使用合适的排序字段
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		// 验证排序字段，防止SQL注入
		validSortFields := map[string]bool{
			"created_at": true,
			"sort_order": true,
		}
		if validSortFields[filter.SortBy] {
			sortOrder := "DESC"
			if filter.SortOrder == "asc" {
				sortOrder = "ASC"
			}
			orderBy = fmt.Sprintf("%s %s", filter.SortBy, sortOrder)
		}
	}

	err := query.
		Order(orderBy).
		Offset(int(offset)).
		Limit(int(filter.PageSize)).
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	return &ProductListResult{
		Products: products,
		Total:    total,
	}, nil
}

func (r *productRepository) OnShelfProduct(ctx context.Context, id string) error {
	//先检查当前商品状态是否为审核通过
	var product model.Product
	err := r.db.Model(&model.Product{}).Where("id = ?", id).First(&product).Error
	if err != nil {
		return err
	}
	if product.ID == "" {
		return errors.New("商品不存在")
	}
	if product.Status != ProductStatusAuditPass {
		return errors.New("商品状态不正确")
	}
	product.Status = ProductStatusOnShelf
	err = r.db.Model(&model.Product{}).Where("id = ?", id).Updates(&product).Error
	if err != nil {
		return err
	}

	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) OffShelfProduct(ctx context.Context, id string) error {
	//先检查当前商品状态是否为上架
	var product model.Product
	err := r.db.Model(&model.Product{}).Where("id = ?", id).First(&product).Error
	if err != nil {
		return err
	}
	if product.ID == "" {
		return errors.New("商品不存在")
	}
	if product.Status != ProductStatusOnShelf {
		return errors.New("商品状态不正确")
	}

	product.Status = ProductStatusOffShelf
	err = r.db.Model(&model.Product{}).Where("id = ?", id).Updates(&product).Error
	if err != nil {
		return err
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) SubmitProductAudit(ctx context.Context, id string) error {
	//先检查当前商品状态是否为草稿
	var product model.Product
	err := r.db.Model(&model.Product{}).Where("id = ?", id).First(&product).Error
	if err != nil {
		return err
	}
	if product.ID == "" {
		return errors.New("商品不存在")
	}
	if product.Status != ProductStatusDraft {
		return errors.New("商品状态不正确")
	}
	product.Status = ProductStatusToAudit
	err = r.db.Model(&model.Product{}).Where("id = ?", id).Updates(&product).Error
	if err != nil {
		return err
	}

	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) AuditProduct(ctx context.Context, id string, result int32) error {
	//先检查当前商品状态是否为待审核
	var product model.Product
	err := r.db.Model(&model.Product{}).Where("id = ?", id).First(&product).Error
	if err != nil {
		return err
	}
	if product.ID == "" {
		return errors.New("商品不存在")
	}
	if product.Status != ProductStatusToAudit {
		return errors.New("商品状态不正确")
	}

	if result == 1 {
		product.Status = ProductStatusAuditPass
	}
	err = r.db.Model(&model.Product{}).Where("id = ?", id).Updates(product).Error
	if err != nil {
		return err
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}
