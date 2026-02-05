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
	"zjMall/pkg/validator"
)

type UserServiceHandler struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService // 依赖注入：业务逻辑层
	rbacService *service.RBACService // RBAC服务
}

func NewUserServiceHandler(userService *service.UserService, rbacService *service.RBACService) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService, // 初始化 service
		rbacService: rbacService, // 初始化 RBAC service
	}
}

func (h *UserServiceHandler) GetSMSCode(ctx context.Context, req *userv1.GetSMSCodeRequest) (*userv1.GetSMSCodeResponse, error) {
	if !validator.IsValidPhone(req.Phone) {
		return &userv1.GetSMSCodeResponse{
			Code:    1,
			Message: "手机号格式错误，请输入11位手机号",
		}, nil
	}
	return h.userService.GetSMSCode(ctx, req)
}

func (h *UserServiceHandler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	// 校验请求参数
	validator := service.NewRegisterRequestValidator(req)
	if err := validator.Validate(); err != nil {
		log.Printf("参数校验失败: %v", err)
		return &userv1.RegisterResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return h.userService.Register(ctx, req)
}

func (h *UserServiceHandler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	validator := service.NewLoginRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.userService.Login(ctx, req)
}

func (h *UserServiceHandler) LoginBySMS(ctx context.Context, req *userv1.LoginBySMSRequest) (*userv1.LoginResponse, error) {
	validator := service.NewLoginBySMSRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.userService.LoginBySMS(ctx, req)
}

func (h *UserServiceHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if req.UserId == 0 {
		return &userv1.GetUserResponse{
			Code:    1,
			Message: "用户ID不能为空",
		}, nil
	}
	return h.userService.GetUser(ctx, req)
}

func (h *UserServiceHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	validator := service.NewUpdateUserRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.UpdateUserResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
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
	// 从 context 中获取用户 ID（由认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &userv1.CreateAddressResponse{
			Code:    1,
			Message: "未登录或用户ID无效",
		}, nil
	}
	return h.userService.CreateAddress(ctx, userID, req)
}

func (h *UserServiceHandler) ListAddresses(ctx context.Context, req *userv1.ListAddressesRequest) (*userv1.ListAddressesResponse, error) {
	// 从 context 中获取用户 ID（由认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &userv1.ListAddressesResponse{
			Code:    1,
			Message: "未登录或用户ID无效",
		}, nil
	}
	return h.userService.ListAddresses(ctx, userID)
}

func (h *UserServiceHandler) UpdateAddress(ctx context.Context, req *userv1.UpdateAddressRequest) (*userv1.UpdateAddressResponse, error) {
	if req.AddressId == "" {
		return &userv1.UpdateAddressResponse{
			Code:    1,
			Message: "地址ID不能为空",
		}, nil
	}
	// 从 context 中获取用户 ID（由认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &userv1.UpdateAddressResponse{
			Code:    1,
			Message: "未登录或用户ID无效",
		}, nil
	}
	return h.userService.UpdateAddress(ctx, userID, req)
}

func (h *UserServiceHandler) DeleteAddress(ctx context.Context, req *userv1.DeleteAddressRequest) (*userv1.DeleteAddressResponse, error) {
	if req.AddressId == "" {
		return &userv1.DeleteAddressResponse{
			Code:    1,
			Message: "地址ID不能为空",
		}, nil
	}
	// 从 context 中获取用户 ID（由认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &userv1.DeleteAddressResponse{
			Code:    1,
			Message: "未登录或用户ID无效",
		}, nil
	}
	return h.userService.DeleteAddress(ctx, userID, req.AddressId)
}

func (h *UserServiceHandler) SetDefaultAddress(ctx context.Context, req *userv1.SetDefaultAddressRequest) (*userv1.SetDefaultAddressResponse, error) {
	if req.AddressId == "" {
		return &userv1.SetDefaultAddressResponse{
			Code:    1,
			Message: "地址ID不能为空",
		}, nil
	}
	// 从 context 中获取用户 ID（由认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == "" {
		return &userv1.SetDefaultAddressResponse{
			Code:    1,
			Message: "未登录或用户ID无效",
		}, nil
	}
	return h.userService.SetDefaultAddress(ctx, userID, req.AddressId)
}

func (h *UserServiceHandler) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*userv1.ChangePasswordResponse, error) {
	validator := service.NewChangePasswordRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	return h.userService.ChangePassword(ctx, req)
}

func (h *UserServiceHandler) BindPhone(ctx context.Context, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error) {
	validator := service.NewBindPhoneRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return h.userService.BindPhone(ctx, req)
}

