// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	userDto "hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
)

type GetUserService struct {
	logx.Logger
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

func (l *GetUserService) GetUser(req *userDto.GetUserReq) (resp *userDto.GetUserResp, err error) {
	// todo: add your logic here and delete this line

	return
}
