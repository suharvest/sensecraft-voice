-- 修复数据库字符集问题
-- 确保数据库和表都使用正确的UTF-8字符集

-- 1. 修改数据库字符集
ALTER DATABASE sensecraft_voice CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 2. 修改chat_messages表的字符集
ALTER TABLE chat_messages CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 3. 修改chat_sessions表的字符集
ALTER TABLE chat_sessions CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 4. 修改chat_stats表的字符集
ALTER TABLE chat_stats CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 5. 检查表字符集
SELECT 
    TABLE_NAME,
    TABLE_COLLATION,
    CHARACTER_SET_NAME
FROM information_schema.TABLES t
JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY ccs
    ON t.TABLE_COLLATION = ccs.COLLATION_NAME
WHERE t.TABLE_SCHEMA = 'sensecraft_voice'
    AND t.TABLE_NAME IN ('chat_messages', 'chat_sessions', 'chat_stats');

-- 6. 检查列字符集
SELECT 
    TABLE_NAME,
    COLUMN_NAME,
    CHARACTER_SET_NAME,
    COLLATION_NAME
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'sensecraft_voice'
    AND TABLE_NAME IN ('chat_messages', 'chat_sessions', 'chat_stats')
    AND CHARACTER_SET_NAME IS NOT NULL;
