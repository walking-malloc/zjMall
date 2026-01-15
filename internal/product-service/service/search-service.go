package service

import (
	"context"
	"fmt"
	"time"
	"zjMall/internal/product-service/model"
	"zjMall/internal/product-service/repository"
)

type SearchService struct {
	searchRepo         repository.SearchRepository
	productRepo        repository.ProductRepository
	categoryRepo       repository.CategoryRepository
	brandRepo          repository.BrandRepository
	tagRepo            repository.TagRepository
	attributeRepo      repository.AttributeRepository
	attributeValueRepo repository.AttributeValueRepository
	skuRepo            repository.SkuRepository
}

func NewSearchService(
	searchRepo repository.SearchRepository,
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	brandRepo repository.BrandRepository,
	tagRepo repository.TagRepository,
	attributeRepo repository.AttributeRepository,
	attributeValueRepo repository.AttributeValueRepository,
	skuRepo repository.SkuRepository,
) *SearchService {
	return &SearchService{
		searchRepo:         searchRepo,
		productRepo:        productRepo,
		categoryRepo:       categoryRepo,
		brandRepo:          brandRepo,
		tagRepo:            tagRepo,
		attributeRepo:      attributeRepo,
		attributeValueRepo: attributeValueRepo,
		skuRepo:            skuRepo,
	}
}

// SearchProducts 搜索商品
func (s *SearchService) SearchProducts(ctx context.Context, keyword string, page, pageSize int32, filters *repository.SearchFilters) (*repository.SearchResult, error) {
	return s.searchRepo.SearchProducts(ctx, keyword, page, pageSize, filters)
}

// SyncProductToES 同步商品到 ES（商品创建/更新时调用）
func (s *SearchService) SyncProductToES(ctx context.Context, productID string) error {
	// 1. 查询商品信息
	product, err := s.productRepo.GetProduct(ctx, productID)
	if err != nil {
		return fmt.Errorf("查询商品失败: %w", err)
	}
	if product == nil {
		return fmt.Errorf("商品不存在")
	}

	// 2. 只索引已上架的商品（状态=4）
	if product.Status != int8(repository.ProductStatusOnShelf) {
		// 删除索引
		return s.searchRepo.DeleteProduct(ctx, productID)
	}

	// 3. 查询关联信息
	var categoryName, brandName string
	if product.CategoryID != "" {
		category, _ := s.categoryRepo.GetCategoryByID(ctx, product.CategoryID)
		if category != nil {
			categoryName = category.Name
		}
	}
	if product.BrandID != "" {
		brand, _ := s.brandRepo.GetBrandByID(ctx, product.BrandID)
		if brand != nil {
			brandName = brand.Name
		}
	}

	// 4. 查询标签
	tags := make([]string, 0)
	productTags, _ := s.productRepo.GetProductTags(ctx, productID)
	for _, tag := range productTags {
		if tag != nil {
			tags = append(tags, tag.Name)
		}
	}
	var skuIDs []string
	//查询SKU列表
	var skus []*model.SKUIndex
	res, _, err := s.skuRepo.ListSkus(ctx, &repository.SkuListFilter{
		ProductID: productID,
		Status:    repository.SkuStatusOnShelf,
	})
	if err != nil {
		return fmt.Errorf("查询SKU列表失败: %w", err)
	}
	for _, sku := range res {
		skus = append(skus, &model.SKUIndex{
			SKUName: sku.Name,
			Price:   sku.Price,
		})
		skuIDs = append(skuIDs, sku.ID)
	}

	//查询属性值列表
	attributeValues, err := s.attributeValueRepo.GetAttributeValueBySkuID(ctx, skuIDs)

	if err != nil {
		return fmt.Errorf("查询属性值列表失败: %w", err)
	}

	attributeValueIndex := make([]string, 0)
	for _, attributeValue := range attributeValues {
		attributeValueIndex = append(attributeValueIndex, attributeValue.Value)
	}
	// 5. 构建索引文档
	// 确保日期格式为 RFC3339 (ISO 8601)
	createdAtStr := product.CreatedAt.Format(time.RFC3339)
	updatedAtStr := product.UpdatedAt.Format(time.RFC3339)

	productIndex := &model.ProductIndex{
		ID:              product.ID,
		Title:           product.Title,
		Subtitle:        product.Subtitle,
		Description:     product.Description,
		CategoryID:      product.CategoryID,
		CategoryName:    categoryName,
		BrandID:         product.BrandID,
		BrandName:       brandName,
		Tags:            tags,
		Status:          product.Status,
		SKUs:            skus,
		AttributeValues: attributeValueIndex,
		CreatedAt:       createdAtStr,
		UpdatedAt:       updatedAtStr,
	}

	// 处理 OnShelfTime（可能为 nil）
	if product.OnShelfTime != nil {
		onShelfTimeStr := product.OnShelfTime.Format(time.RFC3339)
		productIndex.OnShelfTime = &onShelfTimeStr
	}

	// 6. 索引到 ES
	return s.searchRepo.IndexProduct(ctx, productIndex)
}
