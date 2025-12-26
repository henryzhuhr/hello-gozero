package worker

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// ScheduledTask 定时任务接口
type ScheduledTask interface {
	// Execute 执行任务
	Execute(ctx context.Context) error
}

// ScheduledWorker 定时任务后台 Worker
type ScheduledWorker struct {
	name     string
	interval time.Duration
	task     ScheduledTask
	logger   logx.Logger
}

// NewScheduledWorker 创建定时任务 Worker
// name: 任务名称
// interval: 执行间隔
// task: 要执行的任务
func NewScheduledWorker(name string, interval time.Duration, task ScheduledTask, logger logx.Logger) *ScheduledWorker {
	return &ScheduledWorker{
		name:     name,
		interval: interval,
		task:     task,
		logger:   logger,
	}
}

// Name Implements [Worker.Name]
func (w *ScheduledWorker) Name() string {
	return w.name
}

// Start Implements [Worker.Start]
func (w *ScheduledWorker) Start(ctx context.Context) error {
	w.logger.Infof("Scheduled worker [%s] started, interval: %v", w.name, w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 立即执行一次
	if err := w.executeTask(ctx); err != nil {
		w.logger.Errorf("First execution failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Infof("Scheduled worker [%s] received shutdown signal", w.name)
			return nil
		case <-ticker.C:
			if err := w.executeTask(ctx); err != nil {
				w.logger.Errorf("Task execution failed: %v", err)
			}
		}
	}
}

// Stop Implements [Worker.Stop]
func (w *ScheduledWorker) Stop() error {
	w.logger.Infof("Scheduled worker [%s] stopping...", w.name)
	return nil
}

// executeTask 执行任务并记录日志
func (w *ScheduledWorker) executeTask(ctx context.Context) error {
	startTime := time.Now()
	w.logger.Infof("Executing scheduled task [%s]", w.name)

	if err := w.task.Execute(ctx); err != nil {
		w.logger.Errorf("Task [%s] execution failed: %v", w.name, err)
		return err
	}

	duration := time.Since(startTime)
	w.logger.Infof("Task [%s] executed successfully in %v", w.name, duration)
	return nil
}

// ========== 示例定时任务 ==========

// CleanupCacheTask 清理缓存定时任务示例
type CleanupCacheTask struct {
	logger logx.Logger
}

// NewCleanupCacheTask 创建缓存清理任务
func NewCleanupCacheTask(logger logx.Logger) *CleanupCacheTask {
	return &CleanupCacheTask{
		logger: logger,
	}
}

// Execute Implements [ScheduledTask.Execute]
func (t *CleanupCacheTask) Execute(ctx context.Context) error {
	t.logger.Info("Cleaning up expired cache entries...")

	// 这里实现你的清理逻辑
	// 例如：
	// 1. 查找过期的缓存键
	// 2. 删除过期数据
	// 3. 记录清理统计信息

	t.logger.Info("Cache cleanup completed")
	return nil
}
