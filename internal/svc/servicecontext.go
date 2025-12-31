// Package svc provides service context and dependency injection for the application.
package svc

import (
	"context"
	"fmt"

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
	// 全局配置
	Config config.Config
	// 全局日志
	Logger logx.Logger

	// Infra 基础设施配置
	Infra Infra

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

// Infra 结构体，包含所有基础设施连接
type Infra struct {
	// 数据库连接
	MysqlConn *gorm.DB

	// Redis 基础设施封装
	Redis *cache.RedisInfra

	// Kafka 生产者
	KafkaWriter *kafka.Writer
	// Kafka 消费者
	KafkaReader *kafka.Reader
}

// NewServiceContext 创建全局服务上下文实例。
// 返回错误时，调用方应处理该错误（如记录日志并退出程序）
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	ctx := context.Background()
	// 初始化日志
	logger := logx.WithContext(ctx)

	// 初始化 MySQL 连接
	mysqlConn, err := database.NewMySQL(c.Infra.Mysql, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init mysql: %w", err)
	}

	// 初始化 Redis 客户端
	redisInfra, err := cache.NewRedisInfra(ctx, c.Infra.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to init redis: %w", err)
	}

	// 初始化 Kafka 读写器
	kafkaWriter, err := queue.NewKafkaWriter(c.Infra.Kafka)
	if err != nil {
		return nil, fmt.Errorf("failed to init kafka writer: %w", err)
	}
	kafkaReader, err := queue.NewKafkaReader(c.Infra.Kafka)
	if err != nil {
		return nil, fmt.Errorf("failed to init kafka reader: %w", err)
	}

	// 初始化仓库
	user := userRepo.NewUserRepository(mysqlConn)
	cachedUser := userRepo.NewCachedUserRepository(redisInfra, user)

	return &ServiceContext{
		Config: c,
		Logger: logger,
		Infra: Infra{
			MysqlConn:   mysqlConn,
			Redis:       redisInfra,
			KafkaWriter: kafkaWriter,
			KafkaReader: kafkaReader,
		},
		Repository: Repository{
			User:       user,
			CachedUser: cachedUser,
		},
	}, nil
}

// Close 关闭所有资源连接
func (sc *ServiceContext) Close() error {
	if err := database.CloseMysql(sc.Infra.MysqlConn); err != nil {
		return fmt.Errorf("failed to close MySQL: %v", err)
	}
	if err := sc.Infra.Redis.Close(); err != nil {
		return fmt.Errorf("failed to close Redis: %v", err)
	}
	if err := queue.CloseKafkaWriter(sc.Infra.KafkaWriter); err != nil {
		return fmt.Errorf("failed to close Kafka writer: %v", err)
	}
	if err := queue.CloseKafkaReader(sc.Infra.KafkaReader); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %v", err)
	}
	return nil
}
