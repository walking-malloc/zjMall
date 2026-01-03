package service

import (
	"errors"
	productv1 "zjMall/gen/go/api/proto/product"
	"zjMall/pkg/validator"
)

type CreateCategoryRequestValidator struct {
	ParentID  string `validate:"required,uuid" label:"父类目ID"`
	Name      string `validate:"required,min=2,max=100" label:"类目名称"`
	Level     int32  `validate:"required,oneof=1 2 3" label:"类目层级"`
	IsLeaf    bool   `validate:"required" label:"是否为叶子节点"`
	IsVisible bool   `validate:"required" label:"是否在前台展示"`
	SortOrder int32  `validate:"required,min=0" label:"排序权重"`
	Icon      string `validate:"omitempty,url" label:"类目图标URL"`
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
