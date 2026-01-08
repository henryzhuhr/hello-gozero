// Package component 定义统一的组件启动/停止接口规范
package component

import (
	"context"
	"fmt"
	"time"
)

// Component 定义组件的生命周期接口
type Component interface {
	// Name 返回组件名称，用于日志和标识
	Name() string

	// Start 启动组件，启动逻辑应该是非阻塞的
	Start(ctx context.Context) error

	// Ready 返回一个会在组件就绪时关闭的 channel
	// 调用者可以通过 select 等待组件完全启动
	Ready() <-chan struct{}

	// Stop 优雅停止组件
	Stop(ctx context.Context) error
}

// Manager 组件管理器，负责统一管理多个组件的启动和停止
type Manager struct {
	components []Component
	timeout    time.Duration
}

// NewManager 创建组件管理器
func NewManager(timeout time.Duration) *Manager {
	if timeout == 0 {
		timeout = 30 * time.Second // 默认超时时间
	}
	return &Manager{
		components: make([]Component, 0),
		timeout:    timeout,
	}
}

// Register 注册组件
func (m *Manager) Register(comp Component) {
	m.components = append(m.components, comp)
}

// StartAll 按顺序启动所有组件，等待每个组件就绪后再启动下一个
func (m *Manager) StartAll(ctx context.Context) error {
	for i, comp := range m.components {
		fmt.Printf("📍 [%d/%d] Starting %s...\n", i+1, len(m.components), comp.Name())

		// 启动组件
		if err := comp.Start(ctx); err != nil {
			return fmt.Errorf("failed to start %s: %w", comp.Name(), err)
		}

		// 等待组件就绪或超时
		select {
		case <-comp.Ready():
			fmt.Printf("   ✅ %s started successfully\n", comp.Name())
		case <-time.After(m.timeout):
			return fmt.Errorf("timeout waiting for %s to be ready", comp.Name())
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while starting %s", comp.Name())
		}
	}

	fmt.Println("✅ All components started successfully!")
	return nil
}

// StopAll 按逆序停止所有组件（后启动的先停止）
func (m *Manager) StopAll(ctx context.Context) error {
	var lastErr error

	// 逆序停止
	for i := len(m.components) - 1; i >= 0; i-- {
		comp := m.components[i]
		fmt.Printf("   🛑 Stopping %s...\n", comp.Name())

		if err := comp.Stop(ctx); err != nil {
			fmt.Printf("   ⚠️  Failed to stop %s: %v\n", comp.Name(), err)
			lastErr = err
		} else {
			fmt.Printf("   ✅ %s stopped\n", comp.Name())
		}
	}

	return lastErr
}
