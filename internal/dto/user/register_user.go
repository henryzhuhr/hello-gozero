package user

import (
	"fmt"
	"regexp"
	"strings"
)

// 邮箱正则（更严谨的版本，可根据需求调整）
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var invalidUsername = map[string]struct{}{
	"admin":         {},
	"root":          {},
	"system":        {},
	"support":       {},
	"contact":       {},
	"info":          {},
	"administrator": {},
}

// RegisterUserReq 创建用户请求
type RegisterUserReq struct {
	Username         string `json:"username" validate:"required,min=3,max=50"`
	Password         string `json:"password" validate:"required,min=6,max=100"`
	Email            string `json:"email,omitempty"`
	PhoneCountryCode string `json:"phone_country_code" validate:"regexp=^\\+[1-9]\\d{0,3}$"`
	PhoneNumber      string `json:"phone_number" validate:"max=20"`
	Nickname         string `json:"nickname,omitempty" validate:"max=50"`
}

// RegisterUserResp 创建用户响应
type RegisterUserResp struct{}

// RegisterUserValidationError 表示一个字段校验错误
type RegisterUserValidationError struct {
	Field string      // 字段名，如 "email", "username"
	Code  string      // 错误码，如 "invalid_email", "username_reserved"
	Value interface{} // 可选：用于上下文（如非法的用户名）
}

func (e RegisterUserValidationError) Error() string {
	return fmt.Sprintf("invalid field %s: %s", e.Field, e.Code)
}

// ToMap 可选：实现一个方法返回结构化数据
func (e RegisterUserValidationError) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"field": e.Field,
		"code":  e.Code,
		"value": e.Value,
	}
}

func (u *RegisterUserReq) Validate() error {
	// 用户名校验
	if err := u.validateUsername(); err != nil {
		return err
	}

	// 手机号校验
	if err := u.validatePhone(); err != nil {
		return err
	}

	// 邮箱校验
	if err := u.validateEmail(); err != nil {
		return err
	}

	// 可继续添加其他字段...

	return nil
}

// validateUsername 校验用户名是否合法
//
// 支持：
//   - 非空
//   - 黑名单校验（如 "admin", "root" 等）
//   - 最小长度（如 3）
//   - 格式校验（只允许字母、数字、下划线、点等）
//
// 不支持
//   - 中文字符（根据需求可添加）
//   - emoji（根据需求可添加）
func (u *RegisterUserReq) validateUsername() error {
	if u.Username == "" {
		return RegisterUserValidationError{Field: "username", Code: "required"}
	}
	// 黑名单校验（统一转小写避免大小写绕过）
	if _, exists := invalidUsername[strings.ToLower(u.Username)]; exists {
		return RegisterUserValidationError{
			Field: "username",
			Code:  "reserved_username",
			Value: u.Username,
		}
	}
	if len(u.Username) < 3 {
		return RegisterUserValidationError{Field: "username", Code: "too_short", Value: u.Username}
	}
	// 格式校验：只允许字母、数字、下划线、点（根据业务需求调整）
	if !regexp.MustCompile(`^[a-zA-Z0-9_.]+$`).MatchString(u.Username) {
		// return fmt.Errorf("username can only contain letters, numbers, underscores, and dots")
		return RegisterUserValidationError{Field: "username", Code: "invalid_format", Value: u.Username}
	}

	return nil
}

// validatePhone 简单校验手机号格式，需要检查区号和号码的组合
func (u *RegisterUserReq) validatePhone() error {
	if u.PhoneNumber == "" {
		return nil // optional field, skip
	}
	// 简单校验：只允许数字和可选的加号（根据需求调整）
	if !regexp.MustCompile(`^\+?[0-9]+$`).MatchString(u.PhoneNumber) {
		return RegisterUserValidationError{
			Field: "phone",
			Code:  "invalid_phone",
			Value: u.PhoneNumber,
		}
	}
	return nil
}

// validateEmail 简单校验邮箱格式
func (u *RegisterUserReq) validateEmail() error {
	if u.Email == "" {
		return nil // optional field, skip
	}
	if !emailRegex.MatchString(u.Email) {
		// return fmt.Errorf("invalid email format: %s", u.Email)
		return RegisterUserValidationError{
			Field: "email",
			Code:  "invalid_email",
			Value: u.Email,
		}
	}
	return nil
}
