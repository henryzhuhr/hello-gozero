// Package user 用户相关服务
package user

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

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

	// 加密密码（在事务外处理，避免事务过长）
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	hashedPassword := string(hashedPasswordBytes)

	user, err := l.svcCtx.Repository.User.GetByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user.Password != hashedPassword {
		return nil, ErrInvalidCredentials
	}

	return &userDto.LoginResp{
		
	}, nil
}
