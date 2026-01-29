// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"errors"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	userDto "hello-gozero/internal/dto/user"
	userService "hello-gozero/internal/service/user"
	"hello-gozero/internal/svc"
	"hello-gozero/pkg/errno"
	"hello-gozero/pkg/locale"
	"hello-gozero/pkg/response"
)

const registerUser = "register_user"

// RegisterUserHandler 注册用户
func RegisterUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := logx.ContextWithFields(r.Context(), logx.Field("handler", registerUser))
		var req userDto.RegisterUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		// 参数校验
		if err := req.Validate(); err != nil {
			switch v := err.(type) {
			case userDto.RegisterUserValidationError:
				// 返回结构化的校验错误信息
				httpx.WriteJsonCtx(ctx, w, http.StatusBadRequest, map[string]interface{}{"error": v.ToMap()})
			default:
				// 其他错误，返回通用错误信息
				httpx.ErrorCtx(ctx, w, err)
			}
			return
		}

		// 调用服务层
		svc := userService.NewRegisterUserService(ctx, svcCtx)
		resp, err := svc.RegisterUser(&req)
		ctx = svc.GetCtx() // 使用服务层的上下文以包含日志字段
		if err != nil {
			svc.Logger.WithContext(ctx).Errorf("failed to get user: %s", err)

			// 检查错误，返回前端
			var bizErr *errno.BizError
			if errors.As(err, &bizErr) {
				lang := "zh-CN"
				userMsg, err := locale.T(lang, &i18n.LocalizeConfig{MessageID: bizErr.UserMessage})
				if err != nil {
					svc.Logger.WithContext(ctx).Errorf("failed to localize message: %s", err)
					userMsg = bizErr.UserMessage // 回退到默认消息
				}
				httpx.WriteJsonCtx(ctx, w, bizErr.HTTPStatus,
					response.NewUnifiedAPIResponse(bizErr.Code, userMsg),
				)
			} else {
				// 默认情况，内部服务错误
				httpx.ErrorCtx(ctx, w, err)
			}
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
