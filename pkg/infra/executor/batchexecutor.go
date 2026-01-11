package executor

import (
	"context"
	"fmt"
	"sync"
)

// PanicError 表示任务执行过程中发生的 panic
type PanicError struct {
	TaskID string
	Value  interface{}
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("task %s panicked: %v", e.TaskID, e.Value)
}

// RequestTask 定义单个请求任务的接口
// 业务方需要实现此接口来定义具体的请求逻辑
type RequestTask[T any] interface {
	// Execute 执行请求任务
	// 返回结果和可能的错误
	Execute(ctx context.Context) (T, error)

	// GetID 获取任务的唯一标识
	GetID() string
}

// BatchRequestConfig 批量请求的配置
type BatchRequestConfig struct {
	// MaxConcurrency 最大并发数
	// 控制同时执行的任务数量，避免资源耗尽
	MaxConcurrency int
}

// BatchRequestResult 批量请求的单个结果
type BatchRequestResult[T any] struct {
	// ID 任务的唯一标识
	ID string
	// Data 请求返回的数据
	Data T
	// Err 请求过程中的错误
	Err error
}

// BatchRequestExecutor 批量请求执行器
// 提供通用的并发请求处理能力，独立于具体业务逻辑
type BatchRequestExecutor[T any] struct {
	config BatchRequestConfig
}

// NewBatchRequestExecutor 创建批量请求执行器
func NewBatchRequestExecutor[T any](config BatchRequestConfig) *BatchRequestExecutor[T] {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 1
	}
	return &BatchRequestExecutor[T]{
		config: config,
	}
}

// Execute 执行批量请求
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - tasks: 要执行的任务列表
//
// 返回:
//   - []BatchRequestResult[T]: 所有任务的执行结果
//   - error: 执行过程中的严重错误（如 context 取消、panic 等）
//
// 特性:
//   - 使用 semaphore 控制并发数量
//   - 支持通过 context 取消请求
//   - 使用独立的消费者 goroutine 收集结果，避免 channel 阻塞
//   - 所有 panic 都会被捕获并转换为 error
//   - 个别任务失败不会影响整体执行，错误保存在 BatchRequestResult.Err 中
func (e *BatchRequestExecutor[T]) Execute(ctx context.Context, tasks []RequestTask[T]) ([]BatchRequestResult[T], error) {
	if len(tasks) == 0 {
		return []BatchRequestResult[T]{}, nil
	}

	// 用于接收任务执行结果
	// 注释说明: 使用无缓冲 channel 配合独立的消费者 goroutine
	// 这样可以确保生产者不会因为 channel 满而阻塞
	// 消费者会持续读取直到 channel 关闭
	ch := make(chan BatchRequestResult[T])

	// 控制并发数的信号量
	sem := make(chan struct{}, e.config.MaxConcurrency)

	// 等待所有生产者完成
	var wg sync.WaitGroup

	// 启动生产者 goroutines
	for _, task := range tasks {
		wg.Add(1)
		go func(t RequestTask[T]) {
			defer wg.Done()
			result, err := e.executeTask(ctx, t, sem)
			if err != nil {
				// executeTask 层面的严重错误（如 panic）
				result = BatchRequestResult[T]{
					ID:  t.GetID(),
					Err: err,
				}
			}
			// 发送结果到 channel
			select {
			case ch <- result:
			case <-ctx.Done():
				// context 已取消，放弃发送结果
			}
		}(task)
	}

	// 启动消费者 goroutine 收集结果
	resultsChan := make(chan []BatchRequestResult[T], 1)
	go e.collectResults(ctx, ch, resultsChan)

	// 等待所有生产者完成
	wg.Wait()

	// 关闭 channel 通知消费者停止
	close(ch)

	// 等待消费者完成并获取结果
	results := <-resultsChan

	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return results, ctx.Err()
	}

	return results, nil
}

// executeTask 执行单个任务
// 返回任务执行结果和可能的严重错误（如 panic）
// 任务业务逻辑的错误会包含在 BatchRequestResult.Err 中
func (e *BatchRequestExecutor[T]) executeTask(
	ctx context.Context,
	task RequestTask[T],
	sem chan struct{},
) (BatchRequestResult[T], error) {
	var result BatchRequestResult[T]
	var panicErr error

	defer func() {
		if r := recover(); r != nil {
			// 捕获 panic 并转换为错误返回给上层
			panicErr = &PanicError{
				TaskID: task.GetID(),
				Value:  r,
			}
		}
	}()

	// 获取信号量
	sem <- struct{}{}
	defer func() { <-sem }()

	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return BatchRequestResult[T]{
			ID:  task.GetID(),
			Err: fmt.Errorf("task [%s] cancelled: %w", task.GetID(), ctx.Err()),
		}, nil
	default:
	}

	// 执行任务
	data, err := task.Execute(ctx)

	result = BatchRequestResult[T]{
		ID:   task.GetID(),
		Data: data,
		Err:  err,
	}

	// 如果有 panic 错误，返回它
	if panicErr != nil {
		return result, panicErr
	}

	return result, nil
}

// collectResults 收集所有任务的执行结果
func (e *BatchRequestExecutor[T]) collectResults(
	ctx context.Context,
	ch <-chan BatchRequestResult[T],
	resultsChan chan<- []BatchRequestResult[T],
) {
	var results []BatchRequestResult[T]
	defer func() {
		resultsChan <- results
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-ch:
			if !ok {
				return
			}
			results = append(results, result)
		}
	}
}
