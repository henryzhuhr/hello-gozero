// Package user 用户相关路由注册
package user

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
	// 用户管理 / User Management 接口组
	r.addUserMangement()
	r.addAccountStatusManagement()        // 账户状态管理
	r.addUserInformationManagement()      // 用户信息管理
	r.addBatchUserInformationManagement() // 用户批量管理

	// 认证 / Authentication 接口组
	r.addAuthenticationRoutes()

	// 鉴权 / Authorization  接口组
}

// addUserMangement 用户管理 / User Management 接口组
//   - POST /api/v1/users/register - 注册用户 【新增】
func (r *userRouter) addUserMangement() {
	// 注册 v1 接口组
	// - POST /api/v1/users/register - 注册用户 【新增】
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

	// 密码管理 v1 接口组
	// - `PUT /api/v1/users/:username/password` - 修改密码
	// - `POST /api/v1/users/password/reset` - 重置密码（忘记密码）
	// - `POST /api/v1/users/password/reset/verify` - 验证重置密码令牌
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

	// MFA 管理 v1 接口组
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 新增 TOTP MFA 方法
				Method:  http.MethodPost,
				Path:    "/mfa/setup/totp",
				Handler: nil, // TODO
			},
			{
				// 新增 SMS MFA 方法
				Method:  http.MethodPost,
				Path:    "/mfa/setup/sms",
				Handler: nil, // TODO
			},
			{
				// MFA 验证
				Method:  http.MethodPost,
				Path:    "/mfa/verify",
				Handler: user.MFAVerifyHandler(r.serverCtx),
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
				Path:    "/users/list",
				Handler: user.GetUserListHandler(r.serverCtx),
			},
			{
				// 获取用户总数统计
				Method:  http.MethodGet,
				Path:    "/users/count",
				Handler: user.GetUserCountHandler(r.serverCtx),
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

// addAuthenticationRoutes 认证 / Authentication  接口组
func (r *userRouter) addAuthenticationRoutes() {
	// v1 接口组
	r.server.AddRoutes(
		[]rest.Route{
			{
				// 用户登录
				Method:  http.MethodPost,
				Path:    "/auth/login",
				Handler: user.LoginHandler(r.serverCtx),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}
