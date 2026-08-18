package model

func init() {
	register(&KeywordMatch{})
}

// KeywordMatch 关键词匹配记录
type KeywordMatch struct {
	ID          int64   `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	RecordingID int64   `gorm:"column:recording_id;index:idx_recording_id;not null" json:"recording_id"`
	MacAddress  string  `gorm:"column:mac_address;type:varchar(32);index:idx_mac_keyword,priority:1;not null" json:"mac_address"`
	KeywordID   int64   `gorm:"column:keyword_id;index:idx_mac_keyword,priority:2;not null" json:"keyword_id"`
	Keyword     string  `gorm:"column:keyword;type:varchar(50);not null" json:"keyword"`
	MatchedText string  `gorm:"column:matched_text;type:text;not null" json:"matched_text"`                    // 匹配到的具体文本
	MatchType   string  `gorm:"column:match_type;type:varchar(20);default:'exact';not null" json:"match_type"` // exact, synonym
	Confidence  float64 `gorm:"column:confidence;type:decimal(3,2);default:1.00;not null" json:"confidence"`   // 匹配置信度
	Position    int     `gorm:"column:position;not null" json:"position"`                                      // 在文本中的位置
	Length      int     `gorm:"column:length;not null" json:"length"`                                          // 匹配文本长度
	CreatedAt   int64   `gorm:"column:created_at;type:bigint;not null" json:"created_at"`

	// 录音相关字段（通过 JOIN 获取，不参与数据库操作）
	SessionID   string `gorm:"-" json:"session_id,omitempty"`
	AudioID     string `gorm:"-" json:"audio_id,omitempty"`
	SpeakerID   string `gorm:"-" json:"speaker_id,omitempty"`
	SpeakerName string `gorm:"-" json:"speaker_name,omitempty"`
	Text        string `gorm:"-" json:"text,omitempty"`
	DeviceTime  int64  `gorm:"-" json:"device_time,omitempty"`
	Status      int8   `gorm:"-" json:"status,omitempty"`
}

func (km *KeywordMatch) TableName() string { return "keyword_matches" }

// KeywordMatchQuery 用于查询的关键词匹配记录（包含录音信息）
type KeywordMatchQuery struct {
	ID          int64   `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	RecordingID int64   `gorm:"column:recording_id;index:idx_recording_id;not null" json:"recording_id"`
	MacAddress  string  `gorm:"column:mac_address;type:varchar(32);index:idx_mac_keyword,priority:1;not null" json:"mac_address"`
	KeywordID   int64   `gorm:"column:keyword_id;index:idx_mac_keyword,priority:2;not null" json:"keyword_id"`
	Keyword     string  `gorm:"column:keyword;type:varchar(50);not null" json:"keyword"`
	MatchedText string  `gorm:"column:matched_text;type:text;not null" json:"matched_text"`                    // 匹配到的具体文本
	MatchType   string  `gorm:"column:match_type;type:varchar(20);default:'exact';not null" json:"match_type"` // exact, synonym
	Confidence  float64 `gorm:"column:confidence;type:decimal(3,2);default:1.00;not null" json:"confidence"`   // 匹配置信度
	Position    int     `gorm:"column:position;not null" json:"position"`                                      // 在文本中的位置
	Length      int     `gorm:"column:length;not null" json:"length"`                                          // 匹配文本长度
	CreatedAt   int64   `gorm:"column:created_at;type:bigint;not null" json:"created_at"`

	// 录音相关字段（通过 JOIN 获取）
	SessionID   string `gorm:"column:session_id;type:varchar(64)" json:"session_id,omitempty"`
	AudioID     string `gorm:"column:audio_id;type:varchar(64)" json:"audio_id,omitempty"`
	SpeakerID   string `gorm:"column:speaker_id;type:varchar(64)" json:"speaker_id,omitempty"`
	SpeakerName string `gorm:"column:speaker_name;type:varchar(128)" json:"speaker_name,omitempty"`
	Text        string `gorm:"column:text;type:text" json:"text,omitempty"`
	DeviceTime  int64  `gorm:"column:device_time;type:bigint" json:"device_time,omitempty"`
	Status      int8   `gorm:"column:status;type:tinyint" json:"status,omitempty"`
}

func (kmq *KeywordMatchQuery) TableName() string { return "keyword_matches" }

// 匹配类型常量
const (
	MatchTypeExact   = "exact"   // 精确匹配
	MatchTypeSynonym = "synonym" // 近义词匹配
)
