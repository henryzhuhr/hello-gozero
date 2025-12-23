package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `json:"Host"`
	Password string `json:"Password"`
	DB       int    `json:"DB"`

	// 默认缓存过期时间，单位秒
	DefaultTTL int `json:"DefaultTTL"`
	// 缓存过期时间抖动，单位秒
	DefaultJitter int `json:"DefaultJitter"`
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
