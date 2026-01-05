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



-- ============================================================
-- 唯一性约束索引（仅对活跃用户 deleted_at IS NULL）
-- ============================================================
-- 之前的索引存在问题：
-- MySQL 将 NULL 视为不同的值，所以当 deleted_at 为 NULL 时（活跃用户），多个相同 username 可以共存！这导致并发场景下可能创建多个相同用户名的记录，违反唯一性约束的初衷
-- CREATE UNIQUE INDEX `uk_username_deleted` ON `t_user` (`username`, `deleted_at`);
-- CREATE UNIQUE INDEX `uk_email_deleted` ON `t_user` (`email`, `deleted_at`);
-- CREATE UNIQUE INDEX `uk_phone_deleted` ON `t_user` (`phone_country_code`, `phone_number`, `deleted_at`);

-- MySQL 5.7+ 支持函数索引的替代方案：
-- 1. 创建虚拟列 is_active，当 deleted_at IS NULL 时为固定值（如 1）
-- 2. 对虚拟列和业务字段创建唯一索引
-- 3. 这样只有活跃用户会被唯一性约束，已删除用户不受影响

-- 方案一：使用 MySQL 8.0+ 的函数索引（推荐）
-- CREATE UNIQUE INDEX `uk_username_active` ON `t_user` (`username`) WHERE `deleted_at` IS NULL;
-- CREATE UNIQUE INDEX `uk_email_active` ON `t_user` (`email`) WHERE `deleted_at` IS NULL;
-- CREATE UNIQUE INDEX `uk_phone_active` ON `t_user` (`phone_country_code`, `phone_number`) WHERE `deleted_at` IS NULL;

-- 方案二：兼容 MySQL 5.7 的虚拟列方案（当前使用）
-- 注意：deleted_at 为 NULL 时，IFNULL(deleted_at, 1) 返回固定值 1
--       deleted_at 不为 NULL 时，返回 deleted_at 本身（不同的时间戳）
--       这样活跃用户（deleted_at=NULL）会被约束为唯一，已删除用户不受影响
-- CREATE UNIQUE INDEX `uk_username_active` ON `t_user` (`username`, (IFNULL(`deleted_at`, 1)));
-- CREATE UNIQUE INDEX `uk_email_active` ON `t_user` (`email`, (IFNULL(`deleted_at`, 1)));
-- CREATE UNIQUE INDEX `uk_phone_active` ON `t_user` (`phone_country_code`, `phone_number`, (IFNULL(`deleted_at`, 1)));

-- ✅ 正确写法：“索引表达式”。
-- 类比理解：“只监控活着的用户名，死了的随便重名”。符合软删除业务中最常见且合理的唯一性需求。
CREATE UNIQUE INDEX uk_username_active ON t_user (
  (CASE WHEN deleted_at IS NULL THEN username ELSE NULL END)
);

CREATE UNIQUE INDEX uk_email_active ON t_user (
  (CASE WHEN deleted_at IS NULL THEN email ELSE NULL END)
);

CREATE UNIQUE INDEX uk_phone_active ON t_user (
  (CASE WHEN deleted_at IS NULL THEN phone_country_code ELSE NULL END),
  (CASE WHEN deleted_at IS NULL THEN phone_number ELSE NULL END)
);

-- ============================================================
-- 普通索引（提升查询性能）
-- ============================================================
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