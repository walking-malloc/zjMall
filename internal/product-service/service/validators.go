package service

import (
	"errors"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/pkg/validator"
)

// ==============类目管理验证器==============
type CreateCategoryRequestValidator struct {
	ParentID  string `validate:"omitempty" label:"父类目ID"`
	Name      string `validate:"required,min=2,max=100" label:"类目名称"`
	Level     int32  `validate:"required" label:"类目层级"`
	IsLeaf    bool   `validate:"-" label:"是否为叶子节点"` // 不验证，因为 false 是有效值
	IsVisible bool   `validate:"-" label:"是否在前台展示"` // 不验证，因为 false 是有效值
	SortOrder int32  `validate:"required" label:"排序权重"`
	Icon      string `validate:"omitempty,url" label:"类目图标URL"` // 修复：去掉空格
	Status    int32  `validate:"required,oneof=1 2" label:"状态"`
}

func NewCreateCategoryRequestValidator(req *productv1.CreateCategoryRequest) *CreateCategoryRequestValidator {
	return &CreateCategoryRequestValidator{
		ParentID:  req.ParentId,
		Name:      req.Name,
		Level:     req.Level,
		IsLeaf:    req.IsLeaf,
		IsVisible: req.IsVisible,
		SortOrder: req.SortOrder,
		Icon:      req.Icon,
		Status:    req.Status,
	}
}

