// Package user defines the user entity and related database models.
package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserMFA represents a user entity. Use GORM model definitions and tags as needed.
type UserMFA struct {
	ID     []byte `gorm:"primaryKey;type:BINARY(16);not null"`
	UserID []byte `gorm:"primaryKey;type:BINARY(16);not null"`

	// MFA 方式，如 SMS、TOTP、Email 等
	Method string `gorm:"type:varchar(20);not null;column:method" json:"method"`

	// 绑定目标地址
	// 	- sms: 手机号（如 "+8613812345678"）
	// 	- email: 邮箱（如 "a***@example.com"）
	// 	- totp: 设备描述（如 "iPhone Google Authenticator"）
	// 	- webauthn: 凭据 ID（Base64 编码）
	Target string `gorm:"type:varchar(100);not null;column:target" json:"target"`

	// 是否启用该 MFA 方式
	Enabled bool `gorm:"type:boolean;default:false;column:enabled" json:"enabled"`

	// 是否为主用方式
	IsPrimary bool `gorm:"type:boolean;default:false;column:is_primary" json:"is_primary"`

	// 加密存储的敏感数据（如 TOTP 密钥）
	Secret []byte `gorm:"type:varbinary(255);column:secret" json:"-"`

	// 首次验证成功时间
	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`

	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"` // 软删除
}

// TableName specifies the table name for the User model
func (UserMFA) TableName() string {
	return "t_user_mfa"
}

// BeforeCreate GORM hook - generates UUID before creating a new user
func (u *UserMFA) BeforeCreate(tx *gorm.DB) error {
	if len(u.ID) == 0 {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = id[:]
	}
	return nil
}

func (u *UserMFA) GetIDAsString() string {
	id, err := uuid.FromBytes(u.ID)
	if err != nil {
		return ""
	}
	return id.String()
}
