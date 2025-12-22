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

// 删除用户
func NewDeleteUserService(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserService {
	return &DeleteUserService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserService) DeleteUser(req *userDto.DeleteUserReq) (resp *userDto.DeleteUserResp, err error) {
	// todo: add your logic here and delete this line

	return
}
