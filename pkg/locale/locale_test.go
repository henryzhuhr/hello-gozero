package locale

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetLocale 测试设置 locale
func TestSetLocale(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected string
	}{
		{
			name:     "设置英文",
			locale:   "en-US",
			expected: "en-US",
		},
		{
			name:     "设置中文",
			locale:   "zh-CN",
			expected: "zh-CN",
		},
		{
			name:     "设置其他语言",
			locale:   "ja-JP",
			expected: "ja-JP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := SetLocale(ctx, tt.locale)
			
			// 验证值已经设置到 context 中
			value := newCtx.Value(key)
			assert.NotNil(t, value)
			assert.Equal(t, tt.expected, value.(string))
		})
	}
}

// TestGetLocale 测试获取 locale
func TestGetLocale(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		expected Locale
	}{
		{
			name: "获取英文 locale",
			setup: func() context.Context {
				return SetLocale(context.Background(), "en-US")
			},
			expected: LocaleEN,
		},
		{
			name: "获取中文 locale",
			setup: func() context.Context {
				return SetLocale(context.Background(), "zh-CN")
			},
			expected: LocaleZH,
		},
		{
			name: "未设置 locale，返回默认英文",
			setup: func() context.Context {
				return context.Background()
			},
			expected: LocaleEN,
		},
		{
			name: "不支持的 locale，返回默认英文",
			setup: func() context.Context {
				return SetLocale(context.Background(), "fr-FR")
			},
			expected: LocaleEN,
		},
		{
			name: "空字符串 locale，返回默认英文",
			setup: func() context.Context {
				return SetLocale(context.Background(), "")
			},
			expected: LocaleEN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			locale := GetLocale(ctx)
			assert.Equal(t, tt.expected, locale)
		})
	}
}

// TestLocaleConstants 测试 Locale 常量
func TestLocaleConstants(t *testing.T) {
	assert.Equal(t, Locale("en-US"), LocaleEN)
	assert.Equal(t, Locale("zh-CN"), LocaleZH)
}

// TestSetAndGetLocale 测试设置和获取 locale 的完整流程
func TestSetAndGetLocale(t *testing.T) {
	ctx := context.Background()
	
	// 初始状态应该返回默认的英文
	locale := GetLocale(ctx)
	assert.Equal(t, LocaleEN, locale)
	
	// 设置为中文
	ctx = SetLocale(ctx, "zh-CN")
	locale = GetLocale(ctx)
	assert.Equal(t, LocaleZH, locale)
	
	// 重新设置为英文
	ctx = SetLocale(ctx, "en-US")
	locale = GetLocale(ctx)
	assert.Equal(t, LocaleEN, locale)
}

// TestLocaleContextIsolation 测试 context 的隔离性
func TestLocaleContextIsolation(t *testing.T) {
	// 创建父 context
	parentCtx := context.Background()
	
	// 创建两个不同的子 context
	ctx1 := SetLocale(parentCtx, "en-US")
	ctx2 := SetLocale(parentCtx, "zh-CN")
	
	// 验证两个 context 互不影响
	locale1 := GetLocale(ctx1)
	locale2 := GetLocale(ctx2)
	
	assert.Equal(t, LocaleEN, locale1)
	assert.Equal(t, LocaleZH, locale2)
	
	// 父 context 不应该被影响
	parentLocale := GetLocale(parentCtx)
	assert.Equal(t, LocaleEN, parentLocale)
}

// TestGetLocaleString 测试 Locale 类型的字符串值
func TestGetLocaleString(t *testing.T) {
	assert.Equal(t, "en-US", string(LocaleEN))
	assert.Equal(t, "zh-CN", string(LocaleZH))
}
