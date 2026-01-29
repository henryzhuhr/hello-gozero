// Package errno 定义业务逻辑错误类型和常见错误
package errno

import (
	"fmt"
)

// BizError 业务逻辑错误
type BizError struct {
	Code        string // 唯一错误码，如 "AUTH_001"
	Message     string // 错误消息消息（用于日志、调试）
	HTTPStatus  int    // HTTP 状态码（可选）
	UserMessage string // 返回给用户的错误消息，i18n 的键名
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is 支持 errors.Is 比较（按 Code 判断）
func (e *BizError) Is(target error) bool {
	t, ok := target.(*BizError)
	return ok && e.Code == t.Code
}
