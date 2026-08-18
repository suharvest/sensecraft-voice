package model

func init() {
	register(&Keyword{})
}

// Keyword 关键词模型
type Keyword struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	Keyword   string `gorm:"column:keyword;type:varchar(50);uniqueIndex:uk_keywords_keyword;not null" json:"keyword"`
	Synonyms  string `gorm:"column:synonyms;type:varchar(500);not null" json:"synonyms"`
	MarkColor string `gorm:"column:mark_color;type:varchar(7);default:'#ff4d4f';not null" json:"mark_color"`
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (k *Keyword) TableName() string { return "keywords" }
