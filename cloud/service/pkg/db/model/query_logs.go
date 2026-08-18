package model

func init() {
	register(&QueryLog{})
}

// QueryLog 查询日志表，记录 /api/recordings/query 接口的请求和处理结果
type QueryLog struct {
	Id             int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	DeviceID       string `gorm:"column:device_id;type:varchar(64);index:idx_query_logs_device;not null" json:"device_id"`
	StartTimestamp int64  `gorm:"column:start_timestamp;type:bigint;not null" json:"start_timestamp"`
	EndTimestamp   int64  `gorm:"column:end_timestamp;type:bigint;not null" json:"end_timestamp"`
	SessionID      string `gorm:"column:session_id;type:varchar(64);index:idx_query_logs_session;not null" json:"session_id"`

	// 处理结果相关字段
	OpenAIAnswer    string `gorm:"column:openai_answer;type:text;default:null" json:"openai_answer"`         // OpenAI返回的总结
	SeeedAPIStatus  int8   `gorm:"column:seeed_api_status;type:tinyint;default:0;not null" json:"seeed_api_status"` // Seeed API调用状态：0-未处理，1-成功，2-失败
	ProcessingError string `gorm:"column:processing_error;type:text;default:null" json:"processing_error"`   // 处理错误信息（如果有）

	// 时间戳（毫秒）
	CreatedAt int64 `gorm:"column:created_at;type:bigint;not null" json:"created_at"` // 请求创建时间
	UpdatedAt int64 `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"` // 最后更新时间
}

func (q *QueryLog) TableName() string { return "query_logs" }
