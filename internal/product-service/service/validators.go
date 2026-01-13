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

// ==============标签管理验证器==============

type CreateTagRequestValidator struct {
	Name      string `validate:"required,min=1,max=50" label:"标签名称"`
	Type      int32  `validate:"omitempty,oneof=1 2" label:"标签类型"`
	Color     string `validate:"omitempty" label:"标签颜色"`
	SortOrder int32  `validate:"omitempty" label:"排序权重"`
	Status    int32  `validate:"omitempty,oneof=1 2" label:"状态"`
}

func NewCreateTagRequestValidator(req *productv1.CreateTagRequest) *CreateTagRequestValidator {
	return &CreateTagRequestValidator{
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		Status:    req.Status,
	}
}

func (v *CreateTagRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateTagRequestValidator struct {
	TagID     string `validate:"required" label:"标签ID"`
	Name      string `validate:"omitempty,min=1,max=50" label:"标签名称"`
	Type      int32  `validate:"omitempty,oneof=1 2" label:"标签类型"`
	Color     string `validate:"omitempty" label:"标签颜色"`
	SortOrder int32  `validate:"omitempty" label:"排序权重"`
	Status    int32  `validate:"omitempty,oneof=1 2" label:"状态"`
}

func NewUpdateTagRequestValidator(req *productv1.UpdateTagRequest) *UpdateTagRequestValidator {
	return &UpdateTagRequestValidator{
		TagID:     req.TagId,
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		Status:    req.Status,
	}
}

func (v *UpdateTagRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListTagsRequestValidator struct {
	Page     int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize int32  `validate:"omitempty,min=1,max=100" label:"每页数量"`
	Type     int32  `validate:"omitempty,oneof=1 2" label:"标签类型"`
	Status   int32  `validate:"omitempty,oneof=1 2" label:"状态"`
	Keyword  string `validate:"omitempty,min=1,max=50" label:"关键词"`
}

func NewListTagsRequestValidator(req *productv1.ListTagsRequest) *ListTagsRequestValidator {
	return &ListTagsRequestValidator{
		Page:     req.Page,
		PageSize: req.PageSize,
		Type:     req.Type,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
}

func (v *ListTagsRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// ==============SKU 管理验证器==============

type CreateSkuRequestValidator struct {
	ProductID     string  `validate:"required" label:"所属商品ID"`
	SkuCode       string  `validate:"omitempty,min=1,max=50" label:"SKU编码"`
	Barcode       string  `validate:"omitempty,min=1,max=50" label:"条形码"`
	Name          string  `validate:"required,min=1,max=200" label:"SKU名称"`
	Image         string  `validate:"omitempty,url" label:"SKU图片"`
	Weight        float64 `validate:"omitempty,gt=0" label:"重量"`
	Volume        float64 `validate:"omitempty,gt=0" label:"体积"`
	Price         float64 `validate:"required,gt=0" label:"销售价格"`
	CostPrice     float64 `validate:"omitempty,gt=0" label:"成本价格"`
	OriginalPrice float64 `validate:"omitempty,gt=0" label:"原价"`
	Status        int32   `validate:"omitempty,oneof=1 2 3" label:"状态"`
}

func NewCreateSkuRequestValidator(req *productv1.CreateSkuRequest) *CreateSkuRequestValidator {
	return &CreateSkuRequestValidator{
		ProductID:     req.ProductId,
		Price:         req.Price,
		Status:        req.Status,
		SkuCode:       req.SkuCode,
		Barcode:       req.Barcode,
		Name:          req.Name,
		Image:         req.Image,
		Weight:        req.Weight,
		Volume:        req.Volume,
		CostPrice:     req.CostPrice,
		OriginalPrice: req.OriginalPrice,
	}
}

func (v *CreateSkuRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateSkuRequestValidator struct {
	SkuID         string  `validate:"required" label:"SKU ID"`
	SkuCode       string  `validate:"omitempty,min=1,max=50" label:"SKU编码"`
	Barcode       string  `validate:"omitempty,min=1,max=50" label:"条形码"`
	Name          string  `validate:"omitempty,min=1,max=200" label:"SKU名称"`
	Image         string  `validate:"omitempty,url" label:"SKU图片"`
	Weight        float64 `validate:"omitempty,gt=0" label:"重量"`
	Volume        float64 `validate:"omitempty,gt=0" label:"体积"`
	CostPrice     float64 `validate:"omitempty,gt=0" label:"成本价格"`
	OriginalPrice float64 `validate:"omitempty,gt=0" label:"原价"`
	Price         float64 `validate:"omitempty,gt=0" label:"销售价格"`
	Status        int32   `validate:"omitempty,oneof=1 2 3" label:"状态"`
}

func NewUpdateSkuRequestValidator(req *productv1.UpdateSkuRequest) *UpdateSkuRequestValidator {
	return &UpdateSkuRequestValidator{
		SkuID:  req.SkuId,
		Price:  req.Price,
		Status: req.Status,
	}
}

func (v *UpdateSkuRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListSkusRequestValidator struct {
	Page      int32   `validate:"omitempty,min=1" label:"页码"`
	PageSize  int32   `validate:"omitempty,min=1,max=100" label:"每页数量"`
	Status    int32   `validate:"omitempty,oneof=1 2 3" label:"状态"`
	MinPrice  float64 `validate:"omitempty,gte=0" label:"最低价格"`
	MaxPrice  float64 `validate:"omitempty,gte=0" label:"最高价格"`
	Keyword   string  `validate:"omitempty,min=1,max=200" label:"关键词"`
	ProductID string  `validate:"omitempty" label:"商品ID"`
}

func NewListSkusRequestValidator(req *productv1.ListSkusRequest) *ListSkusRequestValidator {
	return &ListSkusRequestValidator{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Status:    req.Status,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		Keyword:   req.Keyword,
		ProductID: req.ProductId,
	}
}

func (v *ListSkusRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	// 额外检查：如果同时传了 min_price 和 max_price，要求 min <= max
	if v.MinPrice > 0 && v.MaxPrice > 0 && v.MinPrice > v.MaxPrice {
		return errors.New("最低价格不能大于最高价格")
	}
	return nil
}

type BatchCreateSkusRequestValidator struct {
	ProductID string `validate:"required" label:"所属商品ID"`
	// skus 数组本身在 service 中进一步校验（如长度、每个元素的必填字段）
}

func NewBatchCreateSkusRequestValidator(req *productv1.BatchCreateSkusRequest) *BatchCreateSkusRequestValidator {
	return &BatchCreateSkusRequestValidator{
		ProductID: req.ProductId,
	}
}

func (v *BatchCreateSkusRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// ==============SKU 属性关联验证器==============

type AddSkuAttributeRequestValidator struct {
	SkuID            string `validate:"required" label:"SKU ID"`
	AttributeValueID string `validate:"required" label:"属性值ID"`
}

func NewAddSkuAttributeRequestValidator(req *productv1.AddSkuAttributeRequest) *AddSkuAttributeRequestValidator {
	return &AddSkuAttributeRequestValidator{
		SkuID:            req.SkuId,
		AttributeValueID: req.AttributeValueId,
	}
}

func (v *AddSkuAttributeRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type RemoveSkuAttributeRequestValidator struct {
	SkuID            string `validate:"required" label:"SKU ID"`
	AttributeValueID string `validate:"required" label:"属性值ID"`
}

func NewRemoveSkuAttributeRequestValidator(req *productv1.RemoveSkuAttributeRequest) *RemoveSkuAttributeRequestValidator {
	return &RemoveSkuAttributeRequestValidator{
		SkuID:            req.SkuId,
		AttributeValueID: req.AttributeValueId,
	}
}

func (v *RemoveSkuAttributeRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type BatchSetSkuAttributesRequestValidator struct {
	SkuID             string   `validate:"required" label:"SKU ID"`
	AttributeValueIDs []string `validate:"required,min=1,dive,required" label:"属性值ID列表"`
}

func NewBatchSetSkuAttributesRequestValidator(req *productv1.BatchSetSkuAttributesRequest) *BatchSetSkuAttributesRequestValidator {
	return &BatchSetSkuAttributesRequestValidator{
		SkuID:             req.SkuId,
		AttributeValueIDs: req.AttributeValueIds,
	}
}

func (v *BatchSetSkuAttributesRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// ==============属性管理验证器==============

type CreateAttributeRequestValidator struct {
	CategoryID string `validate:"required" label:"所属类目ID"`
	Name       string `validate:"required,min=1,max=100" label:"属性名称"`
	Type       int32  `validate:"required,oneof=1 2" label:"属性类型"`
	InputType  int32  `validate:"required,oneof=1 2 3 4" label:"录入方式"`
	IsRequired int32  `validate:"omitempty,oneof=0 1" label:"是否必填"`
	SortOrder  int32  `validate:"omitempty" label:"排序权重"`
}

func NewCreateAttributeRequestValidator(req *productv1.CreateAttributeRequest) *CreateAttributeRequestValidator {
	return &CreateAttributeRequestValidator{
		CategoryID: req.CategoryId,
		Name:       req.Name,
		Type:       req.Type,
		InputType:  req.InputType,
		IsRequired: req.IsRequired,
		SortOrder:  req.SortOrder,
	}
}

func (v *CreateAttributeRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateAttributeRequestValidator struct {
	AttributeID string `validate:"required" label:"属性ID"`
	Name        string `validate:"omitempty,min=1,max=100" label:"属性名称"`
	Type        int32  `validate:"omitempty,oneof=1 2" label:"属性类型"`
	InputType   int32  `validate:"omitempty,oneof=1 2 3 4" label:"录入方式"`
	IsRequired  int32  `validate:"omitempty,oneof=0 1" label:"是否必填"`
	SortOrder   int32  `validate:"omitempty" label:"排序权重"`
}

func NewUpdateAttributeRequestValidator(req *productv1.UpdateAttributeRequest) *UpdateAttributeRequestValidator {
	return &UpdateAttributeRequestValidator{
		AttributeID: req.AttributeId,
		Name:        req.Name,
		Type:        req.Type,
		InputType:   req.InputType,
		IsRequired:  req.IsRequired,
		SortOrder:   req.SortOrder,
	}
}

func (v *UpdateAttributeRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListAttributesRequestValidator struct {
	Page       int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize   int32  `validate:"omitempty,min=1,max=100" label:"每页数量"`
	CategoryID string `validate:"omitempty" label:"类目ID"`
	Type       int32  `validate:"omitempty,oneof=1 2" label:"属性类型"`
	IsRequired int32  `validate:"omitempty,oneof=0 1" label:"是否必填"`
	Keyword    string `validate:"omitempty,min=1,max=100" label:"关键词"`
}

func NewListAttributesRequestValidator(req *productv1.ListAttributesRequest) *ListAttributesRequestValidator {
	return &ListAttributesRequestValidator{
		Page:       req.Page,
		PageSize:   req.PageSize,
		CategoryID: req.CategoryId,
		Type:       req.Type,
		IsRequired: req.IsRequired,
		Keyword:    req.Keyword,
	}
}

func (v *ListAttributesRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// ==============属性值管理验证器==============

type CreateAttributeValueRequestValidator struct {
	AttributeID string `validate:"required" label:"所属属性ID"`
	Value       string `validate:"required,min=1,max=100" label:"属性值名称"`
	SortOrder   int32  `validate:"omitempty" label:"排序权重"`
}

func NewCreateAttributeValueRequestValidator(req *productv1.CreateAttributeValueRequest) *CreateAttributeValueRequestValidator {
	return &CreateAttributeValueRequestValidator{
		AttributeID: req.AttributeId,
		Value:       req.Value,
		SortOrder:   req.SortOrder,
	}
}

func (v *CreateAttributeValueRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateAttributeValueRequestValidator struct {
	AttributeValueID string `validate:"required" label:"属性值ID"`
	Value            string `validate:"omitempty,min=1,max=100" label:"属性值名称"`
	SortOrder        int32  `validate:"omitempty" label:"排序权重"`
}

func NewUpdateAttributeValueRequestValidator(req *productv1.UpdateAttributeValueRequest) *UpdateAttributeValueRequestValidator {
	return &UpdateAttributeValueRequestValidator{
		AttributeValueID: req.AttributeValueId,
		Value:            req.Value,
		SortOrder:        req.SortOrder,
	}
}

func (v *UpdateAttributeValueRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ListAttributeValuesRequestValidator struct {
	Page        int32  `validate:"omitempty,min=1" label:"页码"`
	PageSize    int32  `validate:"omitempty,min=1,max=100" label:"每页数量"`
	AttributeID string `validate:"omitempty" label:"属性ID"`
	Keyword     string `validate:"omitempty,min=1,max=100" label:"关键词"`
}

func NewListAttributeValuesRequestValidator(req *productv1.ListAttributeValuesRequest) *ListAttributeValuesRequestValidator {
	return &ListAttributeValuesRequestValidator{
		Page:        req.Page,
		PageSize:    req.PageSize,
		AttributeID: req.AttributeId,
		Keyword:     req.Keyword,
	}
}

func (v *ListAttributeValuesRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}
