package user

type UpdatePasswordReq struct {
	// 用户名，路径参数
	// 例如: /api/v1/user/{username}/password
	Username string `path:"username"`

	// 旧密码
	OldPassword string `json:"old_password"`

	// 新密码
	NewPassword string `json:"new_password"`
}

type UpdatePasswordResp struct {
	// 更新结果消息
	Message string `json:"message"`
}
