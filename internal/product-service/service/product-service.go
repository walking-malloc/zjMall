package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/internal/product-service/model"
	"zjMall/internal/product-service/repository"
	"zjMall/pkg"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// ProductService 商品服务（业务逻辑层）
type ProductService struct {
	categoryRepo       repository.CategoryRepository
	brandRepo          repository.BrandRepository
	productRepo        repository.ProductRepository
	tagRepo            repository.TagRepository
	skuRepo            repository.SkuRepository
	attributeRepo      repository.AttributeRepository
	attributeValueRepo repository.AttributeValueRepository
}

// NewProductService 创建商品服务实例
func NewProductService(
	categoryRepo repository.CategoryRepository,
	brandRepo repository.BrandRepository,
	productRepo repository.ProductRepository,
	tagRepo repository.TagRepository,
	skuRepo repository.SkuRepository,
	attributeRepo repository.AttributeRepository,
	attributeValueRepo repository.AttributeValueRepository,
) *ProductService {
	return &ProductService{
		categoryRepo:       categoryRepo,
		brandRepo:          brandRepo,
		productRepo:        productRepo,
		tagRepo:            tagRepo,
		skuRepo:            skuRepo,
		attributeRepo:      attributeRepo,
		attributeValueRepo: attributeValueRepo,
	}
}

// ============================================
// 类目管理接口
// ============================================

// CreateCategory 创建类目
func (s *ProductService) CreateCategory(ctx context.Context, req *productv1.CreateCategoryRequest) (*productv1.CreateCategoryResponse, error) {
	category := &model.Category{
		Name:      req.Name,
		ParentID:  req.ParentId,
		Level:     int8(req.Level),
		IsLeaf:    req.IsLeaf,
		IsVisible: req.IsVisible,
		SortOrder: req.SortOrder,
		Icon:      req.Icon,
		Status:    int8(req.Status),
	}
	err := s.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
		return &productv1.CreateCategoryResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &productv1.CreateCategoryResponse{
		Code:    0,
		Message: "创建成功",
		Data:    category.ID,
	}, nil
}

// GetCategory 查询类目详情
func (s *ProductService) GetCategory(ctx context.Context, req *productv1.GetCategoryRequest) (*productv1.GetCategoryResponse, error) {
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryId)
	if err != nil {
		return &productv1.GetCategoryResponse{
			Code:    1,
			Message: fmt.Sprintf("查询类目详情失败: %v", err),
		}, nil
	}

	if category == nil {
		return &productv1.GetCategoryResponse{
			Code:    1,
			Message: "类目不存在",
		}, nil
	}
	return &productv1.GetCategoryResponse{
		Code:    0,
		Message: "查询成功",
		Data: &productv1.CategoryInfo{
			Id:        category.ID,
			Name:      category.Name,
			ParentId:  category.ParentID,
			Level:     int32(category.Level),
			IsLeaf:    category.IsLeaf,
			IsVisible: category.IsVisible,
			SortOrder: category.SortOrder,
			Icon:      category.Icon,
			Status:    int32(category.Status),
			CreatedAt: timestamppb.New(category.CreatedAt),
			UpdatedAt: timestamppb.New(category.UpdatedAt),
		},
	}, nil
}

// UpdateCategory 更新类目
func (s *ProductService) UpdateCategory(ctx context.Context, req *productv1.UpdateCategoryRequest) (*productv1.UpdateCategoryResponse, error) {

	category := &model.Category{
		BaseModel: pkg.BaseModel{
			ID: req.CategoryId,
		},
		Name:      req.Name,
		IsLeaf:    req.IsLeaf,
		IsVisible: req.IsVisible,
		SortOrder: req.SortOrder,
		Icon:      req.Icon,
		Status:    int8(req.Status),
	}
	err := s.categoryRepo.UpdateCategory(ctx, category)
	if err != nil {
		return &productv1.UpdateCategoryResponse{
			Code:    1,
			Message: fmt.Sprintf("更新类目失败: %v", err),
		}, nil
	}
	return &productv1.UpdateCategoryResponse{
		Code:    0,
		Message: "更新成功",
	}, nil
}

