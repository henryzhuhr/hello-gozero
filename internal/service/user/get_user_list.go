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
	// 计算 offset 和 limit
	offset := (req.Page - 1) * req.PageSize
	limit := req.PageSize

	// 从仓库获取用户列表
	var users []*userEntity.User
	var total int64
	if req.Status != nil {
		users, total, err = l.svcCtx.Repository.User.ListWithStatus(l.ctx, offset, limit, *req.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to get user list: %w", err)
		}
	} else {
		users, total, err = l.svcCtx.Repository.User.List(l.ctx, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get user list: %w", err)
		}
	}

	// 转化为 DTO 格式
	userDtos := make([]userDto.User, 0, len(users))
	for _, user := range users {
		dto := userDto.User{
			Username:         user.Username,
			Email:            user.Email,
			PhoneCountryCode: user.PhoneCountryCode,
			PhoneNumber:      user.PhoneNumber,
			Nickname:         user.Nickname,
			Status:           int(user.Status),
		}
		if user.LastLoginTime != nil {
			dto.LastLoginTime = user.LastLoginTime.Format(time.DateTime)
		}
		userDtos = append(userDtos, dto)
	}

	return &userDto.GetUserListResp{List: userDtos, Total: int(total)}, nil
}
