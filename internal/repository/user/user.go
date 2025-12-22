package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	userEntity "hello-gozero/internal/entity/user"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *userEntity.User) error
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error)
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*userEntity.User, error)
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*userEntity.User, error)
	// GetByPhone retrieves a user by phone number
	GetByPhone(ctx context.Context, phone string) (*userEntity.User, error)
	// Update updates an existing user
	Update(ctx context.Context, user *userEntity.User) error
	// UpdateFields updates specific fields of a user
	UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
	// Delete soft deletes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error
	// List retrieves a paginated list of users, returns users list and total count
	List(ctx context.Context, offset, limit int) ([]*userEntity.User, int64, error)
	// Exists checks if a user with the given username exists
	Exists(ctx context.Context, username string) (bool, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// Create Implements [UserRepository.Create]
func (r *userRepositoryImpl) Create(ctx context.Context, user *userEntity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID Implements [UserRepository.GetByID]
func (r *userRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where("id = ?", id[:]).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
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

// GetByEmail Implements [UserRepository.GetByEmail]
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone Implements [UserRepository.GetByPhone]
func (r *userRepositoryImpl) GetByPhone(ctx context.Context, phone string) (*userEntity.User, error) {
	var user userEntity.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update Implements [UserRepository.Update]
func (r *userRepositoryImpl) Update(ctx context.Context, user *userEntity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateFields Implements [UserRepository.UpdateFields]
func (r *userRepositoryImpl) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&userEntity.User{}).Where("id = ?", id[:]).Updates(fields).Error
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

// Exists Implements [UserRepository.Exists]
func (r *userRepositoryImpl) Exists(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userEntity.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
