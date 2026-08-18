-- 数据库：sensecraftVoice（如需其他库名请在连接时指定或先执行 CREATE DATABASE）

-- 设备表 devices
CREATE TABLE IF NOT EXISTS `devices` (
  `id`                bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `mac_address`       varchar(32) COLLATE utf8mb4_bin NOT NULL COMMENT '设备 MAC 地址（小写、冒号分隔）',
  `name`              varchar(128) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '设备名称',
  `version`           varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '设备/固件版本',
  `cpu_usage_percent` decimal(5,2) NOT NULL DEFAULT 0 COMMENT 'CPU 占用率百分比 0-100.00',
  `memory_used_bytes` bigint NOT NULL DEFAULT 0 COMMENT '内存使用字节数',
  `disk_used_bytes`   bigint NOT NULL DEFAULT 0 COMMENT '磁盘使用字节数',
  `swap_used_bytes`   bigint NOT NULL DEFAULT 0 COMMENT 'Swap 使用字节数',
  `created_at`        bigint NOT NULL COMMENT '创建时间 ms',
  `updated_at`        bigint NOT NULL COMMENT '更新时间 ms',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_devices_mac` (`mac_address`),
  KEY `idx_devices_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='设备表';

-- 录音记录表 recordings
CREATE TABLE IF NOT EXISTS `recordings` (
  `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `mac_address`  varchar(32) COLLATE utf8mb4_bin NOT NULL COMMENT '来源设备 MAC',
  `speaker_id`   varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '说话人 ID',
  `speaker_name` varchar(128) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '说话人名称',
  `text`         text COLLATE utf8mb4_bin NOT NULL COMMENT '识别文本',
  `status`       tinyint NOT NULL DEFAULT 0 COMMENT '0=中间结果,1=最终结果',
  `created_at`   bigint NOT NULL COMMENT '服务端入库时间 ms',
  `device_time`  bigint NOT NULL COMMENT '设备采集时间 ms',
  PRIMARY KEY (`id`),
  KEY `idx_recordings_mac_created` (`mac_address`, `created_at`),
  KEY `idx_recordings_device_time` (`device_time`),
  KEY `idx_recordings_status` (`status`),
  CONSTRAINT `fk_recordings_device_mac`
    FOREIGN KEY (`mac_address`) REFERENCES `devices` (`mac_address`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='录音记录表';

-- 可选：扩展列（若后续需要，可执行 ALTER TABLE 增加）
-- ALTER TABLE `recordings` ADD COLUMN `confidence` decimal(6,5) NULL COMMENT '模型置信度';
-- ALTER TABLE `recordings` ADD COLUMN `identified` tinyint(1) NULL COMMENT '是否识别到说话人';
-- ALTER TABLE `recordings` ADD COLUMN `type` varchar(16) COLLATE utf8mb4_bin NULL COMMENT '片段类型 partial/final';
-- ALTER TABLE `recordings` ADD COLUMN `text_length` int NULL COMMENT '文本长度';
-- ALTER TABLE `recordings` ADD COLUMN `word_count` int NULL COMMENT '分词数量';


