package model

import "time"

func init() {
	register(&AudioSession{})
}

// AudioSession 音频会话表
type AudioSession struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	SessionID   string `gorm:"column:session_id;type:varchar(64);uniqueIndex;not null" json:"session_id"`
	DeviceID    string `gorm:"column:device_id;type:varchar(32);index:idx_device_time,priority:1;not null" json:"device_id"`
	StartTime   int64  `gorm:"column:start_time;type:bigint;index:idx_device_time,priority:2;not null" json:"start_time"`
	EndTime     *int64 `gorm:"column:end_time;type:bigint;index:idx_end_time" json:"end_time"`
	TotalChunks int    `gorm:"column:total_chunks;type:int;default:0;not null" json:"total_chunks"`
	Status      int8   `gorm:"column:status;type:tinyint;default:0;not null" json:"status"` // 0=进行中,1=已完成,2=已失败
	CreatedAt   int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (a *AudioSession) TableName() string { return "audio_sessions" }

// AudioChunk 音频块表
type AudioChunk struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	SessionID  string `gorm:"column:session_id;type:varchar(64);index:idx_session_chunk,priority:1;not null" json:"session_id"`
	ChunkIndex int    `gorm:"column:chunk_index;type:int;index:idx_session_chunk,priority:2;not null" json:"chunk_index"`
	DeviceID   string `gorm:"column:device_id;type:varchar(32);index:idx_device_time,priority:1;not null" json:"device_id"`
	StartTime  int64  `gorm:"column:start_time;type:bigint;index:idx_device_time,priority:2;not null" json:"start_time"`
	EndTime    int64  `gorm:"column:end_time;type:bigint;index:idx_time_range,priority:1;not null" json:"end_time"`
	Duration   int    `gorm:"column:duration;type:int;not null" json:"duration"` // 音频时长(毫秒)
	FileSize   int64  `gorm:"column:file_size;type:bigint;not null" json:"file_size"`
	Format     string `gorm:"column:format;type:varchar(16);not null" json:"format"`
	SampleRate int    `gorm:"column:sample_rate;type:int;not null" json:"sample_rate"`
	Channels   int8   `gorm:"column:channels;type:tinyint;not null" json:"channels"`
	MinIOPath  string `gorm:"column:minio_path;type:varchar(255);not null" json:"minio_path"`
	Checksum   string `gorm:"column:checksum;type:varchar(32);not null" json:"checksum"`
	Status     int8   `gorm:"column:status;type:tinyint;default:0;not null" json:"status"` // 0=上传中,1=已完成,2=已失败
	CreatedAt  int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
}

func (a *AudioChunk) TableName() string { return "audio_chunks" }

// TimeSync 时间同步表
type TimeSync struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	DeviceID   string `gorm:"column:device_id;type:varchar(32);uniqueIndex;not null" json:"device_id"`
	DeviceTime int64  `gorm:"column:device_time;type:bigint;not null" json:"device_time"`
	ServerTime int64  `gorm:"column:server_time;type:bigint;not null" json:"server_time"`
	Offset     int64  `gorm:"column:offset;type:bigint;not null" json:"offset"` // 设备时间 - 服务端时间
	LastSync   int64  `gorm:"column:last_sync;type:bigint;not null" json:"last_sync"`
	CreatedAt  int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (t *TimeSync) TableName() string { return "time_sync" }

// 会话状态常量
const (
	AudioSessionStatusActive    = 0 // 进行中
	AudioSessionStatusCompleted = 1 // 已完成
	AudioSessionStatusFailed    = 2 // 已失败
)

// 音频块状态常量
const (
	AudioChunkStatusUploading = 0 // 上传中
	AudioChunkStatusCompleted = 1 // 已完成
	AudioChunkStatusFailed    = 2 // 已失败
)

// 获取当前时间戳(毫秒)
func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// 验证时间戳是否在合理范围内
func ValidateTimestamp(timestamp int64, tolerance int64) bool {
	now := time.Now().UnixMilli()
	return timestamp >= now-tolerance && timestamp <= now+tolerance
}
