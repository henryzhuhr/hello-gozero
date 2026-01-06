// Package password 密码检查
package password

import "regexp"

// -------------------------- 重构密码服务，聚合所有规则 --------------------------

// PasswordChecker 密码检查接口
type PasswordChecker interface {
	// Check 执行所有密码校验规则
	Check(password string) error

	// CheckStrength 密码强度检测
	CheckStrength(password string) string
}

// passwordChecker 密码服务实现（不再持有大配置，而是持有规则列表）
type passwordChecker struct {
	rules []Rule // 所有需要执行的校验规则
}

// NewPasswordChecker 创建密码服务实例，传入自定义规则列表
func NewPasswordChecker(rules ...Rule) PasswordChecker {
	return &passwordChecker{rules: rules}
}

// NewDefaultPasswordChecker 创建使用默认规则的密码服务实例
// 默认规则：长度8-32位，要求数字、大小写字母，禁止简单密码
func NewDefaultPasswordChecker() PasswordChecker {
	return NewPasswordChecker(
		NewLengthRule(LengthConfig{Min: 8, Max: 32}),
		NewCharTypeRule(CharTypeConfig{
			RequireDigit:  true,
			RequireUpper:  true,
			RequireLower:  true,
			RequireSymbol: false,
		}),
		NewSimplePasswordRule(SimplePasswordConfig{BanSimple: true}),
	)
}

// Check Implements [PasswordChecker.Check]
func (p *passwordChecker) Check(password string) error {
	// 依次执行每个规则，只要有一个失败就返回错误
	for _, rule := range p.rules {
		if err := rule.Check(password); err != nil {
			return err
		}
	}
	return nil
}

// CheckStrength Implements [PasswordChecker.CheckStrength]
func (p *passwordChecker) CheckStrength(password string) string {
	strength := 0
	if len(password) >= 8 {
		strength++
	}
	if regexp.MustCompile(`[0-9]`).MatchString(password) {
		strength++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		strength++
	}
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		strength++
	}
	if regexp.MustCompile(`[~!@#$%^&*()_+-=[]{}|;:,.<>?]`).MatchString(password) {
		strength++
	}

	switch strength {
	case 0, 1, 2:
		return "弱"
	case 3, 4:
		return "中"
	default:
		return "强"
	}
}

// -------------------------- 辅助函数 --------------------------
func isConsecutive(s string) bool {
	// 数字连续
	numConsec := true
	for i := 1; i < len(s); i++ {
		if int(s[i]) != int(s[i-1])+1 && int(s[i]) != int(s[i-1])-1 {
			numConsec = false
			break
		}
	}
	if numConsec {
		return true
	}

	// 字母连续
	letterConsec := true
	for i := 1; i < len(s); i++ {
		if (s[i] >= 'a' && s[i] <= 'z' && s[i-1] >= 'a' && s[i-1] <= 'z') ||
			(s[i] >= 'A' && s[i] <= 'Z' && s[i-1] >= 'A' && s[i-1] <= 'Z') {
			if int(s[i]) != int(s[i-1])+1 && int(s[i]) != int(s[i-1])-1 {
				letterConsec = false
				break
			}
		} else {
			letterConsec = false
			break
		}
	}
	return letterConsec
}

func isRepeated(s string) bool {
	if len(s) <= 1 {
		return false
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}
