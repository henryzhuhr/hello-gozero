// Package user defines the user entity and related database models.
package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MFAMethod represents a user entity. Use GORM model definitions and tags as needed.
type MFAMethod struct {
	ID     []byte `gorm:"primaryKey;type:BINARY(16);column:id;not null"`
	UserID []byte `gorm:"type:BINARY(16);column:user_id;not null"`

	// MFA 类型，如 SMS、TOTP、Email 等
	Type string `gorm:"type:varchar(20);not null;column:type" json:"type"`

	// MFA 方式标签（用户自定义，如 "个人手机"、"工作邮箱"）
	Label string `gorm:"type:varchar(50);column:label" json:"label"`

	// 绑定目标地址
	// 	- sms: 手机号（如 "+8613812345678"）
	// 	- email: 邮箱（如 "a***@example.com"）
	// 	- totp: 设备描述（如 "iPhone Google Authenticator"）
	// 	- webauthn: 凭据 ID（Base64 编码）
	Target string `gorm:"type:varchar(100);not null;column:target" json:"target"`

	// 是否启用该 MFA 方式
	Enabled bool `gorm:"type:boolean;default:false;column:enabled" json:"enabled"`

	// 是否为主用方式
	IsDefault bool `gorm:"type:boolean;default:false;column:is_default" json:"is_default"`

	// 加密存储的敏感数据（如 TOTP 密钥）
	Secret []byte `gorm:"type:varbinary(255);column:secret" json:"-"`

	// 是否已完成验证（防止未验证就启用）
	Verified bool `gorm:"column:verified" json:"verified,omitempty"`

	// 首次验证成功时间
	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`

	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"` // 软删除
}

// TableName specifies the table name for the User model
func (MFAMethod) TableName() string {
	return "mfa_method"
}

// BeforeCreate GORM hook - generates UUID before creating a new user
func (u *MFAMethod) BeforeCreate(tx *gorm.DB) error {
	if len(u.ID) == 0 {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = id[:]
	}
	if len(u.UserID) == 0 {
		return errors.New("UserID is required")
	}
	return nil
}

func (u *MFAMethod) GetIDAsString() string {
	id, err := uuid.FromBytes(u.ID)
	if err != nil {
		return ""
	}
	return id.String()
}
