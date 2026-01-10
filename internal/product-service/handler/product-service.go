package handler

import (
	"context"
	"log"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/internal/product-service/service"
)

type ProductServiceHandler struct {
	productv1.UnimplementedProductServiceServer
	productService *service.ProductService
}

func NewProductServiceHandler(productService *service.ProductService) *ProductServiceHandler {
	return &ProductServiceHandler{
		productService: productService,
	}
}

// ============================================
// 类目管理接口
// ============================================

func (h *ProductServiceHandler) CreateCategory(ctx context.Context, req *productv1.CreateCategoryRequest) (*productv1.CreateCategoryResponse, error) {
	// 参数校验
	validator := service.NewCreateCategoryRequestValidator(req)
	log.Printf("DEBUG - Validator IsLeaf: %v, IsVisible: %v", validator.IsLeaf, validator.IsVisible)
	if err := validator.Validate(); err != nil {
		log.Printf("DEBUG - Validation error: %v", err)
		return &productv1.CreateCategoryResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.CreateCategory(ctx, req)
}

func (h *ProductServiceHandler) GetCategory(ctx context.Context, req *productv1.GetCategoryRequest) (*productv1.GetCategoryResponse, error) {
	if req.CategoryId == "" {
		return &productv1.GetCategoryResponse{
			Code:    1,
			Message: "类目ID不能为空",
		}, nil
	}

	return h.productService.GetCategory(ctx, req)
}

func (h *ProductServiceHandler) UpdateCategory(ctx context.Context, req *productv1.UpdateCategoryRequest) (*productv1.UpdateCategoryResponse, error) {
	validator := service.NewUpdateCategoryRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.UpdateCategoryResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.UpdateCategory(ctx, req)
}

func (h *ProductServiceHandler) DeleteCategory(ctx context.Context, req *productv1.DeleteCategoryRequest) (*productv1.DeleteCategoryResponse, error) {
	if req.CategoryId == "" {
		return &productv1.DeleteCategoryResponse{
			Code:    1,
			Message: "类目ID不能为空",
		}, nil
	}
	return h.productService.DeleteCategory(ctx, req)
}

func (h *ProductServiceHandler) ListCategories(ctx context.Context, req *productv1.ListCategoriesRequest) (*productv1.ListCategoriesResponse, error) {
	log.Printf("DEBUG - ListCategories request: %+v", req)
	validator := service.NewListCategoriesRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.ListCategoriesResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.ListCategories(ctx, req)
}

func (h *ProductServiceHandler) GetCategoryTree(ctx context.Context, req *productv1.GetCategoryTreeRequest) (*productv1.GetCategoryTreeResponse, error) {
	return h.productService.GetCategoryTree(ctx, req)
}

func (h *ProductServiceHandler) GetCategoryChildren(ctx context.Context, req *productv1.GetCategoryChildrenRequest) (*productv1.GetCategoryChildrenResponse, error) {
	validator := service.NewGetCategoryChildrenRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.GetCategoryChildrenResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.GetCategoryChildren(ctx, req)
}

// ============================================
// 品牌管理接口
// ============================================

func (h *ProductServiceHandler) CreateBrand(ctx context.Context, req *productv1.CreateBrandRequest) (*productv1.CreateBrandResponse, error) {
	validator := service.NewCreateBrandRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.CreateBrandResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.CreateBrand(ctx, req)
}

func (h *ProductServiceHandler) GetBrand(ctx context.Context, req *productv1.GetBrandRequest) (*productv1.GetBrandResponse, error) {
	if req.BrandId == "" {
		return &productv1.GetBrandResponse{
			Code:    1,
			Message: "品牌ID不能为空",
		}, nil
	}
	return h.productService.GetBrand(ctx, req)
}

func (h *ProductServiceHandler) UpdateBrand(ctx context.Context, req *productv1.UpdateBrandRequest) (*productv1.UpdateBrandResponse, error) {
	validator := service.NewUpdateBrandRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.UpdateBrandResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.UpdateBrand(ctx, req)
}

func (h *ProductServiceHandler) DeleteBrand(ctx context.Context, req *productv1.DeleteBrandRequest) (*productv1.DeleteBrandResponse, error) {
	if req.BrandId == "" {
		return &productv1.DeleteBrandResponse{
			Code:    1,
			Message: "品牌ID不能为空",
		}, nil
	}
	return h.productService.DeleteBrand(ctx, req)
}

func (h *ProductServiceHandler) ListBrands(ctx context.Context, req *productv1.ListBrandsRequest) (*productv1.ListBrandsResponse, error) {
	validator := service.NewListBrandsRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.ListBrandsResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.ListBrands(ctx, req)
}

func (h *ProductServiceHandler) GetBrandsByFirstLetter(ctx context.Context, req *productv1.GetBrandsByFirstLetterRequest) (*productv1.GetBrandsByFirstLetterResponse, error) {
	return h.productService.GetBrandsByFirstLetter(ctx, req)
}

// ============================================
// 品牌类目关联管理接口
// ============================================

func (h *ProductServiceHandler) AddBrandCategory(ctx context.Context, req *productv1.AddBrandCategoryRequest) (*productv1.AddBrandCategoryResponse, error) {
	if req.BrandId == "" {
		return &productv1.AddBrandCategoryResponse{
			Code:    1,
			Message: "品牌ID不能为空",
		}, nil
	}
	if req.CategoryId == "" {
		return &productv1.AddBrandCategoryResponse{
			Code:    1,
			Message: "类目ID不能为空",
		}, nil
	}
	return h.productService.AddBrandCategory(ctx, req)
}

func (h *ProductServiceHandler) RemoveBrandCategory(ctx context.Context, req *productv1.RemoveBrandCategoryRequest) (*productv1.RemoveBrandCategoryResponse, error) {
	if req.BrandId == "" {
		return &productv1.RemoveBrandCategoryResponse{
			Code:    1,
			Message: "品牌ID不能为空",
		}, nil
	}
	if req.CategoryId == "" {
		return &productv1.RemoveBrandCategoryResponse{
			Code:    1,
			Message: "类目ID不能为空",
		}, nil
	}
	return h.productService.RemoveBrandCategory(ctx, req)
}

func (h *ProductServiceHandler) GetBrandCategories(ctx context.Context, req *productv1.GetBrandCategoriesRequest) (*productv1.GetBrandCategoriesResponse, error) {
	if req.BrandId == "" {
		return &productv1.GetBrandCategoriesResponse{
			Code:    1,
			Message: "品牌ID不能为空",
		}, nil
	}
	return h.productService.GetBrandCategories(ctx, req)
}

func (h *ProductServiceHandler) BatchSetBrandCategories(ctx context.Context, req *productv1.BatchSetBrandCategoriesRequest) (*productv1.BatchSetBrandCategoriesResponse, error) {
	if req.BrandId == "" {
		return &productv1.BatchSetBrandCategoriesResponse{
			Code:    1,
			Message: "品牌ID不能为空",
		}, nil
	}
	if len(req.CategoryIds) == 0 {
		return &productv1.BatchSetBrandCategoriesResponse{
			Code:    1,
			Message: "类目ID不能为空",
		}, nil
	}
	return h.productService.BatchSetBrandCategories(ctx, req)
}

// ============================================
// 商品（SPU）管理接口
// ============================================

func (h *ProductServiceHandler) CreateProduct(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	// 参数校验
	validator := service.NewCreateProductRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.CreateProductResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.CreateProduct(ctx, req)
}

func (h *ProductServiceHandler) GetProduct(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	if req.ProductId == "" {
		return &productv1.GetProductResponse{
			Code:    1,
			Message: "商品ID不能为空",
		}, nil
	}
	return h.productService.GetProduct(ctx, req)
}

func (h *ProductServiceHandler) UpdateProduct(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	// 参数校验
	validator := service.NewUpdateProductRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.UpdateProductResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.UpdateProduct(ctx, req)
}

func (h *ProductServiceHandler) DeleteProduct(ctx context.Context, req *productv1.DeleteProductRequest) (*productv1.DeleteProductResponse, error) {
	if req.ProductId == "" {
		return &productv1.DeleteProductResponse{
			Code:    1,
			Message: "商品ID不能为空",
		}, nil
	}
	return h.productService.DeleteProduct(ctx, req)
}

func (h *ProductServiceHandler) ListProducts(ctx context.Context, req *productv1.ListProductsRequest) (*productv1.ListProductsResponse, error) {
	// 参数校验
	validator := service.NewListProductsRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.ListProductsResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.ListProducts(ctx, req)
}

func (h *ProductServiceHandler) OnShelfProduct(ctx context.Context, req *productv1.OnShelfProductRequest) (*productv1.OnShelfProductResponse, error) {
	if req.ProductId == "" {
		return &productv1.OnShelfProductResponse{
			Code:    1,
			Message: "商品ID不能为空",
		}, nil
	}
	return h.productService.OnShelfProduct(ctx, req)
}

func (h *ProductServiceHandler) OffShelfProduct(ctx context.Context, req *productv1.OffShelfProductRequest) (*productv1.OffShelfProductResponse, error) {
	if req.ProductId == "" {
		return &productv1.OffShelfProductResponse{
			Code:    1,
			Message: "商品ID不能为空",
		}, nil
	}
	return h.productService.OffShelfProduct(ctx, req)
}

func (h *ProductServiceHandler) SubmitProductAudit(ctx context.Context, req *productv1.SubmitProductAuditRequest) (*productv1.SubmitProductAuditResponse, error) {
	if req.ProductId == "" {
		return &productv1.SubmitProductAuditResponse{
			Code:    1,
			Message: "商品ID不能为空",
		}, nil
	}
	return h.productService.SubmitProductAudit(ctx, req)
}

func (h *ProductServiceHandler) AuditProduct(ctx context.Context, req *productv1.AuditProductRequest) (*productv1.AuditProductResponse, error) {
	// 参数校验
	validator := service.NewAuditProductRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &productv1.AuditProductResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.productService.AuditProduct(ctx, req)
}

// ============================================
// 商品标签关联管理接口
// ============================================

func (h *ProductServiceHandler) AddProductTag(ctx context.Context, req *productv1.AddProductTagRequest) (*productv1.AddProductTagResponse, error) {
	return h.productService.AddProductTag(ctx, req)
}

func (h *ProductServiceHandler) RemoveProductTag(ctx context.Context, req *productv1.RemoveProductTagRequest) (*productv1.RemoveProductTagResponse, error) {
	return h.productService.RemoveProductTag(ctx, req)
}

func (h *ProductServiceHandler) GetProductTags(ctx context.Context, req *productv1.GetProductTagsRequest) (*productv1.GetProductTagsResponse, error) {
	return h.productService.GetProductTags(ctx, req)
}

func (h *ProductServiceHandler) BatchSetProductTags(ctx context.Context, req *productv1.BatchSetProductTagsRequest) (*productv1.BatchSetProductTagsResponse, error) {
	return h.productService.BatchSetProductTags(ctx, req)
}

// ============================================
// SKU管理接口
// ============================================

func (h *ProductServiceHandler) CreateSku(ctx context.Context, req *productv1.CreateSkuRequest) (*productv1.CreateSkuResponse, error) {
	return h.productService.CreateSku(ctx, req)
}

func (h *ProductServiceHandler) GetSku(ctx context.Context, req *productv1.GetSkuRequest) (*productv1.GetSkuResponse, error) {
	return h.productService.GetSku(ctx, req)
}

func (h *ProductServiceHandler) UpdateSku(ctx context.Context, req *productv1.UpdateSkuRequest) (*productv1.UpdateSkuResponse, error) {
	return h.productService.UpdateSku(ctx, req)
}

func (h *ProductServiceHandler) DeleteSku(ctx context.Context, req *productv1.DeleteSkuRequest) (*productv1.DeleteSkuResponse, error) {
	return h.productService.DeleteSku(ctx, req)
}

func (h *ProductServiceHandler) ListSkus(ctx context.Context, req *productv1.ListSkusRequest) (*productv1.ListSkusResponse, error) {
	return h.productService.ListSkus(ctx, req)
}

func (h *ProductServiceHandler) BatchCreateSkus(ctx context.Context, req *productv1.BatchCreateSkusRequest) (*productv1.BatchCreateSkusResponse, error) {
	return h.productService.BatchCreateSkus(ctx, req)
}

// ============================================
// SKU属性关联管理接口
// ============================================

func (h *ProductServiceHandler) AddSkuAttribute(ctx context.Context, req *productv1.AddSkuAttributeRequest) (*productv1.AddSkuAttributeResponse, error) {
	return h.productService.AddSkuAttribute(ctx, req)
}

func (h *ProductServiceHandler) RemoveSkuAttribute(ctx context.Context, req *productv1.RemoveSkuAttributeRequest) (*productv1.RemoveSkuAttributeResponse, error) {
	return h.productService.RemoveSkuAttribute(ctx, req)
}

func (h *ProductServiceHandler) GetSkuAttributes(ctx context.Context, req *productv1.GetSkuAttributesRequest) (*productv1.GetSkuAttributesResponse, error) {
	return h.productService.GetSkuAttributes(ctx, req)
}

func (h *ProductServiceHandler) BatchSetSkuAttributes(ctx context.Context, req *productv1.BatchSetSkuAttributesRequest) (*productv1.BatchSetSkuAttributesResponse, error) {
	return h.productService.BatchSetSkuAttributes(ctx, req)
}

// ============================================
// 标签管理接口
// ============================================

func (h *ProductServiceHandler) CreateTag(ctx context.Context, req *productv1.CreateTagRequest) (*productv1.CreateTagResponse, error) {
	return h.productService.CreateTag(ctx, req)
}

func (h *ProductServiceHandler) GetTag(ctx context.Context, req *productv1.GetTagRequest) (*productv1.GetTagResponse, error) {
	return h.productService.GetTag(ctx, req)
}

func (h *ProductServiceHandler) UpdateTag(ctx context.Context, req *productv1.UpdateTagRequest) (*productv1.UpdateTagResponse, error) {
	return h.productService.UpdateTag(ctx, req)
}

func (h *ProductServiceHandler) DeleteTag(ctx context.Context, req *productv1.DeleteTagRequest) (*productv1.DeleteTagResponse, error) {
	return h.productService.DeleteTag(ctx, req)
}

func (h *ProductServiceHandler) ListTags(ctx context.Context, req *productv1.ListTagsRequest) (*productv1.ListTagsResponse, error) {
	return h.productService.ListTags(ctx, req)
}
