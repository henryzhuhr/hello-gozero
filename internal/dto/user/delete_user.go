// Package user
package user

// DeleteUserReq 删除用户请求参数
type DeleteUserReq struct {
	Username string `path:"username" validate:"required"`
}

// DeleteUserResp 删除用户响应参数
type DeleteUserResp struct {
}

func (r *DeleteUserReq) Validate() error {
	// 用户名校验
	if err := validateUsername(r.Username); err != nil {
		return err
	}

	// 可继续添加其他字段...

	return nil
}
