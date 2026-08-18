-- Chat 相关表结构
-- 聊天会话表
CREATE TABLE IF NOT EXISTS `chat_sessions` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `session_id` varchar(64) NOT NULL COMMENT '会话ID',
    `user_id` varchar(64) NOT NULL COMMENT '用户ID',
    `status` varchar(16) NOT NULL DEFAULT 'active' COMMENT '会话状态：active-活跃, closed-关闭',
    `created_at` bigint NOT NULL COMMENT '创建时间（毫秒时间戳）',
    `updated_at` bigint NOT NULL COMMENT '更新时间（毫秒时间戳）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_session_id` (`session_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天会话表';

-- 聊天消息表
CREATE TABLE IF NOT EXISTS `chat_messages` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `session_id` varchar(64) NOT NULL COMMENT '会话ID',
    `message_id` varchar(64) NOT NULL COMMENT '消息ID',
    `event` varchar(32) NOT NULL COMMENT '事件类型',
    `content` text COMMENT '消息内容',
    `data` longtext COMMENT '额外数据（JSON格式）',
    `created_at` bigint NOT NULL COMMENT '创建时间（毫秒时间戳）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_message_id` (`message_id`),
    KEY `idx_session_id` (`session_id`),
    KEY `idx_event` (`event`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';

-- 聊天统计表
CREATE TABLE IF NOT EXISTS `chat_stats` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `session_id` varchar(64) NOT NULL COMMENT '会话ID',
    `total_tokens` bigint NOT NULL DEFAULT 0 COMMENT '总token数量',
    `total_price` decimal(10,6) NOT NULL DEFAULT 0.000000 COMMENT '总价格',
    `currency` varchar(8) NOT NULL DEFAULT 'USD' COMMENT '货币类型',
    `latency` decimal(10,6) NOT NULL DEFAULT 0.000000 COMMENT '延迟时间（秒）',
    `elapsed_time` decimal(10,6) NOT NULL DEFAULT 0.000000 COMMENT '总耗时（秒）',
    `total_steps` int NOT NULL DEFAULT 0 COMMENT '总步骤数',
    `conversation_id` varchar(64) NOT NULL COMMENT '对话ID',
    `created_at` bigint NOT NULL COMMENT '创建时间（毫秒时间戳）',
    PRIMARY KEY (`id`),
    KEY `idx_session_id` (`session_id`),
    KEY `idx_conversation_id` (`conversation_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天统计表';

-- 插入示例数据
INSERT INTO `chat_sessions` (`session_id`, `user_id`, `status`, `created_at`, `updated_at`) VALUES
('session_1703123456789', 'user_001', 'active', 1703123456789, 1703123456789),
('session_1703123456790', 'user_002', 'active', 1703123456790, 1703123456790);

INSERT INTO `chat_messages` (`session_id`, `message_id`, `event`, `content`, `data`, `created_at`) VALUES
('session_1703123456789', 'msg_1703123456789', 'message', 'Hello, how are you?', '{"answer": "Hello, how are you?", "timestamp": 1703123456789}', 1703123456789),
('session_1703123456789', 'msg_1703123456790', 'message_end', 'Message completed', '{"status": "completed", "total_tokens": 15}', 1703123456790);

INSERT INTO `chat_stats` (`session_id`, `total_tokens`, `total_price`, `currency`, `latency`, `elapsed_time`, `total_steps`, `conversation_id`, `created_at`) VALUES
('session_1703123456789', 15, 0.000015, 'USD', 0.5, 1.2, 1, 'conv_1703123456789', 1703123456789);
