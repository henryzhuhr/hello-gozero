package password

import "testing"

func TestLengthRule(t *testing.T) {
	rule := NewLengthRule(LengthConfig{Min: 8, Max: 16})

	cases := []struct {
		name     string
		password string
		ok       bool
	}{
		{"too short", "1234567", false},
		{"too long", "12345678901234567", false},
		{"in range", "12345678", true},
		{"boundary min", "12345678", true},
		{"boundary max", "1234567890123456", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Check(tc.password)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestCharTypeRule(t *testing.T) {
	rule := NewCharTypeRule(CharTypeConfig{
		RequireDigit:  true,
		RequireUpper:  true,
		RequireLower:  true,
		RequireSymbol: true,
	})

	cases := []struct {
		name     string
		password string
		ok       bool
	}{
		{"missing digit", "Abcdef!@", false},
		{"missing upper", "abc123!@", false},
		{"missing lower", "ABC123!@", false},
		{"missing symbol", "Abc12345", false},
		{"all satisfied", "Abc123!@", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Check(tc.password)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestSimplePasswordRule(t *testing.T) {
	banRule := NewSimplePasswordRule(SimplePasswordConfig{BanSimple: true})
	allowRule := NewSimplePasswordRule(SimplePasswordConfig{BanSimple: false})

	cases := []struct {
		name     string
		rule     Rule
		password string
		ok       bool
	}{
		{"pure digits banned", banRule, "12345678", false},
		{"pure letters banned", banRule, "abcdefgh", false},
		{"consecutive letters banned", banRule, "abcdefg", false},
		{"consecutive digits banned", banRule, "1234567", false},
		{"repeated char banned", banRule, "aaaaaaaa", false},
		{"mix allowed under ban", banRule, "Abcd1234", true},
		{"any allowed when not banning", allowRule, "11111111", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.rule.Check(tc.password)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestPasswordService_Check(t *testing.T) {
	// 使用默认服务：长度[8,32]，要求数字/大小写字母，禁止简单密码
	svc := NewDefaultPasswordChecker()

	cases := []struct {
		name     string
		password string
		ok       bool
	}{
		{"too short", "Ab1", false},
		{"no digit", "Abcdefgh", false},
		{"no upper", "abc12345", false},
		{"no lower", "ABC12345", false},
		{"repeated simple banned", "AAAAAAAA", false},
		{"valid default", "Password1", true},
		{"valid with symbol", "Passw0rd!", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Check(tc.password)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestPasswordService_CustomRules(t *testing.T) {
	// 自定义规则：只要求长度，不要求字符类型
	svc := NewPasswordChecker(
		NewLengthRule(LengthConfig{Min: 6, Max: 20}),
	)

	cases := []struct {
		name     string
		password string
		ok       bool
	}{
		{"too short", "12345", false},
		{"too long", "123456789012345678901", false},
		{"pure digits ok", "123456", true},
		{"pure letters ok", "abcdef", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Check(tc.password)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestPasswordService_CheckStrength(t *testing.T) {
	svc := NewDefaultPasswordChecker()

	cases := []struct {
		password string
		want     string
	}{
		{"a", "弱"},         // 仅小写 -> 弱
		{"abcdEFGH", "中"},  // 长度+大小写=3 -> 中
		{"Password1", "中"}, // 4 -> 中
		{"Passw0rd!", "中"}, // 当前实现判断为中
		{"12345678", "弱"},  // 长度+数字=2 -> 弱
	}

	for i, tc := range cases {
		got := svc.CheckStrength(tc.password)
		if got != tc.want {
			t.Fatalf("case %d: want %s, got %s", i, tc.want, got)
		}
	}
}
