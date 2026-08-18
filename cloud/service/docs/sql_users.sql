-- 用户管理系统的数据库表结构

-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码(MD5加密)',
  `created_at` bigint NOT NULL COMMENT '创建时间(毫秒)',
  `updated_at` bigint NOT NULL COMMENT '更新时间(毫秒)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  KEY `idx_users_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 插入示例数据
INSERT INTO `users` (`username`, `password`, `created_at`, `updated_at`) VALUES
('admin', MD5('admin123'), UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
('testuser', MD5('test123'), UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000);

-- 查看表结构
DESCRIBE `users`;

-- 查看示例数据
SELECT id, username, created_at, updated_at FROM `users`;
