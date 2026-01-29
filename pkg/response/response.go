// Package response 定义返回给前端的标准响应格式
package response

// UnifiedAPIResponse 是返回给前端的标准格式
type UnifiedAPIResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`        // 已本地化的用户消息
	Data    interface{} `json:"data,omitempty"` // 业务数据，成功时填充
}

func NewUnifiedAPIResponse(code, message string) *UnifiedAPIResponse {
	return &UnifiedAPIResponse{
		Code:    code,
		Message: message,
	}
}
