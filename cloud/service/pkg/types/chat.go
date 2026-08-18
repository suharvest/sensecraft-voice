package types

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query"`
	ResponseMode   string                 `json:"response_mode"`
	ConversationID string                 `json:"conversation_id"`
	User           string                 `json:"user"`
	Files          []ChatFile             `json:"files,omitempty"`
}

// ChatFile 聊天文件结构
type ChatFile struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	URL            string `json:"url"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Raw   string                 `json:"raw,omitempty"`
}

// ChatConfig 聊天配置
type ChatConfig struct {
	BaseURL     string `json:"base_url"`
	APIKey      string `json:"api_key"`
	Timeout     int    `json:"timeout"` // 秒
	EnableDebug bool   `json:"enable_debug"`
}

// ChatStats 聊天统计
type ChatStats struct {
	ID             int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID      string  `json:"session_id" gorm:"index"`
	TotalTokens    int64   `json:"total_tokens"`
	TotalPrice     float64 `json:"total_price"`
	Currency       string  `json:"currency"`
	Latency        float64 `json:"latency"`
	ElapsedTime    float64 `json:"elapsed_time"`
	TotalSteps     int     `json:"total_steps"`
	ConversationID string  `json:"conversation_id"`
	CreatedAt      int64   `json:"created_at" gorm:"index"`
}

// ChatSession 聊天会话
type ChatSession struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID string `json:"session_id" gorm:"uniqueIndex;not null"`
	UserID    string `json:"user_id" gorm:"index"`
	Title     string `json:"title" gorm:"size:255;default:''"` // 会话标题
	Status    string `json:"status"`                           // active, closed
	CreatedAt int64  `json:"created_at" gorm:"index"`
	UpdatedAt int64  `json:"updated_at" gorm:"index"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID string `json:"session_id" gorm:"index"`
	MessageID string `json:"message_id" gorm:"uniqueIndex"`
	Event     string `json:"event"`
	Content   string `json:"content"`
	Data      string `json:"data"` // JSON格式的额外数据
	CreatedAt int64  `json:"created_at" gorm:"index"`
}
