// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"hello-gozero/infra/cache"
	"hello-gozero/infra/database"
	"hello-gozero/infra/queue"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Infra Infra `json:"Infra"`
}

// Infra 结构体，包含所有基础设施配置
type Infra struct {
	Mysql database.MysqlConfig `json:"Mysql"`
	Redis cache.RedisConfig    `json:"Redis"`
	Kafka queue.KafkaConfig    `json:"Kafka"`
}
