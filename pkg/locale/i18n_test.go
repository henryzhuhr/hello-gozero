package locale

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadFromDir 测试从目录加载语言文件
func TestLoadFromDir(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试用的语言文件
	enContent := []byte(`test.hello: "Hello"
test.welcome: "Welcome, {{.Name}}"
test.count:
  one: "{{.Count}} item"
  other: "{{.Count}} items"`)

	zhContent := []byte(`test.hello: "你好"
test.welcome: "欢迎, {{.Name}}"
test.count:
  one: "{{.Count}} 项"
  other: "{{.Count}} 项"`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), enContent, 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "message.zh-CN.yaml"), zhContent, 0644)
	require.NoError(t, err)

	// 测试加载
	err = LoadFromDir(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, globalBundle)

	// 验证语言包是否加载成功
	languages := GetSupportedLanguages()
	assert.True(t, len(languages) >= 2, "应该至少加载两种语言")
}

// TestLoadFromDir_DirectoryNotExist 测试目录不存在的情况
func TestLoadFromDir_DirectoryNotExist(t *testing.T) {
	err := LoadFromDir("/non/existent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read i18n directory")
}

// TestLoadFromDir_InvalidYAML 测试无效的 YAML 文件
func TestLoadFromDir_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()

	// 创建无效的 YAML 文件
	invalidContent := []byte(`invalid: yaml: content:
  - this is wrong`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en.yaml"), invalidContent, 0644)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	assert.Error(t, err)
}

// TestT 测试翻译函数
func TestT(t *testing.T) {
	// 准备测试环境
	tempDir := t.TempDir()
	enContent := []byte(`test.hello: "Hello"
test.welcome: "Welcome, {{.Name}}"`)
	zhContent := []byte(`test.hello: "你好"
test.welcome: "欢迎, {{.Name}}"`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), enContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "message.zh-CN.yaml"), zhContent, 0644)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name     string
		lang     string
		config   *i18n.LocalizeConfig
		expected string
		wantErr  bool
	}{
		{
			name: "英文简单翻译",
			lang: "en-US",
			config: &i18n.LocalizeConfig{
				MessageID: "test.hello",
			},
			expected: "Hello",
			wantErr:  false,
		},
		{
			name: "中文简单翻译",
			lang: "zh-CN",
			config: &i18n.LocalizeConfig{
				MessageID: "test.hello",
			},
			expected: "你好",
			wantErr:  false,
		},
		{
			name: "英文带模板数据",
			lang: "en-US",
			config: &i18n.LocalizeConfig{
				MessageID: "test.welcome",
				TemplateData: map[string]interface{}{
					"Name": "Alice",
				},
			},
			expected: "Welcome, Alice",
			wantErr:  false,
		},
		{
			name: "中文带模板数据",
			lang: "zh-CN",
			config: &i18n.LocalizeConfig{
				MessageID: "test.welcome",
				TemplateData: map[string]interface{}{
					"Name": "张三",
				},
			},
			expected: "欢迎, 张三",
			wantErr:  false,
		},
		{
			name: "消息ID不存在",
			lang: "en-US",
			config: &i18n.LocalizeConfig{
				MessageID: "test.nonexistent",
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := T(tt.lang, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestT_BundleNotInitialized 测试未初始化的情况
func TestT_BundleNotInitialized(t *testing.T) {
	// 保存原始的 globalBundle
	originalBundle := globalBundle
	defer func() {
		globalBundle = originalBundle
	}()

	// 将 globalBundle 设置为 nil
	globalBundle = nil

	_, err := T("en-US", &i18n.LocalizeConfig{
		MessageID: "test.hello",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "i18n bundle not initialized")
}

// TestMustT 测试 MustT 函数
func TestMustT(t *testing.T) {
	// 准备测试环境
	tempDir := t.TempDir()
	enContent := []byte(`test.hello: "Hello"`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), enContent, 0644)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	require.NoError(t, err)

	// 测试正常情况
	result := MustT("en-US", &i18n.LocalizeConfig{
		MessageID: "test.hello",
	})
	assert.Equal(t, "Hello", result)
}

// TestMustT_Panic 测试 MustT 在出错时是否会 panic
func TestMustT_Panic(t *testing.T) {
	// 准备测试环境
	tempDir := t.TempDir()
	enContent := []byte(`test.hello: "Hello"`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), enContent, 0644)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	require.NoError(t, err)

	// 测试 panic 情况
	assert.Panics(t, func() {
		MustT("en-US", &i18n.LocalizeConfig{
			MessageID: "test.nonexistent",
		})
	})
}

// TestGetSupportedLanguages 测试获取支持的语言列表
func TestGetSupportedLanguages(t *testing.T) {
	// 准备测试环境
	tempDir := t.TempDir()
	enContent := []byte(`test.hello: "Hello"`)
	zhContent := []byte(`test.hello: "你好"`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), enContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "message.zh-CN.yaml"), zhContent, 0644)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	require.NoError(t, err)

	languages := GetSupportedLanguages()
	assert.True(t, len(languages) >= 2, "应该至少有两种语言")

	// 验证语言标签
	hasEnglish := false
	hasChinese := false
	for _, lang := range languages {
		if lang.String() == "en-US" {
			hasEnglish = true
		}
		if lang.String() == "zh-CN" {
			hasChinese = true
		}
	}
	assert.True(t, hasEnglish, "应该包含英语")
	assert.True(t, hasChinese, "应该包含中文")
}

// TestGetSupportedLanguages_NotInitialized 测试未初始化时返回 nil
func TestGetSupportedLanguages_NotInitialized(t *testing.T) {
	// 保存原始的 globalBundle
	originalBundle := globalBundle
	defer func() {
		globalBundle = originalBundle
	}()

	globalBundle = nil

	languages := GetSupportedLanguages()
	assert.Nil(t, languages)
}

// TestLoadFromDir_IgnoreNonYAMLFiles 测试忽略非 YAML 文件
func TestLoadFromDir_IgnoreNonYAMLFiles(t *testing.T) {
	tempDir := t.TempDir()

	// 创建 YAML 文件
	yamlContent := []byte(`test.hello: "Hello"`)
	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), yamlContent, 0644)
	require.NoError(t, err)

	// 创建非 YAML 文件
	err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# README"), 0644)
	require.NoError(t, err)

	// 创建子目录
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	assert.NoError(t, err)
}

