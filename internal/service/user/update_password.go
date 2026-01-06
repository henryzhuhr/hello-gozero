package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	"hello-gozero/infra/cache"
	"hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
	passwordUtil "hello-gozero/internal/utils/password"
)

type UpdatePasswordService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdatePasswordService 更新用户密码
func NewUpdatePasswordService(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordService {
	return &UpdatePasswordService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (s *UpdatePasswordService) GetCtx() context.Context {
	return s.ctx
}

// UpdatePassword 更新用户密码
func (s *UpdatePasswordService) UpdatePassword(req *user.UpdatePasswordReq) (*user.UpdatePasswordResp, error) {
	if req.Username == "" {
		return nil, ErrMissingUsername
	}

	// ============================================================
	// 使用 Redis 分布式锁避免并发修改密码冲突
	// ============================================================
	// 优势：
	//   1. 防止同一用户的密码被并发修改导致数据不一致
	//   2. 在应用层面序列化并发请求，保证修改的原子性
	//   3. 避免读取-修改-写入的竞态条件
	// 锁的粒度：基于用户名
	lockKey := fmt.Sprintf("lock:user:password:%s", req.Username)
	lockValue := uuid.New().String() // 锁的唯一标识
	lockTTL := 10 * time.Second      // 锁的过期时间（防止死锁）

	// 使用 Redis 锁保护密码更新逻辑
	err := cache.WithLock(s.ctx, s.svcCtx.Infra.Redis.Client, lockKey, lockValue, lockTTL, func() error {
		return s.updatePasswordWithinLock(req)
	})
	if err != nil {
		return nil, err
	}

	return &user.UpdatePasswordResp{
		Message: "password updated successfully",
	}, nil
}

// updatePasswordWithinLock 在分布式锁保护下执行密码更新逻辑
func (s *UpdatePasswordService) updatePasswordWithinLock(req *user.UpdatePasswordReq) error {
	// 查找用户
	existUser, err := s.svcCtx.Repository.User.GetByUsername(s.ctx, req.Username)
	if err != nil {
		return fmt.Errorf("failed to get user by name(%s): %w", req.Username, err)
	}
	if existUser == nil {
		return ErrUserNotFound
	}

	// 对比旧密码（使用bcrypt比较哈希值）
	err = bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(req.OldPassword))
	if err != nil {
		return ErrOldPasswordMismatch
	}

	// 检查新密码是否与旧密码相同
	if req.OldPassword == req.NewPassword {
		return ErrNewPasswordSameAsOld
	}

	// 密码检查
	checker := passwordUtil.NewDefaultPasswordChecker()
	if err := checker.Check(req.NewPassword); err != nil {
		return fmt.Errorf("new password does not meet complexity requirements: %w", err)
	}

	// 对新密码进行哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新新的密码
	existUser.Password = string(hashedPassword)
	if err := s.svcCtx.Repository.User.Update(s.ctx, existUser); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}