func (h *UserServiceHandler) GetUserAddress(ctx context.Context, req *userv1.GetUserAddressRequest) (*userv1.GetUserAddressResponse, error) {
	// 从 context 中获取用户 ID（由认证中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	log.Printf("req.AddressId: %s", req.AddressId)
	log.Printf("userID: %s", userID)
	if userID == "" {
		return &userv1.GetUserAddressResponse{
			Code:    1,
			Message: "未登录或用户ID无效",
		}, nil
	}
	return h.userService.GetUserAddress(ctx, userID, req.AddressId)
}

// ========== RBAC相关方法 ==========

// AssignRole 为用户分配角色
func (h *UserServiceHandler) AssignRole(ctx context.Context, req *userv1.AssignRoleRequest) (*userv1.AssignRoleResponse, error) {
	if req.UserId == "" {
		return &userv1.AssignRoleResponse{
			Code:    1,
			Message: "用户ID不能为空",
		}, nil
	}
	if req.RoleCode == "" {
		return &userv1.AssignRoleResponse{
			Code:    1,
			Message: "角色代码不能为空",
		}, nil
	}

	err := h.rbacService.AssignRole(ctx, req.UserId, req.RoleCode)
	if err != nil {
		return &userv1.AssignRoleResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &userv1.AssignRoleResponse{
		Code:    0,
		Message: "分配角色成功",
	}, nil
}

// RemoveRole 移除用户角色
func (h *UserServiceHandler) RemoveRole(ctx context.Context, req *userv1.RemoveRoleRequest) (*userv1.RemoveRoleResponse, error) {
	if req.UserId == "" {
		return &userv1.RemoveRoleResponse{
			Code:    1,
			Message: "用户ID不能为空",
		}, nil
	}
	if req.RoleCode == "" {
		return &userv1.RemoveRoleResponse{
			Code:    1,
			Message: "角色代码不能为空",
		}, nil
	}

	err := h.rbacService.RemoveRole(ctx, req.UserId, req.RoleCode)
	if err != nil {
		return &userv1.RemoveRoleResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &userv1.RemoveRoleResponse{
		Code:    0,
		Message: "移除角色成功",
	}, nil
}

// GetUserRoles 查询用户角色列表
func (h *UserServiceHandler) GetUserRoles(ctx context.Context, req *userv1.GetUserRolesRequest) (*userv1.GetUserRolesResponse, error) {
	if req.UserId == "" {
		return &userv1.GetUserRolesResponse{
			Code:    1,
			Message: "用户ID不能为空",
		}, nil
	}

	roles, err := h.rbacService.GetUserRoles(ctx, req.UserId)
	if err != nil {
		return &userv1.GetUserRolesResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &userv1.GetUserRolesResponse{
		Code:    0,
		Message: "查询成功",
		Data:    roles,
	}, nil
}

// GetUserPermissions 查询用户权限列表
func (h *UserServiceHandler) GetUserPermissions(ctx context.Context, req *userv1.GetUserPermissionsRequest) (*userv1.GetUserPermissionsResponse, error) {
	if req.UserId == "" {
		return &userv1.GetUserPermissionsResponse{
			Code:    1,
			Message: "用户ID不能为空",
		}, nil
	}

	permissions, err := h.rbacService.GetUserPermissions(ctx, req.UserId)
	if err != nil {
		return &userv1.GetUserPermissionsResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &userv1.GetUserPermissionsResponse{
		Code:    0,
		Message: "查询成功",
		Data:    permissions,
	}, nil
}

// ListRoles 查询所有角色列表
func (h *UserServiceHandler) ListRoles(ctx context.Context, req *userv1.ListRolesRequest) (*userv1.ListRolesResponse, error) {
	var status *int32
	if req.Status > 0 {
		status = &req.Status
	}

	roles, err := h.rbacService.ListRoles(ctx, status)
	if err != nil {
		return &userv1.ListRolesResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &userv1.ListRolesResponse{
		Code:    0,
		Message: "查询成功",
		Data:    roles,
	}, nil
}

// ListPermissions 查询所有权限列表
func (h *UserServiceHandler) ListPermissions(ctx context.Context, req *userv1.ListPermissionsRequest) (*userv1.ListPermissionsResponse, error) {
	var resource *string
	if req.Resource != "" {
		resource = &req.Resource
	}
	var status *int32
	if req.Status > 0 {
		status = &req.Status
	}

	permissions, err := h.rbacService.ListPermissions(ctx, resource, status)
	if err != nil {
		return &userv1.ListPermissionsResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	return &userv1.ListPermissionsResponse{
		Code:    0,
		Message: "查询成功",
		Data:    permissions,
	}, nil
}
