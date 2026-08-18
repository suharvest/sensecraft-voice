package model

func init() {
	register(&Device{})
}

type Device struct {
	Id               int64   `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	MacAddress       string  `gorm:"column:mac_address;type:varchar(32);uniqueIndex:uk_devices_mac;not null" json:"mac_address"`
	Name             string  `gorm:"column:name;type:varchar(128);default:'';not null" json:"name"`
	IpAddress        string  `gorm:"column:ip_address;type:varchar(64)" json:"ip_address"`
	Version          string  `gorm:"column:version;type:varchar(64);default:'';not null" json:"version"`
	CpuUsagePercent  float64 `gorm:"column:cpu_usage_percent;type:decimal(5,2);default:0;not null" json:"cpu_usage_percent"`
	MemoryUsedBytes  int64   `gorm:"column:memory_used_bytes;type:bigint;default:0;not null" json:"memory_used_bytes"`
	MemoryTotalBytes int64   `gorm:"column:memory_total_bytes;type:bigint;default:0;not null" json:"memory_total_bytes"`
	DiskUsedBytes    int64   `gorm:"column:disk_used_bytes;type:bigint;default:0;not null" json:"disk_used_bytes"`
	DiskTotalBytes   int64   `gorm:"column:disk_total_bytes;type:bigint;default:0;not null" json:"disk_total_bytes"`
	SwapUsedBytes    int64   `gorm:"column:swap_used_bytes;type:bigint;default:0;not null" json:"swap_used_bytes"`
	SwapTotalBytes   int64   `gorm:"column:swap_total_bytes;type:bigint;default:0;not null" json:"swap_total_bytes"`
	LocationId       int64   `gorm:"column:location_id;index:idx_devices_location_id;default:0" json:"location_id"`
	StoreId          int64   `gorm:"column:store_id;index:idx_devices_store_id;default:0" json:"store_id"`
	// 设备认证与在线状态
	TokenHash  string `gorm:"column:token_hash;type:varchar(64);index:idx_devices_token_hash;default:'';not null" json:"-"`
	LastSeenAt int64  `gorm:"column:last_seen_at;type:bigint;default:0;not null" json:"last_seen_at"`

	// ASR 配置下发
	AsrServerId        int64  `gorm:"column:asr_server_id;index:idx_devices_asr_server_id;default:0;not null" json:"asr_server_id"`
	AsrConfigVersion   int    `gorm:"column:asr_config_version;type:int;default:0;not null" json:"asr_config_version"`
	AsrConfigAppliedAt int64  `gorm:"column:asr_config_applied_at;type:bigint;default:0;not null" json:"asr_config_applied_at"`
	AsrConfigError     string `gorm:"column:asr_config_error;type:varchar(512);default:'';not null" json:"asr_config_error"`

	Online              bool   `gorm:"-" json:"online"`                             // 不存储到数据库，查询时按 last_seen_at 计算
	LocationName        string `gorm:"->;-:migration" json:"location_name"`         // 不存储到数据库，仅用于API响应
	StoreName           string `gorm:"->;-:migration" json:"store_name"`            // 不存储到数据库，仅用于API响应
	LatestRecordingTime *int64 `gorm:"->;-:migration" json:"latest_recording_time"` // 最新recording的创建时间
	CreatedAt           int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt           int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (d *Device) TableName() string { return "devices" }