func (v *CreateCategoryRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateCategoryRequestValidator struct {
	CategoryID string `validate:"required" label:"类目ID"`
	Name       string `validate:"omitempty,min=2,max=100" label:"类目名称"`
	IsLeaf     bool   `validate:"-" label:"是否为叶子节点"`
	IsVisible  bool   `validate:"-" label:"是否在前台展示"`
	SortOrder  int32  `validate:"-" label:"排序权重"`
	Icon       string `validate:"omitempty,url" label:"类目图标URL"`
	Status     int32  `validate:"omitempty,oneof=1 2" label:"状态"`
}

func NewUpdateCategoryRequestValidator(req *productv1.UpdateCategoryRequest) *UpdateCategoryRequestValidator {
	return &UpdateCategoryRequestValidator{
		CategoryID: req.CategoryId,
		Name:       req.Name,
		IsLeaf:     req.IsLeaf,
		IsVisible:  req.IsVisible,
		SortOrder:  req.SortOrder,
		Icon:       req.Icon,
		Status:     req.Status,
	}
}

func (v *UpdateCategoryRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListCategoriesRequestValidator struct {
	Page      int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize  int32  `validate:"omitempty,min=1" label:"每页条数"`
	Level     int32  `validate:"omitempty,min=1" label:"类目层级"`
	Status    int32  `validate:"required,oneof=1 2" label:"状态"`
	IsVisible bool   `validate:"-" label:"是否在前台展示"`
	Keyword   string `validate:"omitempty,min=2,max=100" label:"关键词"`
}

func NewListCategoriesRequestValidator(req *productv1.ListCategoriesRequest) *ListCategoriesRequestValidator {
	return &ListCategoriesRequestValidator{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Level:     req.Level,
		Status:    req.Status,
		IsVisible: req.IsVisible,
		Keyword:   req.Keyword,
	}
}

func (v *ListCategoriesRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type GetCategoryChildrenRequestValidator struct {
	ParentID  string `validate:"-" label:"父类目ID"`
	Status    int32  `validate:"required,oneof=1 2" label:"状态"`
	IsVisible bool   `validate:"-" label:"是否在前台展示"`
	Page      int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize  int32  `validate:"omitempty,min=1" label:"每页条数"`
}

func NewGetCategoryChildrenRequestValidator(req *productv1.GetCategoryChildrenRequest) *GetCategoryChildrenRequestValidator {
	return &GetCategoryChildrenRequestValidator{
		ParentID:  req.ParentId,
		Status:    req.Status,
		IsVisible: req.OnlyVisible,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
}
func (v *GetCategoryChildrenRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// ==============品牌管理验证器==============
type CreateBrandRequestValidator struct {
	Name        string `validate:"required,min=2,max=100" label:"品牌名称"`
	LogoURL     string `validate:"omitempty,url" label:"品牌LogoURL"`
	Country     string `validate:"omitempty,min=2,max=100" label:"国家"`
	Description string `validate:"omitempty,min=2,max=1000" label:"品牌描述"`
	FirstLetter string `validate:"required,min=1,max=1" label:"首字母"`
	SortOrder   int32  `validate:"required" label:"排序权重"`
	Status      int32  `validate:"required,oneof=1 2" label:"状态"`
}

func NewCreateBrandRequestValidator(req *productv1.CreateBrandRequest) *CreateBrandRequestValidator {
	return &CreateBrandRequestValidator{
		Name:        req.Name,
		LogoURL:     req.LogoUrl,
		Country:     req.Country,
		Description: req.Description,
		FirstLetter: req.FirstLetter,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}
}
func (v *CreateBrandRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateBrandRequestValidator struct {
	BrandID     string `validate:"required" label:"品牌ID"`
	Name        string `validate:"omitempty,min=2,max=100" label:"品牌名称"`
	LogoURL     string `validate:"omitempty,url" label:"品牌LogoURL"`
	Country     string `validate:"omitempty,min=2,max=100" label:"国家"`
	Description string `validate:"omitempty,min=2,max=1000" label:"品牌描述"`
	FirstLetter string `validate:"omitempty,min=1,max=1" label:"首字母"`
	SortOrder   int32  `validate:"-" label:"排序权重"`
	Status      int32  `validate:"omitempty,oneof=1 2" label:"状态"`
}

func NewUpdateBrandRequestValidator(req *productv1.UpdateBrandRequest) *UpdateBrandRequestValidator {
	return &UpdateBrandRequestValidator{
		BrandID:     req.BrandId,
		Name:        req.Name,
		LogoURL:     req.LogoUrl,
		Country:     req.Country,
		Description: req.Description,
		FirstLetter: req.FirstLetter,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}
}
func (v *UpdateBrandRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListBrandsRequestValidator struct {
	Page        int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize    int32  `validate:"omitempty,min=1" label:"每页条数"`
	Status      int32  `validate:"required,oneof=1 2" label:"状态"`
	Keyword     string `validate:"omitempty,min=2,max=100" label:"关键词"`
	FirstLetter string `validate:"omitempty,min=1,max=1" label:"首字母"`
	Country     string `validate:"omitempty,min=2,max=100" label:"国家"`
}

func NewListBrandsRequestValidator(req *productv1.ListBrandsRequest) *ListBrandsRequestValidator {
	return &ListBrandsRequestValidator{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Status:      req.Status,
		Keyword:     req.Keyword,
		FirstLetter: req.FirstLetter,
		Country:     req.Country,
	}
}
func (v *ListBrandsRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// ==============商品管理验证器==============
type CreateProductRequestValidator struct {
	CategoryID  string `validate:"required" label:"所属类目ID"`
	BrandID     string `validate:"omitempty" label:"品牌ID"`
	Title       string `validate:"required,min=1,max=200" label:"商品标题"`
	Subtitle    string `validate:"omitempty,max=200" label:"商品副标题"`
	MainImage   string `validate:"required,url" label:"主图URL"`
	Description string `validate:"omitempty" label:"商品详情"`
	Status      int32  `validate:"omitempty,oneof=1 2" label:"状态"`
	// OnShelfTime 在 proto 中是可选字段，不需要在 validator 中校验
}

func NewCreateProductRequestValidator(req *productv1.CreateProductRequest) *CreateProductRequestValidator {
	return &CreateProductRequestValidator{
		CategoryID:  req.CategoryId,
		BrandID:     req.BrandId,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		MainImage:   req.MainImage,
		Description: req.Description,
		Status:      req.Status,
	}
}

func (v *CreateProductRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateProductRequestValidator struct {
	ProductID   string `validate:"required" label:"商品ID"`
	CategoryID  string `validate:"omitempty" label:"所属类目ID"`
	BrandID     string `validate:"omitempty" label:"品牌ID"`
	Title       string `validate:"omitempty,min=1,max=200" label:"商品标题"`
	Subtitle    string `validate:"omitempty,max=200" label:"商品副标题"`
	MainImage   string `validate:"omitempty,url" label:"主图URL"`
	Description string `validate:"omitempty" label:"商品详情"`
	Status      int32  `validate:"omitempty,oneof=1 2 3 4 5" label:"状态"`
}

func NewUpdateProductRequestValidator(req *productv1.UpdateProductRequest) *UpdateProductRequestValidator {
	return &UpdateProductRequestValidator{
		ProductID:   req.ProductId,
		CategoryID:  req.CategoryId,
		BrandID:     req.BrandId,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		MainImage:   req.MainImage,
		Description: req.Description,
		Status:      req.Status,
	}
}

func (v *UpdateProductRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListProductsRequestValidator struct {
	Page       int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize   int32  `validate:"omitempty,min=1,max=100" label:"每页数量"`
	CategoryID string `validate:"omitempty" label:"类目ID"`
	BrandID    string `validate:"omitempty" label:"品牌ID"`
	Status     int32  `validate:"omitempty,oneof=1 2 3 4 5" label:"状态"`
	Keyword    string `validate:"omitempty,min=1,max=200" label:"关键词"`
	SortBy     string `validate:"omitempty,oneof=created_at on_shelf_time" label:"排序字段"`
	SortOrder  string `validate:"omitempty,oneof=asc desc" label:"排序方向"`
}

func NewListProductsRequestValidator(req *productv1.ListProductsRequest) *ListProductsRequestValidator {
	return &ListProductsRequestValidator{
		Page:       req.Page,
		PageSize:   req.PageSize,
		CategoryID: req.CategoryId,
		BrandID:    req.BrandId,
		Status:     req.Status,
		Keyword:    req.Keyword,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}
}

func (v *ListProductsRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type AuditProductRequestValidator struct {
	ProductID    string `validate:"required" label:"商品ID"`
	Result       int32  `validate:"required,oneof=1 2" label:"审核结果"`
	Reason       string `validate:"omitempty,max=500" label:"审核原因"`
	OperatorID   string `validate:"omitempty" label:"操作人ID"`
	OperatorName string `validate:"omitempty,max=50" label:"操作人姓名"`
}

func NewAuditProductRequestValidator(req *productv1.AuditProductRequest) *AuditProductRequestValidator {
	return &AuditProductRequestValidator{
		ProductID:    req.ProductId,
		Result:       req.Result,
		Reason:       req.Reason,
		OperatorID:   req.OperatorId,
		OperatorName: req.OperatorName,
	}
}

func (v *AuditProductRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}
