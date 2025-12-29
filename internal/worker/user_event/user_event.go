// Package userevent provides example message handlers for different business scenarios.
package userevent

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"

	userRepo "hello-gozero/internal/repository/user"
	kafkaconsumer "hello-gozero/internal/worker/kafka_consumer"
)

// UserEvent 用户事件结构
type UserEvent struct {
	EventType string                 `json:"event_type"` // 事件类型：user_registered, user_updated, user_deleted
	UserID    string                 `json:"user_id"`    // UUID 格式的用户 ID
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// UserEventHandler 用户事件消息处理器
// 用于处理用户相关的事件消息，如：用户注册、用户更新等
type UserEventHandler struct {
	logger logx.Logger

	// 用户仓库
	User userRepo.UserRepository

	// 带缓存的用户仓库
	CachedUser userRepo.CachedUserRepository
}

// NewUserEventHandler 创建用户事件消息处理器
// 需要注入用户仓库实例 [userRepo.UserRepository] 和 带缓存的用户仓库实例 [userRepo.CachedUserRepository] 以便处理业务逻辑
func NewUserEventHandler(
	userRepoInstance userRepo.UserRepository,
	cachedUserRepoInstance userRepo.CachedUserRepository,
) kafkaconsumer.MessageHandler {
	return &UserEventHandler{
		logger:     logx.WithContext(context.Background()),
		User:       userRepoInstance,
		CachedUser: cachedUserRepoInstance,
	}
}

// Handle Implements [MessageHandler.Handle]
func (h *UserEventHandler) Handle(ctx context.Context, message kafka.Message) error {
	// 每一个事件注入 trace_id 和其他元信息，方便日志追踪
	// ctx = logx.ContextWithFields(ctx, logx.Field("trace_id", uuid.New().String()))

	var event UserEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		h.logger.Errorf("Failed to unmarshal user event: %v", err)
		return nil // 返回 nil 以提交 offset，避免重复消费无效消息
	}

	h.logger.WithContext(ctx).Infof("Processing user event: type=%s, user_id=%d", event.EventType, event.UserID)

	// 根据事件类型处理不同的业务逻辑
	switch event.EventType {
	case "user_registered":
		return h.handleUserRegistered(ctx, event)
	case "user_updated":
		return h.handleUserUpdated(ctx, event)
	case "user_deleted":
		return h.handleUserDeleted(ctx, event)
	default:
		h.logger.Infof("Unknown event type: %s", event.EventType)
		return nil
	}
}

// handleUserRegistered 处理用户注册事件
//
// 幂等性保证：
// 在分布式系统中，虽然使用了 Consumer Group，但在极端情况下（如 Rebalance）
// 仍可能出现消息重复消费。因此需要确保业务逻辑的幂等性：
//
// 方案一：使用唯一键约束
// - 在数据库操作中使用 INSERT IGNORE 或 ON DUPLICATE KEY UPDATE
// - 使用 event.UserID 或 message offset 作为幂等键
//
// 方案二：先查询再操作
// - 处理前先检查该事件是否已处理（查询处理记录表）
// - 使用分布式锁防止并发重复处理
//
// 方案三：使用消息去重表
// - 维护一个消息 ID 去重表（包含 message offset/key）
// - 处理前插入去重记录，利用唯一索引防止重复
func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event UserEvent) error {
	h.logger.WithContext(ctx).Infof("User registered: user_id=%d, data=%+v", event.UserID, event.Data)

	// 示例业务逻辑：
	// 1. 发送欢迎邮件 - 需确保邮件服务商的幂等性或记录已发送状态
	// 2. 初始化用户默认设置 - 使用 INSERT IGNORE 或先检查是否存在
	// 3. 触发推荐系统 - 推荐系统应支持重复触发（幂等）
	// 4. 记录用户行为日志 - 使用唯一键（user_id + event_type + timestamp）
	// etc.

	// 幂等性实现示例：
	// 1. 先检查是否已处理
	// exists, err := h.checkEventProcessed(ctx, event.UserID, event.Timestamp)
	// if err != nil {
	//     return fmt.Errorf("failed to check event: %w", err)
	// }
	// if exists {
	//     h.logger.Infof("Event already processed, skipping")
	//     return nil
	// }
	//
	// 2. 处理业务逻辑
	// user, err := h.User.GetByID(ctx, event.UserID)
	// if err != nil {
	//     return fmt.Errorf("failed to get user: %w", err)
	// }
	//
	// 3. 记录已处理标记（使用事务确保原子性）
	// if err := h.markEventProcessed(ctx, event.UserID, event.Timestamp); err != nil {
	//     return fmt.Errorf("failed to mark event: %w", err)
	// }

	return nil
}

// handleUserUpdated 处理用户更新事件
func (h *UserEventHandler) handleUserUpdated(ctx context.Context, event UserEvent) error {
	h.logger.WithContext(ctx).Infof("User updated: user_id=%d, data=%+v", event.UserID, event.Data)

	// 示例业务逻辑：
	// 1. 更新缓存
	// 2. 同步到其他系统
	// 3. 触发相关业务流程
	// etc.

	return nil
}

// handleUserDeleted 处理用户删除事件
func (h *UserEventHandler) handleUserDeleted(ctx context.Context, event UserEvent) error {
	h.logger.WithContext(ctx).Infof("User deleted: user_id=%d", event.UserID)

	// 示例业务逻辑：
	// 1. 清理用户相关数据
	// 2. 清理缓存
	// 3. 通知相关系统
	// etc.

	return nil
}
