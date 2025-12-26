# 后台任务系统使用指南

本项目集成了一个灵活的后台任务管理系统，支持 Kafka 消息消费、定时任务等多种后台任务类型。

## 架构说明

### 核心组件

1. **Worker Interface** (`internal/worker/manager.go`)
   - 所有后台任务的统一接口
   - 提供 Start、Stop、Name 三个方法

2. **Manager** (`internal/worker/manager.go`)
   - 后台任务管理器
   - 负责任务的注册、启动、停止
   - 管理任务的生命周期

3. **KafkaConsumerWorker** (`internal/worker/kafka_consumer.go`)
   - Kafka 消息消费者 Worker
   - 自动处理消息获取、提交 offset
   - 支持优雅关闭

4. **ScheduledWorker** (`internal/worker/scheduled_worker.go`)
   - 定时任务 Worker
   - 支持周期性执行任务
   - 可配置执行间隔

## 快速开始

### 1. 创建消息处理器

实现 `MessageHandler` 接口来处理 Kafka 消息：

```go
type MyMessageHandler struct {
    svcCtx *svc.ServiceContext
    logger logx.Logger
}

func NewMyMessageHandler(svcCtx *svc.ServiceContext) *MyMessageHandler {
    return &MyMessageHandler{
        svcCtx: svcCtx,
        logger: logx.WithContext(context.Background()),
    }
}

// Handle 实现 MessageHandler 接口
func (h *MyMessageHandler) Handle(ctx context.Context, message kafka.Message) error {
    // 解析消息
    var data MyData
    if err := json.Unmarshal(message.Value, &data); err != nil {
        return err
    }
    
    // 处理业务逻辑
    // ...
    
    return nil
}
```

### 2. 注册 Kafka 消费者任务

在 `app/main.go` 的 `setupWorkers` 函数中注册：

```go
func setupWorkers(svrCtx *svc.ServiceContext) *worker.Manager {
    logger := logx.WithContext(context.Background())
    manager := worker.NewManager(logger)
    
    // 注册 Kafka 消费者
    handler := NewMyMessageHandler(svrCtx)
    kafkaWorker := worker.NewKafkaConsumerWorker(
        "my-consumer-worker",      // Worker 名称
        svrCtx.Infra.KafkaReader,  // Kafka Reader
        handler,                    // 消息处理器
        logger,                     // 日志器
    )
    manager.Register(kafkaWorker)
    
    return manager
}
```

### 3. 创建定时任务

实现 `ScheduledTask` 接口：

```go
type MyScheduledTask struct {
    svcCtx *svc.ServiceContext
    logger logx.Logger
}

func NewMyScheduledTask(svcCtx *svc.ServiceContext) *MyScheduledTask {
    return &MyScheduledTask{
        svcCtx: svcCtx,
        logger: logx.WithContext(context.Background()),
    }
}

// Execute 实现 ScheduledTask 接口
func (t *MyScheduledTask) Execute(ctx context.Context) error {
    t.logger.Info("Executing scheduled task...")
    
    // 执行定时任务逻辑
    // 例如：清理过期数据、生成报表、同步数据等
    
    return nil
}
```

### 4. 注册定时任务

```go
func setupWorkers(svrCtx *svc.ServiceContext) *worker.Manager {
    logger := logx.WithContext(context.Background())
    manager := worker.NewManager(logger)
    
    // 注册定时任务（每 5 分钟执行一次）
    task := NewMyScheduledTask(svrCtx)
    scheduledWorker := worker.NewScheduledWorker(
        "my-scheduled-task",    // 任务名称
        5*time.Minute,          // 执行间隔
        task,                   // 任务实现
        logger,                 // 日志器
    )
    manager.Register(scheduledWorker)
    
    return manager
}
```

## 多个 Kafka 消费者示例

如果需要消费多个 Kafka Topic，需要为每个 Topic 创建独立的 Reader：

```go
// 在 ServiceContext 中添加多个 Reader
type Infra struct {
    // ...existing fields...
    
    // 用户事件消费者
    UserEventReader *kafka.Reader
    // 订单事件消费者
    OrderEventReader *kafka.Reader
}

// 在 NewServiceContext 中初始化
func NewServiceContext(c config.Config) (*ServiceContext, error) {
    // ...existing code...
    
    // 为不同的 Topic 创建不同的 Reader
    userEventReader := queue.MustNewKafkaReader(queue.KafkaConfig{
        Brokers: c.Infra.Kafka.Brokers,
        Topic:   "user-events",
        Group:   "user-event-consumer-group",
    })
    
    orderEventReader := queue.MustNewKafkaReader(queue.KafkaConfig{
        Brokers: c.Infra.Kafka.Brokers,
        Topic:   "order-events",
        Group:   "order-event-consumer-group",
    })
    
    // ...
}

// 在 setupWorkers 中注册多个消费者
func setupWorkers(svrCtx *svc.ServiceContext) *worker.Manager {
    logger := logx.WithContext(context.Background())
    manager := worker.NewManager(logger)
    
    // 用户事件消费者
    userHandler := worker.NewUserEventHandler(svrCtx)
    userWorker := worker.NewKafkaConsumerWorker(
        "user-event-consumer",
        svrCtx.Infra.UserEventReader,
        userHandler,
        logger,
    )
    manager.Register(userWorker)
    
    // 订单事件消费者
    orderHandler := NewOrderEventHandler(svrCtx)
    orderWorker := worker.NewKafkaConsumerWorker(
        "order-event-consumer",
        svrCtx.Infra.OrderEventReader,
        orderHandler,
        logger,
    )
    manager.Register(orderWorker)
    
    return manager
}
```

