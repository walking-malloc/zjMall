package pkg

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}, message string) Result {
	if message == "" {
		message = GetMessage(SuccessCode)
	}
	return Result{
		Code:    SuccessCode,
		Message: message,
		Data:    data,
	}
}

// Fail 失败响应
func Fail(code int, message string) Result {
	if message == "" {
		message = GetMessage(code)
	}
	return Result{
		Code:    code,
		Message: message,
	}
}
