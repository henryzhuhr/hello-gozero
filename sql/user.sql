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
  `phone` VARCHAR(20) DEFAULT '' COMMENT '手机号',
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
CREATE UNIQUE INDEX `uk_phone_deleted` ON `t_user` (`phone`, `deleted_at`);

CREATE INDEX `idx_status` ON `t_user` (`status`);
CREATE INDEX `idx_created_at` ON `t_user` (`created_at`);
CREATE INDEX `idx_deleted_at` ON `t_user` (`deleted_at`);


-- -- 应用层生成 UUIDv7 字符串，然后插入
-- INSERT INTO `t_user` (id, username, password)
-- VALUES (
--   UUID_TO_BIN('0373b8c0-d5b6-7abc-9def-123456789abc'),  -- 注意第13位是 '7'
--   'bob',
--   '...'
-- );


-- -- 查询时还原为标准 UUID 字符串
-- SELECT BIN_TO_UUID(id) AS id, username, email FROM `t_user`;

-- -- 根据 UUID 字符串查询
-- SELECT BIN_TO_UUID(id) AS id, username
-- FROM `t_user`
-- WHERE id = UUID_TO_BIN('f81d4fae-7dec-11d0-a765-00a0c91e6bf6');