package locale

import (
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// globalBundle 全局语言包
var globalBundle *i18n.Bundle

// LoadFromDir 从指定目录加载所有语言文件
// 目录下的每个文件应为 YAML 格式的语言文件，例如：message.en-US.yaml、message.zh-CN.yaml
//
// 📝 注意事项
// 文件命名建议：messages.en.yaml, messages.zh-CN.yaml，扩展名必须是 .yaml。
// go-i18n 会自动从文件名中提取语言标签（如 xx.yy.yaml → xx-YY）。
// 如果你使用 message.zh.yaml，它会被识别为 zh；message.zh-CN.yaml → zh-CN。
// 确保 YAML 文件结构是扁平的或嵌套的 message 对象（见前文示例）。
func LoadFromDir(dir string) error {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yamlUnmarshal)

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read i18n directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// 可选：只加载 .yaml 文件
		if len(file.Name()) < 5 || file.Name()[len(file.Name())-5:] != ".yaml" {
			continue
		}
		if err := loadLangFile(bundle, dir+"/"+file.Name()); err != nil {
			return err
		}
	}

	globalBundle = bundle
	return nil
}

// T 是一个便捷函数，用于获取翻译文本
// lang: 语言标签，如 "en-US", "zh-CN"
// config: LocalizeConfig，包含 MessageID、TemplateData、PluralCount 等
func T(lang string, config *i18n.LocalizeConfig) (string, error) {
	if globalBundle == nil {
		return "", fmt.Errorf("i18n bundle not initialized. Call LoadFromDir first")
	}
	localizer := i18n.NewLocalizer(globalBundle, lang)
	return localizer.Localize(config)
}

// MustT 类似 T，但在出错时 panic（适用于初始化阶段或测试）
func MustT(lang string, config *i18n.LocalizeConfig) string {
	s, err := T(lang, config)
	if err != nil {
		panic(err)
	}
	return s
}

// GetSupportedLanguages 返回已加载的所有语言标签
func GetSupportedLanguages() []language.Tag {
	if globalBundle == nil {
		return nil
	}
	return globalBundle.LanguageTags()
}

// yamlUnmarshal 是 go-i18n 要求的反序列化函数签名
func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// loadLangFile 辅助函数：安全加载语言文件
func loadLangFile(bundle *i18n.Bundle, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read language file %s: %v", path, err)
	}
	_, err = bundle.ParseMessageFileBytes(data, path)
	if err != nil {
		return fmt.Errorf("failed to parse language file %s: %v", path, err)
	}
	return nil
}
