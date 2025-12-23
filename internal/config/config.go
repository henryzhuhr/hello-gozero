// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"hello-gozero/infra/cache"
	"hello-gozero/infra/queue"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Mysql struct {
		DataSource string `json:"DataSource"`
	} `json:"Mysql"`
	Redis cache.RedisConfig `json:"Redis"`
	Kafka queue.KafkaConfig `json:"Kafka"`
}
