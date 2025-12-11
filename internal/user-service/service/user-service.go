package service

import (
	"context"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/user-service/repository"
)

// UserService 用户服务（业务逻辑层）
type UserService struct {
	userRepo repository.UserRepository // 数据访问（内部封装查询缓存）
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	// 校验请求参数
	validator := NewRegisterRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	//检查手机号是否已注册
	user, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	if user != nil {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "手机号已注册",
		}, nil
	}
	// todo 验证校验码
	return &userv1.RegisterResponse{
		Code:    0,
		Message: "注册成功",
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	validator := NewLoginRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	// todo 实现登录逻辑
	return &userv1.LoginResponse{
		Code:    0,
		Message: "登录成功",
	}, nil
}
