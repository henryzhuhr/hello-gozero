// Package user 用户登录相关的 DTO 定义
/*
+---------------------+
|      开始            |
+---------------------+
          |
          v
+---------------------+
| 输入用户名 + 密码     |
+---------------------+
          |
          v
+---------------------+     +----------------------+
| 验证账号密码          | No  | 返回 401: 凭据无效   |
+---------------------+---->+----------------------+
          | Yes
          v
+---------------------+
| 获取用户MFA选项       |
| (如: 手机、邮箱)      |
+---------------------+
          |
          v
+---------------------+
| 默认选择手机          |
| → 发送短信验证码       |
+---------------------+
          |
          v
+----------------------------------+
| 弹出 MFA 验证界面                  |
| - 显示: "验证码已发至 +86****1234"  |
| - [切换为邮箱] 按钮                |
+----------------------------------+
          |
          |<-----------------------+
          |                        |
          | 用户点击"切换为邮箱"      |
          v                        |
+---------------------+            |
| 取消当前短信流程    |               |
| → 发送邮箱验证码    |---------------+
+---------------------+
          |
          v
+----------------------------------+
| 用户输入6位验证码                |
+----------------------------------+
          |
          v
+---------------------+     +----------------------+
| 验证验证码 + 方式   | Fail| 返回 403: 验证码错误 |
+---------------------+---->+ (含重试次数限制)     |
          | Pass
          v
+---------------------+
| 颁发访问令牌          |
| → 登录成功           |
+---------------------+
*/
package user

// AuthFlowStatus 表示认证流程中的状态（仅用于 /login 响应）
type AuthFlowStatus string

const (
	AuthStatusSuccess       AuthFlowStatus = "success"        // 无需 MFA，直接登录成功
	AuthStatusMFARequired   AuthFlowStatus = "mfa_required"   // 需要完成 MFA
	AuthStatusAccountLocked AuthFlowStatus = "account_locked" // 账户被锁定（因安全策略，非用户状态暴露）
	// 注意：这里 "account_locked" 是认证流程结果，不是用户模型字段！
)

// IsValid 提供合法性校验（防止非法值序列化）
func (s AuthFlowStatus) IsValid() bool {
	switch s {
	case AuthStatusSuccess, AuthStatusMFARequired, AuthStatusAccountLocked:
		return true
	default:
		return false
	}
}

type MFAMethod string

const (
	MFAMethodSMS   MFAMethod = "sms"
	MFAMethodEmail MFAMethod = "email"
)

// LoginReq 登录请求体
type LoginReq struct {
	// 用户名，路径参数
	Username string `json:"username"`

	// 密码
	Password string `json:"password"`
}

// LoginResp 登录响应体
type LoginResp struct {
	// 账户状态
	// 	- 某些用户没开 MFA → 直接返回 token（status: "success"）
	// 	- 用户开了 MFA → 返回 status: "mfa_required"
	// 	- 账号被锁定 → status: "account_locked"
	Status AuthFlowStatus `json:"status"`

	// 临时 MFA 会话标识（短期有效，如5分钟）
	// 应为一次性、短期有效的随机字符串（如 UUID + 时间戳签名）
	MFAToken string `json:"mfa_token"`

	// 当前选中的方式
	CurrentMethod MFAMethod `json:"mfa_method"`

	// MFA 令牌过期时间，单位：分钟
	MFATokenExpiredTime int `json:"mfa_token_expired_time"`

	// 脱敏后的值，例如 "+86****1234" 或 "a***@example.com"，用于前端提示“验证码已发送至... +86 131****1111”
	MaskedValue string `json:"masked_value"`

	// 可用的 MFA 方式列表["sms", "email"]
	AvailableMethods []MFAMethod `json:"available_methods"`
}
