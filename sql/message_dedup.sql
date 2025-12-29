-- 消息去重表
-- 用于防止 Kafka 消息在 Rebalance 等极端情况下的重复消费

CREATE TABLE IF NOT EXISTS message_dedup (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增主键',
    
    -- Kafka 消息标识
    topic VARCHAR(100) NOT NULL COMMENT 'Kafka Topic 名称',
    partition_id INT NOT NULL COMMENT '分区 ID',
    offset_id BIGINT NOT NULL COMMENT '消息 Offset',
    message_key VARCHAR(255) DEFAULT NULL COMMENT '消息 Key (可选)',
    
    -- 处理信息
    event_type VARCHAR(50) DEFAULT NULL COMMENT '事件类型 (如: user_registered)',
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '处理时间',
    consumer_id VARCHAR(100) DEFAULT NULL COMMENT '消费者 ID (Pod 名称)',
    
    -- 索引
    UNIQUE KEY uk_partition_offset (topic, partition_id, offset_id) COMMENT '防止重复消费的唯一索引',
    KEY idx_message_key (message_key) COMMENT '按消息 Key 查询',
    KEY idx_processed_at (processed_at) COMMENT '按处理时间查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Kafka 消息去重表';

-- 定期清理历史数据 (保留最近 7 天)
-- 建议创建定时任务执行
-- DELETE FROM message_dedup WHERE processed_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
