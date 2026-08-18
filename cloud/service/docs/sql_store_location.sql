-- 门店和点位管理系统的数据库表结构

-- 创建门店表
CREATE TABLE IF NOT EXISTS `stores` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '门店ID',
  `name` varchar(128) NOT NULL COMMENT '门店名称',
  `code` varchar(32) NOT NULL COMMENT '门店代码',
  `address` varchar(256) DEFAULT '' COMMENT '门店地址',
  `contact` varchar(64) DEFAULT '' COMMENT '联系人',
  `phone` varchar(20) DEFAULT '' COMMENT '联系电话',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1-正常，0-停用',
  `created_at` bigint NOT NULL COMMENT '创建时间(毫秒)',
  `updated_at` bigint NOT NULL COMMENT '更新时间(毫秒)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stores_code` (`code`),
  KEY `idx_stores_status` (`status`),
  KEY `idx_stores_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='门店表';

-- 创建点位表
CREATE TABLE IF NOT EXISTS `locations` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '点位ID',
  `store_id` bigint NOT NULL COMMENT '所属门店ID',
  `name` varchar(128) NOT NULL COMMENT '点位名称',
  `code` varchar(32) NOT NULL COMMENT '点位代码',
  `description` varchar(256) DEFAULT '' COMMENT '点位描述',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1-正常，0-停用',
  `created_at` bigint NOT NULL COMMENT '创建时间(毫秒)',
  `updated_at` bigint NOT NULL COMMENT '更新时间(毫秒)',
  PRIMARY KEY (`id`),
  KEY `idx_locations_store_id` (`store_id`),
  KEY `idx_locations_status` (`status`),
  KEY `idx_locations_created_at` (`created_at`),
  CONSTRAINT `fk_locations_store_id` FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点位表';

-- 为设备表添加门店和点位关联字段
ALTER TABLE `devices` 
ADD COLUMN `location_id` bigint DEFAULT 0 COMMENT '所属点位ID' AFTER `swap_used_bytes`,
ADD COLUMN `store_id` bigint DEFAULT 0 COMMENT '所属门店ID' AFTER `location_id`,
ADD INDEX `idx_devices_location_id` (`location_id`),
ADD INDEX `idx_devices_store_id` (`store_id`);

-- 添加外键约束（可选，根据实际需求决定）
-- ALTER TABLE `devices` 
-- ADD CONSTRAINT `fk_devices_location_id` FOREIGN KEY (`location_id`) REFERENCES `locations` (`id`) ON DELETE SET NULL,
-- ADD CONSTRAINT `fk_devices_store_id` FOREIGN KEY (`store_id`) REFERENCES `stores` (`id`) ON DELETE SET NULL;

-- 插入示例数据
INSERT INTO `stores` (`name`, `code`, `address`, `contact`, `phone`, `status`, `created_at`, `updated_at`) VALUES
('北京朝阳店', 'BJ_CY_001', '北京市朝阳区xxx街道xxx号', '张经理', '13800138000', 1, UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
('北京海淀店', 'BJ_HD_001', '北京市海淀区xxx街道xxx号', '李经理', '13800138001', 1, UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000);

INSERT INTO `locations` (`store_id`, `name`, `code`, `description`, `status`, `created_at`, `updated_at`) VALUES
(1, '一楼大厅', 'L1_HALL', '一楼主要营业区域', 1, UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
(1, '二楼办公区', 'L2_OFFICE', '二楼办公区域', 1, UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000),
(2, '一楼营业厅', 'L1_BUSINESS', '一楼营业区域', 1, UNIX_TIMESTAMP()*1000, UNIX_TIMESTAMP()*1000);

-- 查看表结构
DESCRIBE `stores`;
DESCRIBE `locations`;
DESCRIBE `devices`;

-- 查看示例数据
SELECT * FROM `stores`;
SELECT * FROM `locations`;
SELECT 
    d.id,
    d.name as device_name,
    d.mac_address,
    l.name as location_name,
    s.name as store_name
FROM `devices` d
LEFT JOIN `locations` l ON d.location_id = l.id
LEFT JOIN `stores` s ON d.store_id = s.id
LIMIT 10;
