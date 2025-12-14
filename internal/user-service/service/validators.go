package service

import (
	"errors"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/pkg/validator"
)

// RegisterRequestValidator 注册请求校验器
type RegisterRequestValidator struct {
	Phone           string `validate:"required,phone" label:"手机号"`
	SMSCode         string `validate:"required,sms_code" label:"短信验证码"`
	Password        string `validate:"required,password" label:"密码"`
	ConfirmPassword string `validate:"required,eqfield=Password" label:"确认密码"`
}

// NewRegisterRequestValidator 从 protobuf 请求创建校验器
func NewRegisterRequestValidator(req *userv1.RegisterRequest) *RegisterRequestValidator {
	return &RegisterRequestValidator{
		Phone:           req.Phone,
		SMSCode:         req.SmsCode,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}
}

// Validate 执行校验
func (r *RegisterRequestValidator) Validate() error {
	if err := validator.ValidateStruct(r); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

// LoginRequestValidator 登录请求校验器
type LoginRequestValidator struct {
	Phone    string `validate:"required,phone" label:"手机号"`
	Password string `validate:"required,password" label:"密码"`
}

// NewLoginRequestValidator 从 protobuf 请求创建校验器
func NewLoginRequestValidator(req *userv1.LoginRequest) *LoginRequestValidator {
	return &LoginRequestValidator{
		Phone:    req.Phone,
		Password: req.Password,
	}
}

// Validate 执行校验
func (l *LoginRequestValidator) Validate() error {
	if err := validator.ValidateStruct(l); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type LoginBySMSRequestValidator struct {
	Phone   string `validate:"required,phone" label:"手机号"`
	SmsCode string `validate:"required,sms_code" label:"短信验证码"`
}

func NewLoginBySMSRequestValidator(req *userv1.LoginBySMSRequest) *LoginBySMSRequestValidator {
	return &LoginBySMSRequestValidator{
		Phone:   req.Phone,
		SmsCode: req.SmsCode,
	}
}
func (l *LoginBySMSRequestValidator) Validate() error {
	if err := validator.ValidateStruct(l); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type UpdateUserRequestValidator struct {
	Email    string `validate:"required,email" label:"邮箱"`
	Gender   int32  `validate:"required,oneof=0 1 2" label:"性别"`
	Birthday string `validate:"required,date" label:"生日"`
}

func NewUpdateUserRequestValidator(req *userv1.UpdateUserRequest) *UpdateUserRequestValidator {
	return &UpdateUserRequestValidator{
		Email:    req.Email,
		Gender:   req.Gender,
		Birthday: req.Birthday,
	}
}
func (u *UpdateUserRequestValidator) Validate() error {
	if err := validator.ValidateStruct(u); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type ChangePasswordRequestValidator struct {
	OldPassword     string `validate:"required,password" label:"旧密码"`
	NewPassword     string `validate:"required,password" label:"新密码"`
	ConfirmPassword string `validate:"required,eqfield=NewPassword" label:"确认新密码"`
}

func NewChangePasswordRequestValidator(req *userv1.ChangePasswordRequest) *ChangePasswordRequestValidator {
	return &ChangePasswordRequestValidator{
		OldPassword:     req.OldPassword,
		NewPassword:     req.NewPassword,
		ConfirmPassword: req.ConfirmPassword,
	}
}
func (c *ChangePasswordRequestValidator) Validate() error {
	if err := validator.ValidateStruct(c); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}

type BindPhoneRequestValidator struct {
	NewPhone   string `validate:"required,phone" label:"新手机号"`
	NewSmsCode string `validate:"required,sms_code" label:"新短信验证码"`
}

func NewBindPhoneRequestValidator(req *userv1.BindPhoneRequest) *BindPhoneRequestValidator {
	return &BindPhoneRequestValidator{
		NewPhone:   req.NewPhone,
		NewSmsCode: req.NewSmsCode,
	}
}
func (b *BindPhoneRequestValidator) Validate() error {
	if err := validator.ValidateStruct(b); err != nil {
		return errors.New(validator.FormatError(err))
	}
	return nil
}
