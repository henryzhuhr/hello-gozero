package component

import (
	"context"
	"fmt"

	"hello-gozero/internal/svc"
	"hello-gozero/internal/worker"
	kafkaconsumer "hello-gozero/internal/worker/kafka_consumer"
	userevent "hello-gozero/internal/worker/user_event"

	"github.com/zeromicro/go-zero/core/logx"
)

// WorkerComponent 后台任务组件
type WorkerComponent struct {
	svcCtx        *svc.ServiceContext
	manager       *worker.Manager
	cancelFunc    context.CancelFunc
	workerContext context.Context
	ready         chan struct{}
}

// NewWorkerComponent 创建后台任务组件
func NewWorkerComponent(svcCtx *svc.ServiceContext) *WorkerComponent {
	return &WorkerComponent{
		svcCtx: svcCtx,
		ready:  make(chan struct{}),
	}
}

// Name Implements [Component.Name]
func (w *WorkerComponent) Name() string {
	return "Background Workers"
}

// Start Implements [Component.Start]
func (w *WorkerComponent) Start(ctx context.Context) error {
	// 创建后台任务管理器
	w.manager = w.setupWorkers()

	// 创建用于控制 worker 的 context
	w.workerContext, w.cancelFunc = context.WithCancel(context.Background())

	// 启动后台任务
	if err := w.manager.Start(w.workerContext); err != nil {
		return fmt.Errorf("failed to start workers: %w", err)
	}

	// 标记为就绪
	close(w.ready)
	return nil
}

// Ready Implements [Component.Ready]
func (w *WorkerComponent) Ready() <-chan struct{} {
	return w.ready
}

// Stop Implements [Component.Stop]
func (w *WorkerComponent) Stop(ctx context.Context) error {
	if w.cancelFunc != nil {
		w.cancelFunc()
	}

	if w.manager != nil {
		return w.manager.Stop()
	}

	return nil
}

// setupWorkers 配置并返回后台任务管理器
func (w *WorkerComponent) setupWorkers() *worker.Manager {
	logger := logx.WithContext(context.Background())
	manager := worker.NewManager(logger)

	// 注册 Kafka 消费者任务 - 用户事件处理
	userEventHandler := userevent.NewUserEventHandler(
		w.svcCtx.Repository.User,
		w.svcCtx.Repository.CachedUser,
	)
	userEventWorker := kafkaconsumer.NewKafkaConsumerWorker(
		"user-event-consumer",
		w.svcCtx.Infra.KafkaReader,
		userEventHandler,
		logger,
	)
	manager.Register(userEventWorker)

	// 可以注册更多的后台任务
	// 例如：定时任务、另一个 Kafka 消费者等

	return manager
}
