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

	// 商品标签关联方法
	AddProductTag(ctx context.Context, productID, tagID string) error
	RemoveProductTag(ctx context.Context, productID, tagID string) error
	GetProductTags(ctx context.Context, productID string) ([]*model.Tag, error)
	BatchSetProductTags(ctx context.Context, productID string, tagIDs []string) error
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

		//再检查brandId是否存在
		var brand model.Brand
		err = tx.Model(&model.Brand{}).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", product.BrandID).First(&brand).Error
		if err != nil {
			return err
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
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//先删除商品和tag关联的数据
		err := tx.Where("product_id = ?", id).Delete(&model.ProductTag{}).Error
		if err != nil {
			return err
		}
		//再删除商品数据
		err = tx.Where("id = ?", id).Delete(&model.Product{}).Error
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	go func() {
		ctx2 := context.Background()
		r.cacheRepo.Delete(ctx2, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx2, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
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
		safeKeyword := strings.ReplaceAll(filter.Keyword, `\`, `\\`)
		safeKeyword = strings.ReplaceAll(filter.Keyword, "%", "\\%")
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

	// 在更新时同时检查状态，保证原子性
	result := r.db.Model(&model.Product{}).
		Where("id = ? AND status = ?", id, ProductStatusAuditPass).
		Update("status", ProductStatusOnShelf)
	if result.Error != nil {
		return result.Error
	}

	// 如果更新失败，检查是商品不存在还是状态不正确
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.Model(&model.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("商品不存在")
		}
		return errors.New("商品状态不正确")
	}

	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) OffShelfProduct(ctx context.Context, id string) error {
	// 在更新时同时检查状态，保证原子性
	result := r.db.Model(&model.Product{}).
		Where("id = ? AND status = ?", id, ProductStatusOnShelf).
		Update("status", ProductStatusOffShelf)
	if result.Error != nil {
		return result.Error
	}

	// 如果更新失败，检查是商品不存在还是状态不正确
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.Model(&model.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("商品不存在")
		}
		return errors.New("商品状态不正确")
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) SubmitProductAudit(ctx context.Context, id string) error {
	// 在更新时同时检查状态，保证原子性
	result := r.db.Model(&model.Product{}).
		Where("id = ? AND status = ?", id, ProductStatusDraft).
		Update("status", ProductStatusToAudit)
	if result.Error != nil {
		return result.Error
	}

	// 如果更新失败，检查是商品不存在还是状态不正确
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.Model(&model.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("商品不存在")
		}
		return errors.New("商品状态不正确")
	}

	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

func (r *productRepository) AuditProduct(ctx context.Context, id string, result int32) error {
	// 在更新时同时检查状态，保证原子性
	// result: 1-通过（上架），2-驳回（草稿）
	var newStatus int8
	if result == 1 {
		newStatus = ProductStatusAuditPass
	} else {
		newStatus = ProductStatusDraft
	}

	updateResult := r.db.Model(&model.Product{}).
		Where("id = ? AND status = ?", id, ProductStatusToAudit).
		Update("status", newStatus)
	if updateResult.Error != nil {
		return updateResult.Error
	}

	// 如果更新失败，检查是商品不存在还是状态不正确
	if updateResult.RowsAffected == 0 {
		var count int64
		if err := r.db.Model(&model.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("商品不存在")
		}
		return errors.New("商品状态不正确")
	}
	go func() {
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductDetailCachedKey, id)) //删除缓存商品信息
		r.cacheRepo.Delete(ctx, fmt.Sprintf(ProductNullCachedKey, id))   //删除空值缓存
	}()
	return nil
}

// AddProductTag 添加商品标签关联
func (r *productRepository) AddProductTag(ctx context.Context, productID, tagID string) error {
	// 使用事务 + Clauses(clause.Locking{Strength: "SHARE"}) 锁定记录，防止在检查后被删除
	// 这样可以避免"检查时存在，插入时已被删除"的并发问题
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查商品是否存在，并锁定记录（FOR SHARE）
		var product model.Product
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", productID).
			First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("商品不存在: %s", productID)
			}
			return err
		}

		// 检查标签是否存在且启用，并锁定记录（FOR SHARE）
		var tag model.Tag
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ? AND status = ?", tagID, 1).
			First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("标签不存在或已停用: %s", tagID)
			}
			return err
		}

		// 创建关联（唯一索引会自动防止重复）
		productTag := &model.ProductTag{
			ProductID: productID,
			TagID:     tagID,
		}
		if err := tx.Create(productTag).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") ||
				strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("该商品已关联该标签")
			}
			return err
		}

		return nil
	})

	return err
}

// RemoveProductTag 删除商品标签关联
func (r *productRepository) RemoveProductTag(ctx context.Context, productID, tagID string) error {
	result := r.db.WithContext(ctx).
		Where("product_id = ? AND tag_id = ?", productID, tagID).
		Delete(&model.ProductTag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("商品标签关联不存在")
	}
	return nil
}

// GetProductTags 查询商品的标签列表
func (r *productRepository) GetProductTags(ctx context.Context, productID string) ([]*model.Tag, error) {
	var tags []*model.Tag

	// 通过关联表查询标签
	err := r.db.WithContext(ctx).
		Table("tags").
		Joins("INNER JOIN product_tags ON tags.id = product_tags.tag_id").
		Where("product_tags.product_id = ? AND product_tags.deleted_at IS NULL", productID).
		Where("tags.deleted_at IS NULL").
		Order("tags.sort_order DESC, tags.created_at ASC").
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

// BatchSetProductTags 批量设置商品标签关联（先删除旧的，再插入新的）
func (r *productRepository) BatchSetProductTags(ctx context.Context, productID string, tagIDs []string) error {
	// 使用事务批量替换，并在事务中检查商品和标签是否存在
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查商品是否存在，并锁定记录（FOR SHARE）
		var product model.Product
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ?", productID).
			First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("商品不存在: %s", productID)
			}
			return err
		}

		// 检查所有标签是否存在且启用
		if len(tagIDs) > 0 {
			// 去重
			tagIDMap := make(map[string]struct{}, len(tagIDs))
			uniqueTagIDs := make([]string, 0, len(tagIDs))
			for _, id := range tagIDs {
				if _, exists := tagIDMap[id]; !exists {
					tagIDMap[id] = struct{}{}
					uniqueTagIDs = append(uniqueTagIDs, id)
				}
			}
			tagIDs = uniqueTagIDs

			var count int64
			err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Model(&model.Tag{}).
				Where("id IN ? AND status = ?", tagIDs, 1).
				Count(&count).Error
			if err != nil {
				return err
			}
			if int(count) != len(tagIDs) {
				return fmt.Errorf("部分标签不存在或已停用")
			}
		}

		// 1. 删除该商品的所有旧关联
		if err := tx.Where("product_id = ?", productID).Delete(&model.ProductTag{}).Error; err != nil {
			return err
		}

		// 2. 如果有关联要添加，批量插入新关联
		if len(tagIDs) > 0 {
			productTags := make([]*model.ProductTag, 0, len(tagIDs))
			for _, tagID := range tagIDs {
				productTags = append(productTags, &model.ProductTag{
					ProductID: productID,
					TagID:     tagID,
				})
			}
			if err := tx.CreateInBatches(productTags, 50).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
