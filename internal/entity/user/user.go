// Package user defines the user entity and related database models.
package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity. Use GORM model definitions and tags as needed.
type User struct {
	ID []byte `gorm:"primaryKey;type:BINARY(16);not null"`

	Username         string `gorm:"type:varchar(50);not null;column:username" json:"username"`
	Password         string `gorm:"type:varchar(255);not null;column:password" json:"-"` // 不在 JSON 中返回密码
	Email            string `gorm:"type:varchar(100);default:'';column:email" json:"email"`
	PhoneCountryCode string `gorm:"type:varchar(10);default:'';column:phone_country_code" json:"phone_country_code"`
	PhoneNumber      string `gorm:"type:varchar(20);default:'';column:phone_number" json:"phone_number"`
	Nickname         string `gorm:"type:varchar(50);default:'';column:nickname" json:"nickname"`

	Status        int8       `gorm:"type:tinyint;default:1;column:status" json:"status"` // 0-禁用，1-正常
	LastLoginTime *time.Time `gorm:"column:last_login_time" json:"last_login_time,omitempty"`

	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"` // 软删除
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "t_user"
}

// BeforeCreate GORM hook - generates UUID before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if len(u.ID) == 0 {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = id[:]
	}
	return nil
}

func (u *User) GetIDAsString() string {
	id, err := uuid.FromBytes(u.ID)
	if err != nil {
		return ""
	}
	return id.String()
}
