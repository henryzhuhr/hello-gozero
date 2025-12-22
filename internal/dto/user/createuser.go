package user

import (
	"fmt"
	"regexp"
)

// 邮箱正则（更严谨的版本，可根据需求调整）
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type CreateUserReq struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,max=20"`
	Nickname string `json:"nickname,omitempty" validate:"max=50"`
}

func (u CreateUserReq) Validate() error {
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format: %s", u.Email)
	}
	return nil
}

type CreateUserResp struct {
	Id string `json:"id"`
}
