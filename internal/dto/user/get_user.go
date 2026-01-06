// Package user 获取用户的请求
package user

// GetUserReq 获取单个用户请求
type GetUserReq struct {
	// 用户名，路径参数
	// 例如: /api/v1/user/{username}
	Username string `path:"username"`
}

type GetUserResp struct {
	User User `json:"user"`
}
