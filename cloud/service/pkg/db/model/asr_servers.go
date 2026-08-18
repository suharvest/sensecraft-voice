package model

func init() {
	register(&AsrServer{})
}

// ASR 服务器状态
const (
	AsrServerStatusUnknown = "unknown"
	AsrServerStatusUp      = "up"
	// AsrServerStatusBusy /readyz 503 且原因仅为 sessions_full：解码占满限流器，服务器健康
	AsrServerStatusBusy = "busy"
	AsrServerStatusDown = "down"
)

// AsrServer 独立部署的 ASR 算力服务器（OVS / SenseVoice）
type AsrServer struct {
	Id       int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	Name     string `gorm:"column:name;type:varchar(128);default:'';not null" json:"name"`
	BaseUrl  string `gorm:"column:base_url;type:varchar(255);not null" json:"base_url"`
	Platform string `gorm:"column:platform;type:varchar(32);default:'';not null" json:"platform"` // rk3576 | rk3588 | jetson

	// ApiKeyCipher 该服务器的 API key，AES-GCM 密文落库，绝不出现在 API 响应里
	ApiKeyCipher string `gorm:"column:api_key_cipher;type:varchar(512);default:'';not null" json:"-"`

	LocationId int64 `gorm:"column:location_id;index:idx_asr_servers_location_id;default:0;not null" json:"location_id"`

	// Status unknown | up | busy | down
	Status       string `gorm:"column:status;type:varchar(16);default:'unknown';not null" json:"status"`
	LastProbeAt  int64  `gorm:"column:last_probe_at;type:bigint;default:0;not null" json:"last_probe_at"`
	FailCount    int    `gorm:"column:fail_count;type:int;default:0;not null" json:"fail_count"`
	LastError    string `gorm:"column:last_error;type:varchar(512);default:'';not null" json:"last_error"`
	Backend      string `gorm:"column:backend;type:varchar(64);default:'';not null" json:"backend"`
	Capabilities string `gorm:"column:capabilities;type:varchar(255);default:'';not null" json:"capabilities"` // 逗号分隔
	SampleRate   int    `gorm:"column:sample_rate;type:int;default:0;not null" json:"sample_rate"`

	CreatedAt int64 `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt int64 `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`

	// HasApiKey 不落库，仅告知前端是否已配置 api_key
	HasApiKey bool `gorm:"-" json:"has_api_key"`
	// DeviceCount 不落库，列表接口回填已分配设备数
	DeviceCount int64 `gorm:"-" json:"device_count"`
}

func (a *AsrServer) TableName() string { return "asr_servers" }

// TableOptions 覆盖 migrator 的默认建表选项。
// migrator 默认 CHARSET=utf8，name 字段含中文（4 字节 emoji / 部分生僻字）会踩 MySQL Error 1366。
func (a *AsrServer) TableOptions() string {
	return "AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci"
}
