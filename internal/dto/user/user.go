package user

type User struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Status        int    `json:"status"`
	LastLoginTime string `json:"lastLoginTime,omitempty"`
}


type DeleteUserReq struct {
	Id string `path:"id" validate:"required"`
}

type DeleteUserResp struct {
	Message string `json:"message"`
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

type GetUserReq struct {
	Id string `path:"id" validate:"required"`
}

type GetUserResp struct {
	User User `json:"user"`
}
