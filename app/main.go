// Main entry point for the Hello GoZero application.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof" // 导入 pprof
	"os"
	"os/signal"
	"syscall"
	"time"

	"hello-gozero/internal/config"
	"hello-gozero/internal/middleware"
	"hello-gozero/internal/routes"
	"hello-gozero/internal/svc"
	"hello-gozero/internal/worker"
	kafkaconsumer "hello-gozero/internal/worker/kafka_consumer"
	userevent "hello-gozero/internal/worker/user_event"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

// 全局配置文件路径
var configFile = flag.String("f", "etc/hellogozero.yaml", "the config file")

func main() {
	flag.Parse() // 加载配置文件

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	svcCtx, err := svc.NewServiceContext(c)
	if err != nil {
		fmt.Printf("failed to create service context: %v\n", err)
		return
	}
	defer svcCtx.Close()

	// ========== 按顺序启动各个组件 ==========

	// 1. 启动 pprof 性能分析服务
	fmt.Println("📍 [1/3] Starting pprof server...")
	fmt.Println("📍 [1/3] 启动 pprof 服务...")
	if err := startPprofServer(c.Pprof); err != nil {
		fmt.Printf("❌ Failed to start pprof: %v\n", err)
		return
	}

	// 2. 启动后台 Worker 任务
	fmt.Println("📍 [2/3] Starting background workers...")
	fmt.Println("📍 [2/3] 启动后台任务...")
	cancelWorkers, workerManager := startWorkers(svcCtx)
	if workerManager == nil {
		fmt.Println("❌ Failed to start workers")
		return
	}
	defer cancelWorkers()
	defer workerManager.Stop()

	// 3. 启动 GoZero HTTP 服务
	fmt.Println("📍 [3/3] Starting HTTP server...")
	fmt.Println("📍 [3/3] 启动 HTTP 服务...")
	server, err := startHTTPServer(c, svcCtx)
	if err != nil {
		fmt.Printf("❌ Failed to start HTTP server: %v\n", err)
		cancelWorkers()
		workerManager.Stop()
		return
	}
	defer server.Stop()

	fmt.Println("✅ All components started successfully!")

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down gracefully...")

	// 优雅关闭：先停止接收新请求，再停止后台任务
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 关闭 HTTP 服务
	server.Stop()

	// 取消 worker context，通知所有 worker 停止
	cancelWorkers()

	// 等待后台任务完成
	if err := workerManager.Stop(); err != nil {
		fmt.Printf("failed to stop workers: %v\n", err)
	}

	select {
	case <-shutdownCtx.Done():
		fmt.Println("⚠️ Shutdown timeout exceeded")
	default:
		fmt.Println("✅ Server stopped successfully")
	}
}

// startPprofServer 启动 pprof 性能分析服务
func startPprofServer(pprofConf config.PprofConfig) error {
	if !pprofConf.Enabled {
		fmt.Println("   ⏭️  Pprof disabled, skipping...")
		return nil
	}

	pprofAddr := fmt.Sprintf(":%d", pprofConf.Port)
	go func() {
		fmt.Printf("   ✅ Pprof server started at http://localhost%s/debug/pprof/\n", pprofAddr)
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			logx.Errorf("pprof server failed: %v", err)
		}
	}()
	// 等待一小段时间确保服务启动
	time.Sleep(100 * time.Millisecond)
	return nil
}

// startWorkers 启动后台 Worker 任务
func startWorkers(svcCtx *svc.ServiceContext) (context.CancelFunc, *worker.Manager) {
	// 创建后台任务管理器
	workerManager := setupWorkers(svcCtx)

	// 创建用于控制 worker 的 context
	workerCtx, cancelWorkers := context.WithCancel(context.Background())

	// 启动后台任务
	if err := workerManager.Start(workerCtx); err != nil {
		fmt.Printf("   ❌ Failed to start workers: %v\n", err)
		cancelWorkers()
		return nil, nil
	}

	// 等待一小段时间确保 workers 完全启动
	time.Sleep(100 * time.Millisecond)
	fmt.Println("   ✅ All workers started successfully")

	return cancelWorkers, workerManager
}

// startHTTPServer 启动 GoZero HTTP 服务
func startHTTPServer(c config.Config, svcCtx *svc.ServiceContext) (*rest.Server, error) {
	// 创建 HTTP 服务
	server := rest.MustNewServer(c.RestConf)

	// 注册全局中间件
	server.Use(middleware.NewUserAgentMiddleware().Handle)

	// 注册路由
	routes.RegisterHandlers(server, svcCtx)

	// 使用 channel 等待服务启动完成
	started := make(chan error, 1)

	// 启动 HTTP 服务（非阻塞）
	go func() {
		defer close(started)
		server.Start()
	}()

	// 等待服务启动并验证
	time.Sleep(200 * time.Millisecond)

	// 健康检查：尝试连接服务端口
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	if c.Host == "" || c.Host == "0.0.0.0" {
		addr = fmt.Sprintf("localhost:%d", c.Port)
	}

	healthURL := fmt.Sprintf("http://%s/health", addr)
	resp, err := http.Get(healthURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP server health check failed: %w", err)
	}
	resp.Body.Close()

	fmt.Printf("   ✅ HTTP server started at %s:%d (health check passed)\n", c.Host, c.Port)
	return server, nil
}

// setupWorkers 配置并返回后台任务管理器
func setupWorkers(svcCtx *svc.ServiceContext) *worker.Manager {
	logger := logx.WithContext(context.Background())
	manager := worker.NewManager(logger)

	// 示例 1: 注册 Kafka 消费者任务 - 用户事件处理
	userEventHandler := userevent.NewUserEventHandler(
		svcCtx.Repository.User,
		svcCtx.Repository.CachedUser,
	)
	userEventWorker := kafkaconsumer.NewKafkaConsumerWorker(
		"user-event-consumer",
		svcCtx.Infra.KafkaReader,
		userEventHandler,
		logger,
	)
	manager.Register(userEventWorker)

	// 示例 2: 可以注册更多的后台任务
	// 例如：定时任务、另一个 Kafka 消费者等
	// exampleHandler := worker.NewExampleMessageHandler(logger)
	// exampleWorker := worker.NewKafkaConsumerWorker(
	// 	"example-consumer",
	// 	anotherKafkaReader,
	// 	exampleHandler,
	// 	logger,
	// )
	// manager.Register(exampleWorker)

	return manager
}
