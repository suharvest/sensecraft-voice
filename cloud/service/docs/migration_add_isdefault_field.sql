-- 为 system_prompts 表添加 is_default 字段的迁移脚本
-- 执行时间：2024年

-- 1. 添加 is_default 字段
ALTER TABLE system_prompts ADD COLUMN is_default BOOLEAN DEFAULT FALSE;

-- 2. 为现有数据设置默认值
UPDATE system_prompts SET is_default = FALSE WHERE is_default IS NULL;

-- 3. 将现有的 name='default' 且 is_active=true 的记录设为默认
UPDATE system_prompts 
SET is_default = TRUE 
WHERE name = 'default' AND is_active = TRUE;

-- 4. 确保只有一个默认记录（如果存在多个 name='default' 的记录，只保留最新的一个）
-- 首先找出所有 name='default' 的记录
-- 然后只保留 ID 最大的那个设为默认，其他的设为非默认
UPDATE system_prompts 
SET is_default = FALSE 
WHERE name = 'default' 
  AND id NOT IN (
    SELECT max_id FROM (
      SELECT MAX(id) as max_id 
      FROM system_prompts 
      WHERE name = 'default' AND is_active = TRUE
    ) as temp
  );

-- 5. 添加索引以提高查询性能
CREATE INDEX idx_system_prompts_is_default ON system_prompts(is_default);

-- 6. 添加复合索引用于快速查找默认且激活的记录
CREATE INDEX idx_system_prompts_default_active ON system_prompts(is_default, is_active);

-- 验证迁移结果
SELECT 
  id, 
  name, 
  is_active, 
  is_default, 
  created_at 
FROM system_prompts 
WHERE is_default = TRUE 
ORDER BY id;
