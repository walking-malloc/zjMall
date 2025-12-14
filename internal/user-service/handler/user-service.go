package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/common/middleware"
	"zjMall/internal/user-service/service"
)

type UserServiceHandler struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService // 依赖注入：业务逻辑层
}

func NewUserServiceHandler(userService *service.UserService) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService, // 初始化 service
	}
}

func (h *UserServiceHandler) GetSMSCode(ctx context.Context, req *userv1.GetSMSCodeRequest) (*userv1.GetSMSCodeResponse, error) {
	return h.userService.GetSMSCode(ctx, req)
}

func (h *UserServiceHandler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	return h.userService.Register(ctx, req)
}

func (h *UserServiceHandler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return h.userService.Login(ctx, req)
}

func (h *UserServiceHandler) LoginBySMS(ctx context.Context, req *userv1.LoginBySMSRequest) (*userv1.LoginResponse, error) {
	return h.userService.LoginBySMS(ctx, req)
}

func (h *UserServiceHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	return h.userService.GetUser(ctx, req)
}

func (h *UserServiceHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	return h.userService.UpdateUser(ctx, req)
}

// UploadAvatarHTTP 处理头像上传的HTTP请求（multipart/form-data）
func (h *UserServiceHandler) UploadAvatarHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("UploadAvatarHTTP 被调用: Method=%s, Path=%s", r.Method, r.URL.Path)

	// 1. 只接受POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 从Context获取用户ID（认证中间件已注入）
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, `{"code":1,"message":"未登录或Token无效"}`, http.StatusUnauthorized)
		return
	}

	// 3. 解析multipart/form-data（限制10MB）
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, `{"code":1,"message":"解析表单失败"}`, http.StatusBadRequest)
		return
	}

	// 4. 获取上传的文件
	file, header, err := r.FormFile("avatar") //获取前端对应的字段
	if err != nil {
		http.Error(w, `{"code":1,"message":"请选择图片文件"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 5. 校验文件类型（可选，更严格可以校验文件头）
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, `{"code":1,"message":"文件必须是图片格式"}`, http.StatusBadRequest)
		return
	}

	// 6. 调用Service层上传
	resp, err := h.userService.UploadAvatarFromReader(r.Context(), userID, file, header.Filename)
	if err != nil {
		log.Printf("上传头像失败: %v", err)
		http.Error(w, `{"code":1,"message":"上传失败，请稍后重试"}`, http.StatusInternalServerError)
		return
	}

	// 7. 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserServiceHandler) CreateAddress(ctx context.Context, req *userv1.CreateAddressRequest) (*userv1.CreateAddressResponse, error) {
	return h.userService.CreateAddress(ctx, req)
}

func (h *UserServiceHandler) ListAddresses(ctx context.Context, req *userv1.ListAddressesRequest) (*userv1.ListAddressesResponse, error) {
	return h.userService.ListAddresses(ctx, req)
}

func (h *UserServiceHandler) UpdateAddress(ctx context.Context, req *userv1.UpdateAddressRequest) (*userv1.UpdateAddressResponse, error) {
	return h.userService.UpdateAddress(ctx, req)
}

func (h *UserServiceHandler) DeleteAddress(ctx context.Context, req *userv1.DeleteAddressRequest) (*userv1.DeleteAddressResponse, error) {
	return h.userService.DeleteAddress(ctx, req)
}

func (h *UserServiceHandler) SetDefaultAddress(ctx context.Context, req *userv1.SetDefaultAddressRequest) (*userv1.SetDefaultAddressResponse, error) {
	return h.userService.SetDefaultAddress(ctx, req)
}

func (h *UserServiceHandler) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*userv1.ChangePasswordResponse, error) {
	return h.userService.ChangePassword(ctx, req)
}

func (h *UserServiceHandler) BindPhone(ctx context.Context, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error) {
	return h.userService.BindPhone(ctx, req)
}
