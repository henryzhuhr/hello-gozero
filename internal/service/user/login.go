// Package user 用户相关服务
package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	userDto "hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
)

type LoginService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginService 用户登录
func NewLoginService(ctx context.Context, svcCtx *svc.ServiceContext) *LoginService {
	return &LoginService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginService) GetCtx() context.Context {
	return l.ctx
}

func (l *LoginService) Login(req *userDto.LoginReq) (resp *userDto.LoginResp, err error) {
	if req == nil || req.Username == "" {
		return nil, ErrMissingUsername
	}
	return &userDto.LoginResp{}, nil
}
