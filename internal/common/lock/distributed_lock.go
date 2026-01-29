package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	// Lock 尝试获取锁，如果获取失败返回错误
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	// Unlock 释放锁
	Unlock(ctx context.Context, key string) error
	// TryLock 尝试获取锁，立即返回结果（不阻塞）
	TryLock(ctx context.Context, key string, expiration time.Duration) (bool, error)
}

// RedisDistributedLock Redis 实现的分布式锁
type RedisDistributedLock struct {
	client *redis.Client
}

// NewDistributedLock 创建分布式锁实例
func NewDistributedLock(client *redis.Client) DistributedLock {
	return &RedisDistributedLock{
		client: client,
	}
}

// Lock 尝试获取锁（使用 SET NX EX）
func (l *RedisDistributedLock) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)
	result, err := l.client.SetNX(ctx, lockKey, "1", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("获取锁失败: %w", err)
	}
	return result, nil
}

// TryLock 尝试获取锁（立即返回，不阻塞）
func (l *RedisDistributedLock) TryLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return l.Lock(ctx, key, expiration)
}

// Unlock 释放锁
func (l *RedisDistributedLock) Unlock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	if err := l.client.Del(ctx, lockKey).Err(); err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}
	return nil
}

