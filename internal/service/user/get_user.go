// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	userDto "hello-gozero/internal/dto/user"
	userEntity "hello-gozero/internal/entity/user"
	"hello-gozero/internal/svc"
)

type GetUserService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取单个用户
func NewGetUserService(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserService {
	return &GetUserService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *GetUserService) GetCtx() context.Context {
	return l.ctx
}

func (l *GetUserService) GetUser(req *userDto.GetUserReq) (resp *userDto.GetUserResp, err error) {
	if req == nil || req.Username == "" {
		return nil, fmt.Errorf("missing username")
	}

	cachedEntity, err := l.svcCtx.Repository.CachedUser.GetByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	l.ctx = logx.ContextWithFields(l.ctx, logx.Field("source", cachedEntity.DataSource))
	l.Logger.WithContext(l.ctx).Debugf("GetUser: fetched user '%s' from %s", req.Username, cachedEntity.DataSource)
	if cachedEntity.User == nil {
		// 用户不存在
		return nil, nil
	}

	return l.userEntityToResp(cachedEntity.User), nil
}

// userEntityToResp 将用户实体转换为响应 DTO
func (l *GetUserService) userEntityToResp(user *userEntity.User) *userDto.GetUserResp {
	var lastLogin string
	if user.LastLoginTime != nil {
		lastLogin = user.LastLoginTime.Format(time.RFC3339)
	}

	return &userDto.GetUserResp{
		User: userDto.User{
			Username:         user.Username,
			Email:            user.Email,
			PhoneCountryCode: user.PhoneCountryCode,
			PhoneNumber:      user.PhoneNumber,
			Nickname:         user.Nickname,
			Status:           int(user.Status),
			LastLoginTime:    lastLogin,
		},
	}
}
