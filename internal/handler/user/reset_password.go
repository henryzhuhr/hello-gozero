package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	"hello-gozero/internal/svc"
)

// ResetPasswordHandler 重制用户密码
func ResetPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.ResetPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.Logger.WithContext(r.Context()).Errorf("failed to parse reset password request: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// l := userService.NewGetUserService(r.Context(), svcCtx)
		// resp, err := l.GetUser(&req)
		// ctx := l.GetCtx() // 使用服务层的上下文以包含日志字段
		// if err != nil {
		// 	l.Logger.WithContext(ctx).Errorf("failed to get user: %v", err)
		// 	if errors.Is(err, userService.ErrMissingUsername) {
		// 		// 用户名缺失错误，返回标准错误响应
		// 		httpx.ErrorCtx(ctx, w, err)
		// 	} else if errors.Is(err, userService.ErrUserNotFound) {
		// 		// 用户不存在错误，返回 404 状态码和自定义错误信息
		// 		w.WriteHeader(http.StatusNotFound)
		// 		httpx.WriteJsonCtx(ctx, w, http.StatusNotFound, map[string]interface{}{
		// 			"code": http.StatusNotFound,
		// 			"msg":  "user not found",
		// 		})
		// 	} else {
		// 		// 其他未知错误，返回标准错误响应
		// 		httpx.ErrorCtx(ctx, w, err)
		// 	}
		// } else {
		// 	httpx.OkJsonCtx(ctx, w, resp)
		// }
	}
}

// VerifyResetPasswordTokenHandler 验证重置密码令牌
func VerifyResetPasswordTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.VerifyResetPasswordTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.Logger.WithContext(r.Context()).Errorf("failed to parse reset password request: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// l := userService.NewGetUserService(r.Context(), svcCtx)
		// resp, err := l.GetUser(&req)
		// ctx := l.GetCtx() // 使用服务层的上下文以包含日志字段
		// if err != nil {
		// 	l.Logger.WithContext(ctx).Errorf("failed to get user: %v", err)
		// 	if errors.Is(err, userService.ErrMissingUsername) {
		// 		// 用户名缺失错误，返回标准错误响应
		// 		httpx.ErrorCtx(ctx, w, err)
		// 	} else if errors.Is(err, userService.ErrUserNotFound) {
		// 		// 用户不存在错误，返回 404 状态码和自定义错误信息
		// 		w.WriteHeader(http.StatusNotFound)
		// 		httpx.WriteJsonCtx(ctx, w, http.StatusNotFound, map[string]interface{}{
		// 			"code": http.StatusNotFound,
		// 			"msg":  "user not found",
		// 		})
		// 	} else {
		// 		// 其他未知错误，返回标准错误响应
		// 		httpx.ErrorCtx(ctx, w, err)
		// 	}
		// } else {
		// 	httpx.OkJsonCtx(ctx, w, resp)
		// }
	}
}
