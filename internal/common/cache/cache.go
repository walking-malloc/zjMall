package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheRepository 通用缓存仓库接口（基础方法）
type CacheRepository interface {
	// 基础方法
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

// RedisCacheRepository Redis 缓存仓库实现
type RedisCacheRepository struct {
	client *redis.Client
}

// NewCacheRepository 创建通用缓存仓库
func NewCacheRepository(client *redis.Client) CacheRepository {
	return &RedisCacheRepository{
		client: client,
	}
}

// Set 设置缓存
func (r *RedisCacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取缓存
func (r *RedisCacheRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete 删除缓存
func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists 检查 key 是否存在
func (r *RedisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	result := r.client.Exists(ctx, key)
	return result.Val() > 0, result.Err()
}

// SetNX 设置缓存（如果不存在）
func (r *RedisCacheRepository) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Expire 设置过期时间
func (r *RedisCacheRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}
