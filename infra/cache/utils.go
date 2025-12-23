package cache

import (
	"math/rand"
	"time"
)

// RandomTTL 生成一个带有随机抖动的 TTL，防止缓存雪崩
func RandomTTL(base time.Duration, jitter time.Duration) time.Duration {
	return base + time.Duration(rand.Int63n(int64(jitter)))
}
