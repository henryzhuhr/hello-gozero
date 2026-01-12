package user

import (
	"regexp"
	"strings"
)

var invalidUsername = map[string]struct{}{
	"admin":         {},
	"root":          {},
	"system":        {},
	"support":       {},
	"contact":       {},
	"info":          {},
	"administrator": {},
}

// validateUsername 校验用户名是否合法的公共方法
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
func validateUsername(username string) error {
	if username == "" {
		return RegisterUserValidationError{Field: "username", Code: "required"}
	}
	// 黑名单校验（统一转小写避免大小写绕过）
	if _, exists := invalidUsername[strings.ToLower(username)]; exists {
		return RegisterUserValidationError{
			Field: "username",
			Code:  "reserved_username",
			Value: username,
		}
	}
	if len(username) < 3 {
		return RegisterUserValidationError{Field: "username", Code: "too_short", Value: username}
	}
	// 格式校验：只允许字母、数字、下划线、点（根据业务需求调整）
	if !regexp.MustCompile(`^[a-zA-Z0-9_.]+$`).MatchString(username) {
		// return fmt.Errorf("username can only contain letters, numbers, underscores, and dots")
		return RegisterUserValidationError{Field: "username", Code: "invalid_format", Value: username}
	}

	return nil
}
