// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	userDto "hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
)

type DeleteUserService struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteUserService 删除用户
func NewDeleteUserService(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserService {
	return &DeleteUserService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *DeleteUserService) GetCtx() context.Context {
	return l.ctx
}
func (l *DeleteUserService) DeleteUser(req *userDto.DeleteUserReq) (resp *userDto.DeleteUserResp, err error) {

	// 删除用户
	err = l.svcCtx.Repository.User.DeleteByUsername(l.ctx, req.Username)
	if err != nil {
		return &userDto.DeleteUserResp{}, nil
	}

	// 检查缓存，如果存在则删除缓存
	_, err = l.svcCtx.Repository.CachedUser.GetByUsername(l.ctx, req.Username)

	return &userDto.DeleteUserResp{}, nil
}
