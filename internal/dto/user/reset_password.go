package user

type ResetPasswordReq struct {
	// 邮箱
	Email string `json:"email"`

	// 新密码
	NewPassword string `json:"new_password"`

	// 重置密码的验证码
	ResetCode string `json:"reset_code"`
}

type ResetPasswordResp struct {
	// 重置结果消息
	Message string `json:"message"`
}

type VerifyResetPasswordTokenReq struct {
	// 重置密码的验证码
	ResetCode string `json:"reset_code"`
}

type VerifyResetPasswordTokenResp struct {
	// 验证结果消息
	Message string `json:"message"`
}
