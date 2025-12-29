// Package kafkaconsumer provides a Kafka consumer worker implementation for processing messages from Kafka topics.
package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hello-gozero/internal/worker"

	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// Handle 处理消息
	Handle(ctx context.Context, message kafka.Message) error
}

// KafkaConsumerWorker Kafka 消费者后台任务
type KafkaConsumerWorker struct {
	name    string
	reader  *kafka.Reader
	handler MessageHandler
	logger  logx.Logger
}

// NewKafkaConsumerWorker 创建 Kafka 消费者后台任务
func NewKafkaConsumerWorker(name string, reader *kafka.Reader, handler MessageHandler, logger logx.Logger) worker.Worker {
	return &KafkaConsumerWorker{
		name:    name,
		reader:  reader,
		handler: handler,
		logger:  logger,
	}
}

// Name 返回 Worker 名称
func (w *KafkaConsumerWorker) Name() string {
	return w.name
}

// Start 启动 Kafka 消费者
func (w *KafkaConsumerWorker) Start(ctx context.Context) error {
	w.logger.Infof("Kafka consumer worker [%s] started, topic: %s", w.name, w.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			w.logger.Infof("Kafka consumer worker [%s] received shutdown signal", w.name)
			return nil
		default:
			// 设置读取超时
			readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			message, err := w.reader.FetchMessage(readCtx)
			cancel()

			if err != nil {
				// 如果是上下文取消，直接返回
				if err == context.Canceled || err == context.DeadlineExceeded {
					continue
				}
				w.logger.Errorf("Failed to fetch message: %v", err)
				time.Sleep(time.Second) // 出错后等待一秒再重试
				continue
			}

			// 处理消息
			// 注意：在分布式环境下，虽然使用了 Consumer Group 避免正常情况下的重复消费，
			// 但在以下情况仍可能重复消费：
			// 1. Consumer Rebalance 期间
			// 2. 消息处理完成但 offset 提交失败
			// 3. Consumer 崩溃重启
			// 因此，业务处理必须设计为幂等操作
			if err := w.processMessage(ctx, message); err != nil {
				w.logger.Errorf("Failed to process message: %v, offset: %d, partition: %d",
					err, message.Offset, message.Partition)
				// 根据业务需求决定是否提交 offset
				// 这里选择提交，避免重复消费导致堆积
				// 如果需要严格的 at-least-once 语义，可以不提交让消息重试
			}

			// 提交 offset
			if err := w.reader.CommitMessages(ctx, message); err != nil {
				w.logger.Errorf("Failed to commit message: %v", err)
			}
		}
	}
}

// Stop 停止 Kafka 消费者
func (w *KafkaConsumerWorker) Stop() error {
	w.logger.Infof("Kafka consumer worker [%s] stopping...", w.name)
	// reader 的关闭由 ServiceContext 统一管理
	return nil
}

// processMessage 处理单条消息
func (w *KafkaConsumerWorker) processMessage(ctx context.Context, message kafka.Message) error {
	startTime := time.Now()

	w.logger.Infof("Processing message - Topic: %s, Partition: %d, Offset: %d, Key: %s",
		message.Topic, message.Partition, message.Offset, string(message.Key))

	// 调用业务处理器
	if err := w.handler.Handle(ctx, message); err != nil {
		return fmt.Errorf("handler error: %w", err)
	}

	duration := time.Since(startTime)
	w.logger.Infof("Message processed successfully in %v - Offset: %d", duration, message.Offset)

	return nil
}

// ========== 示例消息处理器 ==========

// ExampleMessageHandler 示例消息处理器
type ExampleMessageHandler struct {
	logger logx.Logger
}

// NewExampleMessageHandler 创建示例消息处理器
func NewExampleMessageHandler(logger logx.Logger) *ExampleMessageHandler {
	return &ExampleMessageHandler{
		logger: logger,
	}
}

// Handle 处理消息
func (h *ExampleMessageHandler) Handle(ctx context.Context, message kafka.Message) error {
	// 解析消息（这里假设是 JSON 格式）
	var data map[string]interface{}
	if err := json.Unmarshal(message.Value, &data); err != nil {
		h.logger.Errorf("Failed to unmarshal message: %v", err)
		// 可以选择返回 error 或者记录后继续
		return nil
	}

	// 处理业务逻辑
	h.logger.Infof("Received message data: %+v", data)

	// 这里添加你的业务逻辑
	// 例如：
	// 1. 更新数据库
	// 2. 调用其他服务
	// 3. 发送通知
	// etc.

	return nil
}
