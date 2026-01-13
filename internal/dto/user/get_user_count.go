package user

// GetUserCountReq 获取用户总数请求参数
type GetUserCountReq struct {
	// 可选：是否包含软删除的用户
	IncludeDeleted bool `form:"include_deleted,optional,default=false"`
}

// GetUserCountResp 获取用户总数响应结果
type GetUserCountResp struct {
	// 用户总数
	Total int64 `json:"total"`
	// 活跃用户数（不包含软删除）
	Active int64 `json:"active,omitempty"`
	// 已删除用户数
	Deleted int64 `json:"deleted,omitempty"`
}
