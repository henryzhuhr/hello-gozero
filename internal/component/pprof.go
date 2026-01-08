package component

import (
	"context"
	"fmt"
	"net/http"

	"hello-gozero/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

// PprofComponent pprof 性能分析服务组件
type PprofComponent struct {
	config config.PprofConfig
	ready  chan struct{}
	server *http.Server
}

// NewPprofComponent 创建 pprof 组件
func NewPprofComponent(config config.PprofConfig) *PprofComponent {
	return &PprofComponent{
		config: config,
		ready:  make(chan struct{}),
	}
}

// Name Implements [Component.Name]
func (p *PprofComponent) Name() string {
	return "Pprof Server"
}

// Start Implements [Component.Start]
func (p *PprofComponent) Start(ctx context.Context) error {
	if !p.config.Enabled {
		fmt.Println("   ⏭️  Pprof disabled, skipping...")
		close(p.ready) // 标记为就绪
		return nil
	}

	pprofAddr := fmt.Sprintf(":%d", p.config.Port)
	p.server = &http.Server{
		Addr: pprofAddr,
	}

	go func() {
		// 标记为就绪（HTTP 服务器会在 ListenAndServe 时立即监听端口）
		close(p.ready)

		fmt.Printf("   ✅ Pprof server listening at http://localhost%s/debug/pprof/\n", pprofAddr)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Errorf("pprof server failed: %v", err)
		}
	}()

	return nil
}

// Ready Implements [Component.Ready]
func (p *PprofComponent) Ready() <-chan struct{} {
	return p.ready
}

// Stop Implements [Component.Stop]
func (p *PprofComponent) Stop(ctx context.Context) error {
	if p.server == nil {
		return nil
	}
	return p.server.Shutdown(ctx)
}
