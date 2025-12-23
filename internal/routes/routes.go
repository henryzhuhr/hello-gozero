package routes

import (
	"net/http"

	hello "hello-gozero/internal/handler/hello"
	user "hello-gozero/internal/handler/user"
	"hello-gozero/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
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
		rest.WithPrefix("/api/v1"),
	)

	registerV1Handlers(server, serverCtx)
}

// registerV1Handlers 注册 v1 版本的路由
func registerV1Handlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				// 注册用户
				Method:  http.MethodPost,
				Path:    "/users/register",
				Handler: user.RegisterUserHandler(serverCtx),
			},
			{
				// 获取单个用户
				Method:  http.MethodGet,
				Path:    "/users/:username",
				Handler: user.GetUserHandler(serverCtx),
			},
			{
				// 获取用户列表
				Method:  http.MethodGet,
				Path:    "/users",
				Handler: user.GetUserListHandler(serverCtx),
			},
			{
				// 删除用户
				Method:  http.MethodDelete,
				Path:    "/users/:username",
				Handler: user.DeleteUserHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}
