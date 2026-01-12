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
