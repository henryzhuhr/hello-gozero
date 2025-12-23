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

// RegisterUserHandler 注册用户
func RegisterUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.RegisterUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 参数校验
		if err := req.Validate(); err != nil {
			switch v := err.(type) {
			case userDto.RegisterUserValidationError:
				// 返回结构化的校验错误信息
				httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, map[string]interface{}{"error": v.ToMap()})
			default:
				// 其他错误，返回通用错误信息
				httpx.ErrorCtx(r.Context(), w, err)
			}
			return
		}

		l := userService.NewRegisterUserService(r.Context(), svcCtx)
		resp, err := l.RegisterUser(&req)
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
