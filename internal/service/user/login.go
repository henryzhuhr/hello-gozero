// Package user 用户相关服务
package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	userDto "hello-gozero/internal/dto/user"
	userEntity "hello-gozero/internal/entity/user"
	"hello-gozero/internal/svc"
)

const (
	defaultMFATokenExpiredTime = 5 // 分钟
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

	// 获取用户信息
	user, err := l.svcCtx.Repository.User.GetByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 检查账户状态
	if user.Status == userEntity.UserStatusDisabled {
		return nil, ErrAccountDisabled
	}

	currentMethod := userDto.MFAMethodSMS // 默认 MFA 方式为 SMS
	maskedValue := user.PhoneNumber

	// 目前暂不支持 MFA，直接返回登录成功
	// TODO: 后续需要实现 MFA 功能时，需要：
	// 1. 在 User 实体中添加 MFA 相关字段（如 mfa_enabled） 
	// 2. 检查用户是否启用了 MFA
	// 3. 如果启用，生成 MFA token 并发送验证码，返回 mfa_required 状态
	// 4. 如果未启用，直接返回 success 状态
	return &userDto.LoginResp{
		Status:              userDto.AuthStatusSuccess,
		MFAToken:            "",
		CurrentMethod:       currentMethod,
		MFATokenExpiredTime: defaultMFATokenExpiredTime,
		MaskedValue:         maskedValue,
		AvailableMethods:    []userDto.MFAMethod{},
	}, nil
}
