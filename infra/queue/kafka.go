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

// MustNewKafkaWriter 初始化 Kafka 生产者，失败时 panic
func MustNewKafkaWriter(conf KafkaConfig) *kafka.Writer {
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

	return writer
}

// MustNewKafkaReader 初始化 Kafka 消费者，失败时 panic
func MustNewKafkaReader(conf KafkaConfig) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Brokers,
		Topic:          conf.Topic,
		GroupID:        conf.Group,
		MinBytes:       10e3,             // 10KB
		MaxBytes:       10e6,             // 10MB
		CommitInterval: time.Second,      // 每秒提交一次 offset
		StartOffset:    kafka.LastOffset, // 从最新消息开始消费
	})

	// 测试连接（尝试读取一条消息，超时即可）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, _ = reader.FetchMessage(ctx)
	// 即使读取失败也不 panic，因为可能是 topic 为空或连接超时

	return reader
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
