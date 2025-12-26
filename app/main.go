// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
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

var configFile = flag.String("f", "etc/hellogozero.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	svcCtx, err := svc.NewServiceContext(c)
	if err != nil {
		fmt.Printf("failed to create service context: %v\n", err)
		return
	}
	defer svcCtx.Close()

	// 创建后台任务管理器
	workerManager := setupWorkers(svcCtx)

	// 创建用于控制 worker 的 context
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	// 启动后台任务
	if err := workerManager.Start(workerCtx); err != nil {
		fmt.Printf("failed to start workers: %v\n", err)
		return
	}
	defer workerManager.Stop()

	// 创建 HTTP 服务
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册全局中间件
	server.Use(middleware.NewUserAgentMiddleware().Handle)

	// 注册路由
	routes.RegisterHandlers(server, svcCtx)

	// 启动 HTTP 服务（非阻塞）
	go func() {
		// 启动服务
		fmt.Printf("🚀 Starting server at %s:%d...\n", c.Host, c.Port)
		server.Start()
	}()

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
