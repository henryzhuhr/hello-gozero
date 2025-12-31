// Package user provides HTTP handlers for user-related operations.
package user

import (
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
)

// DeleteUserHandler 删除用户
// 请求参数 [userDto.DeleteUserReq] 中的 `username` 对应路径参数 `:username`
// 例如，DELETE /users/johndoe 会将 `johndoe` 作为 `username` 参数传递
func DeleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.DeleteUserReq
		if err := httpx.Parse(r, &req); err != nil {
			svcCtx.Logger.WithContext(r.Context()).Errorf("failed to parse delete user request: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := userService.NewDeleteUserService(r.Context(), svcCtx)
		resp, err := l.DeleteUser(&req)
		ctx := l.GetCtx() // 使用服务层的上下文以包含日志字段
		if err != nil {
			l.Logger.WithContext(ctx).Errorf("failed to delete user (req: %+v): %v", req, err)
			if errors.Is(err, userService.ErrUserNotFound) {
				// 用户不存在错误，返回 404 状态码和自定义错误信息
				w.WriteHeader(http.StatusNotFound)
				httpx.WriteJsonCtx(ctx, w, http.StatusNotFound, map[string]interface{}{
					"code": http.StatusNotFound,
					"msg":  "user not found",
				})
				return
			} else {
				httpx.ErrorCtx(r.Context(), w, err)
			}
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
