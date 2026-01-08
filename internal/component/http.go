package component

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hello-gozero/internal/config"
	"hello-gozero/internal/middleware"
	"hello-gozero/internal/routes"
	"hello-gozero/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// HTTPServerComponent GoZero HTTP 服务组件
type HTTPServerComponent struct {
	config config.Config
	svcCtx *svc.ServiceContext
	server *rest.Server
	ready  chan struct{}
}

// NewHTTPServerComponent 创建 HTTP 服务组件
func NewHTTPServerComponent(config config.Config, svcCtx *svc.ServiceContext) *HTTPServerComponent {
	return &HTTPServerComponent{
		config: config,
		svcCtx: svcCtx,
		ready:  make(chan struct{}),
	}
}

// Name Implements [Component.Name]
func (h *HTTPServerComponent) Name() string {
	return "HTTP Server"
}

// Start Implements [Component.Start]
func (h *HTTPServerComponent) Start(ctx context.Context) error {
	// 创建 HTTP 服务
	h.server = rest.MustNewServer(h.config.RestConf)

	// 注册全局中间件
	h.server.Use(middleware.NewUserAgentMiddleware().Handle)

	// 注册路由
	routes.RegisterHandlers(h.server, h.svcCtx)

	// 启动 HTTP 服务（非阻塞）
	go func() {
		h.server.Start()
	}()

	// 等待服务启动
	time.Sleep(200 * time.Millisecond)

	// 健康检查：确保服务真正可用
	if err := h.healthCheck(); err != nil {
		return fmt.Errorf("HTTP server health check failed: %w", err)
	}

	fmt.Printf("   ✅ HTTP server listening at %s:%d\n", h.config.Host, h.config.Port)

	// 标记为就绪
	close(h.ready)
	return nil
}

// Ready Implements [Component.Ready]
func (h *HTTPServerComponent) Ready() <-chan struct{} {
	return h.ready
}

// Stop Implements [Component.Stop]
func (h *HTTPServerComponent) Stop(ctx context.Context) error {
	if h.server != nil {
		h.server.Stop()
	}
	return nil
}

// healthCheck 执行健康检查
func (h *HTTPServerComponent) healthCheck() error {
	addr := fmt.Sprintf("%s:%d", h.config.Host, h.config.Port)
	if h.config.Host == "" || h.config.Host == "0.0.0.0" {
		addr = fmt.Sprintf("localhost:%d", h.config.Port)
	}

	// 健康检查路由在 /api/health
	healthURL := fmt.Sprintf("http://%s/api/health", addr)
	resp, err := http.Get(healthURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
