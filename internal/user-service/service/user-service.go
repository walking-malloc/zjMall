package service

import (
	"context"
	"fmt"
	"log"
	"time"
	userv1 "zjMall/gen/go/api/proto/user"
	"zjMall/internal/config"
	"zjMall/internal/sms"
	"zjMall/internal/user-service/repository"
	"zjMall/pkg"

	"github.com/go-redis/redis/v8"
)

// UserService 用户服务（业务逻辑层）
type UserService struct {
	userRepo  repository.UserRepository // 数据访问（内部封装查询缓存）
	smsClient sms.SMSClient
	smsConfig config.SMSConfig
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, smsClient sms.SMSClient, smsConfig config.SMSConfig) *UserService {
	return &UserService{
		userRepo:  userRepo,
		smsClient: smsClient,
		smsConfig: smsConfig,
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

// 获取短信验证码
func (s *UserService) GetSMSCode(ctx context.Context, req *userv1.GetSMSCodeRequest) (*userv1.GetSMSCodeResponse, error) {
	validator := NewGetSMSCodeRequestValidator(req)
	if err := validator.Validate(); err != nil {
		return &userv1.GetSMSCodeResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	smsConfig := s.smsConfig
	//检查发送频率
	err := s.userRepo.CheckSMSCodeRateLimit(ctx, req.Phone, int64(s.smsConfig.SendInterval), int64(s.smsConfig.MaxSendCount))
	if err != nil {
		return &userv1.GetSMSCodeResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	//获取验证码
	storedCode, err := s.userRepo.GetSMSCode(ctx, req.Phone)
	if err != nil && err != redis.Nil {
		return &userv1.GetSMSCodeResponse{
			Code:    1,
			Message: "获取验证码失败",
		}, nil
	}
	if err == nil && storedCode != "" {
		// 发送短信验证码
		go func() {
			err := s.smsClient.SendCode(req.Phone, storedCode)
			if err != nil {
				log.Printf("发送短信验证码失败: %v", err)
			}
		}()
		return &userv1.GetSMSCodeResponse{
			Code:    0,
			Message: "验证码已发送",
		}, nil
	} //如果验证码存在直接返回验证码

	// 生成短信验证码
	code, err := pkg.GenerateSMSCode()
	if err != nil {
		return &userv1.GetSMSCodeResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	//存入缓存
	err = s.userRepo.SetSMSCode(ctx, req.Phone, code, time.Duration(smsConfig.ExpireTime)*time.Second)
	if err != nil {
		return &userv1.GetSMSCodeResponse{
			Code:    1,
			Message: "存入缓存失败",
		}, nil
	}
	// 发送短信验证码
	go func() {
		err := s.smsClient.SendCode(req.Phone, code)
		if err != nil {
			log.Printf("发送短信验证码失败: %v", err)
		}
	}()
	return &userv1.GetSMSCodeResponse{
		Code:    0,
		Message: "获取短信验证码成功",
	}, nil
}

func (s *UserService) VerifySMSCode(ctx context.Context, phone, code string) error {
	//获取缓存中的验证码
	storedCode, err := s.userRepo.GetSMSCode(ctx, phone)
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("验证码已过期或不存在")
		}
		return err
	}
	//如果验证码不存在
	if storedCode == "" {
		return fmt.Errorf("验证码已过期或不存在")
	}
	//如果验证码错误
	if storedCode != code {
		return fmt.Errorf("验证码错误")
	}

	//删除缓存中的验证码
	err = s.userRepo.DeleteSMSCode(ctx, phone)
	if err != nil {
		log.Printf("删除缓存中的验证码失败: %v", err)
	}
	return nil
}
