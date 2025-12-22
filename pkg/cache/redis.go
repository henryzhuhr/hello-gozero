package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Password string
	DB       int
}

// MustNewRedis 初始化 Redis 连接，失败时 panic
func MustNewRedis(conf RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Password,
		DB:       conf.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return client
}

// CloseRedis 关闭 Redis 连接
func CloseRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
