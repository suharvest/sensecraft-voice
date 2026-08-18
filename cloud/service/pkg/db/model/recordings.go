package model

func init() {
	register(&Recording{})
}

type Recording struct {
	Id          int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	SessionID   string `gorm:"column:session_id;type:varchar(64);index:idx_recordings_session_created,priority:1;not null" json:"session_id"`
	AudioID     string `gorm:"column:audio_id;type:varchar(64);index:idx_recordings_audio_created,priority:2;not null" json:"audio_id"`
	MacAddress  string `gorm:"column:mac_address;type:varchar(32);index:idx_recordings_mac_created,priority:1;not null" json:"mac_address"`
	SpeakerId   string `gorm:"column:speaker_id;type:varchar(64);default:'';not null" json:"speaker_id"`
	SpeakerName string `gorm:"column:speaker_name;type:varchar(128);default:'';not null" json:"speaker_name"`
	Text        string `gorm:"column:text;type:text;not null" json:"text"`
	Status      int8   `gorm:"column:status;type:tinyint;default:0;not null" json:"status"`

	// 与业务要求的毫秒时间戳对应，这里保留为 bigint
	CreatedAtMs int64 `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	DeviceTime  int64 `gorm:"column:device_time;type:bigint;not null" json:"device_time"`

	// 播放链接（不存储到数据库，仅用于API响应）
	PlayURL string `gorm:"-" json:"play_url"`
}

func (r *Recording) TableName() string { return "recordings" }
