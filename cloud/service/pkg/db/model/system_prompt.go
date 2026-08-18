package model

// SystemPrompt 系统提示词模型
type SystemPrompt struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"uniqueIndex;size:128"`
	Role      string `gorm:"size:64"`
	Content   string `gorm:"type:longtext;charset:utf8mb4"`
	Tags      string `gorm:"type:json"`
	IsActive  bool
	IsDefault bool  `gorm:"default:false"`
	Version   int   `gorm:"default:1"`
	CreatedAt int64 `gorm:"index"`
	UpdatedAt int64 `gorm:"index"`
}

func (SystemPrompt) TableName() string { return "system_prompts" }

func init() {
	register(&SystemPrompt{})
}
