// Package password 密码规则
package password

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// -------------------------- 定义校验规则接口 --------------------------

// Rule 单个校验规则的接口，每种规则实现自己的Check方法
type Rule interface {
	Check(password string) error
}

// -------------------------- 实现各细分规则 --------------------------

// -------- LengthRule 长度校验规则 --------

// LengthConfig 长度校验配置
type LengthConfig struct {
	Min int
	Max int
}

// LengthRule 长度校验规则
type LengthRule struct {
	config LengthConfig
}

func NewLengthRule(config LengthConfig) Rule {
	return &LengthRule{config: config}
}

func (r *LengthRule) Check(password string) error {
	length := len(password)
	if length < r.config.Min {
		return fmt.Errorf("密码长度不能小于%d位", r.config.Min)
	}
	if length > r.config.Max {
		return fmt.Errorf("密码长度不能大于%d位", r.config.Max)
	}
	return nil
}

// -------- CharTypeRule 字符类型校验规则 --------

// CharTypeConfig 字符类型校验配置（职责单一）
type CharTypeConfig struct {
	RequireDigit  bool // 是否要求数字
	RequireUpper  bool // 是否要求大写字母
	RequireLower  bool // 是否要求小写字母
	RequireSymbol bool // 是否要求特殊符号
}

// CharTypeRule 字符类型校验规则
type CharTypeRule struct {
	config CharTypeConfig
}

func NewCharTypeRule(config CharTypeConfig) Rule {
	return &CharTypeRule{config: config}
}

func (r *CharTypeRule) Check(password string) error {
	var (
		hasDigit  = false
		hasUpper  = false
		hasLower  = false
		hasSymbol = false
	)

	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case strings.ContainsRune("~!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSymbol = true
		}
	}

	if r.config.RequireDigit && !hasDigit {
		return errors.New("密码必须包含数字")
	}
	if r.config.RequireUpper && !hasUpper {
		return errors.New("密码必须包含大写字母")
	}
	if r.config.RequireLower && !hasLower {
		return errors.New("密码必须包含小写字母")
	}
	if r.config.RequireSymbol && !hasSymbol {
		return errors.New("密码必须包含特殊符号（~!@#$%^&*()_+-=[]{}|;:,.<>?）")
	}
	return nil
}

// -------- SimplePasswordRule 简单密码校验规则 --------

// SimplePasswordConfig 简单密码校验配置（职责单一）
type SimplePasswordConfig struct {
	BanSimple bool // 是否禁止简单密码
}

// SimplePasswordRule 简单密码校验规则
type SimplePasswordRule struct {
	config SimplePasswordConfig
}

func NewSimplePasswordRule(config SimplePasswordConfig) Rule {
	return &SimplePasswordRule{config: config}
}

func (r *SimplePasswordRule) Check(password string) error {
	if !r.config.BanSimple {
		return nil
	}

	// 纯数字/纯字母校验
	numRegex := regexp.MustCompile(`^[0-9]+$`)
	letterRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if numRegex.MatchString(password) || letterRegex.MatchString(password) {
		return errors.New("密码不能为纯数字或纯字母")
	}

	// 连续字符校验
	if isConsecutive(password) {
		return errors.New("密码不能包含连续的数字或字母")
	}

	// 重复字符校验
	if isRepeated(password) {
		return errors.New("密码不能包含重复的字符")
	}
	return nil
}
