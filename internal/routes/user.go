// Package routes 用户相关路由注册
package routes

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	user "hello-gozero/internal/handler/user"
	"hello-gozero/internal/svc"
)

type userRouter struct {
	server    *rest.Server
	serverCtx *svc.ServiceContext
}

func NewUserRouter(server *rest.Server, serverCtx *svc.ServiceContext) *userRouter {
	return &userRouter{
		server:    server,
		serverCtx: serverCtx,
	}
}

func (r *userRouter) Register() {
	r.addRegisterUser()                   // 用户注册
	r.addAccountStatusManagement()        // 账户状态管理
	r.addUserInformationManagement()      // 用户信息管理
	r.addBatchUserInformationManagement() // 用户批量管理
	r.addPasswordManagement()             // 密码管理
}

// addRegisterUser 用户注册
//   - POST /api/v1/users/register - 注册用户
func (r *userRouter) addRegisterUser() {
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 注册用户
				Method:  http.MethodPost,
				Path:    "/users/register",
				Handler: user.RegisterUserHandler(r.serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}

// addUserInformationManagement 用户信息管理
//   - GET /api/v1/users/:username - 获取单个用户基础信息 【新增】
//   - PUT /api/v1/users/:username - 更新用户信息（完整更新）
//   - PATCH /api/v1/users/:username - 部分更新用户信息
//   - GET /api/v1/users/:username/profile - 获取用户详细资料
func (r *userRouter) addUserInformationManagement() {
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 获取单个用户
				Method:  http.MethodGet,
				Path:    "/users/:username",
				Handler: user.GetUserHandler(r.serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}

// addBatchUserInformationManagement 用户批量管理
func (r *userRouter) addBatchUserInformationManagement() {
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 获取用户列表
				Method:  http.MethodGet,
				Path:    "/users",
				Handler: user.GetUserListHandler(r.serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}

// addAccountStatusManagement 账户状态管理
//   - DELETE /api/v1/users/:username - 删除用户 【新增】
//   - PUT /api/v1/users/:username/status - 更新用户状态（启用/禁用/锁定）
//   - PUT /api/v1/users/:username/activate - 激活用户
//   - PUT /api/v1/users/:username/deactivate - 停用用户
func (r *userRouter) addAccountStatusManagement() {
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 删除用户
				Method:  http.MethodDelete,
				Path:    "/users/:username",
				Handler: user.DeleteUserHandler(r.serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}

// addPasswordManagement 密码管理
// - `PUT /api/v1/users/:username/password` - 修改密码
// - `POST /api/v1/users/password/reset` - 重置密码（忘记密码）
// - `POST /api/v1/users/password/reset/verify` - 验证重置密码令牌
func (r *userRouter) addPasswordManagement() {
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 修改密码
				Method:  http.MethodPut,
				Path:    "/users/:username/password",
				Handler: user.UpdatePasswordHandler(r.serverCtx),
			},
			{
				// 重置密码（忘记密码）
				Method:  http.MethodPost,
				Path:    "/users/password/reset",
				Handler: user.ResetPasswordHandler(r.serverCtx),
			},
			{
				// 验证重置密码令牌
				Method:  http.MethodPost,
				Path:    "/users/password/reset/verify",
				Handler: user.VerifyResetPasswordTokenHandler(r.serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}
