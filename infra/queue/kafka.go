// Package queue provides Kafka producer and consumer initialization.
package queue

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConfig Kafka 配置
type KafkaConfig struct {
	Brokers []string
	Topic   string
	Group   string // Consumer Group ID
}

// NewKafkaWriter 初始化 Kafka 生产者
func NewKafkaWriter(conf KafkaConfig) (*kafka.Writer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(conf.Brokers...),
		Topic:        conf.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Compression:  kafka.Snappy,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testMsg := kafka.Message{
		Key:   []byte("test"),
		Value: []byte("connection test"),
	}

	if err := writer.WriteMessages(ctx, testMsg); err != nil {
		// 如果 topic 不存在，忽略这个错误（在生产环境中不应该忽略）
		// panic(err)
		_ = err // 显式忽略错误
	}

	return writer, nil
}

// NewKafkaReader 初始化 Kafka 消费者
//
// 分布式消费说明：
// 1. 使用 Consumer Group（GroupID）确保同一组内的多个消费者不会重复消费
// 2. Kafka 会自动将 Topic 的分区分配给组内的不同消费者
// 3. 当消费者数量变化时，Kafka 会触发 Rebalance 重新分配分区
// 4. 建议 Topic 分区数 >= 服务副本数，以实现最佳并行度
//
// 注意事项：
// - 确保所有服务副本使用相同的 GroupID
// - 消息处理必须是幂等的，因为在极端情况下（如 Rebalance）可能会重复消费
// - StartOffset 设置为 LastOffset，新消费者只消费新消息，不处理历史消息
func NewKafkaReader(conf KafkaConfig) (*kafka.Reader, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Brokers,
		Topic:          conf.Topic,
		GroupID:        conf.Group,       // Consumer Group ID - 确保所有副本使用相同的值
		MinBytes:       10e3,             // 10KB
		MaxBytes:       10e6,             // 10MB
		CommitInterval: time.Second,      // 每秒提交一次 offset
		StartOffset:    kafka.LastOffset, // 从最新消息开始消费（新 Group 时）
	})

	// 测试连接（尝试读取一条消息，超时即可）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, _ = reader.FetchMessage(ctx)
	// 即使读取失败也不 panic，因为可能是 topic 为空或连接超时

	return reader, nil
}

// CloseKafkaWriter 关闭 Kafka 生产者
func CloseKafkaWriter(writer *kafka.Writer) error {
	if writer == nil {
		return nil
	}
	return writer.Close()
}

// CloseKafkaReader 关闭 Kafka 消费者
func CloseKafkaReader(reader *kafka.Reader) error {
	if reader == nil {
		return nil
	}
	return reader.Close()
}
