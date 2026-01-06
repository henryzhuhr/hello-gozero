package password

import "testing"

// 测试 isConsecutive 辅助函数
func TestIsConsecutive(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", true}, // 实现中空字符串返回 true (numConsec 初始为 true)
		{"single char", "a", true}, // 单字符无循环，返回 true
		{"ascending digits", "12345", true},
		{"descending digits", "54321", true},
		{"ascending lowercase", "abcde", true},
		{"descending lowercase", "edcba", true},
		{"ascending uppercase", "ABCDE", true},
		{"descending uppercase", "EDCBA", true},
		{"non consecutive digits", "13579", false},
		{"non consecutive letters", "acegi", false},
		{"mixed case non consecutive", "AbCdE", false},
		{"letters and digits mixed", "a1b2c", false},
		{"repeated char", "aaaa", false},
		{"partial consecutive", "abc9", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isConsecutive(tc.input)
			if got != tc.want {
				t.Errorf("isConsecutive(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// 测试 isRepeated 辅助函数
func TestIsRepeated(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", false},
		{"single char", "a", false},
		{"all same digit", "1111", true},
		{"all same lowercase", "aaaa", true},
		{"all same uppercase", "AAAA", true},
		{"two same chars", "aa", true},
		{"different chars", "abcd", false},
		{"mostly same with one different", "aaaab", false},
		{"alternating chars", "abab", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isRepeated(tc.input)
			if got != tc.want {
				t.Errorf("isRepeated(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// 测试空规则列表的情况
func TestPasswordService_EmptyRules(t *testing.T) {
	svc := NewPasswordChecker()

	// 空规则列表应该接受任何密码
	passwords := []string{
		"",
		"a",
		"123",
		"anything goes",
	}

	for _, pwd := range passwords {
		if err := svc.Check(pwd); err != nil {
			t.Errorf("Empty rules should accept %q, but got error: %v", pwd, err)
		}
	}
}

// 测试 CheckStrength 的边界情况
func TestPasswordService_CheckStrength_EdgeCases(t *testing.T) {
	svc := NewDefaultPasswordChecker()

	cases := []struct {
		name     string
		password string
		want     string
	}{
		{"empty", "", "弱"},
		{"single char", "a", "弱"},
		{"only digits", "12345678", "弱"},                 // 长度+数字=2
		{"only lowercase", "abcdefgh", "弱"},              // 长度+小写=2
		{"only uppercase", "ABCDEFGH", "弱"},              // 长度+大写=2
		{"digits and lower", "abc12345", "中"},            // 长度+数字+小写=3
		{"digits and upper", "ABC12345", "中"},            // 长度+数字+大写=3
		{"upper and lower", "AbcdEfgh", "中"},             // 长度+大小写=3
		{"all but symbol", "Abcd1234", "中"},              // 长度+数字+大小写=4
		{"all types", "Abcd123!", "中"},                   // 长度+数字+大小写+符号=4（长度<8不加分）
		{"short with all", "Ab1!", "中"},                  // 数字+大小写+符号=4
		{"long pure digit", "12345678901234567890", "弱"}, // 长度+数字=2
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.CheckStrength(tc.password)
			if got != tc.want {
				t.Errorf("CheckStrength(%q) = %s, want %s", tc.password, got, tc.want)
			}
		})
	}
}

// 测试单规则场景
func TestPasswordService_SingleRule(t *testing.T) {
	// 只有长度规则
	svc := NewPasswordChecker(
		NewLengthRule(LengthConfig{Min: 5, Max: 10}),
	)

	if err := svc.Check("abc"); err == nil {
		t.Error("Expected error for too short password")
	}

	if err := svc.Check("12345678901"); err == nil {
		t.Error("Expected error for too long password")
	}

	if err := svc.Check("12345"); err != nil {
		t.Errorf("Expected success for valid length, got: %v", err)
	}
}

// 测试规则执行顺序（遇到第一个失败就返回）
func TestPasswordService_RuleExecutionOrder(t *testing.T) {
	svc := NewPasswordChecker(
		NewLengthRule(LengthConfig{Min: 8, Max: 32}),
		NewCharTypeRule(CharTypeConfig{RequireDigit: true}),
	)

	// 太短会先被长度规则拒绝
	err := svc.Check("abc")
	if err == nil {
		t.Fatal("Expected error for short password")
	}
	if err.Error() != "密码长度不能小于8位" {
		t.Errorf("Expected length error first, got: %v", err)
	}

	// 长度满足但没有数字会被字符类型规则拒绝
	err = svc.Check("abcdefgh")
	if err == nil {
		t.Fatal("Expected error for password without digit")
	}
	if err.Error() != "密码必须包含数字" {
		t.Errorf("Expected digit requirement error, got: %v", err)
	}
}

// 测试多个字符类型要求
func TestPasswordService_MultipleCharTypeRequirements(t *testing.T) {
	svc := NewPasswordChecker(
		NewCharTypeRule(CharTypeConfig{
			RequireDigit:  true,
			RequireUpper:  true,
			RequireLower:  true,
			RequireSymbol: true,
		}),
	)

	cases := []struct {
		password string
		ok       bool
		desc     string
	}{
		{"", false, "empty"},
		{"abc", false, "only lower"},
		{"ABC", false, "only upper"},
		{"123", false, "only digit"},
		{"!@#", false, "only symbol"},
		{"Abc", false, "lower+upper"},
		{"Abc1", false, "lower+upper+digit"},
		{"Abc1!", true, "all types"},
		{"Password123!", true, "all types longer"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := svc.Check(tc.password)
			if tc.ok && err != nil {
				t.Errorf("Expected success for %q, got: %v", tc.password, err)
			}
			if !tc.ok && err == nil {
				t.Errorf("Expected error for %q", tc.password)
			}
		})
	}
}
