// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"context"

	"hello-gozero/internal/svc"
	"hello-gozero/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HelloLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHelloLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HelloLogic {
	return &HelloLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HelloLogic) Hello() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	l.Logger.Infof("hello: logic 调用成功")
	return
}
