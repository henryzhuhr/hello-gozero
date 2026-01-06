// Package routes 路由注册
package routes

import (
	"net/http"

	hello "hello-gozero/internal/handler/hello"
	"hello-gozero/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	registerGlobalHandlers(server, serverCtx)

	// 注册用户相关路由
	userRouter := NewUserRouter(server, serverCtx)
	userRouter.Register()
}

// registerGlobalHandlers 注册全局路由
func registerGlobalHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				// 健康检查
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: hello.HealthHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/hello",
				Handler: hello.HelloHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api"),
	)
}
