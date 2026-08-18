package model

func init() {
	register(&SeeedAPILog{})
}

// SeeedAPILog Seeed API调用日志表，记录每次调用Seeed API的详细信息
type SeeedAPILog struct {
	Id         int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	QueryLogID int64  `gorm:"column:query_log_id;type:bigint;index:idx_seeed_logs_query;not null" json:"query_log_id"` // 关联的查询日志ID
	SessionID  string `gorm:"column:session_id;type:varchar(64);index:idx_seeed_logs_session;not null" json:"session_id"`

	// 请求信息
	RequestToken   string `gorm:"column:request_token;type:varchar(64);not null" json:"request_token"`       // 请求Token
	RequestContent string `gorm:"column:request_content;type:text;not null" json:"request_content"`          // 发送的内容（OpenAI回复）
	RequestURL     string `gorm:"column:request_url;type:varchar(255);not null" json:"request_url"`          // 请求URL

	// 响应信息
	ResponseCode    int    `gorm:"column:response_code;type:int;default:0;not null" json:"response_code"`       // Seeed API返回码
	ResponseMessage string `gorm:"column:response_message;type:text;default:null" json:"response_message"`      // Seeed API返回消息
	ResponseBody    string `gorm:"column:response_body;type:text;default:null" json:"response_body"`            // 完整响应体（JSON）

	// 调用状态
	RetryCount   int    `gorm:"column:retry_count;type:int;default:0;not null" json:"retry_count"`     // 重试次数（0表示首次调用）
	IsSuccess    bool   `gorm:"column:is_success;type:boolean;default:false;not null" json:"is_success"` // 是否成功
	ErrorMessage string `gorm:"column:error_message;type:text;default:null" json:"error_message"`       // 错误信息

	// 时间戳（毫秒）
	CreatedAt int64 `gorm:"column:created_at;type:bigint;not null" json:"created_at"` // 调用时间
}

func (s *SeeedAPILog) TableName() string { return "seeed_api_logs" }