// DeleteCategory 删除类目
func (s *ProductService) DeleteCategory(ctx context.Context, req *productv1.DeleteCategoryRequest) (*productv1.DeleteCategoryResponse, error) {
	if err := s.categoryRepo.DeleteCategory(ctx, req.CategoryId); err != nil {
		return &productv1.DeleteCategoryResponse{
			Code:    1,
			Message: fmt.Sprintf("删除类目失败: %v", err),
		}, nil
	}

	return &productv1.DeleteCategoryResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// ListCategories 查询类目列表
func (s *ProductService) ListCategories(ctx context.Context, req *productv1.ListCategoriesRequest) (*productv1.ListCategoriesResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := repository.CategoryListFliter{
		Level:     req.Level,
		Status:    req.Status,
		IsVisible: req.IsVisible,
		Keyword:   req.Keyword,
		Offset:    (page - 1) * pageSize,
		Limit:     pageSize,
	}
	categories, err := s.categoryRepo.ListCategories(ctx, &filter)
	if err != nil {
		return &productv1.ListCategoriesResponse{
			Code:    1,
			Message: fmt.Sprintf("查询类目列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	categoryList := make([]*productv1.CategoryInfo, 0, len(categories))
	for _, category := range categories {
		categoryList = append(categoryList, &productv1.CategoryInfo{
			Id:        category.ID,
			Name:      category.Name,
			ParentId:  category.ParentID,
			Level:     int32(category.Level),
			IsLeaf:    category.IsLeaf,
			IsVisible: category.IsVisible,
			SortOrder: category.SortOrder,
			Icon:      category.Icon,
			Status:    int32(category.Status),
		})
	}

	return &productv1.ListCategoriesResponse{
		Code:    0,
		Message: "查询成功",
		Data:    categoryList,
		Total:   int64(len(categories)),
	}, nil
}

// GetCategoryTree 查询类目树
func (s *ProductService) GetCategoryTree(ctx context.Context, req *productv1.GetCategoryTreeRequest) (*productv1.GetCategoryTreeResponse, error) {
	tree, err := s.categoryRepo.GetCategoryTree(ctx, req.MaxLevel, req.OnlyVisible, req.Status)
	if err != nil {
		return &productv1.GetCategoryTreeResponse{
			Code:    1,
			Message: fmt.Sprintf("查询类目树失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	treeNodes := convertTreeNodesToProto(tree)

	return &productv1.GetCategoryTreeResponse{
		Code:    0,
		Message: "查询成功",
		Data:    treeNodes,
	}, nil
}

// 辅助函数：转换树节点为 proto 格式
func convertTreeNodesToProto(nodes []*repository.CategoryTreeNode) []*productv1.CategoryTreeNode {
	result := make([]*productv1.CategoryTreeNode, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, &productv1.CategoryTreeNode{
			Category: &productv1.CategoryInfo{
				Id:        node.ID,
				Name:      node.Name,
				Level:     int32(node.Level),
				IsLeaf:    node.IsLeaf,
				IsVisible: node.IsVisible,
				SortOrder: node.SortOrder,
				Icon:      node.Icon,
				Status:    int32(node.Status),
				CreatedAt: timestamppb.New(node.CreatedAt),
				UpdatedAt: timestamppb.New(node.UpdatedAt),
			},
			Children: convertTreeNodesToProto(node.Children),
		})
	}
	return result
}

// GetCategoryChildren 查询子类目列表
func (s *ProductService) GetCategoryChildren(ctx context.Context, req *productv1.GetCategoryChildrenRequest) (*productv1.GetCategoryChildrenResponse, error) {
	// 和 ListCategories 类似，只是参数来源不同
	categories, err := s.categoryRepo.ListCategories(ctx, &repository.CategoryListFliter{
		ParentID:  req.ParentId,
		Status:    req.Status,
		IsVisible: req.OnlyVisible,
		Offset:    (req.Page - 1) * req.PageSize,
		Limit:     req.PageSize,
	})
	if err != nil {
		return &productv1.GetCategoryChildrenResponse{
			Code:    1,
			Message: fmt.Sprintf("查询子类目列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	categoryList := make([]*productv1.CategoryInfo, 0, len(categories))
	for _, category := range categories {
		categoryList = append(categoryList, &productv1.CategoryInfo{
			Id:        category.ID,
			Name:      category.Name,
			ParentId:  category.ParentID,
			Level:     int32(category.Level),
			IsLeaf:    category.IsLeaf,
			IsVisible: category.IsVisible,
			SortOrder: category.SortOrder,
			Icon:      category.Icon,
			Status:    int32(category.Status),
		})
	}

	return &productv1.GetCategoryChildrenResponse{
		Code:    0,
		Message: "查询成功",
		Data:    categoryList,
	}, nil
}

// ============================================
// 品牌管理接口
// ============================================

// CreateBrand 创建品牌
func (s *ProductService) CreateBrand(ctx context.Context, req *productv1.CreateBrandRequest) (*productv1.CreateBrandResponse, error) {
	err := s.brandRepo.CreateBrand(ctx, &model.Brand{
		Name:        req.Name,
		LogoURL:     req.LogoUrl,
		Country:     req.Country,
		Description: req.Description,
		FirstLetter: req.FirstLetter,
		SortOrder:   int(req.SortOrder),
		Status:      int(req.Status),
	})
	if err != nil {
		return &productv1.CreateBrandResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return &productv1.CreateBrandResponse{
		Code:    0,
		Message: "创建品牌成功",
	}, nil
}

// GetBrand 查询品牌详情
func (s *ProductService) GetBrand(ctx context.Context, req *productv1.GetBrandRequest) (*productv1.GetBrandResponse, error) {
	brand, err := s.brandRepo.GetBrandByID(ctx, req.BrandId)
	if err != nil {
		return &productv1.GetBrandResponse{
			Code:    1,
			Message: fmt.Sprintf("查询品牌详情失败: %v", err),
		}, nil
	}
	if brand == nil {
		return &productv1.GetBrandResponse{
			Code:    1,
			Message: "品牌不存在",
		}, nil
	}
	return &productv1.GetBrandResponse{
		Code:    0,
		Message: "查询品牌详情成功",
		Data: &productv1.BrandInfo{
			Id:          brand.ID,
			Name:        brand.Name,
			LogoUrl:     brand.LogoURL,
			Country:     brand.Country,
			Description: brand.Description,
			FirstLetter: brand.FirstLetter,
			SortOrder:   int32(brand.SortOrder),
			Status:      int32(brand.Status),
			CreatedAt:   timestamppb.New(brand.CreatedAt),
			UpdatedAt:   timestamppb.New(brand.UpdatedAt),
		},
	}, nil
}

// UpdateBrand 更新品牌
func (s *ProductService) UpdateBrand(ctx context.Context, req *productv1.UpdateBrandRequest) (*productv1.UpdateBrandResponse, error) {
	err := s.brandRepo.UpdateBrand(ctx, &model.Brand{
		BaseModel: pkg.BaseModel{
			ID: req.BrandId,
		},
		Name:        req.Name,
		LogoURL:     req.LogoUrl,
		Country:     req.Country,
		Description: req.Description,
		FirstLetter: req.FirstLetter,
		SortOrder:   int(req.SortOrder),
		Status:      int(req.Status),
	})
	if err != nil {
		// 检查是否是版本冲突错误
		if strings.Contains(err.Error(), "version mismatch") {
			return &productv1.UpdateBrandResponse{
				Code:    1,
				Message: "数据已被其他请求修改，请刷新页面后重试",
			}, nil
		}

		return &productv1.UpdateBrandResponse{
			Code:    1,
			Message: fmt.Sprintf("更新品牌失败: %v", err),
		}, nil
	}
	return &productv1.UpdateBrandResponse{
		Code:    0,
		Message: "更新品牌成功",
	}, nil
}

// DeleteBrand 删除品牌
func (s *ProductService) DeleteBrand(ctx context.Context, req *productv1.DeleteBrandRequest) (*productv1.DeleteBrandResponse, error) {
	err := s.brandRepo.DeleteBrand(ctx, req.BrandId)
	if err != nil {
		return &productv1.DeleteBrandResponse{
			Code:    1,
			Message: fmt.Sprintf("删除品牌失败: %v", err),
		}, nil
	}
	return &productv1.DeleteBrandResponse{
		Code:    0,
		Message: "删除品牌成功",
	}, nil
}

// ListBrands 查询品牌列表
func (s *ProductService) ListBrands(ctx context.Context, req *productv1.ListBrandsRequest) (*productv1.ListBrandsResponse, error) {
	brands, err := s.brandRepo.ListBrands(ctx, &repository.BrandListFliter{
		Limit:       int(req.PageSize),
		Offset:      int(req.Page-1) * int(req.PageSize),
		Status:      int(req.Status),
		Keyword:     req.Keyword,
		FirstLetter: req.FirstLetter,
		Country:     req.Country,
	})
	if err != nil {
		return &productv1.ListBrandsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询品牌列表失败: %v", err),
		}, nil
	}
	if len(brands) == 0 {
		return &productv1.ListBrandsResponse{
			Code:    0,
			Message: "查询品牌列表成功",
			Data:    nil,
		}, nil
	}
	brandList := make([]*productv1.BrandInfo, 0, len(brands))
	for _, brand := range brands {
		brandList = append(brandList, &productv1.BrandInfo{
			Id:          brand.ID,
			Name:        brand.Name,
			LogoUrl:     brand.LogoURL,
			Country:     brand.Country,
			Description: brand.Description,
			FirstLetter: brand.FirstLetter,
			SortOrder:   int32(brand.SortOrder),
			Status:      int32(brand.Status),
			CreatedAt:   timestamppb.New(brand.CreatedAt),
			UpdatedAt:   timestamppb.New(brand.UpdatedAt),
		})
	}
	return &productv1.ListBrandsResponse{
		Code:    0,
		Message: "查询品牌列表成功",
		Data:    brandList,
		Total:   int64(len(brands)),
	}, nil
}

// GetBrandsByFirstLetter 按首字母分组查询品牌列表
func (s *ProductService) GetBrandsByFirstLetter(ctx context.Context, req *productv1.GetBrandsByFirstLetterRequest) (*productv1.GetBrandsByFirstLetterResponse, error) {
	groups, err := s.brandRepo.GetBrandsByFirstLetter(ctx, req.Status)
	if err != nil {
		return &productv1.GetBrandsByFirstLetterResponse{
			Code:    1,
			Message: fmt.Sprintf("按首字母分组查询品牌列表失败: %v", err),
		}, nil
	}
	if len(groups) == 0 {
		return &productv1.GetBrandsByFirstLetterResponse{
			Code:    0,
			Message: "查询品牌列表成功",
			Data:    []*productv1.BrandGroupByLetter{},
		}, nil
	}

	// 转换为proto格式
	groupList := make([]*productv1.BrandGroupByLetter, 0, len(groups))
	for _, group := range groups {
		brandList := make([]*productv1.BrandInfo, 0, len(group.Brands))
		for _, brand := range group.Brands {
			brandList = append(brandList, &productv1.BrandInfo{
				Id:          brand.ID,
				Name:        brand.Name,
				LogoUrl:     brand.LogoURL,
				Country:     brand.Country,
				Description: brand.Description,
				FirstLetter: brand.FirstLetter,
				SortOrder:   int32(brand.SortOrder),
				Status:      int32(brand.Status),
				CreatedAt:   timestamppb.New(brand.CreatedAt),
				UpdatedAt:   timestamppb.New(brand.UpdatedAt),
			})
		}
		groupList = append(groupList, &productv1.BrandGroupByLetter{
			FirstLetter: group.FirstLetter,
			Brands:      brandList,
		})
	}

	return &productv1.GetBrandsByFirstLetterResponse{
		Code:    0,
		Message: "查询品牌列表成功",
		Data:    groupList,
		Total:   int64(len(groupList)),
	}, nil
}

// ============================================
// 品牌类目关联管理接口
// ============================================

// AddBrandCategory 添加品牌类目关联
func (s *ProductService) AddBrandCategory(ctx context.Context, req *productv1.AddBrandCategoryRequest) (*productv1.AddBrandCategoryResponse, error) {
	// Repository 层已做完善的校验（品牌存在、类目存在、唯一性检查）
	err := s.brandRepo.AddBrandCategory(ctx, req.BrandId, req.CategoryId)
	if err != nil {
		return &productv1.AddBrandCategoryResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.AddBrandCategoryResponse{
		Code:    0,
		Message: "添加成功",
	}, nil
}

// RemoveBrandCategory 删除品牌类目关联
func (s *ProductService) RemoveBrandCategory(ctx context.Context, req *productv1.RemoveBrandCategoryRequest) (*productv1.RemoveBrandCategoryResponse, error) {
	// Repository 层已做完善的校验（关联是否存在）
	err := s.brandRepo.RemoveBrandCategory(ctx, req.BrandId, req.CategoryId)
	if err != nil {
		return &productv1.RemoveBrandCategoryResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.RemoveBrandCategoryResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// GetBrandCategories 查询品牌的类目列表
func (s *ProductService) GetBrandCategories(ctx context.Context, req *productv1.GetBrandCategoriesRequest) (*productv1.GetBrandCategoriesResponse, error) {
	// Repository 层直接查询，无需额外校验
	categories, err := s.brandRepo.GetBrandCategories(ctx, req.BrandId)
	if err != nil {
		return &productv1.GetBrandCategoriesResponse{
			Code:    1,
			Message: fmt.Sprintf("查询失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	categoryList := make([]*productv1.CategoryInfo, 0, len(categories))
	for _, category := range categories {
		categoryList = append(categoryList, &productv1.CategoryInfo{
			Id:        category.ID,
			Name:      category.Name,
			ParentId:  category.ParentID,
			Level:     int32(category.Level),
			IsLeaf:    category.IsLeaf,
			IsVisible: category.IsVisible,
			SortOrder: category.SortOrder,
			Icon:      category.Icon,
			Status:    int32(category.Status),
			CreatedAt: timestamppb.New(category.CreatedAt),
			UpdatedAt: timestamppb.New(category.UpdatedAt),
		})
	}

	return &productv1.GetBrandCategoriesResponse{
		Code:       0,
		Message:    "查询成功",
		Categories: categoryList,
	}, nil
}

// BatchSetBrandCategories 批量设置品牌类目关联
func (s *ProductService) BatchSetBrandCategories(ctx context.Context, req *productv1.BatchSetBrandCategoriesRequest) (*productv1.BatchSetBrandCategoriesResponse, error) {
	// 去重 category_ids（业务逻辑，保留在 service 层）
	categoryIDMap := make(map[string]bool)
	uniqueCategoryIDs := make([]string, 0, len(req.CategoryIds))
	for _, categoryID := range req.CategoryIds {
		if categoryID != "" && !categoryIDMap[categoryID] {
			categoryIDMap[categoryID] = true
			uniqueCategoryIDs = append(uniqueCategoryIDs, categoryID)
		}
	}

	// Repository 层已做完善的校验（品牌存在、所有类目存在）
	err := s.brandRepo.BatchSetBrandCategories(ctx, req.BrandId, uniqueCategoryIDs)
	if err != nil {
		return &productv1.BatchSetBrandCategoriesResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.BatchSetBrandCategoriesResponse{
		Code:    0,
		Message: "批量设置成功",
	}, nil
}

// ============================================
// 商品（SPU）管理接口
// ============================================

// convertProductToProto 将 model.Product 转换为 productv1.ProductInfo
func convertProductToProto(product *model.Product) *productv1.ProductInfo {
	var images []string
	if product.Images != "" {
		_ = json.Unmarshal([]byte(product.Images), &images)
	}

	productInfo := &productv1.ProductInfo{
		Id:          product.ID,
		CategoryId:  product.CategoryID,
		BrandId:     product.BrandID,
		Title:       product.Title,
		Subtitle:    product.Subtitle,
		MainImage:   product.MainImage,
		Images:      images,
		Description: product.Description,
		Status:      int32(product.Status),
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}

	if product.OnShelfTime != nil {
		productInfo.OnShelfTime = timestamppb.New(*product.OnShelfTime)
	}
	if product.OffShelfTime != nil {
		productInfo.OffShelfTime = timestamppb.New(*product.OffShelfTime)
	}

	return productInfo
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	// 处理图片列表
	var imagesJSON string
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return &productv1.CreateProductResponse{
				Code:    1,
				Message: fmt.Sprintf("图片列表格式错误: %v", err),
			}, nil
		}
		imagesJSON = string(imagesBytes)
	}

	// 设置默认状态
	status := int8(1) // 默认草稿
	if req.Status > 0 {
		status = int8(req.Status)
	}

	// 处理上架时间
	var onShelfTime *time.Time
	if req.OnShelfTime != nil {
		t := req.OnShelfTime.AsTime()
		onShelfTime = &t
	}

	product := &model.Product{
		CategoryID:  req.CategoryId,
		BrandID:     req.BrandId,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		MainImage:   req.MainImage,
		Images:      imagesJSON,
		Description: req.Description,
		Status:      status,
		OnShelfTime: onShelfTime, //TODO:定期上线
	}

	err := s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		return &productv1.CreateProductResponse{
			Code:    1,
			Message: fmt.Sprintf("创建商品失败: %v", err),
		}, nil
	}

	return &productv1.CreateProductResponse{
		Code:    0,
		Message: "创建成功",
		Data:    product.ID,
	}, nil
}

// GetProduct 查询商品详情
func (s *ProductService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	product, err := s.productRepo.GetProduct(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &productv1.GetProductResponse{
				Code:    1,
				Message: "商品不存在",
			}, nil
		}
		return &productv1.GetProductResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品详情失败: %v", err),
		}, nil
	}

	if product == nil {
		return &productv1.GetProductResponse{
			Code:    1,
			Message: "商品不存在",
		}, nil
	}

	productInfo := convertProductToProto(product)

	response := &productv1.GetProductResponse{
		Code:    0,
		Message: "查询成功",
		Product: productInfo,
	}

	// TODO: 如果需要包含SKU列表和标签列表，需要调用相应的repository方法
	// if req.IncludeSkus {
	//     skus, err := s.skuRepo.ListSkusByProductID(ctx, req.ProductId)
	//     ...
	// }
	// if req.IncludeTags {
	//     tags, err := s.tagRepo.GetProductTags(ctx, req.ProductId)
	//     ...
	// }

	return response, nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	// 先查询现有商品
	existingProduct, err := s.productRepo.GetProduct(ctx, req.ProductId)
	if err != nil {
		return &productv1.UpdateProductResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品失败: %v", err),
		}, nil
	}

	if existingProduct == nil {
		return &productv1.UpdateProductResponse{
			Code:    1,
			Message: "商品不存在",
		}, nil
	}

	// 构建更新数据 - 只更新提供的字段
	product := &model.Product{
		BaseModel: pkg.BaseModel{
			ID: req.ProductId,
		},
	}

	// 使用现有值作为默认值，只有当请求中提供了新值时才更新
	product.CategoryID = existingProduct.CategoryID
	product.BrandID = existingProduct.BrandID
	product.Title = existingProduct.Title
	product.Subtitle = existingProduct.Subtitle
	product.MainImage = existingProduct.MainImage
	product.Description = existingProduct.Description
	product.Status = existingProduct.Status
	product.Images = existingProduct.Images

	// 只更新提供的字段
	if req.CategoryId != "" {
		product.CategoryID = req.CategoryId
	}
	if req.BrandId != "" {
		product.BrandID = req.BrandId
	}
	if req.Title != "" {
		product.Title = req.Title
	}
	if req.Subtitle != "" {
		product.Subtitle = req.Subtitle
	}
	if req.MainImage != "" {
		product.MainImage = req.MainImage
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Status > 0 {
		product.Status = int8(req.Status)
	}
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return &productv1.UpdateProductResponse{
				Code:    1,
				Message: fmt.Sprintf("图片列表格式错误: %v", err),
			}, nil
		}
		product.Images = string(imagesBytes)
	}

	err = s.productRepo.UpdateProduct(ctx, product)
	if err != nil {
		return &productv1.UpdateProductResponse{
			Code:    1,
			Message: fmt.Sprintf("更新商品失败: %v", err),
		}, nil
	}

	return &productv1.UpdateProductResponse{
		Code:    0,
		Message: "更新成功",
		Data:    req.ProductId,
	}, nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*productv1.DeleteProductResponse, error) {
	err := s.productRepo.DeleteProduct(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &productv1.DeleteProductResponse{
				Code:    1,
				Message: "商品不存在",
			}, nil
		}
		return &productv1.DeleteProductResponse{
			Code:    1,
			Message: fmt.Sprintf("删除商品失败: %v", err),
		}, nil
	}

	return &productv1.DeleteProductResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.ProductId,
	}, nil
}

// ListProducts 查询商品列表
func (s *ProductService) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	// 处理时间范围
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t := req.StartTime.AsTime()
		startTime = &t
	}
	if req.EndTime != nil {
		t := req.EndTime.AsTime()
		endTime = &t
	}

	filter := &repository.ProductListFliter{
		Page:       req.Page,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
		BrandId:    req.BrandId,
		Status:     req.Status,
		Keyword:    req.Keyword,
		StartTime:  startTime,
		EndTime:    endTime,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}
	result, err := s.productRepo.ListProducts(ctx, filter)
	if err != nil {
		return &productv1.ListProductsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品列表失败: %v", err),
		}, nil
	}

	productList := make([]*productv1.ProductInfo, 0, len(result.Products))
	for _, product := range result.Products {
		productList = append(productList, convertProductToProto(product))
	}
	return &productv1.ListProductsResponse{
		Code:     0,
		Message:  "查询成功",
		Total:    result.Total,
		Products: productList,
	}, nil
}

// OnShelfProduct 上架商品
func (s *ProductService) OnShelfProduct(ctx context.Context, req *productv1.OnShelfProductRequest) (*productv1.OnShelfProductResponse, error) {

	err := s.productRepo.OnShelfProduct(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &productv1.OnShelfProductResponse{
				Code:    1,
				Message: "商品不存在",
			}, nil
		}
		return &productv1.OnShelfProductResponse{
			Code:    1,
			Message: fmt.Sprintf("上架商品失败: %v", err),
		}, nil
	}

	return &productv1.OnShelfProductResponse{
		Code:    0,
		Message: "上架成功",
		Data:    req.ProductId,
	}, nil
}

// OffShelfProduct 下架商品
func (s *ProductService) OffShelfProduct(ctx context.Context, req *productv1.OffShelfProductRequest) (*productv1.OffShelfProductResponse, error) {
	err := s.productRepo.OffShelfProduct(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &productv1.OffShelfProductResponse{
				Code:    1,
				Message: "商品不存在",
			}, nil
		}
		return &productv1.OffShelfProductResponse{
			Code:    1,
			Message: fmt.Sprintf("下架商品失败: %v", err),
		}, nil
	}

	return &productv1.OffShelfProductResponse{
		Code:    0,
		Message: "下架成功",
		Data:    req.ProductId,
	}, nil
}

// SubmitProductAudit 提交审核
func (s *ProductService) SubmitProductAudit(ctx context.Context, req *productv1.SubmitProductAuditRequest) (*productv1.SubmitProductAuditResponse, error) {

	err := s.productRepo.SubmitProductAudit(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &productv1.SubmitProductAuditResponse{
				Code:    1,
				Message: "商品不存在",
			}, nil
		}
		return &productv1.SubmitProductAuditResponse{
			Code:    1,
			Message: fmt.Sprintf("提交审核失败: %v", err),
		}, nil
	}

	return &productv1.SubmitProductAuditResponse{
		Code:    0,
		Message: "提交审核成功",
		Data:    req.ProductId,
	}, nil
}

// AuditProduct 审核商品
func (s *ProductService) AuditProduct(ctx context.Context, req *productv1.AuditProductRequest) (*productv1.AuditProductResponse, error) {
	// 检查商品是否存在
	product, err := s.productRepo.GetProduct(ctx, req.ProductId)
	if err != nil {
		return &productv1.AuditProductResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品失败: %v", err),
		}, nil
	}

	if product == nil {
		return &productv1.AuditProductResponse{
			Code:    1,
			Message: "商品不存在",
		}, nil
	}

	// 验证审核结果
	if req.Result != 1 && req.Result != 2 {
		return &productv1.AuditProductResponse{
			Code:    1,
			Message: "审核结果无效，必须为1（通过）或2（驳回）",
		}, nil
	}

	err = s.productRepo.AuditProduct(ctx, req.ProductId, req.Result)
	if err != nil {
		return &productv1.AuditProductResponse{
			Code:    1,
			Message: fmt.Sprintf("审核商品失败: %v", err),
		}, nil
	}

	message := "审核通过"
	if req.Result == 2 {
		message = "审核驳回"
	}

	return &productv1.AuditProductResponse{
		Code:    0,
		Message: message,
		Data:    req.ProductId,
	}, nil
}

