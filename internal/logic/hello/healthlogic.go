// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"context"

	"hello-gozero/internal/svc"
	"hello-gozero/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	l.Logger.Infof("health: logic 调用成功")

	return
}
