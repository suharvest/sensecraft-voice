-- 为chat_sessions表添加title字段
-- 添加会话标题字段，用于存储AI生成的会话标题

ALTER TABLE `chat_sessions` 
ADD COLUMN `title` varchar(255) NOT NULL DEFAULT '' COMMENT '会话标题' AFTER `user_id`;

-- 为title字段添加索引（可选，如果需要按标题搜索）
-- ALTER TABLE `chat_sessions` ADD INDEX `idx_title` (`title`);

-- 更新现有记录的title字段为默认值（如果需要的话）
-- UPDATE `chat_sessions` SET `title` = '' WHERE `title` IS NULL;
