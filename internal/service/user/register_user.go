// Package user provides user-related service implementations.
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"hello-gozero/infra/cache"
	userConstant "hello-gozero/internal/constant/user"
	userDto "hello-gozero/internal/dto/user"
	userEntity "hello-gozero/internal/entity/user"
	userRepo "hello-gozero/internal/repository/user"
	"hello-gozero/internal/svc"
)

type RegisterUserService struct {
	Logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterUserService 创建用户
func NewRegisterUserService(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterUserService {
	return &RegisterUserService{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (s *RegisterUserService) GetCtx() context.Context {
	return s.ctx
}
func (s *RegisterUserService) RegisterUser(req *userDto.RegisterUserReq) (resp *userDto.RegisterUserResp, err error) {
	// 加密密码（在事务外处理，避免事务过长）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 创建用户实体
	user := &userEntity.User{
		// ID 由 [userEntity.User.BeforeCreate] hook 自动生成，不应该显式设置
		Username:         req.Username,
		Password:         string(hashedPassword),
		Email:            req.Email,
		PhoneCountryCode: req.PhoneCountryCode,
		PhoneNumber:      req.PhoneNumber,
		Nickname:         req.Nickname,
		Status:           userConstant.StatusActive, // 默认正常状态
	}

	// ============================================================
	// 使用 Redis 分布式锁避免并发注册冲突
	// ============================================================
	// 优势：
	//   1. 解耦业务代码和数据库实现细节（不依赖索引名称判断）
	//   2. 在应用层面序列化并发请求，减少数据库压力
	//   3. 数据库唯一索引作为最后防线，保证数据完整性
	// 锁的粒度：基于用户名（可以根据需求调整为邮箱/手机号）
	lockKey := fmt.Sprintf("lock:user:register:%s", req.Username)
	lockValue := uuid.New().String() // 锁的唯一标识
	lockTTL := 10 * time.Second      // 锁的过期时间（防止死锁）

	// 使用 Redis 锁保护注册逻辑
	err = cache.WithLock(s.ctx, s.svcCtx.Infra.Redis.Client, lockKey, lockValue, lockTTL, func() error {
		// 使用事务确保数据一致性
		return s.svcCtx.Repository.User.Transaction(s.ctx, func(txRepo userRepo.UserRepository) error {
			// 检查用户名是否已存在（事务内查询，保证一致性）
			exists, err := txRepo.ExistsByUsername(s.ctx, req.Username)
			if err != nil {
				return fmt.Errorf("failed to check username existence: %w", err)
			}
			if exists {
				return ErrUsernameExists
			}

			// 检查邮箱是否已存在（如果提供）
			if req.Email != "" {
				existingUser, err := txRepo.GetByEmail(s.ctx, req.Email)
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("failed to check email existence: %w", err)
				}
				if existingUser != nil {
					return ErrEmailExists
				}
			}

			// 检查手机号是否已存在
			exists, err = txRepo.ExistsByPhone(s.ctx, req.PhoneCountryCode, req.PhoneNumber)
			if err != nil {
				return fmt.Errorf("failed to check phone existence: %w", err)
			}
			if exists {
				return ErrPhoneExists
			}

			// 创建用户
			if err := txRepo.Create(s.ctx, user); err != nil {
				// 如果仍然发生唯一性冲突（极端情况：锁失效或数据库约束）
				// 数据库唯一索引作为最后防线
				if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
					// 通过再次查询确定是哪个字段冲突（不依赖索引名称）
					if exists, _ := txRepo.ExistsByUsername(s.ctx, req.Username); exists {
						return ErrUsernameExists
					}
					if req.Email != "" {
						if u, _ := txRepo.GetByEmail(s.ctx, req.Email); u != nil {
							return ErrEmailExists
						}
					}
					if exists, _ := txRepo.ExistsByPhone(s.ctx, req.PhoneCountryCode, req.PhoneNumber); exists {
						return ErrPhoneExists
					}
					// 通用唯一性冲突（无法确定具体字段）
					return errors.New("user already exists")
				}
				return fmt.Errorf("failed to create user: %w", err)
			}

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	// 返回结果
	return &userDto.RegisterUserResp{}, nil
}
