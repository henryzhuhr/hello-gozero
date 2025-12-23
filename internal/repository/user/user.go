package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userEntity "hello-gozero/internal/entity/user"
)

// UserRepository 定义用户数据操作的接口
type UserRepository interface {
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

	// List 分页获取用户列表，返回用户切片和总数
	List(ctx context.Context, offset, limit int) ([]*userEntity.User, int64, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建一个新的 UserRepository 实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// Create Implements [UserRepository.Create]
func (r *userRepositoryImpl) Create(ctx context.Context, user *userEntity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByUsername Implements [UserRepository.GetByUsername]
func (r *userRepositoryImpl) GetByUsername(ctx context.Context, username string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername Implements [UserRepository.ExistsByUsername]
func (r *userRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userEntity.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByEmail Implements [UserRepository.GetByEmail]
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where(`email = ?`, email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone Implements [UserRepository.GetByPhone]
func (r *userRepositoryImpl) GetByPhone(ctx context.Context, phoneCountryCode, phoneNumber string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).
		Where(`phone_country_code = ? AND phone_number = ?`, phoneCountryCode, phoneNumber).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByPhone Implements [UserRepository.ExistsByPhone]
func (r *userRepositoryImpl) ExistsByPhone(ctx context.Context, phoneCountryCode, phoneNumber string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&userEntity.User{}).
		Where("phone_country_code = ? AND phone_number = ?", phoneCountryCode, phoneNumber).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update Implements [UserRepository.Update]
func (r *userRepositoryImpl) Update(ctx context.Context, user *userEntity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete Implements [UserRepository.Delete]
func (r *userRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id[:]).Delete(&userEntity.User{}).Error
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
