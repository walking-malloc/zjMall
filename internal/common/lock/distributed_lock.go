package lock

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type DistributedLockService interface {
	AcquireLock(ctx context.Context, key string, expireTime time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
}

// Redis实现
type RedisLockService struct {
	client *redis.Client
}

func NewRedisLockService(client *redis.Client) DistributedLockService {
	return &RedisLockService{client: client}
}

func (r *RedisLockService) AcquireLock(ctx context.Context, key string, expireTime time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, "1", expireTime).Result()
}

func (r *RedisLockService) ReleaseLock(ctx context.Context, key string) error {
	// 使用Lua脚本保证原子性：只有锁的持有者才能释放
	luaScript := `
        local tokenKey = KEYS[1]
        local value = redis.call('GET', tokenKey)
        if value == '1' then
            redis.call('DEL', tokenKey)
            return 1
        else
            return 0
        end
    `
	_, err := r.client.Eval(ctx, luaScript, []string{key}).Result()
	return err
}
