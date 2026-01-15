package service

import (
	"errors"
	promotionv1 "zjMall/gen/go/api/proto/promotion"
	"zjMall/pkg/validator"
)

type CreatePromotionRequestValidator struct {
	Name           string `validate:"required,min=2,max=100" label:"促销名称"`
	Type           int32  `validate:"required,oneof=1 2 3 4" label:"促销类型"`
	ConditionValue string `validate:"required" label:"条件值"`
	DiscountValue  string `validate:"required" label:"优惠值"`
}

func NewCreatePromotionRequestValidator(req *promotionv1.CreatePromotionRequest) *CreatePromotionRequestValidator {
	return &CreatePromotionRequestValidator{
		Name:           req.Name,
		Type:           int32(req.Type),
		ConditionValue: req.ConditionValue,
		DiscountValue:  req.DiscountValue,
	}
}

func (v *CreatePromotionRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}
