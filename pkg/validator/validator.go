package validator

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func Init() {
	v = validator.New()

	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("sms_code", validateSMSCode)

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("label")
		if name == "" {
			return fld.Name
		}
		return name
	})
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return match
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 || len(password) > 20 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasDigit
}

// validateSMSCode 校验短信验证码（6位数字）
func validateSMSCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	matched, _ := regexp.MatchString(`^\d{6}$`, code)
	return matched
}

// ValidateStruct 校验结构体
func ValidateStruct(s interface{}) error {
	if v == nil {
		Init() // 如果未初始化，自动初始化
	}
	return v.Struct(s)
}

// FormatError 格式化错误信息为中文
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	for _, e := range validationErrors {
		return getErrorMessage(e)
	}
	return "参数校验失败"
}

// getErrorMessage 根据校验规则返回中文错误信息
func getErrorMessage(e validator.FieldError) string {
	fieldName := e.Field() // 使用 label 标签的值

	switch e.Tag() {
	case "required":
		return fieldName + "不能为空"
	case "phone":
		return fieldName + "格式错误，请输入11位手机号"
	case "password":
		return fieldName + "必须包含字母和数字，长度6-20位"
	case "sms_code":
		return fieldName + "必须为6位数字"
	case "min":
		return fieldName + "长度不能小于" + e.Param() + "位"
	case "max":
		return fieldName + "长度不能大于" + e.Param() + "位"
	case "len":
		return fieldName + "长度必须为" + e.Param() + "位"
	case "eqfield":
		return fieldName + "必须与" + e.Param() + "一致"
	case "email":
		return fieldName + "格式错误"
	default:
		return fieldName + "校验失败"
	}
}
