// Package worker 提供后台任务管理功能
package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

// Worker 后台任务接口
type Worker interface {
	// Start 启动后台任务
	Start(ctx context.Context) error
	// Stop 停止后台任务
	Stop() error
	// Name 返回任务名称
	Name() string
}

// Manager 后台任务管理器
type Manager struct {
	workers []Worker
	wg      sync.WaitGroup
	logger  logx.Logger
}

// NewManager 创建后台任务管理器
func NewManager(logger logx.Logger) *Manager {
	return &Manager{
		workers: make([]Worker, 0),
		logger:  logger,
	}
}

// Register 注册后台任务
func (m *Manager) Register(worker Worker) {
	m.workers = append(m.workers, worker)
}

// Start 启动所有后台任务
func (m *Manager) Start(ctx context.Context) error {
	if len(m.workers) == 0 {
		m.logger.Info("No workers to start")
		return nil
	}

	m.logger.Infof("Starting %d workers...", len(m.workers))

	for _, worker := range m.workers {
		w := worker // 避免闭包问题
		m.wg.Add(1)

		go func() {
			defer m.wg.Done()

			m.logger.Infof("Worker [%s] starting...", w.Name())
			if err := w.Start(ctx); err != nil {
				m.logger.Errorf("Worker [%s] stopped with error: %v", w.Name(), err)
			} else {
				m.logger.Infof("Worker [%s] stopped gracefully", w.Name())
			}
		}()
	}

	return nil
}

// Stop 停止所有后台任务
func (m *Manager) Stop() error {
	m.logger.Info("Stopping all workers...")

	// 等待所有任务完成
	m.wg.Wait()

	// 调用每个 worker 的 Stop 方法进行清理
	var errs []error
	for _, worker := range m.workers {
		if err := worker.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("worker [%s] stop error: %w", worker.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some workers failed to stop: %v", errs)
	}

	m.logger.Info("All workers stopped successfully")
	return nil
}
