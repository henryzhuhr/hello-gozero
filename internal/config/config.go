// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

// Package config provides configuration structures for the application.
package config

import (
	"hello-gozero/pkg/infra/cache"
	"hello-gozero/pkg/infra/database"
	"hello-gozero/pkg/infra/queue"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Infra Infra       `json:"Infra"`
	Pprof PprofConfig `json:"Pprof,optional"`

	LocaleDir string `json:"LocaleDir,optional"` // 国际化资源文件目录
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
