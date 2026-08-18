package model

// ChatSession 聊天会话模型
type ChatSession struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	SessionID string `gorm:"uniqueIndex;size:64"` // 允许为空，因为可能从API获取
	UserID    string `gorm:"index;size:64"`
	Title     string `gorm:"size:255;default:''"` // 会话标题
	Status    string `gorm:"size:16;default:active"`
	CreatedAt int64  `gorm:"index"`
	UpdatedAt int64  `gorm:"index"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	SessionID string `gorm:"index;size:64"`
	MessageID string `gorm:"uniqueIndex;size:64"`
	Event     string `gorm:"size:32;check:event IN ('user', 'assistant', 'system', 'metadata')"` // 角色类型约束
	Content   string `gorm:"type:text;charset:utf8mb4"`
	Data      string `gorm:"type:longtext;charset:utf8mb4"`

	// 新增字段
	MessageType  string `gorm:"size:16;default:chat"` // 消息类型：chat, metadata, system
	TokenCount   int64  `gorm:"default:0"`            // token 数量
	ModelName    string `gorm:"size:64"`              // 模型名称
	ResponseTime int64  `gorm:"default:0"`            // 响应时间(毫秒)
	Quality      string `gorm:"size:16"`              // 质量评分

	CreatedAt int64 `gorm:"index"`
}

// TableName 指定表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// ChatStats 聊天统计模型
type ChatStats struct {
	ID             int64   `gorm:"primaryKey;autoIncrement"`
	SessionID      string  `gorm:"index;size:64"`
	TotalTokens    int64   `gorm:"default:0"`
	TotalPrice     float64 `gorm:"type:decimal(10,6);default:0.000000"`
	Currency       string  `gorm:"size:8;default:USD"`
	Latency        float64 `gorm:"type:decimal(10,6);default:0.000000"`
	ElapsedTime    float64 `gorm:"type:decimal(10,6);default:0.000000"`
	TotalSteps     int     `gorm:"default:0"`
	ConversationID string  `gorm:"size:64"`
	CreatedAt      int64   `gorm:"index"`
}

// TableName 指定表名
func (ChatStats) TableName() string {
	return "chat_stats"
}

// 注册模型到迁移系统
func init() {
	register(&ChatSession{})
	register(&ChatMessage{})
	register(&ChatStats{})
}
