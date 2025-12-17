// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"net/http"

	"hello-gozero/internal/logic/hello"
	"hello-gozero/internal/svc"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 日志注入 user_id 字段
		ctx := logx.ContextWithFields(r.Context(), logx.Field("user_id", uuid.New().String()))
		l := hello.NewHealthLogic(ctx, svcCtx)
		resp, err := l.Health()
		l.Logger.Infof("resp: %+v", resp)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