// ============================================
// 商品标签关联管理接口
// ============================================

// AddProductTag 添加商品标签关联
func (s *ProductService) AddProductTag(ctx context.Context, req *productv1.AddProductTagRequest) (*productv1.AddProductTagResponse, error) {
	err := s.productRepo.AddProductTag(ctx, req.ProductId, req.TagId)
	if err != nil {
		return &productv1.AddProductTagResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.AddProductTagResponse{
		Code:    0,
		Message: "添加成功",
		Data:    req.ProductId,
	}, nil
}

// RemoveProductTag 删除商品标签关联
func (s *ProductService) RemoveProductTag(ctx context.Context, req *productv1.RemoveProductTagRequest) (*productv1.RemoveProductTagResponse, error) {
	err := s.productRepo.RemoveProductTag(ctx, req.ProductId, req.TagId)
	if err != nil {
		return &productv1.RemoveProductTagResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.RemoveProductTagResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.ProductId,
	}, nil
}

// GetProductTags 查询商品的标签列表
func (s *ProductService) GetProductTags(ctx context.Context, req *productv1.GetProductTagsRequest) (*productv1.GetProductTagsResponse, error) {
	tags, err := s.productRepo.GetProductTags(ctx, req.ProductId)
	if err != nil {
		return &productv1.GetProductTagsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品标签列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	tagList := make([]*productv1.TagInfo, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, &productv1.TagInfo{
			Id:        tag.ID,
			Name:      tag.Name,
			Type:      int32(tag.Type),
			Color:     tag.Color,
			SortOrder: tag.SortOrder,
			Status:    int32(tag.Status),
			CreatedAt: timestamppb.New(tag.CreatedAt),
			UpdatedAt: timestamppb.New(tag.UpdatedAt),
		})
	}

	return &productv1.GetProductTagsResponse{
		Code:    0,
		Message: "查询成功",
		Tags:    tagList,
	}, nil
}

// BatchSetProductTags 批量设置商品标签关联
func (s *ProductService) BatchSetProductTags(ctx context.Context, req *productv1.BatchSetProductTagsRequest) (*productv1.BatchSetProductTagsResponse, error) {
	// 去重 tag_ids（业务逻辑，保留在 service 层）
	tagIDMap := make(map[string]bool)
	uniqueTagIDs := make([]string, 0, len(req.TagIds))
	for _, tagID := range req.TagIds {
		if tagID != "" && !tagIDMap[tagID] {
			tagIDMap[tagID] = true
			uniqueTagIDs = append(uniqueTagIDs, tagID)
		}
	}

	err := s.productRepo.BatchSetProductTags(ctx, req.ProductId, uniqueTagIDs)
	if err != nil {
		return &productv1.BatchSetProductTagsResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.BatchSetProductTagsResponse{
		Code:    0,
		Message: "批量设置成功",
		Data:    req.ProductId,
	}, nil
}

// ============================================
// SKU管理接口
// ============================================

// convertSkuToProto 将 model.Sku 转换为 productv1.SkuInfo
func convertSkuToProto(sku *model.Sku) *productv1.SkuInfo {
	return &productv1.SkuInfo{
		Id:            sku.ID,
		ProductId:     sku.ProductID,
		SkuCode:       sku.SkuCode,
		Barcode:       sku.Barcode,
		Name:          sku.Name,
		Price:         sku.Price,
		OriginalPrice: sku.OriginalPrice,
		CostPrice:     sku.CostPrice,
		Weight:        sku.Weight,
		Volume:        sku.Volume,
		Image:         sku.Image,
		Status:        int32(sku.Status),
		CreatedAt:     timestamppb.New(sku.CreatedAt),
		UpdatedAt:     timestamppb.New(sku.UpdatedAt),
	}
}

// CreateSku 创建SKU
func (s *ProductService) CreateSku(ctx context.Context, req *productv1.CreateSkuRequest) (*productv1.CreateSkuResponse, error) {
	// 检查商品是否存在
	product, err := s.productRepo.GetProduct(ctx, req.ProductId)
	if err != nil {
		return &productv1.CreateSkuResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品失败: %v", err),
		}, nil
	}
	if product == nil {
		return &productv1.CreateSkuResponse{
			Code:    1,
			Message: "商品不存在",
		}, nil
	}

	// 设置默认状态
	status := int8(1) // 默认上架
	if req.Status > 0 {
		status = int8(req.Status)
	}

	sku := &model.Sku{
		ProductID:     req.ProductId,
		SkuCode:       req.SkuCode,
		Barcode:       req.Barcode,
		Name:          req.Name,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		CostPrice:     req.CostPrice,
		Weight:        req.Weight,
		Volume:        req.Volume,
		Image:         req.Image,
		Status:        status,
	}

	// 使用事务同时创建SKU和设置属性关联
	err = s.skuRepo.CreateSkuWithAttributes(ctx, sku, req.AttributeValueIds)
	if err != nil {
		return &productv1.CreateSkuResponse{
			Code:    1,
			Message: fmt.Sprintf("创建SKU失败: %v", err),
		}, nil
	}

	return &productv1.CreateSkuResponse{
		Code:    0,
		Message: "创建成功",
		Data:    sku.ID,
	}, nil
}

// GetSku 查询SKU详情
func (s *ProductService) GetSku(ctx context.Context, req *productv1.GetSkuRequest) (*productv1.GetSkuResponse, error) {
	sku, err := s.skuRepo.GetSkuByID(ctx, req.SkuId)
	if err != nil {
		return &productv1.GetSkuResponse{
			Code:    1,
			Message: fmt.Sprintf("查询SKU详情失败: %v", err),
		}, nil
	}

	if sku == nil {
		return &productv1.GetSkuResponse{
			Code:    1,
			Message: "SKU不存在",
		}, nil
	}

	response := &productv1.GetSkuResponse{
		Code:    0,
		Message: "查询成功",
		Sku:     convertSkuToProto(sku),
	}

	if req.IncludeAttributes {
		// 获取 SKU 的属性值列表
		attributeValues, err := s.skuRepo.GetSkuAttributes(ctx, req.SkuId)
		if err == nil && len(attributeValues) > 0 {
			// 批量获取属性信息（用于显示属性名称）
			attributeMap := make(map[string]*model.Attribute)
			attributeIDs := make([]string, 0, len(attributeValues))
			seen := make(map[string]struct{})
			for _, av := range attributeValues {
				if _, exists := seen[av.AttributeID]; !exists {
					attributeIDs = append(attributeIDs, av.AttributeID)
					seen[av.AttributeID] = struct{}{}
				}
			}

			// 批量查询属性
			for _, attrID := range attributeIDs {
				attr, err := s.attributeRepo.GetAttributeByID(ctx, attrID)
				if err == nil && attr != nil {
					attributeMap[attrID] = attr
				}
			}

			// 转换为响应格式
			for _, attributeValue := range attributeValues {
				info := &productv1.AttributeValueInfo{
					Id:          attributeValue.ID,
					AttributeId: attributeValue.AttributeID,
					Value:       attributeValue.Value,
					SortOrder:   attributeValue.SortOrder,
					CreatedAt:   timestamppb.New(attributeValue.CreatedAt),
					UpdatedAt:   timestamppb.New(attributeValue.UpdatedAt),
				}

				// 填充属性名称
				if attr, exists := attributeMap[attributeValue.AttributeID]; exists {
					info.AttributeName = attr.Name
				}

				response.Attributes = append(response.Attributes, info)
			}
		}
		// 如果查询属性失败或没有属性，不返回错误，只返回空的属性列表
	}

	return response, nil
}

// UpdateSku 更新SKU
func (s *ProductService) UpdateSku(ctx context.Context, req *productv1.UpdateSkuRequest) (*productv1.UpdateSkuResponse, error) {
	// 先查询现有SKU
	existingSku, err := s.skuRepo.GetSkuByID(ctx, req.SkuId)
	if err != nil {
		return &productv1.UpdateSkuResponse{
			Code:    1,
			Message: fmt.Sprintf("查询SKU失败: %v", err),
		}, nil
	}

	if existingSku == nil {
		return &productv1.UpdateSkuResponse{
			Code:    1,
			Message: "SKU不存在",
		}, nil
	}

	// 构建更新数据 - 只更新提供的字段
	sku := &model.Sku{
		BaseModel: pkg.BaseModel{
			ID: req.SkuId,
		},
		SkuCode:       existingSku.SkuCode,
		Barcode:       existingSku.Barcode,
		Name:          existingSku.Name,
		Price:         existingSku.Price,
		OriginalPrice: existingSku.OriginalPrice,
		CostPrice:     existingSku.CostPrice,
		Weight:        existingSku.Weight,
		Volume:        existingSku.Volume,
		Image:         existingSku.Image,
		Status:        existingSku.Status,
	}

	// 只更新提供的字段
	if req.SkuCode != "" {
		sku.SkuCode = req.SkuCode
	}
	if req.Barcode != "" {
		sku.Barcode = req.Barcode
	}
	if req.Name != "" {
		sku.Name = req.Name
	}
	if req.Price > 0 {
		sku.Price = req.Price
	}
	if req.OriginalPrice > 0 {
		sku.OriginalPrice = req.OriginalPrice
	}
	if req.CostPrice > 0 {
		sku.CostPrice = req.CostPrice
	}
	if req.Weight > 0 {
		sku.Weight = req.Weight
	}
	if req.Volume > 0 {
		sku.Volume = req.Volume
	}
	if req.Image != "" {
		sku.Image = req.Image
	}
	if req.Status > 0 {
		sku.Status = int8(req.Status)
	}

	err = s.skuRepo.UpdateSku(ctx, sku)
	if err != nil {
		return &productv1.UpdateSkuResponse{
			Code:    1,
			Message: fmt.Sprintf("更新SKU失败: %v", err),
		}, nil
	}

	return &productv1.UpdateSkuResponse{
		Code:    0,
		Message: "更新成功",
		Data:    req.SkuId,
	}, nil
}

// DeleteSku 删除SKU
func (s *ProductService) DeleteSku(ctx context.Context, req *productv1.DeleteSkuRequest) (*productv1.DeleteSkuResponse, error) {
	err := s.skuRepo.DeleteSku(ctx, req.SkuId)
	if err != nil {
		return &productv1.DeleteSkuResponse{
			Code:    1,
			Message: fmt.Sprintf("删除SKU失败: %v", err),
		}, nil
	}

	return &productv1.DeleteSkuResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.SkuId,
	}, nil
}

