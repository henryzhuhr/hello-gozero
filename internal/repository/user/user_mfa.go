package user

import (
	"context"

	"gorm.io/gorm"

	userEntity "hello-gozero/internal/entity/user"
)

// UserMFARepository 定义用户多因素认证数据操作的接口
type UserMFARepository interface {
	// Transaction 执行事务操作
	// 接受一个函数，该函数接收事务版本的 Repository 并执行业务逻辑
	// 如果函数返回 error，事务回滚；否则提交
	//
	// 示例：
	//   err := userMFARepo.Transaction(ctx, func(txRepo UserMFARepository) error {
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
	Transaction(ctx context.Context, fn func(repo UserMFARepository) error) error

	// Create 创建新用户 MFA 记录
	Create(ctx context.Context, userMFA *userEntity.UserMFA) error

	// GetByUserID 获取指定用户的所有 MFA 记录
	GetByUserID(ctx context.Context, userID []byte) ([]*userEntity.UserMFA, error)
}

type userMFARepositoryImpl struct {
	db *gorm.DB
}

// NewUserMFARepository 创建一个新的 UserMFARepository 实例
func NewUserMFARepository(db *gorm.DB) UserMFARepository {
	return &userMFARepositoryImpl{db: db}
}

// Transaction Implements [UserMFARepository.Transaction]
func (r *userMFARepositoryImpl) Transaction(ctx context.Context, fn func(repo UserMFARepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &userMFARepositoryImpl{db: tx} // 直接构造一个带事务的实例
		return fn(txRepo)
	})
}

// Create implements [UserMFARepository.Create].
func (r *userMFARepositoryImpl) Create(ctx context.Context, userMFA *userEntity.UserMFA) error {
	return r.db.WithContext(ctx).Create(userMFA).Error
}

// GetByUserID implements [UserMFARepository.GetByUserID].
func (r *userMFARepositoryImpl) GetByUserID(ctx context.Context, userID []byte) ([]*userEntity.UserMFA, error) {
	var userMFAs []*userEntity.UserMFA
	if err := r.db.WithContext(ctx).
		Where(&userEntity.UserMFA{UserID: userID}).
		Find(&userMFAs).
		Error; err != nil {
		return nil, err
	}
	return userMFAs, nil
}
