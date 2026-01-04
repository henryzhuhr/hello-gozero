// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"

	userDto "hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
)

const (
	// 延迟双删的延迟时间
	// 这个时间应该大于一次数据库写操作的时间
	cacheDeleteDelay = 500 * time.Millisecond

	// 第二次删除缓存的超时时间（秒）
	secondDeleteTimeoutSec = 3
)

type DeleteUserService struct {
	Logger logx.Logger
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

	// 延迟双删策略（Delayed Double Delete）
	// 目的：解决并发场景下的缓存一致性问题
	//
	// 场景：删除数据库期间，有并发读请求可能将旧数据重新写入缓存
	// 方案：
	// 1. 第一次删除缓存：确保后续读请求回源数据库
	// 2. 删除数据库：执行主要删除操作
	// 3. 延迟后再次删除缓存：清除可能在步骤1-2之间被并发请求写入的旧数据

	// 第一次删除缓存
	err = l.svcCtx.Repository.CachedUser.DeleteByUsername(l.ctx, req.Username)
	if err != nil {
		// 缓存删除失败，返回错误，不继续删除数据库
		return nil, fmt.Errorf("first cache delete failed for user(%s): %v", req.Username, err)
	}
	l.Logger.Debugf("first cache delete success for user(%s)", req.Username)

	// 删除数据库中的用户
	err = l.svcCtx.Repository.User.DeleteByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user(%s) from database: %v", req.Username, err)
	}
	l.Logger.Debugf("database delete success for user(%s)", req.Username)

	// 延迟双删：异步延迟后再次删除缓存
	// 使用 goroutine 异步执行，不阻塞主流程
	threading.GoSafe(func() {
		// 延迟一段时间（通常 100-500ms）
		// 这个时间应该大于一次数据库写操作的时间
		time.Sleep(cacheDeleteDelay)

		// 第二次删除缓存，带超时控制，避免Goroutine 泄漏（超时确保 goroutine 能正常退出）
		ctx, cancel := context.WithTimeout(context.Background(), secondDeleteTimeoutSec*time.Second)
		defer cancel()

		err := l.svcCtx.Repository.CachedUser.DeleteByUsername(ctx, req.Username)
		if err != nil {
			// 记录错误但不影响主流程（主流程已经返回成功）
			l.Logger.Errorf("second cache delete failed for user(%s): %v", req.Username, err)
		} else {
			l.Logger.Debugf("second cache delete success for user(%s)", req.Username)
		}
	})

	return &userDto.DeleteUserResp{}, nil
}
