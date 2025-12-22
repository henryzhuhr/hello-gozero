// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"hello-gozero/pkg/cache"
	"hello-gozero/pkg/queue"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Mysql struct {
		DataSource string
	}
	Redis cache.RedisConfig
	Kafka queue.KafkaConfig
}