// ListSkus 查询SKU列表
func (s *ProductService) ListSkus(ctx context.Context, req *productv1.ListSkusRequest) (*productv1.ListSkusResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &repository.SkuListFilter{
		Page:      page,
		PageSize:  pageSize,
		ProductID: req.ProductId,
		Status:    req.Status,
		Keyword:   req.Keyword,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		Offset:    int((page - 1) * pageSize),
		Limit:     int(pageSize),
	}

	skus, total, err := s.skuRepo.ListSkus(ctx, filter)
	if err != nil {
		return &productv1.ListSkusResponse{
			Code:    1,
			Message: fmt.Sprintf("查询SKU列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	skuList := make([]*productv1.SkuInfo, 0, len(skus))
	for _, sku := range skus {
		skuList = append(skuList, convertSkuToProto(sku))
	}

	return &productv1.ListSkusResponse{
		Code:    0,
		Message: "查询成功",
		Total:   total,
		Skus:    skuList,
	}, nil
}

// BatchCreateSkus 批量创建SKU
func (s *ProductService) BatchCreateSkus(ctx context.Context, req *productv1.BatchCreateSkusRequest) (*productv1.BatchCreateSkusResponse, error) {
	// 检查商品是否存在
	product, err := s.productRepo.GetProduct(ctx, req.ProductId)
	if err != nil {
		return &productv1.BatchCreateSkusResponse{
			Code:    1,
			Message: fmt.Sprintf("查询商品失败: %v", err),
		}, nil
	}
	if product == nil {
		return &productv1.BatchCreateSkusResponse{
			Code:    1,
			Message: "商品不存在",
		}, nil
	}

	if len(req.Skus) == 0 {
		return &productv1.BatchCreateSkusResponse{
			Code:    1,
			Message: "SKU列表不能为空",
		}, nil
	}

	// 转换为 model.Sku 列表
	skus := make([]*model.Sku, 0, len(req.Skus))
	for _, reqSku := range req.Skus {
		status := int8(1) // 默认上架
		if reqSku.Status > 0 {
			status = int8(reqSku.Status)
		}

		sku := &model.Sku{
			ProductID:     req.ProductId,
			SkuCode:       reqSku.SkuCode,
			Barcode:       reqSku.Barcode,
			Name:          reqSku.Name,
			Price:         reqSku.Price,
			OriginalPrice: reqSku.OriginalPrice,
			CostPrice:     reqSku.CostPrice,
			Weight:        reqSku.Weight,
			Volume:        reqSku.Volume,
			Image:         reqSku.Image,
			Status:        status,
		}
		skus = append(skus, sku)
	}

	err = s.skuRepo.BatchCreateSkus(ctx, req.ProductId, skus)
	if err != nil {
		return &productv1.BatchCreateSkusResponse{
			Code:    1,
			Message: fmt.Sprintf("批量创建SKU失败: %v", err),
		}, nil
	}

	// 返回创建的SKU ID列表
	skuIDs := make([]string, 0, len(skus))
	for _, sku := range skus {
		skuIDs = append(skuIDs, sku.ID)
	}

	return &productv1.BatchCreateSkusResponse{
		Code:    0,
		Message: "批量创建成功",
		SkuIds:  skuIDs,
	}, nil
}

// convertAttributeValueToProto 将 model.AttributeValue 转换为 AttributeValueInfo
func convertAttributeValueToProto(attr *model.AttributeValue) *productv1.AttributeValueInfo {
	return &productv1.AttributeValueInfo{
		Id:            attr.ID,
		AttributeId:   attr.AttributeID,
		AttributeName: "", // 属性名称暂未从 attributes 表加载，后续可扩展
		Value:         attr.Value,
		SortOrder:     attr.SortOrder,
		CreatedAt:     timestamppb.New(attr.CreatedAt),
		UpdatedAt:     timestamppb.New(attr.UpdatedAt),
	}
}

// ============================================
// SKU属性关联管理接口
// ============================================

// AddSkuAttribute 添加SKU属性关联
func (s *ProductService) AddSkuAttribute(ctx context.Context, req *productv1.AddSkuAttributeRequest) (*productv1.AddSkuAttributeResponse, error) {
	err := s.skuRepo.AddSkuAttribute(ctx, req.SkuId, req.AttributeValueId)
	if err != nil {
		return &productv1.AddSkuAttributeResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.AddSkuAttributeResponse{
		Code:    0,
		Message: "添加成功",
		Data:    req.SkuId,
	}, nil
}

// RemoveSkuAttribute 删除SKU属性关联
func (s *ProductService) RemoveSkuAttribute(ctx context.Context, req *productv1.RemoveSkuAttributeRequest) (*productv1.RemoveSkuAttributeResponse, error) {
	err := s.skuRepo.RemoveSkuAttribute(ctx, req.SkuId, req.AttributeValueId)
	if err != nil {
		return &productv1.RemoveSkuAttributeResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.RemoveSkuAttributeResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.SkuId,
	}, nil
}

// GetSkuAttributes 查询SKU的属性列表
func (s *ProductService) GetSkuAttributes(ctx context.Context, req *productv1.GetSkuAttributesRequest) (*productv1.GetSkuAttributesResponse, error) {
	attrs, err := s.skuRepo.GetSkuAttributes(ctx, req.SkuId)
	if err != nil {
		return &productv1.GetSkuAttributesResponse{
			Code:    1,
			Message: fmt.Sprintf("查询SKU属性列表失败: %v", err),
		}, nil
	}

	list := make([]*productv1.AttributeValueInfo, 0, len(attrs))
	for _, a := range attrs {
		list = append(list, convertAttributeValueToProto(a))
	}

	return &productv1.GetSkuAttributesResponse{
		Code:       0,
		Message:    "查询成功",
		Attributes: list,
	}, nil
}

// BatchSetSkuAttributes 批量设置SKU属性关联
func (s *ProductService) BatchSetSkuAttributes(ctx context.Context, req *productv1.BatchSetSkuAttributesRequest) (*productv1.BatchSetSkuAttributesResponse, error) {
	// 去重 attribute_value_ids
	idMap := make(map[string]struct{}, len(req.AttributeValueIds))
	uniqueIDs := make([]string, 0, len(req.AttributeValueIds))
	for _, id := range req.AttributeValueIds {
		if id == "" {
			continue
		}
		if _, exists := idMap[id]; !exists {
			idMap[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	err := s.skuRepo.BatchSetSkuAttributes(ctx, req.SkuId, uniqueIDs)
	if err != nil {
		return &productv1.BatchSetSkuAttributesResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.BatchSetSkuAttributesResponse{
		Code:    0,
		Message: "批量设置成功",
		Data:    req.SkuId,
	}, nil
}

// ============================================
// 标签管理接口
// ============================================

// convertTagToProto 将 model.Tag 转换为 productv1.TagInfo
func convertTagToProto(tag *model.Tag) *productv1.TagInfo {
	return &productv1.TagInfo{
		Id:        tag.ID,
		Name:      tag.Name,
		Type:      int32(tag.Type),
		Color:     tag.Color,
		SortOrder: tag.SortOrder,
		Status:    int32(tag.Status),
		CreatedAt: timestamppb.New(tag.CreatedAt),
		UpdatedAt: timestamppb.New(tag.UpdatedAt),
	}
}

// CreateTag 创建标签
func (s *ProductService) CreateTag(ctx context.Context, req *productv1.CreateTagRequest) (*productv1.CreateTagResponse, error) {
	// 设置默认值
	status := int8(1) // 默认启用
	if req.Status > 0 {
		status = int8(req.Status)
	}

	tagType := int8(2) // 默认运营标签
	if req.Type > 0 {
		tagType = int8(req.Type)
	}

	tag := &model.Tag{
		Name:      req.Name,
		Type:      tagType,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		Status:    status,
	}

	err := s.tagRepo.CreateTag(ctx, tag)
	if err != nil {
		return &productv1.CreateTagResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &productv1.CreateTagResponse{
		Code:    0,
		Message: "创建成功",
		Data:    tag.ID,
	}, nil
}

// GetTag 查询标签详情
func (s *ProductService) GetTag(ctx context.Context, req *productv1.GetTagRequest) (*productv1.GetTagResponse, error) {
	tag, err := s.tagRepo.GetTagByID(ctx, req.TagId)
	if err != nil {
		return &productv1.GetTagResponse{
			Code:    1,
			Message: fmt.Sprintf("查询标签详情失败: %v", err),
		}, nil
	}

	if tag == nil {
		return &productv1.GetTagResponse{
			Code:    1,
			Message: "标签不存在",
		}, nil
	}

	return &productv1.GetTagResponse{
		Code:    0,
		Message: "查询成功",
		Tag:     convertTagToProto(tag),
	}, nil
}

// UpdateTag 更新标签
func (s *ProductService) UpdateTag(ctx context.Context, req *productv1.UpdateTagRequest) (*productv1.UpdateTagResponse, error) {
	// 先查询现有标签
	existingTag, err := s.tagRepo.GetTagByID(ctx, req.TagId)
	if err != nil {
		return &productv1.UpdateTagResponse{
			Code:    1,
			Message: fmt.Sprintf("查询标签失败: %v", err),
		}, nil
	}

	if existingTag == nil {
		return &productv1.UpdateTagResponse{
			Code:    1,
			Message: "标签不存在",
		}, nil
	}

	// 构建更新数据 - 只更新提供的字段
	tag := &model.Tag{
		BaseModel: pkg.BaseModel{
			ID: req.TagId,
		},
		Name:      existingTag.Name,
		Type:      existingTag.Type,
		Color:     existingTag.Color,
		SortOrder: existingTag.SortOrder,
		Status:    existingTag.Status,
	}

	// 只更新提供的字段
	if req.Name != "" {
		tag.Name = req.Name
	}
	if req.Type > 0 {
		tag.Type = int8(req.Type)
	}
	if req.Color != "" {
		tag.Color = req.Color
	}
	if req.SortOrder > 0 {
		tag.SortOrder = req.SortOrder
	}
	if req.Status > 0 {
		tag.Status = int8(req.Status)
	}

	err = s.tagRepo.UpdateTag(ctx, tag)
	if err != nil {
		return &productv1.UpdateTagResponse{
			Code:    1,
			Message: fmt.Sprintf("更新标签失败: %v", err),
		}, nil
	}

	return &productv1.UpdateTagResponse{
		Code:    0,
		Message: "更新成功",
		Data:    req.TagId,
	}, nil
}

// DeleteTag 删除标签
func (s *ProductService) DeleteTag(ctx context.Context, req *productv1.DeleteTagRequest) (*productv1.DeleteTagResponse, error) {
	err := s.tagRepo.DeleteTag(ctx, req.TagId)
	if err != nil {
		return &productv1.DeleteTagResponse{
			Code:    1,
			Message: fmt.Sprintf("删除标签失败: %v", err),
		}, nil
	}

	return &productv1.DeleteTagResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.TagId,
	}, nil
}

// ListTags 查询标签列表
func (s *ProductService) ListTags(ctx context.Context, req *productv1.ListTagsRequest) (*productv1.ListTagsResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &repository.TagListFilter{
		Page:     page,
		PageSize: pageSize,
		Type:     req.Type,
		Status:   req.Status,
		Keyword:  req.Keyword,
		Offset:   int((page - 1) * pageSize),
		Limit:    int(pageSize),
	}

	tags, total, err := s.tagRepo.ListTags(ctx, filter)
	if err != nil {
		return &productv1.ListTagsResponse{
			Code:    1,
			Message: fmt.Sprintf("查询标签列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	tagList := make([]*productv1.TagInfo, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, convertTagToProto(tag))
	}

	return &productv1.ListTagsResponse{
		Code:    0,
		Message: "查询成功",
		Total:   total,
		Tags:    tagList,
	}, nil
}

// ============================================
// 属性管理接口
// ============================================

// convertAttributeToProto 将 model.Attribute 转换为 productv1.AttributeInfo
func convertAttributeToProto(attr *model.Attribute) *productv1.AttributeInfo {
	return &productv1.AttributeInfo{
		Id:         attr.ID,
		CategoryId: attr.CategoryID,
		Name:       attr.Name,
		Type:       int32(attr.Type),
		InputType:  int32(attr.InputType),
		IsRequired: int32(attr.IsRequired),
		SortOrder:  attr.SortOrder,
		CreatedAt:  timestamppb.New(attr.CreatedAt),
		UpdatedAt:  timestamppb.New(attr.UpdatedAt),
	}
}

// CreateAttribute 创建属性
func (s *ProductService) CreateAttribute(ctx context.Context, req *productv1.CreateAttributeRequest) (*productv1.CreateAttributeResponse, error) {
	// 检查类目是否存在
	category, err := s.categoryRepo.GetCategoryByID(ctx, req.CategoryId)
	if err != nil {
		return &productv1.CreateAttributeResponse{
			Code:    1,
			Message: fmt.Sprintf("查询类目失败: %v", err),
		}, nil
	}
	if category == nil {
		return &productv1.CreateAttributeResponse{
			Code:    1,
			Message: "类目不存在",
		}, nil
	}

	attribute := &model.Attribute{
		CategoryID: req.CategoryId,
		Name:       req.Name,
		Type:       int8(req.Type),
		InputType:  int8(req.InputType),
		IsRequired: int8(req.IsRequired),
		SortOrder:  req.SortOrder,
	}

	err = s.attributeRepo.CreateAttribute(ctx, attribute)
	if err != nil {
		return &productv1.CreateAttributeResponse{
			Code:    1,
			Message: fmt.Sprintf("创建属性失败: %v", err),
		}, nil
	}

	return &productv1.CreateAttributeResponse{
		Code:    0,
		Message: "创建成功",
		Data:    attribute.ID,
	}, nil
}

// GetAttribute 查询属性详情
func (s *ProductService) GetAttribute(ctx context.Context, req *productv1.GetAttributeRequest) (*productv1.GetAttributeResponse, error) {
	attribute, err := s.attributeRepo.GetAttributeByID(ctx, req.AttributeId)
	if err != nil {
		return &productv1.GetAttributeResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性详情失败: %v", err),
		}, nil
	}

	if attribute == nil {
		return &productv1.GetAttributeResponse{
			Code:    1,
			Message: "属性不存在",
		}, nil
	}

	return &productv1.GetAttributeResponse{
		Code:      0,
		Message:   "查询成功",
		Attribute: convertAttributeToProto(attribute),
	}, nil
}

// UpdateAttribute 更新属性
func (s *ProductService) UpdateAttribute(ctx context.Context, req *productv1.UpdateAttributeRequest) (*productv1.UpdateAttributeResponse, error) {
	// 先查询现有属性
	existingAttribute, err := s.attributeRepo.GetAttributeByID(ctx, req.AttributeId)
	if err != nil {
		return &productv1.UpdateAttributeResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性失败: %v", err),
		}, nil
	}

	if existingAttribute == nil {
		return &productv1.UpdateAttributeResponse{
			Code:    1,
			Message: "属性不存在",
		}, nil
	}

	// 构建更新数据 - 只更新提供的字段
	attribute := &model.Attribute{
		BaseModel: pkg.BaseModel{
			ID: req.AttributeId,
		},
		Name:       existingAttribute.Name,
		Type:       existingAttribute.Type,
		InputType:  existingAttribute.InputType,
		IsRequired: existingAttribute.IsRequired,
		SortOrder:  existingAttribute.SortOrder,
	}

	// 只更新提供的字段
	if req.Name != "" {
		attribute.Name = req.Name
	}
	if req.Type > 0 {
		attribute.Type = int8(req.Type)
	}
	if req.InputType > 0 {
		attribute.InputType = int8(req.InputType)
	}
	if req.IsRequired >= 0 {
		attribute.IsRequired = int8(req.IsRequired)
	}
	if req.SortOrder > 0 {
		attribute.SortOrder = req.SortOrder
	}

	err = s.attributeRepo.UpdateAttribute(ctx, attribute)
	if err != nil {
		return &productv1.UpdateAttributeResponse{
			Code:    1,
			Message: fmt.Sprintf("更新属性失败: %v", err),
		}, nil
	}

	return &productv1.UpdateAttributeResponse{
		Code:    0,
		Message: "更新成功",
		Data:    req.AttributeId,
	}, nil
}

// DeleteAttribute 删除属性
func (s *ProductService) DeleteAttribute(ctx context.Context, req *productv1.DeleteAttributeRequest) (*productv1.DeleteAttributeResponse, error) {
	err := s.attributeRepo.DeleteAttribute(ctx, req.AttributeId)
	if err != nil {
		return &productv1.DeleteAttributeResponse{
			Code:    1,
			Message: fmt.Sprintf("删除属性失败: %v", err),
		}, nil
	}

	return &productv1.DeleteAttributeResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.AttributeId,
	}, nil
}

// ListAttributes 查询属性列表
func (s *ProductService) ListAttributes(ctx context.Context, req *productv1.ListAttributesRequest) (*productv1.ListAttributesResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &repository.AttributeListFilter{
		Page:       page,
		PageSize:   pageSize,
		CategoryID: req.CategoryId,
		Type:       req.Type,
		IsRequired: req.IsRequired,
		Keyword:    req.Keyword,
		Offset:     int((page - 1) * pageSize),
		Limit:      int(pageSize),
	}

	attributes, total, err := s.attributeRepo.ListAttributes(ctx, filter)
	if err != nil {
		return &productv1.ListAttributesResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	attributeList := make([]*productv1.AttributeInfo, 0, len(attributes))
	for _, attr := range attributes {
		attributeList = append(attributeList, convertAttributeToProto(attr))
	}

	return &productv1.ListAttributesResponse{
		Code:       0,
		Message:    "查询成功",
		Total:      total,
		Attributes: attributeList,
	}, nil
}

// ============================================
// 属性值管理接口
// ============================================

// CreateAttributeValue 创建属性值
func (s *ProductService) CreateAttributeValue(ctx context.Context, req *productv1.CreateAttributeValueRequest) (*productv1.CreateAttributeValueResponse, error) {

	attributeValue := &model.AttributeValue{
		AttributeID: req.AttributeId,
		Value:       req.Value,
		SortOrder:   req.SortOrder,
	}

	err := s.attributeValueRepo.CreateAttributeValue(ctx, attributeValue)
	if err != nil {
		return &productv1.CreateAttributeValueResponse{
			Code:    1,
			Message: fmt.Sprintf("创建属性值失败: %v", err),
		}, nil
	}

	return &productv1.CreateAttributeValueResponse{
		Code:    0,
		Message: "创建成功",
		Data:    attributeValue.ID,
	}, nil
}

// GetAttributeValue 查询属性值详情
func (s *ProductService) GetAttributeValue(ctx context.Context, req *productv1.GetAttributeValueRequest) (*productv1.GetAttributeValueResponse, error) {
	attributeValue, err := s.attributeValueRepo.GetAttributeValueByID(ctx, req.AttributeValueId)
	if err != nil {
		return &productv1.GetAttributeValueResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性值详情失败: %v", err),
		}, nil
	}

	if attributeValue == nil {
		return &productv1.GetAttributeValueResponse{
			Code:    1,
			Message: "属性值不存在",
		}, nil
	}

	// 获取属性信息（用于显示属性名称）
	attribute, err := s.attributeRepo.GetAttributeByID(ctx, attributeValue.AttributeID)
	if err != nil {
		return &productv1.GetAttributeValueResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性失败: %v", err),
		}, nil
	}

	attributeValueInfo := &productv1.AttributeValueInfo{
		Id:          attributeValue.ID,
		AttributeId: attributeValue.AttributeID,
		Value:       attributeValue.Value,
		SortOrder:   attributeValue.SortOrder,
		CreatedAt:   timestamppb.New(attributeValue.CreatedAt),
		UpdatedAt:   timestamppb.New(attributeValue.UpdatedAt),
	}

	if attribute != nil {
		attributeValueInfo.AttributeName = attribute.Name
	}

	return &productv1.GetAttributeValueResponse{
		Code:           0,
		Message:        "查询成功",
		AttributeValue: attributeValueInfo,
	}, nil
}

// UpdateAttributeValue 更新属性值
func (s *ProductService) UpdateAttributeValue(ctx context.Context, req *productv1.UpdateAttributeValueRequest) (*productv1.UpdateAttributeValueResponse, error) {
	// 先查询现有属性值
	existingAttributeValue, err := s.attributeValueRepo.GetAttributeValueByID(ctx, req.AttributeValueId)
	if err != nil {
		return &productv1.UpdateAttributeValueResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性值失败: %v", err),
		}, nil
	}

	if existingAttributeValue == nil {
		return &productv1.UpdateAttributeValueResponse{
			Code:    1,
			Message: "属性值不存在",
		}, nil
	}

	// 构建更新数据 - 只更新提供的字段
	attributeValue := &model.AttributeValue{
		BaseModel: pkg.BaseModel{
			ID: req.AttributeValueId,
		},
		Value:     existingAttributeValue.Value,
		SortOrder: existingAttributeValue.SortOrder,
	}

	// 只更新提供的字段
	if req.Value != "" {
		attributeValue.Value = req.Value
	}
	if req.SortOrder > 0 {
		attributeValue.SortOrder = req.SortOrder
	}

	err = s.attributeValueRepo.UpdateAttributeValue(ctx, attributeValue)
	if err != nil {
		return &productv1.UpdateAttributeValueResponse{
			Code:    1,
			Message: fmt.Sprintf("更新属性值失败: %v", err),
		}, nil
	}

	return &productv1.UpdateAttributeValueResponse{
		Code:    0,
		Message: "更新成功",
		Data:    req.AttributeValueId,
	}, nil
}

// DeleteAttributeValue 删除属性值
func (s *ProductService) DeleteAttributeValue(ctx context.Context, req *productv1.DeleteAttributeValueRequest) (*productv1.DeleteAttributeValueResponse, error) {
	err := s.attributeValueRepo.DeleteAttributeValue(ctx, req.AttributeValueId)
	if err != nil {
		return &productv1.DeleteAttributeValueResponse{
			Code:    1,
			Message: fmt.Sprintf("删除属性值失败: %v", err),
		}, nil
	}

	return &productv1.DeleteAttributeValueResponse{
		Code:    0,
		Message: "删除成功",
		Data:    req.AttributeValueId,
	}, nil
}

// ListAttributeValues 查询属性值列表
func (s *ProductService) ListAttributeValues(ctx context.Context, req *productv1.ListAttributeValuesRequest) (*productv1.ListAttributeValuesResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &repository.AttributeValueListFilter{
		Page:        page,
		PageSize:    pageSize,
		AttributeID: req.AttributeId,
		Keyword:     req.Keyword,
		Offset:      int((page - 1) * pageSize),
		Limit:       int(pageSize),
	}

	attributeValues, total, err := s.attributeValueRepo.ListAttributeValues(ctx, filter)
	if err != nil {
		return &productv1.ListAttributeValuesResponse{
			Code:    1,
			Message: fmt.Sprintf("查询属性值列表失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	attributeValueList := make([]*productv1.AttributeValueInfo, 0, len(attributeValues))

	// 批量获取属性信息（用于显示属性名称）
	attributeMap := make(map[string]*model.Attribute)
	if len(attributeValues) > 0 {
		attributeIDs := make([]string, 0, len(attributeValues))
		for _, av := range attributeValues {
			if _, exists := attributeMap[av.AttributeID]; !exists {
				attributeIDs = append(attributeIDs, av.AttributeID)
			}
		}

		// 批量查询属性
		for _, attrID := range attributeIDs {
			attr, err := s.attributeRepo.GetAttributeByID(ctx, attrID)
			if err == nil && attr != nil {
				attributeMap[attrID] = attr
			}
		}
	}

	for _, av := range attributeValues {
		info := &productv1.AttributeValueInfo{
			Id:          av.ID,
			AttributeId: av.AttributeID,
			Value:       av.Value,
			SortOrder:   av.SortOrder,
			CreatedAt:   timestamppb.New(av.CreatedAt),
			UpdatedAt:   timestamppb.New(av.UpdatedAt),
		}

		if attr, exists := attributeMap[av.AttributeID]; exists {
			info.AttributeName = attr.Name
		}

		attributeValueList = append(attributeValueList, info)
	}

	return &productv1.ListAttributeValuesResponse{
		Code:            0,
		Message:         "查询成功",
		Total:           total,
		AttributeValues: attributeValueList,
	}, nil
}
