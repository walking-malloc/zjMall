package validator

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func Init() {
	v = validator.New()

	v.RegisterValidation("phone", validatePhone) //注册后会自动获取标签，然后用对应的方法校验
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("sms_code", validateSMSCode)
	v.RegisterValidation("email", validateEmail)

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

// ValidateSMSCode 校验短信验证码（6位数字）
func validateSMSCode(fl validator.FieldLevel) bool {
	code := fl.Field().String()
	matched, _ := regexp.MatchString(`^\d{6}$`, code)
	return matched
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

// ========== 单个字段校验函数（接受 string 参数）==========

// IsValidPhone 校验手机号（接受 string 参数）
// phone: 手机号
// 返回: true 表示有效，false 表示无效
func IsValidPhone(phone string) bool {
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return match
}

// IsValidPassword 校验密码（接受 string 参数）
// password: 密码
// 返回: true 表示有效，false 表示无效
func IsValidPassword(password string) bool {
	if len(password) < 6 || len(password) > 20 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasDigit
}

// IsValidSMSCode 校验短信验证码（接受 string 参数）
// code: 短信验证码
// 返回: true 表示有效，false 表示无效
func IsValidSMSCode(code string) bool {
	matched, _ := regexp.MatchString(`^\d{6}$`, code)
	return matched
}

func IsValidEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
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
		// 调试：打印详细错误信息（可以在开发时启用）
		// fmt.Printf("DEBUG Validation: Field=%s, Tag=%s, Value=%v, Param=%s\n",
		//     e.Field(), e.Tag(), e.Value(), e.Param())
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
