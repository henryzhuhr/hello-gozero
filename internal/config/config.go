// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

// Package config provides configuration structures for the application.
package config

import (
	"hello-gozero/infra/cache"
	"hello-gozero/infra/database"
	"hello-gozero/infra/queue"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf                  // 匿名嵌入
	zrpc.RpcServerConf             // 匿名嵌入
	Infra              Infra       `json:"Infra"`
	Pprof              PprofConfig `json:"Pprof,optional"`
}

// PprofConfig pprof性能分析配置
type PprofConfig struct {
	Enabled bool `json:"Enabled,default=false"` // 是否启用 pprof
	Port    int  `json:"Port,default=6060"`     // pprof 服务端口
}

// Infra 结构体，包含所有基础设施配置
type Infra struct {
	Mysql database.MysqlConfig `json:"Mysql"`
	Redis cache.RedisConfig    `json:"Redis"`
	Kafka queue.KafkaConfig    `json:"Kafka"`
}
