package model

func init() {
	register(&AudioRecording{})
}

// AudioRecording 音频录音表
type AudioRecording struct {
	ID         string `gorm:"column:id;type:varchar(64);primaryKey;not null" json:"id"`
	SessionID  string `gorm:"column:session_id;type:varchar(64);uniqueIndex:idx_session_audio,priority:1;index:idx_session_id;not null" json:"session_id"`
	AudioID    string `gorm:"column:audio_id;type:varchar(64);uniqueIndex:idx_session_audio,priority:2;not null" json:"audio_id"`
	MacAddress string `gorm:"column:mac_address;type:varchar(64);index:idx_mac_address;not null" json:"mac_address"`
	FilePath   string `gorm:"column:file_path;type:varchar(255);not null" json:"file_path"`
	FileSize   int64  `gorm:"column:file_size;type:bigint;not null" json:"file_size"`
	UploadTime int64  `gorm:"column:upload_time;type:bigint;index:idx_upload_time;not null" json:"upload_time"`
	Status     int8   `gorm:"column:status;type:tinyint;default:1;not null" json:"status"` // 1=正常,0=已删除
	CreatedAt  int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (a *AudioRecording) TableName() string { return "audio_recordings" }

// 音频录音状态常量
const (
	AudioRecordingStatusNormal  = 1 // 正常
	AudioRecordingStatusDeleted = 0 // 已删除
)
