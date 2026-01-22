-- 使用/切换到指定数据库
USE hello_gozero_db;

-- 删除表（如果存在）
-- DROP TABLE IF EXISTS `t_user_mfa`;

-- MFA 配置表（一对多 or 一对一，见下文）
-- 一个用户可以有多条 MFA 记录
-- 这种设计带来的核心能力：
-- ✅ 多方式备份	主方式（如手机）丢失时，可用邮箱或 TOTP 恢复账户
-- ✅ 多设备支持	可在手机、平板、电脑分别绑定 TOTP 或安全密钥
-- ✅ 灵活切换	登录时可选择任一已验证的方式（前端展示列表）
-- ✅ 渐进式增强安全	先开短信，再加 TOTP，不强制一步到位
CREATE TABLE `t_user_mfa` (
    `id`            BINARY(16) PRIMARY KEY,
    `user_id`       BINARY(16) NOT NULL REFERENCES `t_user`(id) ON DELETE CASCADE,
    
    -- MFA 类型
    `method`        VARCHAR(20) NOT NULL,
    
    -- 绑定目标地址
    -- - sms: 手机号（如 "+8613812345678"）
    -- - email: 邮箱（如 "a***@example.com"）
    -- - totp: 设备描述（如 "iPhone Google Authenticator"）
    -- - webauthn: 凭据 ID（Base64 编码）
    `target`        VARCHAR(255) NOT NULL,  -- 手机号 / 邮箱 / TOTP URI / Credential ID
    
    -- 是否启用（可选：有些系统允许绑定但不强制）
    -- 是否启用该 MFA 方式 用户关闭短信验证 → 设为 FALSE
    `enabled`       BOOLEAN DEFAULT true,
    
    -- 是否为主用方式（用于登录时默认选中）
    `is_primary`    BOOLEAN DEFAULT false,
    
    -- 敏感数据（如 TOTP 密钥）: 用层加密（如 AES），数据库只存密文
    `secret`        VARBINARY(255),  -- 加密存储！
    
    -- 元数据
    -- 首次验证成功时间 用于判断是否已完成绑定 / 用户输入正确验证码后更新此字段
    `verified_at`   DATETIME,  -- 首次验证成功时间

    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（软删除）'
);

-- 在已存在的表 t_user_mfa 上添加唯一索引
ALTER TABLE `t_user_mfa`
ADD UNIQUE KEY `uk_user_method` (`user_id`, `method`);