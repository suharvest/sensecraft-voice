package model

func init() {
	register(&Location{})
}

type Location struct {
	Id          int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	StoreId     int64  `gorm:"column:store_id;index:idx_locations_store_id;not null" json:"store_id"`
	Name        string `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Code        string `gorm:"column:code;type:varchar(32);not null" json:"code"`
	Description string `gorm:"column:description;type:varchar(256);default:'';not null" json:"description"`
	Status      int8   `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	CreatedAt   int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (l *Location) TableName() string { return "locations" }
