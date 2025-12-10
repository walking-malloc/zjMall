package pkg

const (
	SuccessCode = 200
)

// 客户端错误码
const (
	CodeBadRequest    = 400 // 请求参数错误
	CodeUnauthorized  = 401 // 未授权
	CodeNotFound      = 404 // 资源不存在
	CodeInternalError = 500
)

// 错误消息映射
var ErrorMessages = map[int]string{
	SuccessCode:      "成功",
	CodeBadRequest:   "请求参数错误",
	CodeUnauthorized: "未授权访问",
	CodeNotFound:     "资源不存在",
}

// GetMessage 根据错误码获取错误消息
func GetMessage(code int) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return "未知错误"
}
