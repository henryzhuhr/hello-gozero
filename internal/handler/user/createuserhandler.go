// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
)

// 创建用户
func CreateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.CreateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 参数校验
		if err := req.Validate(); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := userService.NewCreateUserService(r.Context(), svcCtx)
		resp, err := l.CreateUser(&req)
		if err != nil {
			switch {
			case errors.Is(err, userService.ErrUsernameExists):
				// TODO: 塞入 i18n 信息
				httpx.ErrorCtx(r.Context(), w, err)
			default:
				// 默认情况，内部服务错误
				httpx.ErrorCtx(r.Context(), w, err)
			}
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
