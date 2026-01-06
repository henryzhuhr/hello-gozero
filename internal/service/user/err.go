// Package user 错误定义
package user

import "errors"

var (
	// 缺少用户名参数
	ErrMissingUsername = errors.New("missing username")

	// 用户不存在
	ErrUserNotFound = errors.New("user not found")

	// 用户名已存在
	ErrUsernameExists = errors.New("username already exists")
)

var (
	// 邮箱已存在
	ErrEmailExists = errors.New("email already exists")

	// 手机号已存在
	ErrPhoneExists = errors.New("phone already exists")
)

var (
	// 密码过于简单
	ErrWeakPassword = errors.New("password is too weak")

	// 账户被禁用
	ErrAccountDisabled = errors.New("account is disabled")

	// 旧密码不匹配
	ErrOldPasswordMismatch = errors.New("old password does not match")

	// 新旧密码相同
	ErrNewPasswordSameAsOld = errors.New("new password cannot be the same as the old password")
)
