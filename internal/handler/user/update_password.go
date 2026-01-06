package user

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
)

// UpdatePasswordHandler 更新用户密码
func UpdatePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.UpdatePasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.Logger.WithContext(r.Context()).Errorf("failed to parse get user request: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		srv := userService.NewUpdatePasswordService(r.Context(), svcCtx)
		resp, err := srv.UpdatePassword(&req)
		ctx := srv.GetCtx() // 使用服务层的上下文以包含日志字段
		if err != nil {
			srv.Logger.WithContext(ctx).Errorf("failed to update password for user(req: %+v): %v", req, err)
			if errors.Is(err, userService.ErrMissingUsername) {
				// 用户名缺失错误，返回标准错误响应
				httpx.ErrorCtx(ctx, w, err)
			} else if errors.Is(err, userService.ErrUserNotFound) {
				// 用户不存在错误，返回 404 状态码和自定义错误信息
				w.WriteHeader(http.StatusNotFound)
				httpx.WriteJsonCtx(ctx, w, http.StatusNotFound, map[string]interface{}{
					"code": http.StatusNotFound,
					"msg":  "user not found",
				})
			} else if errors.Is(err, userService.ErrOldPasswordMismatch) {
				// 旧密码不匹配错误，返回 400 状态码和自定义错误信息
				w.WriteHeader(http.StatusBadRequest)
				httpx.WriteJsonCtx(ctx, w, http.StatusBadRequest, map[string]interface{}{
					"code": http.StatusBadRequest,
					"msg":  "old password does not match",
				})
			} else if errors.Is(err, userService.ErrNewPasswordSameAsOld) {
				// 新密码与旧密码相同错误，返回 400 状态码和自定义错误信息
				w.WriteHeader(http.StatusBadRequest)
				httpx.WriteJsonCtx(ctx, w, http.StatusBadRequest, map[string]interface{}{
					"code": http.StatusBadRequest,
					"msg":  "new password cannot be the same as the old password",
				})
			} else {
				// 其他未知错误，返回标准错误响应
				httpx.ErrorCtx(ctx, w, err)
			}
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
