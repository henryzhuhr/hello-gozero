// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
)

// GetUserCountHandler 获取用户总数
func GetUserCountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userDto.GetUserCountReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := userService.NewGetUserCountService(r.Context(), svcCtx)
		resp, err := l.GetUserCount(&req)
		ctx := l.GetCtx() // 使用服务层的上下文以包含日志字段
		if err != nil {
			l.Logger.WithContext(ctx).Errorf("failed to get user count: %v", err)
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
