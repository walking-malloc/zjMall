package service

import (
	"errors"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/pkg/validator"
)

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
	Name       string `validate:"required,min=2,max=100" label:"类目名称"`
	Level      int32  `validate:"required,oneof=1 2 3" label:"类目层级"`
	IsLeaf     bool   `validate:"-" label:"是否为叶子节点"`
	IsVisible  bool   `validate:"-" label:"是否在前台展示"`
	SortOrder  int32  `validate:"required" label:"排序权重"`
	Icon       string `validate:"omitempty,url" label:"类目图标URL"`
	Status     int32  `validate:"required,oneof=1 2" label:"状态"`
}

func NewUpdateCategoryRequestValidator(req *productv1.UpdateCategoryRequest) *UpdateCategoryRequestValidator {
	return &UpdateCategoryRequestValidator{
		CategoryID: req.CategoryId,
		Name:       req.Name,
		Level:      req.Level,
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
	ParentID  string `validate:"omitempty" label:"父类目ID"`
	Level     int32  `validate:"required,oneof=1 2 3" label:"类目层级"`
	Status    int32  `validate:"required,oneof=1 2" label:"状态"`
	IsVisible bool   `validate:"-" label:"是否在前台展示"`
	Keyword   string `validate:"omitempty,min=2,max=100" label:"关键词"`
}

func NewListCategoriesRequestValidator(req *productv1.ListCategoriesRequest) *ListCategoriesRequestValidator {
	return &ListCategoriesRequestValidator{
		Page:      req.Page,
		PageSize:  req.PageSize,
		ParentID:  req.ParentId,
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
