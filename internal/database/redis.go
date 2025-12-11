package database

import (
	"context"
	"fmt"
	"time"
	"zjMall/internal/config"

	"github.com/go-redis/redis/v8"
	// adjust to the actual import path of your config package
)

var RedisClient *redis.Client

func InitRedis(config *config.RedisConfig) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
	})
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//测试redis连接
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis连接测试失败: %v", err)
	}
	fmt.Println("Redis 连接成功")
	RedisClient = redisClient
	return redisClient, nil
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
