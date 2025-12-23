-- 使用/切换到指定数据库
USE hello_gozero_db;

-- 删除表（如果存在）
DROP TABLE IF EXISTS `t_user`;

-- DDL 是 Data Definition Language（数据定义语言）的缩写，是 SQL 语言的一个子集，用于定义或修改数据库结构（schema），而不是操作表中的数据。
-- 仅当表不存在时创建（不会删除已有数据！）
CREATE TABLE `t_user` (
  `id` BINARY(16) NOT NULL PRIMARY KEY COMMENT '用户ID (UUID，二进制存储)',
  `username` VARCHAR(50) NOT NULL COMMENT '用户名',
  `password` VARCHAR(255) NOT NULL COMMENT '加密密码',
  `email` VARCHAR(100) DEFAULT '' COMMENT '邮箱',
  `phone_country_code` VARCHAR(6) NOT NULL COMMENT '手机号国际区号（例如：+86）',
  `phone_number` VARCHAR(20) NOT NULL COMMENT '手机号',
  `nickname` VARCHAR(50) DEFAULT '' COMMENT '昵称',
  `status` TINYINT DEFAULT 1 COMMENT '状态：0-禁用，1-正常',
  `last_login_time` DATETIME DEFAULT NULL COMMENT '最后登录时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间（软删除）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 添加索引（如果不存在）
CREATE UNIQUE INDEX `uk_username_deleted` ON `t_user` (`username`, `deleted_at`);
CREATE UNIQUE INDEX `uk_email_deleted` ON `t_user` (`email`, `deleted_at`);
CREATE UNIQUE INDEX `uk_phone_deleted` ON `t_user` (`phone_country_code`, `phone_number`, `deleted_at`);

CREATE INDEX `idx_status` ON `t_user` (`status`);
CREATE INDEX `idx_created_at` ON `t_user` (`created_at`);
CREATE INDEX `idx_deleted_at` ON `t_user` (`deleted_at`);


INSERT INTO `t_user` (
  `id`,
  `username`,
  `password`,
  `email`,
  `phone_country_code`,
  `phone_number`,
  `nickname`,
  `status`,
  `last_login_time`,
  `created_at`,
  `updated_at`,
  `deleted_at`
) VALUES (
  UNHEX(REPLACE(UUID(), '-', '')),          -- UUID 转为 BINARY(16)
  'admin',
  '$2a$10$default_hashed_password_for_test',-- 示例密码哈希（实际应为 bcrypt/scrypt 等）
  'admin@example.com',
  '+86',
  '13800138000',
  '系统管理员',
  1,
  NULL,
  NOW(),
  NOW(),
  NULL
);