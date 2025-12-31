// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	userDto "hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
)

type GetUserListService struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUserListService 获取用户列表
func NewGetUserListService(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListService {
	return &GetUserListService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *GetUserListService) GetCtx() context.Context {
	return l.ctx
}
func (l *GetUserListService) GetUserList(req *userDto.GetUserListReq) (resp *userDto.GetUserListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
