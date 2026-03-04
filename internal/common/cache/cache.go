package cache

import (
	"context"
	"time"

	"zjMall/internal/common/metrics"

	"github.com/go-redis/redis/v8"
)

// CacheRepository 通用缓存仓库接口（基础方法）
type CacheRepository interface {
	// 基础方法
	//string类型
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	//int类型
	SetInt(ctx context.Context, key string, value int64, expiration time.Duration) error
	GetInt(ctx context.Context, key string) (int64, error)
	SetNXInt(ctx context.Context, key string, value int64, expiration time.Duration) (bool, error)

	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Incr(ctx context.Context, key string) error
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

// string类型
func (r *RedisCacheRepository) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	start := time.Now()
	err := r.client.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start).Seconds()

	metrics.RedisOperationDuration.WithLabelValues("set").Observe(duration)
	if err != nil {
		metrics.RedisOperationsTotal.WithLabelValues("set", "error").Inc()
	} else {
		metrics.RedisOperationsTotal.WithLabelValues("set", "success").Inc()
	}
	return err
}

func (r *RedisCacheRepository) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	result, err := r.client.Get(ctx, key).Result()
	duration := time.Since(start).Seconds()

	metrics.RedisOperationDuration.WithLabelValues("get").Observe(duration)
	if err != nil {
		if err == redis.Nil {
			// 缓存未命中
			metrics.RedisCacheMisses.Inc()
			metrics.RedisOperationsTotal.WithLabelValues("get", "miss").Inc()
		} else {
			// 错误
			metrics.RedisOperationsTotal.WithLabelValues("get", "error").Inc()
		}
	} else {
		// 缓存命中
		metrics.RedisCacheHits.Inc()
		metrics.RedisOperationsTotal.WithLabelValues("get", "success").Inc()
	}
	return result, err
}

// Delete 删除缓存
func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := r.client.Del(ctx, key).Err()
	duration := time.Since(start).Seconds()

	metrics.RedisOperationDuration.WithLabelValues("del").Observe(duration)
	if err != nil {
		metrics.RedisOperationsTotal.WithLabelValues("del", "error").Inc()
	} else {
		metrics.RedisOperationsTotal.WithLabelValues("del", "success").Inc()
	}
	return err
}

// Exists 检查 key 是否存在
func (r *RedisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	result := r.client.Exists(ctx, key)
	return result.Val() > 0, result.Err()
}

// SetNX 设置缓存（如果不存在）
func (r *RedisCacheRepository) SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Expire 设置过期时间
func (r *RedisCacheRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// int类型
func (r *RedisCacheRepository) SetInt(ctx context.Context, key string, value int64, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCacheRepository) GetInt(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	result, err := r.client.Get(ctx, key).Int64()
	duration := time.Since(start).Seconds()

	metrics.RedisOperationDuration.WithLabelValues("get").Observe(duration)
	if err != nil {
		if err == redis.Nil {
			metrics.RedisCacheMisses.Inc()
			metrics.RedisOperationsTotal.WithLabelValues("get", "miss").Inc()
		} else {
			metrics.RedisOperationsTotal.WithLabelValues("get", "error").Inc()
		}
	} else {
		metrics.RedisCacheHits.Inc()
		metrics.RedisOperationsTotal.WithLabelValues("get", "success").Inc()
	}
	return result, err
}

func (r *RedisCacheRepository) Incr(ctx context.Context, key string) error {
	return r.client.Incr(ctx, key).Err()
}

func (r *RedisCacheRepository) SetNXInt(ctx context.Context, key string, value int64, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}
