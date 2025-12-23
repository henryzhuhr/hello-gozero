package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
)

// 获取单个用户
func GetUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.GetUserReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.Logger.WithContext(r.Context()).Errorf("failed to parse get user request: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := userService.NewGetUserService(r.Context(), svcCtx)
		resp, err := l.GetUser(&req)
		if err != nil {
			l.Logger.WithContext(r.Context()).Errorf("failed to get user: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
