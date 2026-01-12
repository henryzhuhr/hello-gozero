package user

const (
	getUserListMinPageNum  = 1   // 最小页码
	getUserListMaxPageNum  = 100 // 最大页码
	getUserListMinPageSize = 10  // 最小每页大小
	getUserListMaxPageSize = 100 // 最大每页大小
)

// GetUserListReq 获取用户列表请求参数
type GetUserListReq struct {
	// 当前页码
	Page int `form:"page,default=1"`

	// 每页大小/limit: 表示每一页包含多少条记录
	PageSize int `form:"pageSize,default=10"`

	// 可选的过滤条件：用户状态
	Status int `form:"status,optional"`
}

// GetUserListResp 获取用户列表响应结果
type GetUserListResp struct {
	Total int64  `json:"total"`
	List  []User `json:"list"`
}

// Validate 验证请求参数
func (r *GetUserListReq) Validate() error {
	// 限制页码范围
	if r.Page < getUserListMinPageNum {
		r.Page = getUserListMinPageNum
	}
	if r.Page > getUserListMaxPageNum {
		r.Page = getUserListMaxPageNum
	}

	// 限制每页大小
	if r.PageSize < getUserListMinPageSize {
		r.PageSize = getUserListMinPageSize
	}
	if r.PageSize > getUserListMaxPageSize {
		r.PageSize = getUserListMaxPageSize
	}
	return nil
}
