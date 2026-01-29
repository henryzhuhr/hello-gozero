package user

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
	"hello-gozero/pkg/errno"
	"hello-gozero/pkg/response"
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

		svc := userService.NewLoginService(r.Context(), svcCtx)
		resp, err := svc.Login(&req)
		ctx := svc.GetCtx() // 使用服务层的上下文以包含日志字段
		if err != nil {
			svc.Logger.WithContext(ctx).Errorf("failed to login for user(req: %+v): %v", req, err)

			// 401 Unauthorized（账号/密码错误）
			// 403 Forbidden（账户被锁定、未启用 MFA 等）
			var bizErr *errno.BizError
			if errors.As(err, &bizErr) {
				// userMsg := i18n.T(bizErr.UserMessage, lang)
				userMsg := bizErr.UserMessage
				httpx.WriteJsonCtx(ctx, w, bizErr.HTTPStatus,
					response.NewUnifiedAPIResponse(bizErr.Code, userMsg),
				)
			} else {
				// 其他未知错误，返回标准错误响应
				httpx.ErrorCtx(ctx, w, err)
			}
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
