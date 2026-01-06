// Package password 提供密码加密与验证的工具实现
package password

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const MinPepperLength = 32

// 错误定义
var (
	ErrEmptyPassword      = errors.New("password cannot be empty")          // 密码为空
	ErrEmptyPepper        = errors.New("pepper key cannot be empty")        // pepper密钥为空
	ErrPepperTooShort     = errors.New("pepper key length is insufficient") // pepper密钥长度不足
	ErrHashGenerationFail = errors.New("failed to generate password hash")  // 哈希生成失败
	ErrEmptyStoredHash    = errors.New("stored hash cannot be empty")       // 存储的哈希为空
)

// PasswordCryptoTool 密码加密验证工具接口（对外暴露）
// 核心能力：基于pepper密钥+bcrypt哈希，完成密码串的哈希生成与验证
// 适用场景：任意密码串的加密存储、密码合法性校验（无场景限制）
type PasswordCryptoTool interface {
	// GenerateHash 生成密码串的bcrypt哈希值（含pepper加盐），用于持久化存储
	// 参数：password - 待哈希的密码串（任意格式/来源）
	// 返回：哈希字符串 | 错误（空密码串/哈希失败等）
	GenerateHash(password string) (string, error)

	// VerifyHash 验证密码串与存储的哈希值是否匹配（自动拼接pepper校验）
	// 参数：
	//   password    - 待验证的密码串
	//   storedHash  - 已存储的密码哈希值
	// 返回：匹配返回true，不匹配/参数非法返回false
	VerifyHash(password, storedHash string) bool
}

// passwordCryptoTool 接口私有实现体
type passwordCryptoTool struct {
	bcryptCost int    // bcrypt哈希成本
	pepperKey  string // 额外加盐密钥（pepper），仅存储在后端
}

// NewPasswordCryptoTool 创建密码工具实例（工厂函数，新增pepperKey参数）
// 参数：
//
//	bcryptCost - 哈希成本（4-31，默认10）
//	pepperKey  - 额外加盐密钥（必填，建议32位以上随机字符串）
//
// 返回：接口实例 | 错误
func NewPasswordCryptoTool(bcryptCost int, pepperKey string) (PasswordCryptoTool, error) {
	// 校验pepperKey非空（核心安全配置，不能为空）
	if pepperKey == "" {
		return nil, ErrEmptyPepper
	}

	// 新增：校验 pepper 长度
	if len(pepperKey) < MinPepperLength {
		return nil, fmt.Errorf("%w: minimum required %d characters, current length: %d", ErrPepperTooShort, MinPepperLength, len(pepperKey))
	}

	// 校验哈希成本范围
	if bcryptCost < bcrypt.MinCost || bcryptCost > bcrypt.MaxCost {
		bcryptCost = bcrypt.DefaultCost
	}

	return &passwordCryptoTool{
		bcryptCost: bcryptCost,
		pepperKey:  pepperKey,
	}, nil
}

// GenerateHash Implements [PasswordCryptoTool.GenerateHash]
func (t *passwordCryptoTool) GenerateHash(password string) (string, error) {
	// 空值校验
	if password == "" {
		return "", ErrEmptyPassword
	}

	// 1. 添加pepper额外加盐
	pwdWithPepper := t.addPepper(password)

	// 2. 生成bcrypt哈希（自动生成随机盐值+pepper固定盐值）
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(pwdWithPepper), t.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrHashGenerationFail, err)
	}

	return string(hashBytes), nil
}

// VerifyHash Implements [PasswordCryptoTool.VerifyHash]
func (t *passwordCryptoTool) VerifyHash(password, storedHash string) bool {
	// 空值校验
	if storedHash == "" || password == "" {
		return false
	}

	// 1. 给输入密码串添加相同的pepper
	pwdWithPepper := t.addPepper(password)

	// 2. 比对哈希（bcrypt自动提取随机盐值，结合pepper比对）
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pwdWithPepper))
	return err == nil
}

// addPepper 使用 HMAC-SHA256 混合密码和 pepper（更安全）
func (t *passwordCryptoTool) addPepper(inputPwd string) string {
	h := hmac.New(sha256.New, []byte(t.pepperKey))
	h.Write([]byte(inputPwd))
	// 返回 base64 编码的 HMAC 结果
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
