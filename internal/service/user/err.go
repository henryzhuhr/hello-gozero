// Package user 错误定义
package user

import (
	"net/http"

	"hello-gozero/pkg/errno"
)

// 用户认证相关错误
var (
	// 缺少用户名参数
	ErrMissingUsername = &errno.BizError{
		Code:        "USER_001",
		Message:     "missing username",
		HTTPStatus:  http.StatusBadRequest,
		UserMessage: "user.missing_username",
	}

	// 用户不存在
	ErrUserNotFound = &errno.BizError{
		Code:        "USER_002",
		Message:     "user not found",
		HTTPStatus:  http.StatusNotFound,
		UserMessage: "user.not_found",
	}

	// 用户名已存在
	ErrUsernameExists = &errno.BizError{
		Code:        "USER_003",
		Message:     "username already exists",
		HTTPStatus:  http.StatusConflict,
		UserMessage: "user.username_exists",
	}

	// 旧密码不匹配
	ErrOldPasswordMismatch = &errno.BizError{
		Code:        "USER_006",
		Message:     "old password does not match",
		HTTPStatus:  http.StatusBadRequest,
		UserMessage: "user.old_password_mismatch",
	}

	// 新旧密码相同
	ErrNewPasswordSameAsOld = &errno.BizError{
		Code:        "USER_007",
		Message:     "new password cannot be the same as the old password",
		HTTPStatus:  http.StatusBadRequest,
		UserMessage: "user.new_password_same_as_old",
	}

	// 密码过于简单
	ErrWeakPassword = &errno.BizError{
		Code:        "AUTH_003",
		Message:     "password is too weak",
		HTTPStatus:  http.StatusBadRequest,
		UserMessage: "user.weak_password",
	}

	// 用户名或密码错误
	ErrInvalidCredentials = &errno.BizError{
		Code:        "AUTH_001",
		Message:     "invalid username or password",
		HTTPStatus:  http.StatusUnauthorized,
		UserMessage: "auth.invalid_credentials",
	}
)

// 账户状态相关错误
var (
	// 账户被锁定
	ErrUserLocked = &errno.BizError{
		Code:        "AUTH_002",
		Message:     "account is locked",
		HTTPStatus:  http.StatusForbidden,
		UserMessage: "auth.account_locked",
	}

	// 账户被禁用
	ErrAccountDisabled = &errno.BizError{
		Code:        "AUTH_004",
		Message:     "account is disabled",
		HTTPStatus:  http.StatusForbidden,
		UserMessage: "auth.account_disabled",
	}
)

var (
	// 邮箱已存在
	ErrEmailExists = &errno.BizError{
		Code:        "USER_004",
		Message:     "email already exists",
		HTTPStatus:  http.StatusConflict,
		UserMessage: "user.email_exists",
	}

	// 手机号已存在
	ErrPhoneExists = &errno.BizError{
		Code:        "USER_005",
		Message:     "phone already exists",
		HTTPStatus:  http.StatusConflict,
		UserMessage: "user.phone_exists",
	}
)
