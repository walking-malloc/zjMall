package service

import (
	"context"
	"fmt"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/internal/product-service/model"
	"zjMall/internal/product-service/repository"
	"zjMall/pkg"
)

// ProductService 商品服务（业务逻辑层）
type ProductService struct {
	// TODO: 添加需要的依赖，例如：
	categoryRepo repository.CategoryRepository
	// brandRepo    repository.BrandRepository
	// productRepo  repository.ProductRepository
	// skuRepo      repository.SkuRepository
	// tagRepo      repository.TagRepository
}

// NewProductService 创建商品服务实例
func NewProductService(
	categoryRepo repository.CategoryRepository,

) *ProductService {
	return &ProductService{
		categoryRepo: categoryRepo,
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
			Message: fmt.Sprintf("创建类目失败: %v", err),
		}, nil
	}
	return &productv1.CreateCategoryResponse{
		Code:    1,
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
		Level:     int8(req.Level),
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
	//先查找是否有子类目
	children, err := s.categoryRepo.ListCategories(ctx, &repository.CategoryListFliter{
		ParentID: req.CategoryId,
	})
	if err != nil {
		return &productv1.DeleteCategoryResponse{
			Code:    1,
			Message: fmt.Sprintf("查询子类目列表失败: %v", err),
		}, nil
	}
	if len(children) > 0 {
		return &productv1.DeleteCategoryResponse{
			Code:    1,
			Message: "该类目有子类目，不能删除",
		}, nil
	}
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
	isVisible := &req.IsVisible
	filter := repository.CategoryListFliter{
		ParentID:  req.ParentId,
		Level:     req.Level,
		Status:    req.Status,
		IsVisible: isVisible,
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
		IsVisible: &req.OnlyVisible,
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
	// TODO: 实现创建品牌的业务逻辑
	return &productv1.CreateBrandResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetBrand 查询品牌详情
func (s *ProductService) GetBrand(ctx context.Context, req *productv1.GetBrandRequest) (*productv1.GetBrandResponse, error) {
	// TODO: 实现查询品牌详情的业务逻辑
	return &productv1.GetBrandResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// UpdateBrand 更新品牌
func (s *ProductService) UpdateBrand(ctx context.Context, req *productv1.UpdateBrandRequest) (*productv1.UpdateBrandResponse, error) {
	// TODO: 实现更新品牌的业务逻辑
	return &productv1.UpdateBrandResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// DeleteBrand 删除品牌
func (s *ProductService) DeleteBrand(ctx context.Context, req *productv1.DeleteBrandRequest) (*productv1.DeleteBrandResponse, error) {
	// TODO: 实现删除品牌的业务逻辑
	return &productv1.DeleteBrandResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ListBrands 查询品牌列表
func (s *ProductService) ListBrands(ctx context.Context, req *productv1.ListBrandsRequest) (*productv1.ListBrandsResponse, error) {
	// TODO: 实现查询品牌列表的业务逻辑
	return &productv1.ListBrandsResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetBrandsByFirstLetter 按首字母分组查询品牌列表
func (s *ProductService) GetBrandsByFirstLetter(ctx context.Context, req *productv1.GetBrandsByFirstLetterRequest) (*productv1.GetBrandsByFirstLetterResponse, error) {
	// TODO: 实现按首字母分组查询品牌的业务逻辑
	return &productv1.GetBrandsByFirstLetterResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ============================================
// 品牌类目关联管理接口
// ============================================

// AddBrandCategory 添加品牌类目关联
func (s *ProductService) AddBrandCategory(ctx context.Context, req *productv1.AddBrandCategoryRequest) (*productv1.AddBrandCategoryResponse, error) {
	// TODO: 实现添加品牌类目关联的业务逻辑
	return &productv1.AddBrandCategoryResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// RemoveBrandCategory 删除品牌类目关联
func (s *ProductService) RemoveBrandCategory(ctx context.Context, req *productv1.RemoveBrandCategoryRequest) (*productv1.RemoveBrandCategoryResponse, error) {
	// TODO: 实现删除品牌类目关联的业务逻辑
	return &productv1.RemoveBrandCategoryResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetBrandCategories 查询品牌的类目列表
func (s *ProductService) GetBrandCategories(ctx context.Context, req *productv1.GetBrandCategoriesRequest) (*productv1.GetBrandCategoriesResponse, error) {
	// TODO: 实现查询品牌类目列表的业务逻辑
	return &productv1.GetBrandCategoriesResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// BatchSetBrandCategories 批量设置品牌类目关联
func (s *ProductService) BatchSetBrandCategories(ctx context.Context, req *productv1.BatchSetBrandCategoriesRequest) (*productv1.BatchSetBrandCategoriesResponse, error) {
	// TODO: 实现批量设置品牌类目关联的业务逻辑
	return &productv1.BatchSetBrandCategoriesResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ============================================
// 商品（SPU）管理接口
// ============================================

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	// TODO: 实现创建商品的业务逻辑
	return &productv1.CreateProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetProduct 查询商品详情
func (s *ProductService) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	// TODO: 实现查询商品详情的业务逻辑
	return &productv1.GetProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	// TODO: 实现更新商品的业务逻辑
	return &productv1.UpdateProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*productv1.DeleteProductResponse, error) {
	// TODO: 实现删除商品的业务逻辑
	return &productv1.DeleteProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ListProducts 查询商品列表
func (s *ProductService) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	// TODO: 实现查询商品列表的业务逻辑
	return &productv1.ListProductsResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// OnShelfProduct 上架商品
func (s *ProductService) OnShelfProduct(ctx context.Context, req *productv1.OnShelfProductRequest) (*productv1.OnShelfProductResponse, error) {
	// TODO: 实现上架商品的业务逻辑
	return &productv1.OnShelfProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// OffShelfProduct 下架商品
func (s *ProductService) OffShelfProduct(ctx context.Context, req *productv1.OffShelfProductRequest) (*productv1.OffShelfProductResponse, error) {
	// TODO: 实现下架商品的业务逻辑
	return &productv1.OffShelfProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// SubmitProductAudit 提交审核
func (s *ProductService) SubmitProductAudit(ctx context.Context, req *productv1.SubmitProductAuditRequest) (*productv1.SubmitProductAuditResponse, error) {
	// TODO: 实现提交审核的业务逻辑
	return &productv1.SubmitProductAuditResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// AuditProduct 审核商品
func (s *ProductService) AuditProduct(ctx context.Context, req *productv1.AuditProductRequest) (*productv1.AuditProductResponse, error) {
	// TODO: 实现审核商品的业务逻辑
	return &productv1.AuditProductResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ============================================
// 商品标签关联管理接口
// ============================================

// AddProductTag 添加商品标签关联
func (s *ProductService) AddProductTag(ctx context.Context, req *productv1.AddProductTagRequest) (*productv1.AddProductTagResponse, error) {
	// TODO: 实现添加商品标签关联的业务逻辑
	return &productv1.AddProductTagResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// RemoveProductTag 删除商品标签关联
func (s *ProductService) RemoveProductTag(ctx context.Context, req *productv1.RemoveProductTagRequest) (*productv1.RemoveProductTagResponse, error) {
	// TODO: 实现删除商品标签关联的业务逻辑
	return &productv1.RemoveProductTagResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetProductTags 查询商品的标签列表
func (s *ProductService) GetProductTags(ctx context.Context, req *productv1.GetProductTagsRequest) (*productv1.GetProductTagsResponse, error) {
	// TODO: 实现查询商品标签列表的业务逻辑
	return &productv1.GetProductTagsResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// BatchSetProductTags 批量设置商品标签关联
func (s *ProductService) BatchSetProductTags(ctx context.Context, req *productv1.BatchSetProductTagsRequest) (*productv1.BatchSetProductTagsResponse, error) {
	// TODO: 实现批量设置商品标签关联的业务逻辑
	return &productv1.BatchSetProductTagsResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ============================================
// SKU管理接口
// ============================================

// CreateSku 创建SKU
func (s *ProductService) CreateSku(ctx context.Context, req *productv1.CreateSkuRequest) (*productv1.CreateSkuResponse, error) {
	// TODO: 实现创建SKU的业务逻辑
	return &productv1.CreateSkuResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetSku 查询SKU详情
func (s *ProductService) GetSku(ctx context.Context, req *productv1.GetSkuRequest) (*productv1.GetSkuResponse, error) {
	// TODO: 实现查询SKU详情的业务逻辑
	return &productv1.GetSkuResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// UpdateSku 更新SKU
func (s *ProductService) UpdateSku(ctx context.Context, req *productv1.UpdateSkuRequest) (*productv1.UpdateSkuResponse, error) {
	// TODO: 实现更新SKU的业务逻辑
	return &productv1.UpdateSkuResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// DeleteSku 删除SKU
func (s *ProductService) DeleteSku(ctx context.Context, req *productv1.DeleteSkuRequest) (*productv1.DeleteSkuResponse, error) {
	// TODO: 实现删除SKU的业务逻辑
	return &productv1.DeleteSkuResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ListSkus 查询SKU列表
func (s *ProductService) ListSkus(ctx context.Context, req *productv1.ListSkusRequest) (*productv1.ListSkusResponse, error) {
	// TODO: 实现查询SKU列表的业务逻辑
	return &productv1.ListSkusResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// BatchCreateSkus 批量创建SKU
func (s *ProductService) BatchCreateSkus(ctx context.Context, req *productv1.BatchCreateSkusRequest) (*productv1.BatchCreateSkusResponse, error) {
	// TODO: 实现批量创建SKU的业务逻辑
	return &productv1.BatchCreateSkusResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ============================================
// SKU属性关联管理接口
// ============================================

// AddSkuAttribute 添加SKU属性关联
func (s *ProductService) AddSkuAttribute(ctx context.Context, req *productv1.AddSkuAttributeRequest) (*productv1.AddSkuAttributeResponse, error) {
	// TODO: 实现添加SKU属性关联的业务逻辑
	return &productv1.AddSkuAttributeResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// RemoveSkuAttribute 删除SKU属性关联
func (s *ProductService) RemoveSkuAttribute(ctx context.Context, req *productv1.RemoveSkuAttributeRequest) (*productv1.RemoveSkuAttributeResponse, error) {
	// TODO: 实现删除SKU属性关联的业务逻辑
	return &productv1.RemoveSkuAttributeResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetSkuAttributes 查询SKU的属性列表
func (s *ProductService) GetSkuAttributes(ctx context.Context, req *productv1.GetSkuAttributesRequest) (*productv1.GetSkuAttributesResponse, error) {
	// TODO: 实现查询SKU属性列表的业务逻辑
	return &productv1.GetSkuAttributesResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// BatchSetSkuAttributes 批量设置SKU属性关联
func (s *ProductService) BatchSetSkuAttributes(ctx context.Context, req *productv1.BatchSetSkuAttributesRequest) (*productv1.BatchSetSkuAttributesResponse, error) {
	// TODO: 实现批量设置SKU属性关联的业务逻辑
	return &productv1.BatchSetSkuAttributesResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ============================================
// 标签管理接口
// ============================================

// CreateTag 创建标签
func (s *ProductService) CreateTag(ctx context.Context, req *productv1.CreateTagRequest) (*productv1.CreateTagResponse, error) {
	// TODO: 实现创建标签的业务逻辑
	return &productv1.CreateTagResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// GetTag 查询标签详情
func (s *ProductService) GetTag(ctx context.Context, req *productv1.GetTagRequest) (*productv1.GetTagResponse, error) {
	// TODO: 实现查询标签详情的业务逻辑
	return &productv1.GetTagResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// UpdateTag 更新标签
func (s *ProductService) UpdateTag(ctx context.Context, req *productv1.UpdateTagRequest) (*productv1.UpdateTagResponse, error) {
	// TODO: 实现更新标签的业务逻辑
	return &productv1.UpdateTagResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// DeleteTag 删除标签
func (s *ProductService) DeleteTag(ctx context.Context, req *productv1.DeleteTagRequest) (*productv1.DeleteTagResponse, error) {
	// TODO: 实现删除标签的业务逻辑
	return &productv1.DeleteTagResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}

// ListTags 查询标签列表
func (s *ProductService) ListTags(ctx context.Context, req *productv1.ListTagsRequest) (*productv1.ListTagsResponse, error) {
	// TODO: 实现查询标签列表的业务逻辑
	return &productv1.ListTagsResponse{
		Code:    1,
		Message: "未实现",
	}, nil
}
