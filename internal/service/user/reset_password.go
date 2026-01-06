package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"hello-gozero/internal/svc"
)

type ResetPasswordService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetPasswordService 获取单个用户
func NewResetPasswordService(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordService {
	return &ResetPasswordService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
