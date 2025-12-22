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

	server.AddRoutes(
		[]rest.Route{
			{
				// 创建用户
				Method:  http.MethodPost,
				Path:    "/user",
				Handler: user.CreateUserHandler(serverCtx),
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
				Path:    "/users/:id",
				Handler: user.DeleteUserHandler(serverCtx),
			},
			{
				// 获取单个用户
				Method:  http.MethodGet,
				Path:    "/users/:id",
				Handler: user.GetUserHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}
