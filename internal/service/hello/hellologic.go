// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	helloDto "hello-gozero/internal/dto/hello"
	"hello-gozero/internal/svc"
)

type HelloService struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHelloService(ctx context.Context, svcCtx *svc.ServiceContext) *HelloService {
	return &HelloService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HelloService) Hello() (resp *helloDto.Response, err error) {
	// todo: add your logic here and delete this line
	l.Logger.Infof("hello: logic 调用成功")
	return
}
