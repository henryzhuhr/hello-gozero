// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"net/http"

	"hello-gozero/internal/logic/hello"
	"hello-gozero/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func HelloHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svcCtx.Logger.Infof("hello: handler svcCtx")
		l := hello.NewHelloLogic(r.Context(), svcCtx)
		resp, err := l.Hello()
		l.Logger.Infof("hello: handler调用成功")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
