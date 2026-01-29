package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"zjMall/internal/common/cache"
	"zjMall/internal/user-service/model"

	"gorm.io/gorm"
)

const (
	// Redis Key 前缀常量
	CacheKeyUserByID    = "user:id:%s"           // 用户ID缓存
	CacheKeyUserByPhone = "user:phone:%s"        // 用户手机号缓存
	CacheKeyUserToken   = "user:token:%s"        // 用户Token缓存
	CacheKeySMSCode     = "user:sms:code:%s"     // 短信验证码
	CacheKeySMSRate     = "user:sms:rate:%s"     // 短信频率限制
	CacheKeySMSCount    = "user:sms:count:%s:%s" // 短信每日计数（phone + date）
)

type UserAuthInfo struct {
	ID       string `json:"id" gorm:"column:id"`
	Phone    string `json:"phone" gorm:"column:phone"`
	Password string `json:"password" gorm:"column:password"`
}
type UserRepository interface {
	//单独密码获取操作
	GetUserPasswordByPhone(ctx context.Context, phone string) (*UserAuthInfo, error)
	GetUserPasswordByID(ctx context.Context, id string) (*UserAuthInfo, error)

	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error

	//短信验证码相关操作
	SetSMSCode(ctx context.Context, phone, code string, expiration time.Duration) error
	GetSMSCode(ctx context.Context, phone string) (string, error)
	DeleteSMSCode(ctx context.Context, phone string) error
	CheckSMSCodeRateLimit(ctx context.Context, phone string, interval int64, dailyLimit int64) error

	//地址相关操作
	AddAddress(ctx context.Context, address *model.Address) error
	ListAddresses(ctx context.Context, userID string) ([]*model.Address, error)
	GetAddressByID(ctx context.Context, userID string, addressID string) (*model.Address, error)
	GetDefaultAddress(ctx context.Context, userID string) (*model.Address, error)
	UpdateAddress(ctx context.Context, address *model.Address) error
	DeleteAddress(ctx context.Context, userID string, addressID string) error
	SetDefaultAddress(ctx context.Context, userID string, addressID string) error
	CreateAddressWithDefault(ctx context.Context, address *model.Address) error
}

type userRepository struct {
	db        *gorm.DB
	cacheRepo cache.CacheRepository // 使用字段，更清晰
}

func NewUserRepository(db *gorm.DB, cacheRepo cache.CacheRepository) UserRepository {
	return &userRepository{
		db:        db,
		cacheRepo: cacheRepo,
	}
}

//TODO: 用户信息缓存需要防止缓存击穿

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	// ULID 会在 BeforeCreate 钩子中自动生成
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	// 清除并设置相关缓存
	go func() {
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(CacheKeyUserByID, user.ID))
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(CacheKeyUserByPhone, user.Phone))
		r.setToCache(context.Background(), fmt.Sprintf(CacheKeyUserByID, user.ID), user, 30*time.Minute)
		r.setToCache(context.Background(), fmt.Sprintf(CacheKeyUserByPhone, user.Phone), user, 30*time.Minute)
	}()
	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	//先从redis中获取用户信息
	cacheKey := fmt.Sprintf(CacheKeyUserByID, id)
	if user := r.getFromCache(ctx, cacheKey); user != nil {
		return user, nil
	}
	//如果redis中没有用户信息，则从数据库中获取
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, err
	}

	go r.setToCache(ctx, cacheKey, &user, 30*time.Minute) //异步写入缓存,避免阻塞主流程
	return &user, nil
}

func (r *userRepository) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	// 1. 先查缓存
	cacheKey := fmt.Sprintf(CacheKeyUserByPhone, phone)
	if user := r.getFromCache(ctx, cacheKey); user != nil {
		return user, nil
	}
	//如果redis中没有用户信息，则从数据库中获取
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, err
	}
	go r.setToCache(ctx, cacheKey, &user, 30*time.Minute) //异步写入缓存,避免阻塞主流程
	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *model.User) error {

	//先获取旧用户信息，用于待会删除缓存的操作
	var oldUser model.User
	if err := r.db.WithContext(ctx).Where("id = ?", user.ID).First(&oldUser).Error; err != nil {
		return err
	}

	//更新用户
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).Updates(user).Error
	if err != nil {
		return err
	}
	go func() {
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf(CacheKeyUserByID, oldUser.ID))
		if oldUser.Phone != "" {
			r.cacheRepo.Delete(context.Background(), fmt.Sprintf(CacheKeyUserByPhone, oldUser.Phone))
		}
	}()
	return nil
}

// 私有方法：从缓存获取
func (r *userRepository) getFromCache(ctx context.Context, key string) *model.User {
	data, err := r.cacheRepo.Get(ctx, key)
	if err != nil {
		return nil // 缓存未命中或出错，返回 nil
	}

	var user model.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil // 反序列化失败，返回 nil
	}

	return &user
}

// 私有方法：写入缓存
func (r *userRepository) setToCache(ctx context.Context, key string, user *model.User, expiration time.Duration) {
	data, err := json.Marshal(user)
	if err != nil {
		return // 序列化失败，不写入缓存
	}

	r.cacheRepo.Set(ctx, key, string(data), expiration)
}

