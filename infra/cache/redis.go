package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `json:"Addr"` // e.g., "localhost:6379"
	Password string `json:"Password"`
	DB       int    `json:"DB"`

	// 是否启用 TLS/SSL 加密连接（用于安全通信，常见于云 Redis 服务如 AWS ElastiCache、Azure Cache、阿里云等）
	UseTLS bool `json:"UseTLS"`

	//
	InsecureSkipVerify bool `json:"InsecureSkipVerify"`

	// 建立 TCP 连接（包括 TLS 握手）的超时时间。不包括 DNS 解析（go-redis 使用 net.DialTimeout 内部处理）。
	// 典型值：3s ~ 10s。
	// 单位：秒
	DialTimeout int `json:"DialTimeout" comment:"unit: seconds"`

	// 从 Redis 读取响应的超时时间。如果 Redis 服务器响应慢或网络卡顿，超过此时间会报 i/o timeout。
	// 典型值：1s ~ 5s（根据业务容忍度调整）。
	// 如果设为 0 表示无超时（不推荐生产环境使用）。
	// 单位：秒
	ReadTimeout int `json:"ReadTimeout" comment:"unit: seconds"`

	// 含义：向 Redis 发送命令的写入超时时间。一般比 ReadTimeout 短。
	// 典型值：1s。
	// 单位：秒
	WriteTimeout int `json:"WriteTimeout" comment:"unit: seconds"`

	// 连接池中最大空闲连接数（实际上是最大总连接数）。
	// go-redis 的连接池会按需创建连接，直到达到 PoolSize。
	// 典型值：
	// 	- 单机服务：10 ~ 50
	// 	- 高并发服务：100 ~ 500（需结合 Redis 服务器 maxclients 限制）
	PoolSize int `json:"PoolSize"`

	// 默认缓存过期时间，单位秒
	DefaultTTL int `json:"DefaultTTL" comment:"unit: seconds"`

	// 缓存过期时间抖动，单位秒
	DefaultJitter int `json:"DefaultJitter" comment:"unit: seconds"`
}

// Validate 配置文件校验
func (c *RedisConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("redis config is nil")
	}
	if c.Addr == "" {
		return fmt.Errorf("redis addr is empty")
	}
	if _, _, err := net.SplitHostPort(c.Addr); err != nil {
		return fmt.Errorf("redis addr must be in 'host:port' format: %w", err)
	}
	if c.DB < 0 {
		return fmt.Errorf("redis DB must be >= 0")
	}
	if c.DefaultJitter < 0 {
		return fmt.Errorf("jitter must be non-negative")
	}
	if c.DefaultTTL < 0 {
		return fmt.Errorf("ttl must be non-negative")
	}
	return nil
}

// applyRedisConfigDefaults 应用 Redis 配置默认值
// ✅ 为什么用值传递更好？
// 1. 语义清晰：无副作用（No Side Effects）
// 值传递：函数接收的是 conf 的副本，原配置不会被修改。
// 指针传递：函数可能（也容易）修改原始配置，造成隐蔽的副作用。
func applyRedisConfigDefaults(c RedisConfig) RedisConfig {
	if c.DialTimeout <= 0 {
		c.DialTimeout = 5 // 默认 5 秒
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 3 // 默认 3 秒
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 2 // 默认 2 秒
	}
	if c.PoolSize <= 0 {
		c.PoolSize = 10 // 默认 10 个连接
	}

	// DefaultTTL 允许通过校验并在必要时设置合理默认值（单位：秒）
	if c.DefaultTTL <= 0 {
		// 若未配置，使用 300 秒作为默认缓存过期时间
		c.DefaultTTL = 300
	}
	// DefaultJitter 不应为负数
	if c.DefaultJitter < 0 {
		c.DefaultJitter = 0
	}
	// 如果抖动值大于 TTL，将其截断为 TTL 的一半以避免异常行为
	if c.DefaultJitter > c.DefaultTTL {
		c.DefaultJitter = c.DefaultTTL / 2
	}
	return c
}

