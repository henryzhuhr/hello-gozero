// Package user provides repository implementations for user data operations.
package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userEntity "hello-gozero/internal/entity/user"
)

// UserRepository 定义用户数据操作的接口
type UserRepository interface {
	// Transaction 执行事务操作
	// 接受一个函数，该函数接收事务版本的 Repository 并执行业务逻辑
	// 如果函数返回 error，事务回滚；否则提交
	//
	// 示例：
	//   err := userRepo.Transaction(ctx, func(txRepo UserRepository) error {
	//       // 检查用户是否存在
	//       exists, err := txRepo.ExistsByUsername(ctx, "alice")
	//       if err != nil {
	//           return err // 自动回滚
	//       }
	//       if exists {
	//           return ErrUsernameExists // 自动回滚
	//       }
	//
	//       // 创建用户
	//       user := &User{Username: "alice"}
	//       if err := txRepo.Create(ctx, user); err != nil {
	//           return err // 自动回滚
	//       }
	//
	//       return nil // 自动提交
	//   })
	Transaction(ctx context.Context, fn func(repo UserRepository) error) error

	// Create 创建新用户
	Create(ctx context.Context, user *userEntity.User) error

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*userEntity.User, error)

	// ExistsByUsername 检查指定用户名的用户是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*userEntity.User, error)

	// GetByPhone 根据手机号获取用户
	GetByPhone(ctx context.Context, phoneCountryCode, phoneNumber string) (*userEntity.User, error)

	// ExistsByPhone 根据手机号检查用户是否存在
	// 通常是检查，给定区号和手机号的组合是否已被注册，避免相同手机号注册多个账户
	ExistsByPhone(ctx context.Context, phoneCountryCode, phoneNumber string) (bool, error)

	// Update 更新已有用户
	Update(ctx context.Context, user *userEntity.User) error

	// Delete 通过 ID 软删除用户
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUsername 通过用户名删除用户
	DeleteByUsername(ctx context.Context, username string) error

	// List 分页获取用户列表，返回用户切片和总数
	List(ctx context.Context, offset, limit int) ([]*userEntity.User, int64, error)
	// ListWithStatus 分页获取用户列表，返回用户切片和总数
	ListWithStatus(ctx context.Context, offset, limit, status int) ([]*userEntity.User, int64, error)

	// Count 获取用户总数（包含软删除）
	Count(ctx context.Context) (int64, error)

	// CountWithStatus 获取指定昨天的用户
	CountWithStatus(ctx context.Context, status int) (int64, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建一个新的 UserRepository 实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// Transaction Implements [UserRepository.Transaction]
func (r *userRepositoryImpl) Transaction(ctx context.Context, fn func(repo UserRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &userRepositoryImpl{db: tx} // 直接构造一个带事务的实例
		return fn(txRepo)
	})
}

// Create Implements [UserRepository.Create]
func (r *userRepositoryImpl) Create(ctx context.Context, user *userEntity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByUsername Implements [UserRepository.GetByUsername]
func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where(&userEntity.User{Username: username}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername Implements [UserRepository.ExistsByUsername]
func (r *userRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	// var exists bool
	// 💡 额外提示：考虑使用 Select("1").Limit(1) 优化 EXISTS
	err := r.db.WithContext(ctx).
		Model(&userEntity.User{}).
		Where(&userEntity.User{Username: username}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByEmail Implements [UserRepository.GetByEmail]
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where(&userEntity.User{Email: email}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone Implements [UserRepository.GetByPhone]
func (r *userRepositoryImpl) GetByPhone(ctx context.Context, phoneCountryCode, phoneNumber string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).
		Where(&userEntity.User{
			PhoneCountryCode: phoneCountryCode,
			PhoneNumber:      phoneNumber,
		}).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByPhone Implements [UserRepository.ExistsByPhone]
func (r *userRepositoryImpl) ExistsByPhone(ctx context.Context, phoneCountryCode, phoneNumber string) (bool, error) {
	var exists bool
	// 💡 额外提示：考虑使用 Select("1").Limit(1) 优化 EXISTS
	err := r.db.WithContext(ctx).
		Model(&userEntity.User{}). // ✅ 必要，不能省略。它是 Count 正确执行的前提，且与结构体绑定，支持表名自定义、软删除等 GORM 特性
		Select("1").
		Where(&userEntity.User{PhoneCountryCode: phoneCountryCode, PhoneNumber: phoneNumber}).
		Limit(1).
		Find(&exists).Error // 注意：这里用 Find，但只关心是否找到
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Update Implements [UserRepository.Update]
func (r *userRepositoryImpl) Update(ctx context.Context, user *userEntity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete Implements [UserRepository.Delete]
func (r *userRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id[:]).Delete(&userEntity.User{}).Error
}

// DeleteByUsername implements [UserRepository.DeleteByUsername].
func (r *userRepositoryImpl) DeleteByUsername(ctx context.Context, username string) error {
	return r.db.WithContext(ctx).
		Where(&userEntity.User{Username: username}).
		Delete(&userEntity.User{}).
		Error
}

// List Implements [UserRepository.List]
func (r *userRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*userEntity.User, int64, error) {
	users := make([]*userEntity.User, 0)
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&userEntity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListWithStatus Implements [UserRepository.ListWithStatus]
func (r *userRepositoryImpl) ListWithStatus(ctx context.Context, offset, limit, status int) ([]*userEntity.User, int64, error) {
	users := make([]*userEntity.User, 0)
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&userEntity.User{}).Where(&userEntity.User{Status: int8(status)}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Where(&userEntity.User{Status: int8(status)}).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Count Implements [UserRepository.Count]
// 获取用户总数（包含软删除的用户）
func (r *userRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	// Unscoped() 会包含软删除的记录
	err := r.db.WithContext(ctx).Model(&userEntity.User{}).Unscoped().Count(&count).Error
	return count, err
}

// CountWithStatus Implements [UserRepository.CountWithStatus]
// 获取活跃用户数（不包含软删除的用户）
func (r *userRepositoryImpl) CountWithStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	// 默认不包含软删除的记录
	err := r.db.WithContext(ctx).Model(&userEntity.User{}).Where(&userEntity.User{Status: int8(status)}).Count(&count).Error
	return count, err
}
