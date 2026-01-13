// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	userDto "hello-gozero/internal/dto/user"
	userEntity "hello-gozero/internal/entity/user"
	"hello-gozero/internal/svc"
)

type GetUserCountService struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUserCountService 创建获取用户总数服务
func NewGetUserCountService(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserCountService {
	return &GetUserCountService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserCountService) GetCtx() context.Context {
	return l.ctx
}

func (l *GetUserCountService) GetUserCount(req *userDto.GetUserCountReq) (resp *userDto.GetUserCountResp, err error) {
	// 获取活跃用户数（不包含软删除）
	activeCount, err := l.svcCtx.Repository.User.CountWithStatus(l.ctx, userEntity.UserStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	resp = &userDto.GetUserCountResp{
		Active: activeCount,
	}

	// 如果需要包含已删除的用户统计
	if req.IncludeDeleted {
		// 获取总用户数（包含软删除）
		totalCount, err := l.svcCtx.Repository.User.Count(l.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count total users: %w", err)
		}

		resp.Total = totalCount
		resp.Deleted = totalCount - activeCount
	} else {
		// 不包含删除用户时，总数等于活跃用户数
		resp.Total = activeCount
	}

	return resp, nil
}
