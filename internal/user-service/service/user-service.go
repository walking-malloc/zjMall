package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
	userv1 "zjMall/gen/go/api/proto/user"
	upload "zjMall/internal/common/oss"
	"zjMall/internal/config"
	"zjMall/internal/sms"
	"zjMall/internal/user-service/model"
	"zjMall/internal/user-service/repository"
	"zjMall/pkg"

	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserService 用户服务（业务逻辑层）
type UserService struct {
	userRepo  repository.UserRepository // 数据访问（内部封装查询缓存）
	smsClient sms.SMSClient
	smsConfig config.SMSConfig
	ossClient upload.UploadClient // OSS上传客户端
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, smsClient sms.SMSClient, smsConfig config.SMSConfig, ossClient upload.UploadClient) *UserService {
	return &UserService{
		userRepo:  userRepo,
		smsClient: smsClient,
		smsConfig: smsConfig,
		ossClient: ossClient,
	}
}

// 用户注册接口
func (s *UserService) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {

	//检查手机号是否已注册
	user, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if user != nil {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "手机号已注册",
		}, nil
	}
	//TODO: 校验码不能明文存储，需要加密存储
	//验证校验码
	err = s.VerifySMSCode(ctx, req.Phone, req.SmsCode)
	if err != nil {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "校验码错误",
		}, nil
	}

	//校验密码是否相等
	if req.Password != req.ConfirmPassword {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "密码不一致",
		}, nil
	}

	password, err := pkg.HashPassword(req.Password)
	if err != nil {
		log.Printf("加密密码失败: %v", err)
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}

	user = &model.User{
		Phone:    req.Phone,
		Password: password,
	}

	//创建用户
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		log.Printf("创建用户失败: Phone=%s, Error=%v", req.Phone, err)
		return &userv1.RegisterResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	log.Printf("用户创建成功: ID=%s, Phone=%s", user.ID, user.Phone)

	//生成JWT Token（注册后自动登录，使用默认过期时间）
	expirationTime := 7 * 24 * time.Hour
	token, _, err := pkg.GenerateJWT(user.ID, expirationTime)
	if err != nil {
		return &userv1.RegisterResponse{
			Code:    1,
			Message: "生成 Token 失败",
		}, nil
	}

	// 转换为 UserInfo
	userInfo := s.convertToUserInfo(user)

	return &userv1.RegisterResponse{
		Code:    0,
		Message: "注册成功",
		Data: &userv1.RegisterData{
			User:  userInfo,
			Token: token,
		},
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {

	// 先查看是否用户存在
	userAuthInfo, err := s.userRepo.GetUserPasswordByPhone(ctx, req.Phone)
	if err != nil {
		log.Printf("获取用户密码失败: %v", err)
		return &userv1.LoginResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if userAuthInfo == nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: "用户不存在，请先注册",
		}, nil
	}

	//验证密码是否正确
	ok := pkg.VerifyPassword(userAuthInfo.Password, req.Password)
	if !ok {
		log.Printf("密码错误: Phone=%s", req.Phone)
		return &userv1.LoginResponse{
			Code:    1,
			Message: "密码错误",
		}, nil
	}

	// 根据 RememberMe 生成 Token
	var token string
	var expiresAt int64
	var expirationTime time.Duration

	if req.RememberMe {
		token, expiresAt, err = pkg.GenerateJWTWithRememberMe(userAuthInfo.ID, req.RememberMe)
		// 计算过期时长：从当前时间到过期时间戳的时长
		expirationTime = time.Until(time.Unix(expiresAt, 0))
	} else {
		expirationTime = 7 * 24 * time.Hour
		token, expiresAt, err = pkg.GenerateJWT(userAuthInfo.ID, expirationTime)
	}

	// 先检查错误，再存储 token
	if err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: "生成 Token 失败",
		}, nil
	}

	return &userv1.LoginResponse{
		Code:    0,
		Message: "登录成功",
		Data: &userv1.LoginData{
			User: &userv1.UserInfo{
				Id:    userAuthInfo.ID,
				Phone: s.maskPhone(userAuthInfo.Phone),
			},
			Token:     token,
			ExpiresAt: expiresAt,
		},
	}, nil
}

func (s *UserService) LoginBySMS(ctx context.Context, req *userv1.LoginBySMSRequest) (*userv1.LoginResponse, error) {

	if err := s.VerifySMSCode(ctx, req.Phone, req.SmsCode); err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: err.Error(),
		}, nil
	}

	//获取用户信息
	user, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if user == nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: "用户不存在，请先注册",
		}, nil
	}

	//生成token（短信登录使用默认过期时间）
	expirationTime := 7 * 24 * time.Hour
	token, expiresAt, err := pkg.GenerateJWT(user.ID, expirationTime)
	if err != nil {
		return &userv1.LoginResponse{
			Code:    1,
			Message: "生成 Token 失败",
		}, nil
	}

	//转换为UserInfo
	userInfo := s.convertToUserInfo(user)

	return &userv1.LoginResponse{
		Code:    0,
		Message: "登录成功",
		Data: &userv1.LoginData{
			User:      userInfo,
			Token:     token,
			ExpiresAt: expiresAt,
		},
	}, nil
}