// 设置短信验证码
func (r *userRepository) SetSMSCode(ctx context.Context, phone, code string, expiration time.Duration) error {
	key := fmt.Sprintf(CacheKeySMSCode, phone)
	return r.cacheRepo.Set(ctx, key, code, expiration)
}

// 获取短信验证码
func (r *userRepository) GetSMSCode(ctx context.Context, phone string) (string, error) {
	key := fmt.Sprintf(CacheKeySMSCode, phone)
	log.Printf("GetSMSCode %s:%v", key, phone)
	return r.cacheRepo.Get(ctx, key)
}

// 删除短信验证码
func (r *userRepository) DeleteSMSCode(ctx context.Context, phone string) error {
	key := fmt.Sprintf(CacheKeySMSCode, phone)
	log.Printf("DeleteSMSCode %s:%v", key, phone)
	return r.cacheRepo.Delete(ctx, key)
}

func (r *userRepository) CheckSMSCodeRateLimit(ctx context.Context, phone string, interval int64, dailyLimit int64) error {
	rateKey := fmt.Sprintf(CacheKeySMSRate, phone)
	ok, _ := r.cacheRepo.SetNXInt(ctx, rateKey, 1, time.Duration(interval)*time.Second) //如果key存在，表示在interval秒内已经发送过短信验证码
	if !ok {
		return fmt.Errorf("短信验证码发送频率过高，请%d秒后再试", interval)
	}

	//每日发送次数检查
	today := time.Now().Format("2006-01-02")
	countKey := fmt.Sprintf(CacheKeySMSCount, phone, today)
	count, _ := r.cacheRepo.GetInt(ctx, countKey)
	if count >= dailyLimit {
		return fmt.Errorf("短信验证码发送次数过多，请明日再试")
	}

	if count == 0 {
		r.cacheRepo.SetInt(ctx, countKey, 1, 24*time.Hour)
	} else {
		r.cacheRepo.Incr(ctx, countKey)
	}
	return nil
}

func (r *userRepository) AddAddress(ctx context.Context, address *model.Address) error {
	return r.db.WithContext(ctx).Model(&model.Address{}).Create(address).Error
}

func (r *userRepository) ListAddresses(ctx context.Context, userID string) ([]*model.Address, error) {
	var addresses []*model.Address
	err := r.db.WithContext(ctx).Model(&model.Address{}).Where("user_id = ?", userID).Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *userRepository) GetAddressByID(ctx context.Context, userID string, addressID string) (*model.Address, error) {
	var address model.Address
	err := r.db.WithContext(ctx).Model(&model.Address{}).
		Where("user_id = ? AND id = ?", userID, addressID).
		First(&address).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 地址不存在
	}
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *userRepository) GetDefaultAddress(ctx context.Context, userID string) (*model.Address, error) {
	var address model.Address
	err := r.db.WithContext(ctx).Model(&model.Address{}).
		Where("user_id = ? AND is_default = ?", userID, true).
		First(&address).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 没有默认地址
	}
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *userRepository) UpdateAddress(ctx context.Context, address *model.Address) error {
	// 同时根据user_id和id更新，确保只能更新自己的地址
	return r.db.WithContext(ctx).Model(&model.Address{}).Where("user_id = ? AND id = ?", address.UserID, address.ID).Updates(address).Error
}

func (r *userRepository) DeleteAddress(ctx context.Context, userID string, addressID string) error {
	return r.db.WithContext(ctx).Model(&model.Address{}).Where("user_id = ? AND id = ?", userID, addressID).Delete(&model.Address{}).Error
}

func (r *userRepository) SetDefaultAddress(ctx context.Context, userID string, addressID string) error {
	// 使用事务确保原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先取消该用户所有地址的默认状态
		if err := tx.Model(&model.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		// 2. 设置新的默认地址
		return tx.Model(&model.Address{}).Where("user_id = ? AND id = ?", userID, addressID).Update("is_default", true).Error
	})
}

func (r *userRepository) CreateAddressWithDefault(ctx context.Context, address *model.Address) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建地址（使用事务中的tx）
		if err := tx.WithContext(ctx).Model(&model.Address{}).Create(address).Error; err != nil {
			return err
		}

		// 取消该用户所有地址的默认状态
		if err := tx.WithContext(ctx).Model(&model.Address{}).
			Where("user_id = ?", address.UserID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新地址为默认
		if err := tx.WithContext(ctx).Model(&model.Address{}).
			Where("user_id = ? AND id = ?", address.UserID, address.ID).
			Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
}
func (r *userRepository) GetUserPasswordByPhone(ctx context.Context, phone string) (*UserAuthInfo, error) {
	var userAuthInfo UserAuthInfo
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("phone = ?", phone).Select("id", "phone", "password").First(&userAuthInfo).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, err
	}
	return &userAuthInfo, nil
}
func (r *userRepository) GetUserPasswordByID(ctx context.Context, id string) (*UserAuthInfo, error) {
	var userAuthInfo UserAuthInfo
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Select("id", "phone", "password").First(&userAuthInfo).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, err
	}
	return &userAuthInfo, nil
}
