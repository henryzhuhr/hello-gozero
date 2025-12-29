# Kafka 分布式消费说明

## 问题：多个服务副本会重复消费吗?

**答案：不会！** 只要配置正确，使用相同的 Consumer Group ID，多个服务副本不会重复消费同一条消息。

## 工作原理

### 1. Consumer Group 机制

Kafka 使用 **Consumer Group** 来管理消费者集群：

```yaml
# etc/hellogozero.yaml
Kafka:
  Group: hello-gozero-group  # 所有副本使用相同的 Group ID
```

- **同一个 Consumer Group 内**：每条消息只会被一个消费者处理
- **不同 Consumer Group**：每个 Group 独立消费，可以重复消费

### 2. 分区分配策略

```
Topic: hello-gozero-topic (假设有 3 个分区)

┌─────────────┬─────────────┬─────────────┐
│  Partition 0 │  Partition 1 │  Partition 2 │
└──────┬──────┴──────┬──────┴──────┬───────┘
       │             │             │
┌──────▼──────┬──────▼──────┬──────▼───────┐
│  Consumer 1 │  Consumer 2 │  Consumer 3  │
│  (Pod 1)    │  (Pod 2)    │  (Pod 3)     │
└─────────────┴─────────────┴──────────────┘
       同一个 Consumer Group
```

每个分区只会分配给组内的一个消费者。

### 3. Rebalance 机制

当消费者数量变化时（Pod 扩缩容），Kafka 会自动触发 **Rebalance**：

```
场景 1: 新增副本
Consumer 1: [Partition 0, 1, 2]  →  Consumer 1: [Partition 0, 1]
                                     Consumer 2: [Partition 2]

场景 2: 副本下线
Consumer 1: [Partition 0]  →  Consumer 2: [Partition 0, 1, 2]
Consumer 2: [Partition 1]
Consumer 3: [Partition 2]  (下线)
```

## 当前配置分析

### ✅ 正确的配置

```go
// infra/queue/kafka.go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers:        conf.Brokers,
    Topic:          conf.Topic,
    GroupID:        conf.Group,          // ✅ 使用 Consumer Group
    CommitInterval: time.Second,         // ✅ 自动提交 offset
    StartOffset:    kafka.LastOffset,    // ✅ 新消费者从最新消息开始
})
```

这个配置确保：

1. 多个副本自动协调分区分配
2. Offset 定期提交到 Kafka
3. 新启动的消费者不会处理历史消息

## 仍需注意的问题

### ⚠️ 极端情况下可能重复消费

虽然使用了 Consumer Group，但在以下情况仍可能重复消费：

#### 1. Rebalance 期间

```
时间线:
T1: Consumer 1 读取 Message A
T2: Consumer 1 处理 Message A
T3: Rebalance 触发 (新 Pod 加入)
T4: Consumer 1 尝试提交 offset - 失败 (已失去分区所有权)
T5: Consumer 2 接管分区，从上次提交的 offset 开始
T6: Consumer 2 重新读取 Message A ← 重复消费
```

#### 2. 处理成功但提交失败

```go
// kafka_consumer.go
if err := w.processMessage(ctx, message); err != nil {
    // 处理成功
}

// 提交 offset
if err := w.reader.CommitMessages(ctx, message); err != nil {
    // 提交失败! ← 下次重启会重新消费
    w.logger.Errorf("Failed to commit message: %v", err)
}
```

#### 3. 消费者崩溃

```
Consumer 1 正在处理消息 → 进程崩溃
↓
Consumer 2 接管分区 → 从上次提交的 offset 开始
↓
重新处理未提交的消息
```

### ✅ 解决方案：幂等性设计

**所有消息处理必须设计为幂等操作**，即：重复执行产生相同结果。

#### 方案一：数据库唯一键约束

```go
// 使用 user_id + event_type + timestamp 作为唯一键
func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event UserEvent) error {
    // 使用 INSERT IGNORE 或 ON DUPLICATE KEY UPDATE
    query := `
        INSERT IGNORE INTO user_events (user_id, event_type, timestamp, processed)
        VALUES (?, ?, ?, 1)
    `
    _, err := h.db.ExecContext(ctx, query, event.UserID, event.EventType, event.Timestamp)
    if err != nil {
        return err
    }
    
    // 如果插入成功 (RowsAffected = 1)，执行业务逻辑
    // 如果插入失败 (重复键)，说明已处理，直接返回
    return nil
}
```

#### 方案二：消息去重表

```sql
-- 创建消息去重表
CREATE TABLE message_dedup (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    message_key VARCHAR(255) NOT NULL,      -- Kafka message key
    partition_id INT NOT NULL,
    offset_id BIGINT NOT NULL,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_partition_offset (partition_id, offset_id),
    KEY idx_message_key (message_key)
);
```

```go
func (h *UserEventHandler) Handle(ctx context.Context, message kafka.Message) error {
    // 1. 先尝试插入去重记录
    query := `
        INSERT INTO message_dedup (message_key, partition_id, offset_id)
        VALUES (?, ?, ?)
    `
    _, err := h.db.ExecContext(ctx, query, 
        string(message.Key), 
        message.Partition, 
        message.Offset,
    )
    
    // 如果插入失败 (重复键)，说明已处理过
    if err != nil && isDuplicateKeyError(err) {
        h.logger.Infof("Message already processed, skipping")
        return nil
    }
    
    if err != nil {
        return fmt.Errorf("failed to insert dedup record: %w", err)
    }
    
    // 2. 处理业务逻辑
    return h.processBusinessLogic(ctx, message)
}
```