// TestPluralTranslation 测试复数翻译
func TestPluralTranslation(t *testing.T) {
	tempDir := t.TempDir()
	enContent := []byte(`test.items:
  one: "{{.Count}} item"
  other: "{{.Count}} items"`)
	zhContent := []byte(`test.items:
  one: "{{.Count}} 项"
  other: "{{.Count}} 项"`)

	err := os.WriteFile(filepath.Join(tempDir, "message.en-US.yaml"), enContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "message.zh-CN.yaml"), zhContent, 0644)
	require.NoError(t, err)

	err = LoadFromDir(tempDir)
	require.NoError(t, err)

	// 测试单数
	result, err := T("en-US", &i18n.LocalizeConfig{
		MessageID:   "test.items",
		PluralCount: 1,
		TemplateData: map[string]interface{}{
			"Count": 1,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "1 item", result)

	// 测试复数
	result, err = T("en-US", &i18n.LocalizeConfig{
		MessageID:   "test.items",
		PluralCount: 5,
		TemplateData: map[string]interface{}{
			"Count": 5,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "5 items", result)
}

// TestLoadFromDir_WithRealLocales 使用实际的 locales 目录测试
func TestLoadFromDir_WithRealLocales(t *testing.T) {
	// 检查 locales 目录是否存在
	localesDir := "../../locales"
	if _, err := os.Stat(localesDir); os.IsNotExist(err) {
		t.Skip("locales 目录不存在，跳过此测试")
		return
	}

	err := LoadFromDir(localesDir)
	require.NoError(t, err)

	// 测试实际的消息
	result, err := T("en-US", &i18n.LocalizeConfig{
		MessageID: "user.missing_username",
	})
	require.NoError(t, err)
	assert.Equal(t, "Missing username", result)

	result, err = T("zh-CN", &i18n.LocalizeConfig{
		MessageID: "user.missing_username",
	})
	require.NoError(t, err)
	assert.Equal(t, "用户名缺失", result)
}