// RedisInfra 封装 Redis 客户端及默认配置
type RedisInfra struct {
	// Redis 客户端
	Client *redis.Client

	// 默认缓存过期时间
	DefaultTTL time.Duration

	// 缓存过期时间抖动
	DefaultJitter time.Duration
}

// NewRedisInfra 创建 RedisInfra 实例
func NewRedisInfra(ctx context.Context, conf RedisConfig) (*RedisInfra, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid redis config: %w", err)
	}
	conf = applyRedisConfigDefaults(conf)

	var tlsConfig *tls.Config
	if conf.UseTLS {
		tlsConfig = &tls.Config{
			// 如果不需要证书验证（如内网自签名），可加：
			InsecureSkipVerify: true,
		}
	}
	// 构建 redis.Options（只构建一次，避免重复分配）
	opts := &redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.DB,
		TLSConfig:    tlsConfig,
		DialTimeout:  time.Duration(conf.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.WriteTimeout) * time.Second,
		PoolSize:     conf.PoolSize,
	}

	var (
		client *redis.Client
		infra  *RedisInfra
	)

	// 使用 retry-go 重试 Ping
	err := retry.Do(
		func() error {
			// 每次重试创建新 client（避免连接污染）
			client = redis.NewClient(opts)

			pingTimeout := time.Duration(conf.DialTimeout) * time.Second
			if pingTimeout == 0 {
				pingTimeout = time.Second
			}

			ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
			defer cancel()

			if err := client.Ping(ctx).Err(); err != nil {
				// 关闭失败的 client，避免 goroutine 泄漏
				_ = client.Close()
				return err // retry-go 会捕获并重试
			}

			infra = &RedisInfra{
				Client:        client,
				DefaultTTL:    time.Duration(conf.DefaultTTL) * time.Second,
				DefaultJitter: time.Duration(conf.DefaultJitter) * time.Second,
			}
			return nil
		},
		retry.Context(ctx), // 传递外部 context，支持取消/超时
		retry.Attempts(3),
		retry.Delay(1*time.Second),          // 初始延迟
		retry.MaxDelay(5*time.Second),       // 最大延迟（自动指数退避）
		retry.DelayType(retry.BackOffDelay), // 指数退避
		// 👇 关键：只重试临时错误
		retry.RetryIf(shouldRetryRedisError), // ✅ 精准重试,
	)

	if err != nil {
		return nil, fmt.Errorf("redis init failed after retries: %w", err)
	}

	return infra, nil
}

// Close 关闭 Redis 连接
func (r *RedisInfra) Close() error {
	if r.Client == nil {
		return nil
	}
	return r.Client.Close()
}

// shouldRetryRedisError 判断 Redis 错误是否可重试（仅限临时性故障）
func shouldRetryRedisError(err error) bool {
	if err == nil {
		return false
	}
	// 永久性错误：绝不重试
	if redis.IsAuthError(err) || redis.IsPermissionError(err) || redis.IsOOMError(err) || redis.IsExecAbortError(err) {
		return false
	}

	// Redis 服务端临时状态（可重试）
	if redis.IsLoadingError(err) || // Redis 正在加载 RDB/AOF
		redis.IsTryAgainError(err) || // 服务端建议重试
		redis.IsClusterDownError(err) || // 集群暂时不可用
		redis.IsMasterDownError(err) || // 主节点暂时不可用
		redis.IsMaxClientsError(err) { // 客户端数满（可能瞬时）
		return true
	}

	// 网络错误：基于 net.Error 判断超时/临时网络故障可重试
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
		// 如果实现了 Temporary() 并返回 true，也视为可重试
		type temporary interface{ Temporary() bool }
		if te, ok := netErr.(temporary); ok && te.Temporary() {
			return true
		}
	}

	// context 超时/取消 不应重试
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	// 其他 Redis 协议错误通常表示客户端逻辑错误，不应重试
	var rErr redis.Error
	if errors.As(err, &rErr) {
		return false
	}

	return false
}
