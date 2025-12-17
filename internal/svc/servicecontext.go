// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"hello-gozero/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Logger logx.Logger
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Logger: logx.WithContext(context.Background()),
		Config: c,
	}
}
