// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"hello-gozero/infra/cache"
	"hello-gozero/infra/database"
	"hello-gozero/infra/queue"
	"hello-gozero/internal/config"
	userRepo "hello-gozero/internal/repository/user"
)

type ServiceContext struct {
	// 全局日志
	Logger logx.Logger
	Config config.Config

	MysqlConn   *gorm.DB
	RedisClient *redis.Client
	KafkaWriter *kafka.Writer
	KafkaReader *kafka.Reader

	// Repository
	Repository Repository
}

// Repository 结构体，包含所有仓库接口
type Repository struct {
	// 用户仓库
	User userRepo.UserRepository
	// 用户仓库（带缓存的装饰器，用于特殊场景，如：防重复提交、限流）
	CachedUser userRepo.CachedUserRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	logger := logx.WithContext(context.Background())
	mysqlConn := database.MustNewMysql(c.Mysql.DataSource, logger)
	redisClient := cache.MustNewRedis(c.Redis)

	// 初始化仓库
	user := userRepo.NewUserRepository(mysqlConn)
	RedisDefaultTTL := time.Duration(c.Redis.DefaultTTL) * time.Second
	RedisDefaultJitter := time.Duration(c.Redis.DefaultJitter) * time.Second

	return &ServiceContext{
		Logger:      logger,
		Config:      c,
		MysqlConn:   mysqlConn,
		RedisClient: redisClient,
		KafkaWriter: queue.MustNewKafkaWriter(c.Kafka),
		KafkaReader: queue.MustNewKafkaReader(c.Kafka),
		Repository: Repository{
			User:       user,
			CachedUser: userRepo.NewCachedUserRepository(redisClient, user, RedisDefaultTTL, RedisDefaultJitter),
		},
	}
}

// Close 关闭所有资源连接
func (sc *ServiceContext) Close() error {
	if err := database.CloseMysql(sc.MysqlConn); err != nil {
		logx.Errorf("Failed to close MySQL: %v", err)
	}
	if err := cache.CloseRedis(sc.RedisClient); err != nil {
		logx.Errorf("Failed to close Redis: %v", err)
	}
	if err := queue.CloseKafkaWriter(sc.KafkaWriter); err != nil {
		logx.Errorf("Failed to close Kafka writer: %v", err)
	}
	if err := queue.CloseKafkaReader(sc.KafkaReader); err != nil {
		logx.Errorf("Failed to close Kafka reader: %v", err)
	}
	return nil
}
