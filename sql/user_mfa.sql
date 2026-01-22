-- 使用/切换到指定数据库
USE hello_gozero_db;

-- 删除表（如果存在）
-- DROP TABLE IF EXISTS `t_user_mfa`;

-- MFA 配置表（一对多 or 一对一，见下文）
CREATE TABLE `t_user_mfa` (
    id            BINARY(16) PRIMARY KEY,
    user_id       BINARY(16) NOT NULL REFERENCES `t_user`(id) ON DELETE CASCADE,
    
    -- MFA 类型
    method        VARCHAR(20) NOT CHECK (method IN ('sms', 'email', 'totp', 'webauthn')),
    
    -- 绑定标识（根据 method 决定含义）
    target        VARCHAR(255) NOT NULL,  -- 手机号 / 邮箱 / TOTP URI / Credential ID
    
    -- 是否启用（可选：有些系统允许绑定但不强制）
    is_enabled    BOOLEAN DEFAULT true,
    
    -- 是否为主用方式（用于登录时默认选中）
    is_primary    BOOLEAN DEFAULT false,
    
    -- 敏感数据（如 TOTP 密钥）
    secret        BYTEA,  -- 加密存储！
    
    -- 元数据
    verified_at   TIMESTAMP,  -- 首次验证成功时间
    created_at    TIMESTAMP DEFAULT NOW(),
    
    -- 唯一约束：一个用户对每种 method 只能有一个 enabled 记录
    UNIQUE (user_id, method)
);