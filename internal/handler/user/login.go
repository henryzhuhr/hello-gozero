package user

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
)

// LoginHandler 用户登录
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.Logger.WithContext(r.Context()).Errorf("failed to parse get user request: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		srv := userService.NewLoginService(r.Context(), svcCtx)
		resp, err := srv.Login(&req)
		ctx := srv.GetCtx() // 使用服务层的上下文以包含日志字段
		if err != nil {
			// 401 Unauthorized（账号/密码错误）
			// 403 Forbidden（账户被锁定、未启用 MFA 等）
			srv.Logger.WithContext(ctx).Errorf("failed to login for user(req: %+v): %v", req, err)
			if errors.Is(err, userService.ErrMissingUsername) {
				// 用户名缺失错误，返回标准错误响应
				httpx.ErrorCtx(ctx, w, err)
			} else {
				// 其他未知错误，返回标准错误响应
				httpx.ErrorCtx(ctx, w, err)
			}
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
