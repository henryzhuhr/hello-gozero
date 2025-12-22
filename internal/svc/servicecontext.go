// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"hello-gozero/internal/config"
	"hello-gozero/pkg/cache"
	"hello-gozero/pkg/database"
	"hello-gozero/pkg/queue"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	// 全局日志
	Logger logx.Logger
	Config config.Config

	MysqlConn   *gorm.DB
	RedisClient *redis.Client
	KafkaWriter *kafka.Writer
	KafkaReader *kafka.Reader
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Logger:      logx.WithContext(context.Background()),
		Config:      c,
		MysqlConn:   database.MustNewMysql(c.Mysql.DataSource),
		RedisClient: cache.MustNewRedis(c.Redis),
		KafkaWriter: queue.MustNewKafkaWriter(c.Kafka),
		KafkaReader: queue.MustNewKafkaReader(c.Kafka),
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
