package user

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"hello-gozero/infra/cache"
	userEntity "hello-gozero/internal/entity/user"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	cacheKeyPrefix   = "user:info"
	cacheEmptyTTL    = 60 * time.Second // 缓存空对象的 TTL
	cachedEmptyValue = "null"           // 缓存空对象的特殊标记
)

// CachedUserRepository 定义用户缓存接口
type CachedUserRepository interface {
	// GetByUsername 从缓存获取用户
	GetByUsername(ctx context.Context, username string) (*userEntity.User, error)

	// SetByUsername 将用户信息写入缓存
	SetByUsername(ctx context.Context, user *userEntity.User) error
}

// CachedUserRepositoryImpl Implements [CachedUserRepository]
type CachedUserRepositoryImpl struct {
	// Redis 客户端
	client *redis.Client

	// 包装底层 DB repo
	repo UserRepository

	group singleflight.Group // ← 新增

	ttl    time.Duration
	jitter time.Duration
}

// NewCachedUserRepository
// Creates a new CachedUserRepository instance
// Parameters:
//   - client: Redis 客户端实例
//   - repo: 底层 UserRepository 实例
//   - ttl: 缓存默认过期时间
//   - jitter: 缓存过期时间抖动，防止缓存雪崩
func NewCachedUserRepository(client *redis.Client, repo UserRepository, ttl time.Duration, jitter time.Duration) CachedUserRepository {
	return &CachedUserRepositoryImpl{
		client: client,
		repo:   repo,
		ttl:    ttl,
		jitter: jitter,
	}
}

// GetByUsername Implements [CachedUserRepository.GetByUsername]
//
// 如果缓存命中且成功反序列化，则直接返回用户；
// 如果缓存未命中、反序列化失败或缓存错误，则回源到底层数据库仓库（c.repo）查询，
// 并在查询成功后异步（此处为同步）回写（cache-aside 模式）到缓存中。
// 注意：缓存反序列化失败不会中断流程，会自动降级到数据库。
//
// gob 是 Go 标准库提供的二进制编码格式，专为 Go 设计。
// 项目是纯 Go 服务（无其他语言读缓存），不存在多语言系统（Go + Python/Java），所以选择 gob。
// 如果需要跨语言支持，建议使用 JSON、MessagePack、Protobuf 等通用格式。
func (c *CachedUserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*userEntity.User, error) {
	// 尝试从缓存中读取数据
	key := cacheKeyPrefix + ":" + username
	val, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		// 检查是否是空值标记
		if string(val) == cachedEmptyValue {
			return nil, gorm.ErrRecordNotFound
		}

		// 缓存命中，尝试使用 gob 反序列化为 User 对象
		var user userEntity.User
		buf := bytes.NewBuffer(val)
		if err := gob.NewDecoder(buf).Decode(&user); err == nil {
			// 反序列化成功，直接返回缓存中的用户
			return &user, nil
		}
		// 反序列化失败（如缓存数据损坏或结构变更），继续回源查询
	}

	// Cache miss or error, fallback to DB
	// 缓存未命中或反序列化失败，回源到数据库
	// ⚡ 使用 singleflight：相同 username 的请求会等待首个 DB 查询结果
	result, err, _ := c.group.Do(username, func() (interface{}, error) {
		dbUser, dbErr := c.repo.GetByUsername(ctx, username)
		if dbErr != nil {
			// 如果是“用户不存在”错误，我们缓存空值
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				// 将空值写入缓存（带短 TTL）
				// 注意：这里不能在回调里直接调 c.Set...，因为可能阻塞 singleflight
				// 更安全的方式：让外层处理缓存写入
				return nil, dbErr // 外层判断是否为 ErrUserNotFound
			}
			return nil, dbErr
		}
		return dbUser, nil
	})
	if err != nil {
		// 如果是“用户不存在”，缓存空值
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 缓存空值，TTL 较短（如 60 秒）
			_ = c.setEmptyUserCache(ctx, username, 60*time.Second)
		}
		// 数据库查询失败，直接返回错误（不缓存错误）
		return nil, err
	}

	// 防御性编程：确保 result 是预期类型
	user, ok := result.(*userEntity.User)
	if !ok {
		// 可能是 panic 被 recover、返回了错误类型、或 nil
		if result == nil {
			return nil, nil // 或 errors.New("user is nil")
		}
		return nil, fmt.Errorf("unexpected result type from singleflight: %T", result)
	}
	// 此时 user 可能为 nil（如果 repo 返回了 (*User)(nil)）
	if user == nil {
		return nil, nil
	}

	// Write back to cache
	// 查询成功，将用户数据写入缓存（用于后续请求加速）
	// 注意：这里忽略写缓存的错误，避免因缓存故障影响主业务流程
	_ = c.SetByUsername(ctx, user)

	return user, nil
}

// SetByUsername Implements [CachedUserRepository.SetByUsername]
// 使用 gob 编码以支持任意 Go 结构体（包括非导出字段），但要求接收方结构一致。
// 若 user 为 nil，则跳过写入（避免缓存空对象，除非你明确需要空值缓存）。
// 缓存有效期由 cacheTTL 全局控制。
func (c *CachedUserRepositoryImpl) SetByUsername(ctx context.Context, user *userEntity.User) error {
	if user == nil {
		// 不缓存 nil 值，防止缓存穿透（除非业务需要空值缓存）
		return nil
	}

	// 使用 gob 将 user 序列化为字节流

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(user); err != nil {
		// 序列化失败，返回错误（通常因结构包含不可 gob 编码的类型）
		return err
	}

	key := c.getCachedKey(user.Username)
	// 写入 Redis，设置过期时间（cacheTTL）
	// 	二、缓存雪崩（Cache Avalanche）
	// 🔍 问题表现
	// 大量 key 在同一时间过期（如服务重启后批量加载缓存，TTL 相同）。
	// 缓存集体失效 → 所有请求打到数据库 → DB 连接池耗尽、CPU 打满。
	// 方案：随机 TTL（TTL jitter），缓存过期时间均匀分布，避免集体失效。
	return c.client.Set(ctx, key, buf.Bytes(), cache.RandomTTL(c.ttl, c.jitter)).Err()
}

// setEmptyUserCache 缓存一个“空用户”标记，防止缓存穿透
func (c *CachedUserRepositoryImpl) setEmptyUserCache(ctx context.Context, username string, ttl time.Duration) error {
	key := c.getCachedKey(username)
	// 方式 1：存一个特殊字符串
	return c.client.Set(ctx, key, cachedEmptyValue, ttl).Err()

	// 方式 2：存一个 gob 编码的 nil 或空结构（需 Get 时兼容）
	// var buf bytes.Buffer
	// gob.NewEncoder(&buf).Encode((*userEntity.User)(nil))
	// return c.client.Set(ctx, key, buf.Bytes(), ttl).Err()
}

func (c *CachedUserRepositoryImpl) getCachedKey(username string) string {
	return cacheKeyPrefix + ":" + username
}