## 优雅关闭

系统已经实现了优雅关闭机制：

1. 监听 `SIGINT` 和 `SIGTERM` 信号
2. 收到信号后停止接收新的 HTTP 请求
3. 停止所有后台任务
4. 等待所有任务完成（最多 30 秒）
5. 关闭所有资源连接

```bash
# 发送停止信号
kill -TERM <pid>

# 或使用 Ctrl+C
```

## 最佳实践

### 1. 消息处理幂等性

确保消息处理逻辑是幂等的，因为在某些情况下可能会重复消费同一条消息：

```go
func (h *MyMessageHandler) Handle(ctx context.Context, message kafka.Message) error {
    // 检查消息是否已处理
    messageID := string(message.Key)
    if h.isProcessed(ctx, messageID) {
        h.logger.Infof("Message already processed: %s", messageID)
        return nil
    }
    
    // 处理消息
    if err := h.processMessage(ctx, message); err != nil {
        return err
    }
    
    // 标记消息已处理
    return h.markAsProcessed(ctx, messageID)
}
```

### 2. 错误处理策略

根据错误类型决定是否重试：

```go
func (h *MyMessageHandler) Handle(ctx context.Context, message kafka.Message) error {
    if err := h.processMessage(ctx, message); err != nil {
        // 业务错误：记录但不返回错误，避免阻塞消费
        if errors.Is(err, ErrBusinessLogic) {
            h.logger.Errorf("Business error: %v", err)
            return nil
        }
        
        // 系统错误：返回错误，触发重试
        return fmt.Errorf("system error: %w", err)
    }
    return nil
}
```

### 3. 监控和告警

添加指标收集和告警：

```go
func (h *MyMessageHandler) Handle(ctx context.Context, message kafka.Message) error {
    start := time.Now()
    defer func() {
        // 记录处理时间
        duration := time.Since(start)
        h.metrics.RecordProcessingTime(duration)
    }()
    
    // 处理消息
    if err := h.processMessage(ctx, message); err != nil {
        h.metrics.RecordError()
        return err
    }
    
    h.metrics.RecordSuccess()
    return nil
}
```

### 4. 限流和背压

处理高并发场景：

```go
type RateLimitedHandler struct {
    handler MessageHandler
    limiter *rate.Limiter
    logger  logx.Logger
}

func (h *RateLimitedHandler) Handle(ctx context.Context, message kafka.Message) error {
    // 等待限流器许可
    if err := h.limiter.Wait(ctx); err != nil {
        return fmt.Errorf("rate limiter error: %w", err)
    }
    
    return h.handler.Handle(ctx, message)
}
```

## 故障排查

### 查看日志

后台任务的日志会包含以下信息：

- Worker 启动和停止
- 消息处理状态
- 错误信息

### 常见问题

1. **消息堆积**
   - 检查处理逻辑是否有性能问题
   - 考虑增加消费者实例
   - 检查是否有错误导致消息处理失败

2. **重复消费**
   - 确保消息处理是幂等的
   - 检查 offset 提交是否正常

3. **任务无法停止**
   - 检查任务是否正确处理 context 取消信号
   - 增加关闭超时时间

## 扩展开发

### 自定义 Worker

实现 `Worker` 接口创建自定义后台任务：

```go
type MyCustomWorker struct {
    name   string
    logger logx.Logger
}

func (w *MyCustomWorker) Name() string {
    return w.name
}

func (w *MyCustomWorker) Start(ctx context.Context) error {
    // 实现你的后台任务逻辑
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // 执行任务
            time.Sleep(time.Second)
        }
    }
}

func (w *MyCustomWorker) Stop() error {
    // 清理资源
    return nil
}
```

## 总结

本后台任务系统提供了：

- ✅ 统一的任务管理接口
- ✅ Kafka 消息消费支持
- ✅ 定时任务支持
- ✅ 优雅关闭机制
- ✅ 完善的错误处理
- ✅ 易于扩展

根据业务需求，可以轻松添加新的后台任务类型。
