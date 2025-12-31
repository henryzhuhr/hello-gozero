package user

// User 用户信息，返回给客户端
type User struct {
	Username         string `json:"username"`
	Email            string `json:"email,omitempty"`
	PhoneCountryCode string `json:"phone_country_code,omitempty"`
	PhoneNumber      string `json:"phone_number,omitempty"`
	Nickname         string `json:"nickname,omitempty"`
	Status           int    `json:"status"`
	LastLoginTime    string `json:"lastLoginTime,omitempty"`
}

// DeleteUserReq 删除用户请求参数
type DeleteUserReq struct {
	Username string `path:"username" validate:"required"`
}

// DeleteUserResp 删除用户响应参数
type DeleteUserResp struct {
}

type GetUserListReq struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"pageSize,default=10" validate:"min=1,max=100"`
	Username string `form:"username,optional"`
	Status   int    `form:"status,optional"`
}

type GetUserListResp struct {
	Total int64  `json:"total"`
	List  []User `json:"list"`
}
