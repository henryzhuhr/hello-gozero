// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	helloDto "hello-gozero/internal/dto/hello"
	"hello-gozero/internal/svc"
)

type HealthService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthService(ctx context.Context, svcCtx *svc.ServiceContext) *HealthService {
	return &HealthService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthService) Health() (resp *helloDto.Response, err error) {
	// todo: add your logic here and delete this line
	l.Logger.Infof("health: logic 调用成功")

	return
}
