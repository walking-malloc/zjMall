package service

import (
	"errors"
	"zjMall/pkg/validator"

	paymentv1 "zjMall/gen/go/api/proto/payment"
)

// CreatePaymentRequestValidator 创建支付单请求校验器
type CreatePaymentRequestValidator struct {
	OrderNo    string `validate:"required" label:"订单号"`
	PayChannel string `validate:"required,oneof=wechat alipay balance" label:"支付渠道"` // wechat-微信, alipay-支付宝, balance-余额
	ReturnURL  string `validate:"omitempty,url" label:"返回地址"`
	Token      string `validate:"required" label:"幂等性Token"` // Token 可选，但建议使用
}

// NewCreatePaymentRequestValidator 创建支付单请求校验器
func NewCreatePaymentRequestValidator(req *paymentv1.CreatePaymentRequest) *CreatePaymentRequestValidator {
	return &CreatePaymentRequestValidator{
		OrderNo:    req.OrderNo,
		PayChannel: req.PayChannel,
		ReturnURL:  req.ReturnUrl,
		Token:      req.Token,
	}
}

// Validate 校验请求参数
func (v *CreatePaymentRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// PaymentCallbackRequestValidator 支付回调请求校验器
type PaymentCallbackRequestValidator struct {
	PayChannel  string            `validate:"required,oneof=wechat alipay balance" label:"支付渠道"`
	PaymentNo   string            `validate:"required" label:"支付单号"`
	TradeNo     string            `validate:"required" label:"第三方交易号"`
	Amount      string            `validate:"required" label:"支付金额"`
	Status      string            `validate:"required" label:"支付状态"`
	Sign        string            `validate:"required" label:"签名"`
	ExtraParams map[string]string `validate:"-" label:"扩展参数"`
}

// NewPaymentCallbackRequestValidator 创建支付回调请求校验器
func NewPaymentCallbackRequestValidator(req *paymentv1.PaymentCallbackRequest) *PaymentCallbackRequestValidator {
	return &PaymentCallbackRequestValidator{
		PayChannel:  req.PayChannel,
		PaymentNo:   req.PaymentNo,
		TradeNo:     req.TradeNo,
		Amount:      req.Amount,
		Status:      req.Status,
		Sign:        req.Sign,
		ExtraParams: req.ExtraParams,
	}
}

// Validate 校验请求参数
func (v *PaymentCallbackRequestValidator) Validate() error {
	if err := validator.ValidateStruct(v); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}
