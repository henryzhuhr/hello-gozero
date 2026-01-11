// Package cache/lock.go 基于 Redis 的分布式锁实现
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockFailed 获取锁失败
	ErrLockFailed = errors.New("failed to acquire lock")

	// ErrLockNotHeld 锁未持有或已过期
	ErrLockNotHeld = errors.New("lock not held")
)

// DistributedLock Redis 分布式锁
type DistributedLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

// NewDistributedLock 创建分布式锁
// key: 锁的唯一标识
// value: 锁的持有者标识（用于释放时验证，防止误删）
// ttl: 锁的过期时间（防止死锁）
func NewDistributedLock(client *redis.Client, key, value string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		value:  value,
		ttl:    ttl,
	}
}

// TryLock 尝试获取锁（非阻塞）
// 返回 true 表示成功获取锁，false 表示锁已被其他人持有
func (l *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	// 使用 SET key value NX EX seconds 实现原子操作
	// NX: 只在键不存在时设置
	// EX: 设置过期时间（秒）
	success, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return success, nil
}

// Lock 阻塞式获取锁（带重试）
// maxRetries: 最大重试次数
// retryInterval: 重试间隔
func (l *DistributedLock) Lock(ctx context.Context, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		success, err := l.TryLock(ctx)
		if err != nil {
			return err // TODO 为什么这里失败直接 return 而不是重试
		}
		if success {
			return nil
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// 继续重试
		}
	}
	return ErrLockFailed
}

// Unlock 释放锁（使用 Lua 脚本保证原子性）
// 只有锁的持有者才能释放锁
func (l *DistributedLock) Unlock(ctx context.Context) error {
	// Lua 脚本：检查 value 是否匹配，匹配则删除
	// 保证原子性，避免释放其他人的锁
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, luaScript, []string{l.key}, l.value).Result()
	if err != nil {
		return fmt.Errorf("failed to unlock: %w", err)
	}

	// result == 0 表示锁不存在或已被其他人持有
	if result == int64(0) {
		return ErrLockNotHeld
	}

	return nil
}

// Extend 延长锁的过期时间
func (l *DistributedLock) Extend(ctx context.Context, additionalTTL time.Duration) error {
	// Lua 脚本：检查 value 是否匹配，匹配则延长过期时间
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, luaScript, []string{l.key}, l.value, int(additionalTTL.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result == int64(0) {
		return ErrLockNotHeld
	}

	return nil
}

// WithLock 使用锁保护的函数执行（自动获取和释放锁）
// fn: 需要在锁保护下执行的函数
func WithLock(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration, fn func() error) error {
	lock := NewDistributedLock(client, key, value, ttl)

	// 尝试获取锁（带重试）
	if err := lock.Lock(ctx, 5, 50*time.Millisecond); err != nil {
		return err
	}

	// 确保释放锁
	defer func() {
		// 使用新的 context 释放锁，避免因原 context 取消导致锁无法释放
		unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = lock.Unlock(unlockCtx)
	}()

	// 执行业务逻辑
	return fn()
}