// 获取短信验证码
func (s *UserService) GetSMSCode(ctx context.Context, req *userv1.GetSMSCodeRequest) (*userv1.GetSMSCodeResponse, error) {

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

// convertToUserInfo 将 model.User 转换为 userv1.UserInfo
func (s *UserService) convertToUserInfo(user *model.User) *userv1.UserInfo {
	userInfo := &userv1.UserInfo{
		Id:        user.ID,
		Phone:     s.maskPhone(user.Phone),
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Email:     user.Email,
		Gender:    int32(user.Gender),
		Status:    int32(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
	}

	// 处理生日（可选字段）
	if user.Birthday != nil {
		userInfo.Birthday = user.Birthday.Format(time.DateOnly)
	}

	// 处理最后登录时间（可选字段）
	if user.LastLoginAt != nil {
		userInfo.LastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return userInfo
}

// maskPhone 手机号脱敏（显示前3位和后4位，中间用*代替）
func (s *UserService) maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
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

// todo
func (s *UserService) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	return nil, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, strconv.FormatInt(req.UserId, 10))
	if err != nil {
		return &userv1.GetUserResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if user == nil {
		return &userv1.GetUserResponse{
			Code:    1,
			Message: "用户不存在",
		}, nil
	}

	userInfo := s.convertToUserInfo(user)
	return &userv1.GetUserResponse{
		Code:    0,
		Message: "获取用户信息成功",
		Data:    userInfo,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {

	user, err := s.userRepo.GetUserByID(ctx, strconv.FormatInt(req.UserId, 10))
	if err != nil {
		return &userv1.UpdateUserResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if user == nil {
		return &userv1.UpdateUserResponse{
			Code:    1,
			Message: "用户不存在",
		}, nil
	}

	user.Nickname = req.Nickname
	user.Email = req.Email
	user.Gender = int8(req.Gender)

	birthday, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		return &userv1.UpdateUserResponse{
			Code:    1,
			Message: "生日格式错误，正确格式应为YYYY-MM-DD",
		}, nil
	}
	user.Birthday = &birthday

	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return &userv1.UpdateUserResponse{
			Code:    1,
			Message: "更新用户信息失败",
		}, nil
	}
	return &userv1.UpdateUserResponse{
		Code:    0,
		Message: "更新用户信息成功",
	}, nil
}

// UploadAvatarFromReader 从io.Reader上传头像到OSS（供HTTP Handler调用）
func (s *UserService) UploadAvatarFromReader(ctx context.Context, userID string, file io.Reader, filename string) (*userv1.UploadAvatarResponse, error) {
	// 1. 上传到OSS
	avatarURL, err := s.ossClient.UploadAvatar(userID, file, filename)
	if err != nil {
		log.Printf("上传头像到OSS失败: %v", err)
		return &userv1.UploadAvatarResponse{
			Code:    1,
			Message: "上传头像失败，请稍后重试",
		}, nil
	}

	// 2. 更新数据库
	user := &model.User{
		BaseModel: pkg.BaseModel{ID: userID},
		Avatar:    avatarURL,
	}
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		log.Printf("更新用户头像URL失败: %v", err)
		return &userv1.UploadAvatarResponse{
			Code:    1,
			Message: "更新头像信息时出错，请稍后重试",
		}, err
	}

	// 3. 返回成功响应
	return &userv1.UploadAvatarResponse{
		Code:      0,
		Message:   "上传成功",
		AvatarUrl: avatarURL,
	}, nil
}

// =============== 地址管理 ===============
// 添加单个地址
func (s *UserService) CreateAddress(ctx context.Context, req *userv1.CreateAddressRequest) (*userv1.CreateAddressResponse, error) {

	address := &model.Address{
		UserID:        req.UserId,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
	}

	// 调用Repository层的方法，事务在Repository层处理
	err := s.userRepo.CreateAddressWithDefault(ctx, address)
	if err != nil {
		return &userv1.CreateAddressResponse{
			Code:    1,
			Message: "创建地址失败",
		}, nil
	}

	// 转换为proto格式并返回（proto中Address只定义了部分字段）
	addressData := &userv1.Address{
		Id:            address.ID,
		UserId:        address.UserID,
		ReceiverName:  address.ReceiverName,
		ReceiverPhone: address.ReceiverPhone,
		Province:      address.Province,
		City:          address.City,
		District:      address.District,
		Detail:        address.Detail,
		PostalCode:    address.PostalCode,
		IsDefault:     address.IsDefault,
		CreatedAt:     timestamppb.New(address.CreatedAt),
		UpdatedAt:     timestamppb.New(address.UpdatedAt),
	}

	return &userv1.CreateAddressResponse{
		Code:    0,
		Message: "创建地址成功",
		Data:    addressData,
	}, nil
}

// 查询收货地址列表
func (s *UserService) ListAddresses(ctx context.Context, req *userv1.ListAddressesRequest) (*userv1.ListAddressesResponse, error) {
	addresses, err := s.userRepo.ListAddresses(ctx, req.UserId)
	if err != nil {
		return &userv1.ListAddressesResponse{
			Code:    1,
			Message: "查询地址列表失败",
		}, nil
	}
	addressesData := make([]*userv1.Address, len(addresses))
	for i, address := range addresses {
		addressesData[i] = &userv1.Address{
			Id:            address.ID,
			UserId:        address.UserID,
			ReceiverName:  address.ReceiverName,
			ReceiverPhone: address.ReceiverPhone,
			Province:      address.Province,
			City:          address.City,
			District:      address.District,
			Detail:        address.Detail,
			PostalCode:    address.PostalCode,
			IsDefault:     address.IsDefault,
			CreatedAt:     timestamppb.New(address.CreatedAt),
			UpdatedAt:     timestamppb.New(address.UpdatedAt),
		}
	}
	return &userv1.ListAddressesResponse{
		Code:    0,
		Message: "查询地址列表成功",
		Data:    addressesData,
	}, nil
}

// 更新收货地址
func (s *UserService) UpdateAddress(ctx context.Context, req *userv1.UpdateAddressRequest) (*userv1.UpdateAddressResponse, error) {
	// 如果设置为默认地址，先取消其他地址的默认状态
	if req.IsDefault {
		err := s.userRepo.SetDefaultAddress(ctx, req.UserId, req.AddressId)
		if err != nil {
			return &userv1.UpdateAddressResponse{
				Code:    1,
				Message: "设置默认地址失败",
			}, nil
		}
	}

	address := &model.Address{
		BaseModel:     pkg.BaseModel{ID: req.AddressId}, // 设置地址ID
		UserID:        req.UserId,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Detail:        req.Detail,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
	}
	err := s.userRepo.UpdateAddress(ctx, address)
	if err != nil {
		return &userv1.UpdateAddressResponse{
			Code:    1,
			Message: "更新地址失败",
		}, nil
	}
	return &userv1.UpdateAddressResponse{
		Code:    0,
		Message: "更新地址成功",
	}, nil
}

// 删除收货地址
func (s *UserService) DeleteAddress(ctx context.Context, req *userv1.DeleteAddressRequest) (*userv1.DeleteAddressResponse, error) {
	err := s.userRepo.DeleteAddress(ctx, req.UserId, req.AddressId)
	if err != nil {
		return &userv1.DeleteAddressResponse{
			Code:    1,
			Message: "删除地址失败",
		}, nil
	}
	return &userv1.DeleteAddressResponse{
		Code:    0,
		Message: "删除地址成功",
	}, nil
}

// 设置默认收货地址
func (s *UserService) SetDefaultAddress(ctx context.Context, req *userv1.SetDefaultAddressRequest) (*userv1.SetDefaultAddressResponse, error) {
	err := s.userRepo.SetDefaultAddress(ctx, req.UserId, req.AddressId)
	if err != nil {
		return &userv1.SetDefaultAddressResponse{
			Code:    1,
			Message: "设置默认地址失败",
		}, nil
	}
	return &userv1.SetDefaultAddressResponse{
		Code:    0,
		Message: "设置默认地址成功",
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*userv1.ChangePasswordResponse, error) {

	// 检查用户是否存在
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if user == nil {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "用户不存在",
		}, nil
	}
	// 检查新密码是否一致
	if req.NewPassword != req.ConfirmPassword {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "新密码不一致",
		}, nil
	}
	if req.NewPassword == req.OldPassword {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "新密码不能与旧密码相同",
		}, nil
	}
	// 检查旧密码是否正确
	ok := pkg.VerifyPassword(user.Password, req.OldPassword)
	if !ok {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "旧密码错误",
		}, nil
	}

	newPassword, err := pkg.HashPassword(req.NewPassword)
	if err != nil {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	user.Password = newPassword
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return &userv1.ChangePasswordResponse{
			Code:    1,
			Message: "更新密码失败",
		}, nil
	}
	return &userv1.ChangePasswordResponse{
		Code:    0,
		Message: "更新密码成功",
	}, nil
}

func (s *UserService) BindPhone(ctx context.Context, req *userv1.BindPhoneRequest) (*userv1.BindPhoneResponse, error) {

	//先检查该用户是否存在
	user, err := s.userRepo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if user == nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "用户不存在",
		}, nil
	}
	// 检查新手机号是否与当前手机号相同
	if user.Phone == req.NewPhone {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "新手机号与当前手机号相同",
		}, nil
	}
	//检查新手机号是否已注册
	existingUser, err := s.userRepo.GetUserByPhone(ctx, req.NewPhone)
	if err != nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "系统错误，稍后重试",
		}, nil
	}
	if existingUser != nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "新手机号已注册",
		}, nil
	}
	//验证新短信验证码
	err = s.VerifySMSCode(ctx, req.NewPhone, req.NewSmsCode)
	if err != nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "新短信验证码错误",
		}, nil
	}
	//更新用户手机号
	user.Phone = req.NewPhone
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return &userv1.BindPhoneResponse{
			Code:    1,
			Message: "更新手机号失败",
		}, nil
	}
	return &userv1.BindPhoneResponse{
		Code:    0,
		Message: "更新手机号成功",
	}, nil
}