#### 方案三：分布式锁 + 状态检查

```go
func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event UserEvent) error {
    lockKey := fmt.Sprintf("user:register:%s", event.UserID)
    
    // 1. 获取分布式锁
    lock, err := h.redis.ObtainLock(ctx, lockKey, 30*time.Second)
    if err != nil {
        return err
    }
    defer lock.Release(ctx)
    
    // 2. 检查是否已处理
    processed, err := h.checkEventProcessed(ctx, event.UserID, event.Timestamp)
    if err != nil {
        return err
    }
    if processed {
        h.logger.Infof("Event already processed")
        return nil
    }
    
    // 3. 使用事务处理业务 + 标记已处理
    return h.processWithTransaction(ctx, event)
}
```

## 性能优化建议

### 1. 合理设置分区数

```bash
# Topic 分区数应该 >= 服务副本数
# 例如：3 个 Pod，至少需要 3 个分区

kafka-topics.sh --create \
  --topic hello-gozero-topic \
  --partitions 6 \              # 建议 2x 服务副本数
  --replication-factor 3 \
  --bootstrap-server kafka:9092
```

**原因：**

- 分区数 < 副本数：部分消费者空闲
- 分区数 = 副本数：完美平衡
- 分区数 > 副本数：更好的扩展性

### 2. 调整提交间隔

```go
// 当前配置
CommitInterval: time.Second,  // 每秒提交

// 低延迟场景（减少重复消费风险）
CommitInterval: 100 * time.Millisecond,

// 高吞吐场景（减少提交开销）
CommitInterval: 5 * time.Second,
```

### 3. 手动提交 vs 自动提交

```go
// 当前：自动提交 (处理完立即提交)
if err := w.reader.CommitMessages(ctx, message); err != nil {
    w.logger.Errorf("Failed to commit: %v", err)
}

// 优化：批量提交 (处理 N 条后提交)
var messages []kafka.Message
for i := 0; i < batchSize; i++ {
    msg, _ := reader.FetchMessage(ctx)
    processMessage(msg)
    messages = append(messages, msg)
}
reader.CommitMessages(ctx, messages...)
```

## 监控指标

建议监控以下指标以检测重复消费：

```go
// 1. 消费延迟 (Lag)
SELECT 
    partition_id,
    current_offset,
    log_end_offset,
    (log_end_offset - current_offset) AS lag
FROM kafka_consumer_offsets
WHERE group_id = 'hello-gozero-group';

// 2. 消息处理速率
messages_processed_total{status="success"}
messages_processed_total{status="duplicate"}

// 3. Rebalance 频率
consumer_rebalance_total{group="hello-gozero-group"}
```

## 测试验证

### 1. 正常消费测试

```bash
# 启动 3 个服务副本
docker-compose up --scale app=3

# 发送测试消息
python debug/user/register_user.py

# 检查日志：每条消息只被处理一次
docker-compose logs app | grep "Processing user event"
```

### 2. Rebalance 测试

```bash
# 启动 2 个副本
docker-compose up --scale app=2

# 发送消息（持续发送）
while true; do
    python debug/user/register_user.py
    sleep 0.5
done

# 在另一个终端扩容到 3 个副本
docker-compose up --scale app=3 --no-recreate

# 观察日志：应该看到 Rebalance 但没有重复处理
```

### 3. 崩溃恢复测试

```bash
# 启动服务
docker-compose up app

# 发送消息
python debug/user/register_user.py

# 在消息处理期间强制停止
docker-compose kill app

# 重新启动
docker-compose up app

# 检查：消息应该被重新处理 (幂等性保证安全)
```

## 总结

### ✅ 当前配置是正确的

您的系统配置了正确的 Consumer Group 机制，**不会在正常情况下重复消费**。

### ⚠️ 但需要实现幂等性

由于 Rebalance、网络故障等极端情况，仍可能重复消费，因此：

1. **必须实现业务幂等性**
2. **使用数据库唯一约束**
3. **考虑分布式锁**
4. **记录消息处理状态**

### 📊 推荐的完整方案

```go
func (h *UserEventHandler) Handle(ctx context.Context, message kafka.Message) error {
    // 1. 去重检查 (基于 offset)
    if isDuplicate := h.checkDuplicate(message); isDuplicate {
        return nil
    }
    
    // 2. 业务处理 (幂等设计)
    if err := h.processIdempotent(ctx, message); err != nil {
        return err
    }
    
    // 3. 标记已处理 (与业务在同一事务中)
    return h.markProcessed(message)
}
```

### 参考资料

- [Kafka Consumer Groups](https://kafka.apache.org/documentation/#intro_consumers)
- [Delivery Semantics](https://kafka.apache.org/documentation/#semantics)
- [Rebalance Protocol](https://kafka.apache.org/documentation/#consumerconfigs)
