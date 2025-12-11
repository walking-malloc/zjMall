package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"zjMall/internal/common/cache"
	"zjMall/internal/user-service/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
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

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	// 清除相关缓存
	go func() {
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf("user:id:%d", user.ID))
		r.cacheRepo.Delete(context.Background(), fmt.Sprintf("user:phone:%s", user.Phone))
	}()
	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	//先从redis中获取用户信息
	cacheKey := fmt.Sprintf("user:id:%d", id)
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
	cacheKey := fmt.Sprintf("user:phone:%s", phone)
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

	r.cacheRepo.Set(ctx, key, data, expiration)
}
