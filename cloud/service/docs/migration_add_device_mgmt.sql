-- 设备纳管迁移脚本：设备认证 + 在线状态 + ASR 配置下发
-- 关联设计文档：docs/DEVICE_MGMT_SENSEVOICE_ASR_DESIGN.md §2
-- 说明：AutoMigrate 只建新表，对已存在的表整表跳过（pkg/db/migrator.go），
--       所以 devices 加列必须手工执行本脚本。asr_servers 是新表，由 AutoMigrate 自动创建。
-- 执行：mysql -h<host> -P<port> -u<user> -p<pass> <dbname> < docs/migration_add_device_mgmt.sql
-- 幂等性：MySQL 8 不支持 ADD COLUMN IF NOT EXISTS，重复执行会报 1060 Duplicate column，可忽略。

-- 1. 设备认证凭证（不透明 token 的 sha256 hex，明文只在首次注册响应里返回一次）
ALTER TABLE devices ADD COLUMN token_hash VARCHAR(64) NOT NULL DEFAULT '';
CREATE INDEX idx_devices_token_hash ON devices (token_hash);

-- 2. 心跳时间（毫秒时间戳，与表内其他时间字段一致）
--    在线状态不落列，查询时按 last_seen_at > now - 3×心跳间隔 计算
ALTER TABLE devices ADD COLUMN last_seen_at BIGINT NOT NULL DEFAULT 0;

-- 3. ASR 服务器分配与配置下发状态
ALTER TABLE devices ADD COLUMN asr_server_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE devices ADD COLUMN asr_config_version INT NOT NULL DEFAULT 0;
ALTER TABLE devices ADD COLUMN asr_config_applied_at BIGINT NOT NULL DEFAULT 0;
ALTER TABLE devices ADD COLUMN asr_config_error VARCHAR(512) NOT NULL DEFAULT '';
CREATE INDEX idx_devices_asr_server_id ON devices (asr_server_id);

-- 4. asr_servers 表（AutoMigrate 会自动建，此处仅作为无 AutoMigrate 环境的等价 DDL）
CREATE TABLE IF NOT EXISTS asr_servers (
  id             BIGINT       NOT NULL AUTO_INCREMENT,
  name           VARCHAR(128) NOT NULL DEFAULT '',
  base_url       VARCHAR(255) NOT NULL,
  platform       VARCHAR(32)  NOT NULL DEFAULT '',
  api_key_cipher VARCHAR(512) NOT NULL DEFAULT '',
  location_id    BIGINT       NOT NULL DEFAULT 0,
  status         VARCHAR(16)  NOT NULL DEFAULT 'unknown', -- unknown | up | busy | down
                                                          -- busy = /readyz 503 且原因仅 sessions_full（解码中，健康）
  last_probe_at  BIGINT       NOT NULL DEFAULT 0,
  fail_count     INT          NOT NULL DEFAULT 0,
  last_error     VARCHAR(512) NOT NULL DEFAULT '',
  backend        VARCHAR(64)  NOT NULL DEFAULT '',
  capabilities   VARCHAR(255) NOT NULL DEFAULT '',
  sample_rate    INT          NOT NULL DEFAULT 0,
  created_at     BIGINT       NOT NULL,
  updated_at     BIGINT       NOT NULL,
  PRIMARY KEY (id),
  KEY idx_asr_servers_location_id (location_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
